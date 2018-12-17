package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadTemplate(t *testing.T) {
	response := string(readTestFile(t, "templateResponse.json"))

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/templates/456" {
			t.Error("Bad path")
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		_, err := io.WriteString(rw, response)
		assert.NoError(t, err)
	}

	template, err := skytap.Templates.Get(context.Background(), "456")

	assert.Nil(t, err)
	var templateExpected Template

	err = json.Unmarshal([]byte(response), &templateExpected)

	assert.Equal(t, templateExpected, *template)
}

func TestListTemplates(t *testing.T) {
	response := string(readTestFile(t, "templateResponse.json"))

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/templates" {
			t.Error("Bad path")
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		_, err := io.WriteString(rw, fmt.Sprintf(`[%+v]`, response))
		assert.NoError(t, err)
	}

	result, err := skytap.Templates.List(context.Background())

	assert.Nil(t, err)

	var found = false
	for _, template := range result.Value {
		if *template.Description == "test template" {
			found = true
			break
		}
	}

	assert.True(t, found)
}
