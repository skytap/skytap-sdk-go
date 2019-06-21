package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateService(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	port := 8080
	exampleService := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceResponse.json")), port, port)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services", req.URL.Path, "Bad path")
			assert.Equal(t, "POST", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceRequest.json")), port), string(body), "Bad request body")

			_, err = io.WriteString(rw, exampleService)
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/8080", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, exampleService)
			assert.NoError(t, err)
		}
		requestCounter++
	}
	internalPort := &CreatePublishedServiceRequest{
		InternalPort: intToPtr(port),
	}

	service, err := skytap.PublishedServices.Create(context.Background(), "123", "456", "789", internalPort)
	assert.Nil(t, err, "Bad API method")

	var serviceExpected PublishedService
	err = json.Unmarshal([]byte(exampleService), &serviceExpected)
	assert.Equal(t, serviceExpected, *service, "Bad publishedService")

	assert.Equal(t, 3, requestCounter)
}

func TestReadService(t *testing.T) {
	exampleService := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceResponse.json")), 8080, 8080)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/abc", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, exampleService)
		assert.NoError(t, err)
		requestCounter++
	}

	service, err := skytap.PublishedServices.Get(context.Background(), "123", "456", "789", "abc")
	assert.Nil(t, err, "Bad API method")

	var serviceExpected PublishedService
	err = json.Unmarshal([]byte(exampleService), &serviceExpected)
	assert.Equal(t, serviceExpected, *service, "Bad Interface")

	assert.Equal(t, 1, requestCounter)
}

func TestUpdateService(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	port := 8081
	exampleService := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceResponse.json")), port, port)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var service PublishedService
	err := json.Unmarshal([]byte(exampleService), &service)
	assert.NoError(t, err)

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/abc", req.URL.Path, "Bad path")
			assert.Equal(t, "DELETE", req.Method, "Bad method")
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services", req.URL.Path, "Bad path")
			assert.Equal(t, "POST", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceRequest.json")), port), string(body), "Bad request body")

			_, err = io.WriteString(rw, exampleService)
			assert.NoError(t, err)
		} else if requestCounter == 3 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/8081", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, exampleService)
			assert.NoError(t, err)
		}
		requestCounter++
	}

	opts := &UpdatePublishedServiceRequest{
		CreatePublishedServiceRequest{
			InternalPort: intToPtr(port),
		},
	}
	serviceUpdate, err := skytap.PublishedServices.Update(context.Background(), "123", "456", "789", "abc", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, service, *serviceUpdate, "Bad publishedService")

	assert.Equal(t, 4, requestCounter)
}

func TestDeleteService(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/abc", req.URL.Path, "Bad path")
		assert.Equal(t, "DELETE", req.Method, "Bad method")
		requestCounter++
	}

	err := skytap.PublishedServices.Delete(context.Background(), "123", "456", "789", "abc")
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, 1, requestCounter)
}

func TestListServices(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, string(readTestFile(t, "examplePublishedServiceListResponse.json")))
		assert.NoError(t, err)
		requestCounter++
	}

	result, err := skytap.PublishedServices.List(context.Background(), "123", "456", "789")
	assert.Nil(t, err, "Bad API method")

	var found = false
	for _, service := range result.Value {
		if *service.ID == "8081" {
			found = true
			break
		}
	}
	assert.True(t, found, "PublishedService not found")

	assert.Equal(t, 1, requestCounter)
}

func TestComparePublishedServiceCreateTrue(t *testing.T) {
	examplePublishedService := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceResponse.json")), 789, 8080)

	var service PublishedService
	err := json.Unmarshal([]byte(examplePublishedService), &service)
	assert.NoError(t, err)
	opts := CreatePublishedServiceRequest{
		InternalPort: intToPtr(8080),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, string(examplePublishedService))
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.adapterID = strToPtr("789")
	message, ok := opts.compare(context.Background(), skytap, &service, state)
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestComparePublishedServiceCreateFalse(t *testing.T) {
	examplePublishedService := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceResponse.json")), 789, 8080)

	var service PublishedService
	err := json.Unmarshal([]byte(examplePublishedService), &service)
	assert.NoError(t, err)
	opts := CreatePublishedServiceRequest{
		InternalPort: intToPtr(8081),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		response := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceRequest.json")), 8080)
		_, err := io.WriteString(rw, string(response))
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.adapterID = strToPtr("789")
	message, ok := opts.compare(context.Background(), skytap, &service, state)
	assert.False(t, ok)
	assert.Equal(t, "published service not ready", message)
}

func TestComparePublishedServiceUpdateTrue(t *testing.T) {
	examplePublishedService := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceResponse.json")), 789, 8080)

	var service PublishedService
	err := json.Unmarshal([]byte(examplePublishedService), &service)
	assert.NoError(t, err)
	opts := UpdatePublishedServiceRequest{
		CreatePublishedServiceRequest{
			InternalPort: intToPtr(8081),
		},
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		response := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceRequest.json")), 8081)
		_, err := io.WriteString(rw, string(response))
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.adapterID = strToPtr("789")
	message, ok := opts.compare(context.Background(), skytap, &service, state)
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestComparePublishedServiceUpdateFalse(t *testing.T) {
	examplePublishedService := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceResponse.json")), 789, 8080)

	var service PublishedService
	err := json.Unmarshal([]byte(examplePublishedService), &service)
	assert.NoError(t, err)
	opts := UpdatePublishedServiceRequest{
		CreatePublishedServiceRequest{
			InternalPort: intToPtr(8081),
		},
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		response := fmt.Sprintf(string(readTestFile(t, "examplePublishedServiceRequest.json")), 8080)
		_, err := io.WriteString(rw, string(response))
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.adapterID = strToPtr("789")
	message, ok := opts.compare(context.Background(), skytap, &service, state)
	assert.False(t, ok)
	assert.Equal(t, "published service not ready", message)
}
