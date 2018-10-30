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

const exampleAdapterRequest = `{
    "nic_type": "e1000",
    "network_id": "23917287"
}`

const exampleAdapterRequest2 = `{
    "nic_type": "e1000",
    "network_id": "23917287",
    "ip": "10.0.0.1",
    "hostname": "name.com"
}`

const exampleAdapterResponse = `{
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

const exampleAdapterListResponse = `[
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

func TestCreateAdapter(t *testing.T) {
	exampleAdapter := fmt.Sprintf(exampleAdapterResponse, 456, 123)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces", req.URL.Path, "Bad path")
		assert.Equal(t, "POST", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, exampleAdapterRequest, string(body), "Bad request body")

		io.WriteString(rw, exampleAdapter)
	}
	opts := &CreateAdapterRequest{
		NICType:   nicTypeToPtr(NICTypeE1000),
		NetworkID: strToPtr("23917287"),
	}

	adapter, err := skytap.Adapters.Create(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	var adapterExpected Interface
	err = json.Unmarshal([]byte(exampleAdapter), &adapterExpected)
	assert.Equal(t, adapterExpected, *adapter, "Bad adapter")
}

func TestCreateAdapter2(t *testing.T) {
	exampleAdapter := fmt.Sprintf(exampleAdapterResponse, 456, 123)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces", req.URL.Path, "Bad path")
		assert.Equal(t, "POST", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, exampleAdapterRequest2, string(body), "Bad request body")

		io.WriteString(rw, exampleAdapter)
	}
	opts := &CreateAdapterRequest{
		NICType:   nicTypeToPtr(NICTypeE1000),
		NetworkID: strToPtr("23917287"),
		IP:        strToPtr("10.0.0.1"),
		Hostname:  strToPtr("name.com"),
	}

	adapter, err := skytap.Adapters.Create(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	var adapterExpected Interface
	err = json.Unmarshal([]byte(exampleAdapter), &adapterExpected)
	assert.Equal(t, adapterExpected, *adapter, "Bad adapter")
}

func TestReadAdapter(t *testing.T) {
	exampleAdapter := fmt.Sprintf(exampleAdapterResponse, 456, 123)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, exampleAdapter)
	}

	adapter, err := skytap.Adapters.Get(context.Background(), "123", "456", "789")
	assert.Nil(t, err, "Bad API method")

	var adapterExpected Interface
	err = json.Unmarshal([]byte(exampleAdapter), &adapterExpected)
	assert.Equal(t, adapterExpected, *adapter, "Bad Adapter")
}

func TestUpdateAdapter(t *testing.T) {
	exampleAdapter := fmt.Sprintf(exampleAdapterResponse, 456, 123)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var adapter Interface
	json.Unmarshal([]byte(exampleAdapter), &adapter)
	adapter.Hostname = strToPtr("updated adapter")

	bytes, err := json.Marshal(&adapter)
	assert.Nil(t, err, "Bad adapter")

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
		assert.Equal(t, "PUT", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, `{"hostname": "updated adapter"}`, string(body), "Bad request body")

		io.WriteString(rw, string(bytes))
	}

	opts := &UpdateAdapterRequest{
		Hostname: strToPtr(*adapter.Hostname),
	}
	adapterUpdate, err := skytap.Adapters.Update(context.Background(), "123", "456", "789", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, adapter, *adapterUpdate, "Bad adapter")
}

func TestDeleteAdapter(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces/789", req.URL.Path, "Bad path")
		assert.Equal(t, "DELETE", req.Method, "Bad method")
	}

	err := skytap.Adapters.Delete(context.Background(), "123", "456", "789")
	assert.Nil(t, err, "Bad API method")
}

func TestListAdapters(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456/interfaces", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, exampleAdapterListResponse)
	}

	result, err := skytap.Adapters.List(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")

	var found = false
	for _, adapter := range result.Value {
		if *adapter.ID == "nic-20246343-38367563-5" {
			found = true
			break
		}
	}
	assert.True(t, found, "Adapter not found")
}
