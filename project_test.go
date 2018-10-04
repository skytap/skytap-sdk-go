package skytap

import (
	"context"
	"github.com/opencredo/skytap-sdk-go-internal/options"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateProject(t *testing.T) {
	handler := http.NotFound
	hs := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		handler(rw, req)
	}))
	defer hs.Close()

	skytap, err := NewClient(context.Background(),
		options.WithUser("pegerto.fernandez@opencredo"),
		options.WithAPIToken(""))

	skytap.CreateProject(context.Background(), "test-project", "test project")

	assert.Nil(t, err)

}
