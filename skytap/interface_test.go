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

func TestCreateInterface(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleCreateInterfaceResponse.json")), 456, 123)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(response))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces", req.URL.Path, "Bad path")
			assert.Equal(t, "POST", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, string(readTestFile(t, "exampleCreateInterfaceRequest.json")), string(body), "Bad request body")

			_, err = io.WriteString(rw, exampleInterface)
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/nic-456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, exampleInterface)
			assert.NoError(t, err)
		}
		requestCounter++
	}
	nicType := &CreateInterfaceRequest{
		NICType: nicTypeToPtr(NICTypeE1000),
	}

	networkInterface, err := skytap.Interfaces.Create(context.Background(), "123", "456", nicType)
	assert.Nil(t, err, "Bad API method")

	var interfaceExpected Interface
	err = json.Unmarshal([]byte(exampleInterface), &interfaceExpected)
	assert.Equal(t, interfaceExpected, *networkInterface, "Bad interface")

	assert.Equal(t, 3, requestCounter)
}

func TestAttachInterface(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleAttachInterfaceResponse.json")))

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(response))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, string(readTestFile(t, "exampleAttachInterfaceRequest.json")), string(body), "Bad request body")

			_, err = io.WriteString(rw, exampleInterface)
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/nic-20250403-38374059-4", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, exampleInterface)
			assert.NoError(t, err)
		}
		requestCounter++
	}
	networkID := &AttachInterfaceRequest{
		NetworkID: strToPtr("23917287"),
	}

	networkInterface, err := skytap.Interfaces.Attach(context.Background(), "123", "456", "789", networkID)
	assert.Nil(t, err, "Bad API method")

	var interfaceExpected Interface
	err = json.Unmarshal([]byte(exampleInterface), &interfaceExpected)
	assert.Equal(t, interfaceExpected, *networkInterface, "Bad interface")

	assert.Equal(t, 3, requestCounter)
}

func TestReadInterface(t *testing.T) {
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleCreateInterfaceResponse.json")), 456, 123)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, exampleInterface)
		assert.NoError(t, err)
	}

	networkInterface, err := skytap.Interfaces.Get(context.Background(), "123", "456", "789")
	assert.Nil(t, err, "Bad API method")

	var interfaceExpected Interface
	err = json.Unmarshal([]byte(exampleInterface), &interfaceExpected)
	assert.Equal(t, interfaceExpected, *networkInterface, "Bad Interface")
}

func TestUpdateInterface(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleUpdateInterfaceResponse.json")))

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(response))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, string(readTestFile(t, "exampleUpdateInterfaceRequest.json")), string(body), "Bad request body")

			_, err = io.WriteString(rw, string(readTestFile(t, "exampleUpdateInterfaceResponse.json")))
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/nic-20250403-38374059-4", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			var networkInterface Interface
			err := json.Unmarshal([]byte(exampleInterface), &networkInterface)
			assert.NoError(t, err)
			networkInterface.IP = strToPtr("10.0.0.1")
			networkInterface.Hostname = strToPtr("updated-hostname")
			b, err := json.Marshal(&networkInterface)
			assert.NoError(t, err)
			_, err = io.WriteString(rw, string(b))
			assert.NoError(t, err)
		}
		requestCounter++
	}

	var networkInterface Interface
	err := json.Unmarshal([]byte(exampleInterface), &networkInterface)
	assert.NoError(t, err)
	networkInterface.IP = strToPtr("10.0.0.1")
	networkInterface.Hostname = strToPtr("updated-hostname")
	_, err = json.Marshal(&networkInterface)
	assert.Nil(t, err, "Bad interface")

	opts := &UpdateInterfaceRequest{
		Hostname: strToPtr(*networkInterface.Hostname),
		IP:       strToPtr("10.0.0.1"),
	}
	interfaceUpdate, err := skytap.Interfaces.Update(context.Background(), "123", "456", "789", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, networkInterface, *interfaceUpdate, "Bad interface")

	assert.Equal(t, 3, requestCounter)
}

func TestDeleteInterface(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(response))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			log.Printf("Request: (%d)\n", requestCounter)
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
			assert.Equal(t, "DELETE", req.Method, "Bad method")
		}
		requestCounter++
	}

	err := skytap.Interfaces.Delete(context.Background(), "123", "456", "789")
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, 2, requestCounter)
}

