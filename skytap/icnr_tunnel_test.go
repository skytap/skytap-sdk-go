package skytap

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestICNRTunnelGet(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.RequestURI, "/v2/tunnels/tunnel-123456-789011.json")
		assert.Equal(t, req.Method, "GET")
		_, err := io.WriteString(rw, string(readTestFile(t, "exampleICNRTunnelResponse.json")))
		assert.NoError(t, err)
	}

	tunnel, err := skytap.ICNRTunnel.Get(context.Background(), "tunnel-123456-789011")
	assert.Nil(t, err)
	assert.NotNil(t, tunnel)
	assert.Equal(t, *tunnel.ID, "tunnel-123456-789011")
	assert.Equal(t, *tunnel.Target.ID, "111111")
	assert.Equal(t, *tunnel.Source.ID, "000000")
}

func TestICNRTunnelCreate(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	created := false
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		created = true

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		assert.JSONEq(t, `{"source_network_id": 111111, "target_network_id": 0}`, string(body))

		assert.Equal(t, req.RequestURI, "/v2/tunnels")
		assert.Equal(t, req.Method, "POST")
		_, err = io.WriteString(rw, string(readTestFile(t, "exampleICNRTunnelResponse.json")))
		assert.NoError(t, err)
	}

	tunnel, err := skytap.ICNRTunnel.Create(context.Background(), 111111, 0)
	assert.Nil(t, err)
	assert.NotNil(t, tunnel)
	assert.True(t, created)
	assert.Equal(t, *tunnel.ID, "tunnel-123456-789011")
}

func TestICNRTunnelDelete(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	deleted := false
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		deleted = true
		assert.Equal(t, req.RequestURI, "/v2/tunnels/tunnel-123-456.json")
		assert.Equal(t, req.Method, "DELETE")
	}

	err := skytap.ICNRTunnel.Delete(context.Background(), "tunnel-123-456")
	assert.Nil(t, err)
	assert.True(t, deleted)
}
