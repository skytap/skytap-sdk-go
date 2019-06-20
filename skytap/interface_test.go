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

const exampleCreateInterfaceRequest = `{
    "nic_type": "e1000"
}`

const exampleAttachInterfaceRequest = `{
    "network_id": "23917287"
}`

const exampleUpdateInterfaceRequest = `{
    "ip": "10.0.0.1",
    "hostname": "updated-hostname"
}`

const exampleCreateInterfaceResponse = `{
    "id": "nic-%d",
    "ip": null,
    "hostname": null,
    "mac": "00:50:56:07:40:3F",
    "services_count": 0,
    "services": [],
    "public_ips_count": 0,
    "public_ips": [],
    "vm_id": "%d",
    "vm_name": "Windows Server 2016 Standard",
    "status": "Powered off",
    "nic_type": "e1000",
    "secondary_ips": [],
    "public_ip_attachments": []
}`

const exampleAttachInterfaceResponse = `{
    "id": "nic-20250403-38374059-4",
    "ip": "192.168.0.5",
    "hostname": "host-3",
    "mac": "00:50:56:05:3F:84",
    "services_count": 0,
    "services": [],
    "public_ips_count": 0,
    "public_ips": [],
    "vm_id": "37533321",
    "vm_name": "CentOS 6 Desktop x64",
    "status": "Powered off",
    "network_id": "23922457",
    "network_name": "tftest-network-1",
    "network_url": "https://cloud.skytap.com/v2/configurations/40071754/networks/23922457",
    "network_type": "automatic",
    "network_subnet": "192.168.0.0/16",
    "nic_type": "vmxnet3",
    "secondary_ips": [],
    "public_ip_attachments": []
}`

const exampleInterfaceListResponse = `[
    {
        "id": "nic-20246343-38367563-0",
        "ip": "192.168.0.1",
        "hostname": "wins2016s",
        "mac": "00:50:56:11:7D:D9",
        "services_count": 0,
        "services": [],
        "public_ips_count": 0,
        "public_ips": [],
        "vm_id": "37527239",
        "vm_name": "Windows Server 2016 Standard",
        "status": "Running",
        "network_id": "23917287",
        "network_name": "tftest-network-1",
        "network_url": "https://cloud.skytap.com/v2/configurations/40064014/networks/23917287",
        "network_type": "automatic",
        "network_subnet": "192.168.0.0/16",
        "nic_type": "vmxnet3",
        "secondary_ips": [],
        "public_ip_attachments": []
    },
    {
        "id": "nic-20246343-38367563-5",
        "ip": null,
        "hostname": null,
        "mac": "00:50:56:07:40:3F",
        "services_count": 0,
        "services": [],
        "public_ips_count": 0,
        "public_ips": [],
        "vm_id": "37527239",
        "vm_name": "Windows Server 2016 Standard",
        "status": "Running",
        "nic_type": "e1000",
        "secondary_ips": [],
        "public_ip_attachments": []
    }
]`

func TestCreateInterface(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	exampleInterface := fmt.Sprintf(exampleCreateInterfaceResponse, 456, 123)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	first := true
	second := true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if first {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(response))
			assert.NoError(t, err)
			first = false
		} else {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces", req.URL.Path, "Bad path")
			assert.Equal(t, "POST", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, exampleCreateInterfaceRequest, string(body), "Bad request body")

			io.WriteString(rw, exampleInterface)
			second = false
		}
	}
	nicType := &CreateInterfaceRequest{
		NICType: nicTypeToPtr(NICTypeE1000),
	}

	networkInterface, err := skytap.Interfaces.Create(context.Background(), "123", "456", nicType)
	assert.Nil(t, err, "Bad API method")

	var interfaceExpected Interface
	err = json.Unmarshal([]byte(exampleInterface), &interfaceExpected)
	assert.Equal(t, interfaceExpected, *networkInterface, "Bad interface")

	assert.False(t, first)
	assert.False(t, second)
}

