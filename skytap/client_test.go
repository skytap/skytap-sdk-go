package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testingRetryAfter = 1
	testingRetryCount = 3
)

func createClient(t *testing.T) (*Client, *httptest.Server, *func(rw http.ResponseWriter, req *http.Request)) {
	handler := http.NotFound
	hs := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		handler(rw, req)
	}))

	var user = "SKYTAP_USER"
	var token = "SKYTAP_ACCESS_TOKEN"

	settings := NewDefaultSettings(WithBaseURL(hs.URL), WithCredentialsProvider(NewAPITokenCredentials(user, token)))

	skytap, err := NewClient(settings)
	assert.Nil(t, err)
	skytap.retryCount = testingRetryCount
	skytap.retryAfter = testingRetryAfter

	assert.NotNil(t, skytap)
	return skytap, hs, &handler
}

func TestGetRetryWithFailure(t *testing.T) {
	requestCounter := 0
	skytap, hs, handler := createClient(t)

	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(401)
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	errorResponse := err.(*ErrorResponse)

	assert.Equal(t, 1, requestCounter)
	assert.Equal(t, http.StatusUnauthorized, errorResponse.Response.StatusCode)
}

func TestGetRetryWithBusy409(t *testing.T) {
	requestCounter := 0
	skytap, hs, handler := createClient(t)
	skytap.retryCount = 1
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Retry-After", "2")
		rw.WriteHeader(http.StatusConflict)
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	errorResponse := err.(*ErrorResponse)

	assert.Equal(t, http.StatusConflict, errorResponse.Response.StatusCode)
	assert.Equal(t, 2, *errorResponse.RetryAfter)
	assert.Equal(t, 1, requestCounter)
	assert.Equal(t, 3, testingRetryCount)
}

func TestGetRetryWithBusy423(t *testing.T) {
	requestCounter := 0
	skytap, hs, handler := createClient(t)
	skytap.retryCount = 1
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Retry-After", "2")
		rw.WriteHeader(http.StatusLocked)
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	errorResponse := err.(*ErrorResponse)

	assert.Equal(t, http.StatusLocked, errorResponse.Response.StatusCode)
	assert.Equal(t, 2, *errorResponse.RetryAfter)
	assert.Equal(t, 1, requestCounter)
	assert.Equal(t, 3, testingRetryCount)
}

func TestGetRetryWithBusy423WithBadRetryAfter(t *testing.T) {
	requestCounter := 0
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Retry-After", "xxx")
		rw.WriteHeader(http.StatusLocked)
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	errorResponse := err.(*ErrorResponse)

	assert.Equal(t, testingRetryAfter, *errorResponse.RetryAfter)
	assert.Equal(t, http.StatusLocked, errorResponse.Response.StatusCode)
	assert.Equal(t, 1, requestCounter)
}

func TestGetRetryWithBusy423WithoutRetryAfter(t *testing.T) {
	requestCounter := 0
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusLocked)
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	errorResponse := err.(*ErrorResponse)

	assert.Equal(t, testingRetryAfter, *errorResponse.RetryAfter)
	assert.Equal(t, http.StatusLocked, errorResponse.Response.StatusCode)
	assert.Equal(t, 1, requestCounter)
}

func TestGetRetryWith429(t *testing.T) {
	requestCounter := 0
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Retry-After", "1")
		rw.WriteHeader(http.StatusTooManyRequests)
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	errorResponse := err.(*ErrorResponse)

	assert.Equal(t, http.StatusTooManyRequests, errorResponse.Response.StatusCode)
	assert.Equal(t, 1, requestCounter)
}

func TestGetRetryWith50x(t *testing.T) {
	requestCounter := 0
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Retry-After", "1")
		rw.WriteHeader(http.StatusInternalServerError)
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	errorResponse := err.(*ErrorResponse)

	assert.Equal(t, http.StatusInternalServerError, errorResponse.Response.StatusCode)
	assert.Equal(t, 1, requestCounter)
}

func TestGetPreRequestRunstateNotExpecting(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()
	responseProcessed := false

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		responseProcessed = true
		assert.Equal(t, http.MethodGet, req.Method, "Unexpected method")
		assert.Equal(t, "/v2/projects/12345", req.URL.Path, "Unexpected path")
		_, err := io.WriteString(rw, `{"id": "12345", "name": "test-project", "summary": "test project"}`)
		assert.NoError(t, err)
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	assert.NoError(t, err)
	assert.True(t, responseProcessed)
}

