package skytap

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createClient(t *testing.T) (*Client, *httptest.Server, *func(rw http.ResponseWriter, req *http.Request)) {
	handler := http.NotFound
	hs := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		handler(rw, req)
	}))

	var user = "SKYTAP_USER"
	var token = "SKYTAP_ACCESS_TOKEN"

	settings := NewDefaultSettings(WithBaseUrl(hs.URL), WithCredentialsProvider(NewApiTokenCredentials(user, token)))

	skytap, err := NewClient(settings)

	assert.Nil(t, err)
	assert.NotNil(t, skytap)
	return skytap, hs, &handler
}

func TestCreateProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var createPhase = true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if createPhase {
			if req.URL.Path != "/projects" {
				t.Error("Bad path:", req.URL.Path)
			}
			if req.Method != "POST" {
				t.Error("Bad method")
			}
			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err)
			assert.JSONEq(t, `{"name":"test-project","summary":"test project"}`, string(body))
			io.WriteString(rw, `{"id": "12345", "name": "test-project"}`)
			createPhase = false
		} else {
			if req.URL.Path != "/projects/12345" {
				t.Error("Bad path:", req.URL.Path)
			}
			if req.Method != "PUT" {
				t.Error("Bad method")
			}
			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err)
			assert.JSONEq(t, `{"id": "12345","name":"test-project","summary":"test project"}`, string(body))
			io.WriteString(rw, `{"id": "12345", "name": "test-project", "summary": "test project"}`)
		}
	}

	opts := Project{
		Name:    StringPtr("test-project"),
		Summary: StringPtr("test project"),
	}

	project, err := skytap.Projects.Create(context.Background(), &opts)

	assert.Nil(t, err)
	assert.Equal(t, &Project{Id: project.Id, Name: StringPtr("test-project"), Summary: StringPtr("test project")}, project)
}

func TestReadProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/projects/12345" {
			t.Error("Bad path:", req.URL.Path)
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		io.WriteString(rw, `{"id": "12345", "name": "test-project", "summary": "test project"}`)
	}

	projectRead, err := skytap.Projects.Get(context.Background(), "12345")

	assert.Nil(t, err)
	assert.Equal(t, &Project{Id: StringPtr("12345"), Name: StringPtr("test-project"), Summary: StringPtr("test project")}, projectRead)
}

func TestUpdateProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/projects/12345" {
			t.Error("Bad path:", req.URL.Path)
		}
		if req.Method != "PUT" {
			t.Error("Bad method")
		}
		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		assert.JSONEq(t, `{"id": "12345","name":"updated name","summary":"updated summary"}`, string(body))
		io.WriteString(rw, `{"id": "12345", "name": "updated name", "summary": "updated summary"}`)
	}

	opts := &Project{
		Id:      StringPtr("12345"),
		Name:    StringPtr("updated name"),
		Summary: StringPtr("updated summary"),
	}

	projectUpdate, err := skytap.Projects.Update(context.Background(), "12345", opts)

	expectedResult := &Project{Id: StringPtr("12345"), Name: StringPtr("updated name"), Summary: StringPtr("updated summary")}

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, projectUpdate)
}

func TestDeleteProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/projects/12345" {
			t.Error("Bad path:", req.URL.Path)
		}
		if req.Method != "DELETE" {
			t.Error("Bad method")
		}
	}

	err := skytap.Projects.Delete(context.Background(), "12345")
	assert.Nil(t, err)
}

func TestListProjects(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/projects" {
			t.Error("Bad path:", req.URL.Path)
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		io.WriteString(rw, `[{
        "id": "12345",
        "url": "https://cloud.skytap.com/projects/12345",
        "name": "updated name",
        "summary": "updated summary",
        "show_project_members": true,
        "auto_add_role_name": null
    }]`)
	}

	result, err := skytap.Projects.List(context.Background())

	assert.Nil(t, err)

	var found = false
	for _, project := range result.Value {
		if *project.Name == "updated name" {
			found = true
			break
		}
	}

	assert.True(t, found)
}
