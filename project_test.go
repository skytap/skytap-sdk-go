package skytap

import (
	"context"
	"github.com/opencredo/skytap-sdk-go-internal/options"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func createClient(t *testing.T) (*Client, *httptest.Server, *func(rw http.ResponseWriter, req *http.Request)) {
	handler := http.NotFound
	hs := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		handler(rw, req)
	}))

	var user = "SKYTAP_USER"
	var token = "SKYTAP_ACCESS_TOKEN"

	var url url.URL
	urlPtr, err := url.Parse(hs.URL)

	assert.Nil(t, err)

	skytap, err := NewClient(context.Background(),
		options.WithUser(user),
		options.WithAPIToken(token),
		options.WithScheme(urlPtr.Scheme),
		options.WithHost(urlPtr.Host))

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
				t.Error("Bad path")
			}
			if req.Method != "POST" {
				t.Error("Bad method")
			}
			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err)
			assert.JSONEq(t, `{"name":"test-project"}`, string(body))
			io.WriteString(rw, `{"id": "12345", "name": "test-project"}`)
			createPhase = false
		} else {
			if req.URL.Path != "/projects/12345" {
				t.Error("Bad path")
			}
			if req.Method != "PUT" {
				t.Error("Bad method")
			}
			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err)
			assert.JSONEq(t, `{"name":"test-project","summary":"test project"}`, string(body))
			io.WriteString(rw, `{"id": "12345", "name": "test-project", "summary": "test project"}`)
		}
	}

	project, err := skytap.CreateProject(context.Background(), "test-project", "test project")

	assert.Nil(t, err)
	assert.Equal(t, &Project{Id: project.Id, Name: "test-project", Summary: "test project"}, project)
}

func TestReadProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/projects/12345" {
			t.Error("Bad path")
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		io.WriteString(rw, `{"id": "12345", "name": "test-project", "summary": "test project"}`)
	}

	projectRead, err := skytap.ReadProject(context.Background(), "12345")

	assert.Nil(t, err)
	assert.Equal(t, &Project{Id: "12345", Name: "test-project", Summary: "test project"}, projectRead)
}

func TestUpdateProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/projects/12345" {
			t.Error("Bad path")
		}
		if req.Method != "PUT" {
			t.Error("Bad method")
		}
		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		assert.JSONEq(t, `{"name":"updated name","summary":"updated summary"}`, string(body))
		io.WriteString(rw, `{"id": "12345", "name": "updated name", "summary": "updated summary"}`)
	}

	projectUpdate, err := skytap.UpdateProject(context.Background(), &Project{"12345",
		"updated name", "updated summary"})

	assert.Nil(t, err)
	assert.Equal(t, &Project{Id: "12345", Name: "updated name", Summary: "updated summary"}, projectUpdate)
}

func TestDeleteProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/projects/12345" {
			t.Error("Bad path")
		}
		if req.Method != "DELETE" {
			t.Error("Bad method")
		}
	}

	err := skytap.DeleteProject(context.Background(), "12345")
	assert.Nil(t, err)
}

func TestListProjects(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/projects" {
			t.Error("Bad path")
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

	projects, err := skytap.ListProjects(context.Background())

	assert.Nil(t, err)

	var found = false
	for _, project := range *projects {
		if project.Name == "updated name" {
			found = true
			break
		}
	}

	assert.True(t, found)
}
