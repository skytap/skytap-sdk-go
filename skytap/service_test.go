package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

const exampleCreateServiceRequest = `{
    "internal_port": %d
}`

const exampleCreateServiceResponse = `{
    "id": %d,
    "internal_port": %d,
    "external_ip": "services-uswest.skytap.com",
    "external_port": 26160
}`

const exampleServiceListResponse = `[
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
	exampleService := fmt.Sprintf(exampleCreateServiceResponse, port, port)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services", req.URL.Path, "Bad path")
		assert.Equal(t, "POST", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, fmt.Sprintf(exampleCreateServiceRequest, port), string(body), "Bad request body")

		io.WriteString(rw, exampleService)
	}
	internalPort := &CreateServiceRequest{
		InternalPort: intToPtr(port),
	}

	service, err := skytap.Services.Create(context.Background(), "123", "456", "789", internalPort)
	assert.Nil(t, err, "Bad API method")

	var serviceExpected Service
	err = json.Unmarshal([]byte(exampleService), &serviceExpected)
	assert.Equal(t, serviceExpected, *service, "Bad service")
}

func TestReadService(t *testing.T) {
	exampleService := fmt.Sprintf(exampleCreateServiceResponse, 8080, 8080)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/abc", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, exampleService)
	}

	service, err := skytap.Services.Get(context.Background(), "123", "456", "789", "abc")
	assert.Nil(t, err, "Bad API method")

	var serviceExpected Service
	err = json.Unmarshal([]byte(exampleService), &serviceExpected)
	assert.Equal(t, serviceExpected, *service, "Bad Interface")
}

func TestUpdateService(t *testing.T) {
	port := 8081
	exampleService := fmt.Sprintf(exampleCreateServiceResponse, port, port)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var service Service
	json.Unmarshal([]byte(exampleService), &service)

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
			assert.JSONEq(t, fmt.Sprintf(exampleCreateServiceRequest, port), string(body), "Bad request body")

			io.WriteString(rw, exampleService)
		}
	}

	opts := &UpdateServiceRequest{
		CreateServiceRequest{
			InternalPort: intToPtr(port),
		},
	}
	serviceUpdate, err := skytap.Services.Update(context.Background(), "123", "456", "789", "abc", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, service, *serviceUpdate, "Bad service")
}

func TestDeleteService(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services/abc", req.URL.Path, "Bad path")
		assert.Equal(t, "DELETE", req.Method, "Bad method")
	}

	err := skytap.Services.Delete(context.Background(), "123", "456", "789", "abc")
	assert.Nil(t, err, "Bad API method")
}

func TestListServices(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789/services", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, exampleServiceListResponse)
	}

	result, err := skytap.Services.List(context.Background(), "123", "456", "789")
	assert.Nil(t, err, "Bad API method")

	var found = false
	for _, service := range result.Value {
		if *service.ID == "8081" {
			found = true
			break
		}
	}
	assert.True(t, found, "Service not found")
}
