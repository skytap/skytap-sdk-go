package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const examplePublishedServiceRequest = `{
    "internal_port": %d
}`

const examplePublishedServiceResponse = `{
    "id": %d,
    "internal_port": %d,
    "external_ip": "services-uswest.skytap.com",
    "external_port": 26160
}`

const examplePublishedServiceListResponse = `[
    {
        "id": "8080",
        "internal_port": 8080,
        "external_ip": "services-uswest.skytap.com",
        "external_port": 26160
    },
    {
        "id": "8081",
        "internal_port": 8081,
        "external_ip": "services-uswest.skytap.com",
        "external_port": 17785
    }
]`

func TestCreateService(t *testing.T) {
	port := 8080
	exampleService := fmt.Sprintf(examplePublishedServiceResponse, port, port)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services", req.URL.Path, "Bad path")
		assert.Equal(t, "POST", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, fmt.Sprintf(examplePublishedServiceRequest, port), string(body), "Bad request body")

		_, err = io.WriteString(rw, exampleService)
		assert.NoError(t, err)
	}
	internalPort := &CreatePublishedServiceRequest{
		InternalPort: intToPtr(port),
	}

	service, err := skytap.PublishedServices.Create(context.Background(), "123", "456", "789", internalPort)
	assert.Nil(t, err, "Bad API method")

	var serviceExpected PublishedService
	err = json.Unmarshal([]byte(exampleService), &serviceExpected)
	assert.Equal(t, serviceExpected, *service, "Bad publishedService")
}

func TestReadService(t *testing.T) {
	exampleService := fmt.Sprintf(examplePublishedServiceResponse, 8080, 8080)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/abc", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, exampleService)
		assert.NoError(t, err)
	}

	service, err := skytap.PublishedServices.Get(context.Background(), "123", "456", "789", "abc")
	assert.Nil(t, err, "Bad API method")

	var serviceExpected PublishedService
	err = json.Unmarshal([]byte(exampleService), &serviceExpected)
	assert.Equal(t, serviceExpected, *service, "Bad Interface")
}

func TestUpdateService(t *testing.T) {
	port := 8081
	exampleService := fmt.Sprintf(examplePublishedServiceResponse, port, port)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var service PublishedService
	err := json.Unmarshal([]byte(exampleService), &service)
	assert.NoError(t, err)

	var deletePhase = true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if deletePhase {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/abc", req.URL.Path, "Bad path")
			assert.Equal(t, "DELETE", req.Method, "Bad method")
			deletePhase = false
		} else {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services", req.URL.Path, "Bad path")
			assert.Equal(t, "POST", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, fmt.Sprintf(examplePublishedServiceRequest, port), string(body), "Bad request body")

			_, err = io.WriteString(rw, exampleService)
			assert.NoError(t, err)
		}
	}

	opts := &UpdatePublishedServiceRequest{
		CreatePublishedServiceRequest{
			InternalPort: intToPtr(port),
		},
	}
	serviceUpdate, err := skytap.PublishedServices.Update(context.Background(), "123", "456", "789", "abc", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, service, *serviceUpdate, "Bad publishedService")
}

func TestDeleteService(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/abc", req.URL.Path, "Bad path")
		assert.Equal(t, "DELETE", req.Method, "Bad method")
	}

	err := skytap.PublishedServices.Delete(context.Background(), "123", "456", "789", "abc")
	assert.Nil(t, err, "Bad API method")
}

func TestListServices(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, examplePublishedServiceListResponse)
		assert.NoError(t, err)
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
}