func TestAttachInterface(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	exampleInterface := fmt.Sprintf(exampleAttachInterfaceResponse)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	first := true
	second := true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if first {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(response))
			assert.NoError(t, err)
			first = false
		} else {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, exampleAttachInterfaceRequest, string(body), "Bad request body")

			io.WriteString(rw, exampleInterface)
			second = false
		}
	}
	networkID := &AttachInterfaceRequest{
		NetworkID: strToPtr("23917287"),
	}

	networkInterface, err := skytap.Interfaces.Attach(context.Background(), "123", "456", "789", networkID)
	assert.Nil(t, err, "Bad API method")

	var interfaceExpected Interface
	err = json.Unmarshal([]byte(exampleInterface), &interfaceExpected)
	assert.Equal(t, interfaceExpected, *networkInterface, "Bad interface")

	assert.False(t, first)
	assert.False(t, second)
}

func TestReadInterface(t *testing.T) {
	exampleInterface := fmt.Sprintf(exampleCreateInterfaceResponse, 456, 123)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, exampleInterface)
	}

	networkInterface, err := skytap.Interfaces.Get(context.Background(), "123", "456", "789")
	assert.Nil(t, err, "Bad API method")

	var interfaceExpected Interface
	err = json.Unmarshal([]byte(exampleInterface), &interfaceExpected)
	assert.Equal(t, interfaceExpected, *networkInterface, "Bad Interface")
}

func TestUpdateInterface(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	exampleInterface := fmt.Sprintf(exampleAttachInterfaceResponse)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var networkInterface Interface
	json.Unmarshal([]byte(exampleInterface), &networkInterface)
	networkInterface.Hostname = strToPtr("updated-hostname")

	bytes, err := json.Marshal(&networkInterface)
	assert.Nil(t, err, "Bad interface")

	first := true
	second := true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if first {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(response))
			assert.NoError(t, err)
			first = false
		} else {
			assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, exampleUpdateInterfaceRequest, string(body), "Bad request body")

			io.WriteString(rw, string(bytes))
			second = false
		}
	}

	opts := &UpdateInterfaceRequest{
		Hostname: strToPtr(*networkInterface.Hostname),
		IP:       strToPtr("10.0.0.1"),
	}
	interfaceUpdate, err := skytap.Interfaces.Update(context.Background(), "123", "456", "789", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, networkInterface, *interfaceUpdate, "Bad interface")

	assert.False(t, first)
	assert.False(t, second)
}

func TestDeleteInterface(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
		assert.Equal(t, "DELETE", req.Method, "Bad method")
	}

	err := skytap.Interfaces.Delete(context.Background(), "123", "456", "789")
	assert.Nil(t, err, "Bad API method")
}

func TestListInterfaces(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, exampleInterfaceListResponse)
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

//func TestUpdateInterface422(t *testing.T) {
//	exampleInterface := fmt.Sprintf(exampleAttachInterfaceResponse)
//
//	skytap, hs, handler := createClient(t)
//	defer hs.Close()
//
//	var networkInterface Interface
//	err := json.Unmarshal([]byte(exampleInterface), &networkInterface)
//	assert.NoError(t, err)
//
//	networkInterface.Hostname = strToPtr("updated-hostname")
//
//	response, err := json.Marshal(&networkInterface)
//	assert.Nil(t, err, "Bad interface")
//	requestCounter := 0
//
//	*handler = func(rw http.ResponseWriter, req *http.Request) {
//		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
//		assert.Equal(t, "PUT", req.Method, "Bad method")
//
//		body, err := ioutil.ReadAll(req.Body)
//		assert.Nil(t, err, "Bad request body")
//		assert.JSONEq(t, exampleUpdateInterfaceRequest, string(body), "Bad request body")
//
//		if requestCounter == 0 {
//			rw.WriteHeader(http.StatusUnprocessableEntity)
//		} else {
//			_, err = io.WriteString(rw, string(response))
//			assert.NoError(t, err)
//		}
//		requestCounter++
//	}
//	opts := &UpdateInterfaceRequest{
//		Hostname: strToPtr(*networkInterface.Hostname),
//		IP:       strToPtr("10.0.0.1"),
//	}
//	interfaceUpdate, err := skytap.Interfaces.Update(context.Background(), "123", "456", "789", opts)
//	assert.Nil(t, err, "Bad API method")
//
//	assert.Equal(t, networkInterface, *interfaceUpdate, "Bad interface")
//	assert.Equal(t, 2, requestCounter)
//}
