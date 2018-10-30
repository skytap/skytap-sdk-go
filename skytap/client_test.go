package skytap

import (
	"context"
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
	skytap.retryCount = testingRetryCount
	skytap.retryAfter = testingRetryAfter

	assert.Nil(t, err)
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
			io.WriteString(rw, `{"id": "12345", "name": "test-project", "summary": "test project"}`)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		requestCounter++
	}

	_, err := skytap.Projects.Get(context.Background(), 12345)

	assert.Nil(t, err)
}
