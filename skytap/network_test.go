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

const exampleNetworkRequest = `{"name": 
    "test network",
    "network_type": "automatic", 
    "subnet": "10.0.2.0/24", 
    "domain": "sampledomain.com", 
    "gateway": "10.0.2.254", 
    "tunnelable": true
}`

const exampleNetworkResponse = `{"id": "%d",
    "url": "https://cloud.skytap.com/v2/configurations/%d/networks/%d",
	"name": "test network",
	"network_type": "automatic",
	"subnet": "10.0.2.0/24",
	"subnet_addr": "10.0.2.0",
	"subnet_size": 24,
	"gateway": "10.0.2.254",
	"primary_nameserver": null,
	"secondary_nameserver": null,
	"region": "US-West",
	"domain": "sampledomain.com",
	"vpn_attachments": [],
	"tunnelable": true,
	"tunnels": []
}`

func TestCreateNetwork(t *testing.T) {
	exampleNetwork := fmt.Sprintf(exampleNetworkResponse, 456, 123, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/networks", req.URL.Path, "Bad path")
		assert.Equal(t, "POST", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, exampleNetworkRequest, string(body), "Bad request body")

		io.WriteString(rw, exampleNetwork)
	}
	opts := &CreateNetworkRequest{
		Name:       strToPtr("test network"),
		Subnet:     strToPtr("10.0.2.0/24"),
		Gateway:    strToPtr("10.0.2.254"),
		Tunnelable: boolToPtr(true),
		Domain:     strToPtr("sampledomain.com"),
	}

	network, err := skytap.Networks.Create(context.Background(), "123", opts)
	assert.Nil(t, err, "Bad API method")

	var networkExpected Network
	err = json.Unmarshal([]byte(exampleNetwork), &networkExpected)
	assert.Equal(t, networkExpected, *network, "Bad network")
}

func TestReadNetwork(t *testing.T) {
	exampleNetwork := fmt.Sprintf(exampleNetworkResponse, 456, 123, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/networks/456", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, exampleNetwork)
	}

	network, err := skytap.Networks.Get(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")

	var networkExpected Network
	err = json.Unmarshal([]byte(exampleNetwork), &networkExpected)
	assert.Equal(t, networkExpected, *network, "Bad Network")
}

func TestUpdateNetwork(t *testing.T) {
	exampleNetwork := fmt.Sprintf(exampleNetworkResponse, 456, 123, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var network Network
	json.Unmarshal([]byte(exampleNetwork), &network)
	*network.Name = "updated network"

	bytes, err := json.Marshal(&network)
	assert.Nil(t, err, "Bad network")

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/networks/456", req.URL.Path, "Bad path")
		assert.Equal(t, "PUT", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, `{"name": "updated network"}`, string(body), "Bad request body")

		io.WriteString(rw, string(bytes))
	}

	opts := &UpdateNetworkRequest{
		Name: strToPtr(*network.Name),
	}
	networkUpdate, err := skytap.Networks.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, network, *networkUpdate, "Bad network")
}

func TestDeleteNetwork(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/networks/456", req.URL.Path, "Bad path")
		assert.Equal(t, "DELETE", req.Method, "Bad method")
	}

	err := skytap.Networks.Delete(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")
}

func TestListNetworks(t *testing.T) {
	exampleNetwork := fmt.Sprintf(exampleNetworkResponse, 456, 123, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/networks", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, fmt.Sprintf(`[%+v]`, exampleNetwork))
	}

	result, err := skytap.Networks.List(context.Background(), "123")
	assert.Nil(t, err, "Bad API method")

	var found = false
	for _, network := range result.Value {
		if *network.Name == "test network" {
			found = true
			break
		}
	}
	assert.True(t, found, "Network not found")
}