func TestPutPostPreRequestRunstateNotExpecting2(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()
	responseProcessed := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		responseProcessed++
		method := http.MethodPost
		path := "/projects"
		if responseProcessed >= 2 {
			method = http.MethodPut
			path = "/projects/12345"
		}
		assert.Equal(t, method, req.Method, "Unexpected method")
		assert.Equal(t, path, req.URL.Path, "Unexpected path")
		_, err := io.WriteString(rw, `{"id": "12345", "name": "test-project"}`)
		assert.NoError(t, err)
	}

	project := Project{}
	_, err := skytap.Projects.Create(context.Background(), &project)
	assert.NoError(t, err)
	assert.Equal(t, 2, responseProcessed)
}

func TestPutPostPreRequestRunstate(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()
	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		method := http.MethodGet
		path := "/v2/configurations/123/vms/456"
		if requestCounter == 1 {
			method = http.MethodPost
			path = "/v2/configurations/123/vms/456/interfaces"
		} else if requestCounter >= 2 {
			method = http.MethodGet
			path = "/v2/configurations/123/vms/456/interfaces/456"
		}
		assert.Equal(t, path, req.URL.Path, fmt.Sprintf("Bad path: %d", requestCounter))
		assert.Equal(t, method, req.Method, fmt.Sprintf("Bad method: %d", requestCounter))

		_, err := io.WriteString(rw, response)
		assert.NoError(t, err)
		requestCounter++
	}

	nicType := &CreateInterfaceRequest{
		NICType: nicTypeToPtr(NICTypeE1000),
	}

	_, err := skytap.Interfaces.Create(context.Background(), "123", "456", nicType)
	assert.Nil(t, err)
	assert.Equal(t, 5, requestCounter)
}

func TestPutPostPreRequestRunstate2(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	var vm VM
	err := json.Unmarshal([]byte(response), &vm)
	assert.NoError(t, err)
	*vm.Runstate = VMRunstateBusy
	responseBusy, err := json.Marshal(&vm)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	skytap.retryAfter = 1
	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(responseBusy))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces", req.URL.Path, "Bad path")
			assert.Equal(t, "POST", req.Method, "Bad method")
			exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleCreateInterfaceResponse.json")), 456, 123)
			_, err := io.WriteString(rw, exampleInterface)
			assert.NoError(t, err)
		} else if requestCounter == 3 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/nic-456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")
			exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleCreateInterfaceResponse.json")), 456, 123)
			_, err := io.WriteString(rw, exampleInterface)
			assert.NoError(t, err)
		}
		requestCounter++
	}

	nicType := &CreateInterfaceRequest{
		NICType: nicTypeToPtr(NICTypeE1000),
	}

	_, err = skytap.Interfaces.Create(context.Background(), "123", "456", nicType)
	assert.Nil(t, err)

	assert.Equal(t, 4, requestCounter)
}

func TestGetStatus200(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, string(readTestFile(t, "exampleEnvironment.json")))
		assert.NoError(t, err)
	}

	var environment Environment
	path := fmt.Sprintf("%s/%s", environmentBasePath, "123")
	req, err := skytap.newRequest(context.Background(), "GET", path, nil)
	resp, err := skytap.requestGet(context.Background(), req, &environment)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPutPostDelete(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var vmResponse VM
	err := json.Unmarshal([]byte(response), &vmResponse)
	assert.NoError(t, err)
	vmResponse.Runstate = vmRunStateToPtr(VMRunstateRunning)
	bytesRunning, err := json.Marshal(&vmResponse)
	assert.Nil(t, err, "Bad vm")

	requestCounter := 0
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {

			_, err := io.WriteString(rw, string(bytesRunning))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			vmResponse.Runstate = vmRunStateToPtr(VMRunstateStopped)
			bytesStopped, err := json.Marshal(&vmResponse)
			assert.Nil(t, err, "Bad vm")

			_, err = io.WriteString(rw, string(bytesStopped))
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
		}
		requestCounter++
	}

	update := &UpdateVMRequest{Runstate: vmRunStateToPtr(VMRunstateStopped)}
	req, err := skytap.newRequest(context.Background(), http.MethodPut, "", update)
	assert.NoError(t, err)

	var vm VM
	_, err = skytap.do(context.Background(), req, &vm, vmRunStateNotBusy("123", "456"), update)
	assert.NoError(t, err)

	assert.Equal(t, 3, requestCounter)
}