func TestListInterfaces(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, string(readTestFile(t, "exampleInterfaceListResponse.json")))
		assert.NoError(t, err)
	}

	result, err := skytap.Interfaces.List(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")

	var found = false
	for _, networkInterface := range result.Value {
		if *networkInterface.ID == "nic-20246343-38367563-5" {
			found = true
			break
		}
	}
	assert.True(t, found, "Interface not found")
}

func TestCompareInterfaceCreateTrue(t *testing.T) {
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleCreateInterfaceResponse.json")), 456, 123)

	var adapter Interface
	err := json.Unmarshal([]byte(exampleInterface), &adapter)
	assert.NoError(t, err)
	opts := CreateInterfaceRequest{
		NICType: nicTypeToPtr(NICTypeE1000),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, fmt.Sprintf(exampleInterface))
		assert.NoError(t, err)
	}
	message, ok := opts.compareResponse(context.Background(), skytap, &adapter, vmRequestRunStateStopped("123", "456"))
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestCompareInterfaceCreateFalse(t *testing.T) {
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleCreateInterfaceResponse.json")), 456, 123)

	var adapter Interface
	err := json.Unmarshal([]byte(exampleInterface), &adapter)
	assert.NoError(t, err)
	opts := CreateInterfaceRequest{
		NICType: nicTypeToPtr(NICTypeE1000E),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, fmt.Sprintf(exampleInterface))
		assert.NoError(t, err)
	}
	message, ok := opts.compareResponse(context.Background(), skytap, &adapter, vmRequestRunStateStopped("123", "456"))
	assert.False(t, ok)
	assert.Equal(t, "network adapter not ready", message)
}

func TestCompareInterfaceAttachTrue(t *testing.T) {
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleAttachInterfaceResponse.json")))

	var adapter Interface
	err := json.Unmarshal([]byte(exampleInterface), &adapter)
	assert.NoError(t, err)
	opts := AttachInterfaceRequest{
		strToPtr("23917287"),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, fmt.Sprintf(exampleInterface))
		assert.NoError(t, err)
	}
	message, ok := opts.compareResponse(context.Background(), skytap, &adapter, vmRequestRunStateStopped("123", "456"))
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestCompareInterfaceAttachFalse(t *testing.T) {
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleAttachInterfaceResponse.json")))

	var adapter Interface
	err := json.Unmarshal([]byte(exampleInterface), &adapter)
	assert.NoError(t, err)
	opts := AttachInterfaceRequest{
		strToPtr("123"),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, fmt.Sprintf(exampleInterface))
		assert.NoError(t, err)
	}
	message, ok := opts.compareResponse(context.Background(), skytap, &adapter, vmRequestRunStateStopped("123", "456"))
	assert.False(t, ok)
	assert.Equal(t, "network adapter not ready", message)
}

func TestCompareInterfaceUpdateTrue(t *testing.T) {
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleUpdateInterfaceResponse.json")))

	var adapter Interface
	err := json.Unmarshal([]byte(exampleInterface), &adapter)
	assert.NoError(t, err)
	opts := UpdateInterfaceRequest{
		IP:       strToPtr("10.0.0.1"),
		Hostname: strToPtr("updated-hostname"),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, fmt.Sprintf(exampleInterface))
		assert.NoError(t, err)
	}
	message, ok := opts.compareResponse(context.Background(), skytap, &adapter, vmRequestRunStateStopped("123", "456"))
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestCompareInterfaceUpdateFalse(t *testing.T) {
	exampleInterface := fmt.Sprintf(string(readTestFile(t, "exampleUpdateInterfaceResponse.json")))

	var adapter Interface
	err := json.Unmarshal([]byte(exampleInterface), &adapter)
	assert.NoError(t, err)
	opts := UpdateInterfaceRequest{
		IP:       strToPtr("10.0.0.2"),
		Hostname: strToPtr("updated-hostname"),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, fmt.Sprintf(exampleInterface))
		assert.NoError(t, err)
	}
	message, ok := opts.compareResponse(context.Background(), skytap, &adapter, vmRequestRunStateStopped("123", "456"))
	assert.False(t, ok)
	assert.Equal(t, "network adapter not ready", message)
}
