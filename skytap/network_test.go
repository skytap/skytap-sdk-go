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

func TestCreateNetwork(t *testing.T) {
	exampleNetwork := fmt.Sprintf(string(readTestFile(t, "exampleNetworkResponse.json")), 456, 123, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(readTestFile(t, "exampleEnvironment.json")))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/networks", req.URL.Path, "Bad path")
			assert.Equal(t, "POST", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, string(readTestFile(t, "exampleNetworkRequest.json")), string(body), "Bad request body")

			_, err = io.WriteString(rw, exampleNetwork)
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123/networks/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, exampleNetwork)
			assert.NoError(t, err)
		}
		requestCounter++
	}
	opts := &CreateNetworkRequest{
		Name:        strToPtr("test network"),
		Subnet:      strToPtr("10.0.2.0/24"),
		Gateway:     strToPtr("10.0.2.254"),
		Tunnelable:  boolToPtr(true),
		Domain:      strToPtr("sampledomain.com"),
		NetworkType: networkTypeToPtr(NetworkTypeAutomatic),
	}

	network, err := skytap.Networks.Create(context.Background(), "123", opts)
	assert.Nil(t, err, "Bad API method")

	var networkExpected Network
	err = json.Unmarshal([]byte(exampleNetwork), &networkExpected)
	assert.Equal(t, networkExpected, *network, "Bad network")

	assert.Equal(t, 3, requestCounter)
}

func TestReadNetwork(t *testing.T) {
	exampleNetwork := fmt.Sprintf(string(readTestFile(t, "exampleNetworkResponse.json")), 456, 123, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/networks/456", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, exampleNetwork)
		assert.NoError(t, err)
	}

	network, err := skytap.Networks.Get(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")

	var networkExpected Network
	err = json.Unmarshal([]byte(exampleNetwork), &networkExpected)
	assert.Equal(t, networkExpected, *network, "Bad Network")
}

func TestUpdateNetwork(t *testing.T) {
	exampleNetwork := fmt.Sprintf(string(readTestFile(t, "exampleNetworkResponse.json")), 456, 123, 456)

	var network Network
	err := json.Unmarshal([]byte(exampleNetwork), &network)
	assert.NoError(t, err)
	*network.Name = "updated network"
	b, err := json.Marshal(&network)
	assert.Nil(t, err, "Bad network")

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(readTestFile(t, "exampleEnvironment.json")))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/networks/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"name": "updated network"}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(b))
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123/networks/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(b))
			assert.NoError(t, err)
		}
		requestCounter++
	}

	opts := &UpdateNetworkRequest{
		Name: strToPtr(*network.Name),
	}
	networkUpdate, err := skytap.Networks.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, network, *networkUpdate, "Bad network")

	assert.Equal(t, 3, requestCounter)
}

func TestDeleteNetwork(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		assert.Equal(t, "/v2/configurations/123/networks/456", req.URL.Path, "Bad path")
		assert.Equal(t, "DELETE", req.Method, "Bad method")
		requestCounter++
	}

	err := skytap.Networks.Delete(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, 1, requestCounter)
}

func TestListNetworks(t *testing.T) {
	exampleNetwork := fmt.Sprintf(string(readTestFile(t, "exampleNetworkResponse.json")), 456, 123, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/networks", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, fmt.Sprintf(`[%+v]`, exampleNetwork))
		assert.NoError(t, err)
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

func TestCompareNetworkCreateTrue(t *testing.T) {
	exampleNetwork := fmt.Sprintf(string(readTestFile(t, "exampleNetworkResponse.json")), 456, 123, 456)

	var network Network
	err := json.Unmarshal([]byte(exampleNetwork), &network)
	assert.NoError(t, err)
	opts := CreateNetworkRequest{
		Name:        strToPtr("test network"),
		Subnet:      strToPtr("10.0.2.0/24"),
		Gateway:     strToPtr("10.0.2.254"),
		Tunnelable:  boolToPtr(true),
		Domain:      strToPtr("sampledomain.com"),
		NetworkType: networkTypeToPtr(NetworkTypeAutomatic),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, exampleNetwork)
		assert.NoError(t, err)
	}
	message, ok := opts.compare(context.Background(), skytap, &network, envRunStateNotBusy("123"))
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestCompareNetworkCreateFalse(t *testing.T) {
	exampleNetwork := fmt.Sprintf(string(readTestFile(t, "exampleNetworkResponse.json")), 456, 123, 456)

	var network Network
	err := json.Unmarshal([]byte(exampleNetwork), &network)
	assert.NoError(t, err)
	opts := CreateNetworkRequest{
		Name: strToPtr("test network2"),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, exampleNetwork)
		assert.NoError(t, err)
	}
	message, ok := opts.compare(context.Background(), skytap, &network, envRunStateNotBusy("123"))
	assert.False(t, ok)
	assert.Equal(t, "network not ready", message)
}

func TestCompareNetworkUpdateTrue(t *testing.T) {
	exampleNetwork := fmt.Sprintf(string(readTestFile(t, "exampleNetworkResponse.json")), 456, 123, 456)

	var network Network
	err := json.Unmarshal([]byte(exampleNetwork), &network)
	assert.NoError(t, err)
	opts := UpdateNetworkRequest{
		Name:       strToPtr("test network"),
		Subnet:     strToPtr("10.0.2.0/24"),
		Gateway:    strToPtr("10.0.2.254"),
		Tunnelable: boolToPtr(true),
		Domain:     strToPtr("sampledomain.com"),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, exampleNetwork)
		assert.NoError(t, err)
	}
	message, ok := opts.compare(context.Background(), skytap, &network, envRunStateNotBusy("123"))
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestCompareNetworkUpdateFalse(t *testing.T) {
	exampleNetwork := fmt.Sprintf(string(readTestFile(t, "exampleNetworkResponse.json")), 456, 123, 456)

	var network Network
	err := json.Unmarshal([]byte(exampleNetwork), &network)
	assert.NoError(t, err)
	opts := UpdateNetworkRequest{
		Name: strToPtr("test network2"),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, exampleNetwork)
		assert.NoError(t, err)
	}
	message, ok := opts.compare(context.Background(), skytap, &network, envRunStateNotBusy("123"))
	assert.False(t, ok)
	assert.Equal(t, "network not ready", message)
}
