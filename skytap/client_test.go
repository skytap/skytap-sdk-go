package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func TestRetryWithFailure(t *testing.T) {
	requestCounter := 0
	skytap, hs, handler := createClient(t)

	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(401)
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	errorResponse := err.(*ErrorResponse)

	assert.Nil(t, errorResponse.RetryAfter)
	assert.Equal(t, 1, requestCounter)
	assert.Equal(t, http.StatusUnauthorized, errorResponse.Response.StatusCode)
}

func TestRetryWithBusy409(t *testing.T) {
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
	assert.True(t, errorResponse.RequiresRetry)
	assert.Equal(t, 2, requestCounter)
	assert.Equal(t, 3, testingRetryCount)
}

func TestRetryWithBusy423(t *testing.T) {
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
	assert.True(t, errorResponse.RequiresRetry)
	assert.Equal(t, 2, requestCounter)
	assert.Equal(t, 3, testingRetryCount)
}

func TestRetryWithBusy423WithBadRetryAfter(t *testing.T) {
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
	assert.Equal(t, testingRetryCount+1, requestCounter)
}

func TestRetryWithBusy423WithoutRetryAfter(t *testing.T) {
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
	assert.Equal(t, testingRetryCount+1, requestCounter)
}

func TestRetryWith429(t *testing.T) {
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
	assert.Equal(t, testingRetryCount+1, requestCounter)
}

func TestRetryWith50x(t *testing.T) {
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
	assert.Equal(t, testingRetryCount+1, requestCounter)
}

func TestRetryWith50xResolves(t *testing.T) {
	requestCounter := 0
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Retry-After", "1")
		if requestCounter == 3 {
			_, err := io.WriteString(rw, `{"id": "12345", "name": "test-project", "summary": "test project"}`)
			assert.NoError(t, err)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)

	assert.Nil(t, err)
}

func TestPreRequestPutPostRunstateNotExpecting(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()
	responseProcessed := false

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		responseProcessed = true
		assert.Equal(t, http.MethodGet, req.Method, "Unexpected method")
		assert.Equal(t, "/v2/projects/12345", req.URL.Path, "Unexpected path")
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)
	assert.NoError(t, err)
	assert.True(t, responseProcessed)
}

func TestPreRequestPutPostRunstateNotExpecting2(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()
	responseProcessed := false

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if !responseProcessed {
			responseProcessed = true
			assert.Equal(t, http.MethodPost, req.Method, "Unexpected method")
			assert.Equal(t, "/projects", req.URL.Path, "Unexpected path")
			_, err := io.WriteString(rw, `{"id": "12345", "name": "test-project"}`)
			assert.NoError(t, err)
		}
	}

	project := Project{}
	_, err := skytap.Projects.Create(context.Background(), &project)
	assert.NoError(t, err)
	assert.True(t, responseProcessed)
}

func TestPreRequestPutPostRunstate(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()
	responseProcessed := false

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if !responseProcessed {
			responseProcessed = true
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
		}
	}

	nicType := &CreateInterfaceRequest{
		NICType: nicTypeToPtr(NICTypeE1000),
	}

	_, err := skytap.Interfaces.Create(context.Background(), "123", "456", nicType)
	assert.Nil(t, err)
	assert.True(t, responseProcessed)
}

func TestPreRequestPutPostRunstate2(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	var vm VM
	err := json.Unmarshal([]byte(response), &vm)
	assert.NoError(t, err)
	*vm.Runstate = VMRunstateBusy
	responseBusy, err := json.Marshal(&vm)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	first := true
	second := true
	third := true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if first {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(responseBusy))
			assert.NoError(t, err)
			first = false
		} else if second {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
			second = false
		} else if third {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces", req.URL.Path, "Bad path")
			assert.Equal(t, "POST", req.Method, "Bad method")
			third = false
		}
	}

	nicType := &CreateInterfaceRequest{
		NICType: nicTypeToPtr(NICTypeE1000),
	}

	_, err = skytap.Interfaces.Create(context.Background(), "123", "456", nicType)
	assert.Nil(t, err)

	assert.False(t, first)
	assert.False(t, second)
	assert.False(t, third)
}
