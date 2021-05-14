package skytap

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			_, err = io.WriteString(rw, `{"id": "12345", "name": "test-project"}`)
			assert.NoError(t, err)
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
			_, err = io.WriteString(rw, `{"id": "12345", "name": "test-project", "summary": "test project"}`)
			assert.NoError(t, err)
		}
	}

	opts := Project{
		Name:    strToPtr("test-project"),
		Summary: strToPtr("test project"),
	}

	project, err := skytap.Projects.Create(context.Background(), &opts)

	assert.Nil(t, err)
	assert.Equal(t, &Project{ID: project.ID, Name: strToPtr("test-project"), Summary: strToPtr("test project")}, project)
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
		_, err := io.WriteString(rw, `{"id": "12345", "name": "test-project", "summary": "test project"}`)
		assert.NoError(t, err)
	}

	projectRead, err := skytap.Projects.Get(context.Background(), 12345)

	assert.Nil(t, err)
	assert.Equal(t, &Project{ID: intToPtr(12345), Name: strToPtr("test-project"), Summary: strToPtr("test project")}, projectRead)
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
		_, err = io.WriteString(rw, `{"id": "12345", "name": "updated name", "summary": "updated summary"}`)
		assert.NoError(t, err)
	}

	opts := &Project{
		ID:      intToPtr(12345),
		Name:    strToPtr("updated name"),
		Summary: strToPtr("updated summary"),
	}

	projectUpdate, err := skytap.Projects.Update(context.Background(), 12345, opts)

	expectedResult := &Project{ID: intToPtr(12345), Name: strToPtr("updated name"), Summary: strToPtr("updated summary")}

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

	err := skytap.Projects.Delete(context.Background(), 12345)
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
		_, err := io.WriteString(rw, `[{
        "id": "12345",
        "url": "https://cloud.skytap.com/projects/12345",
        "name": "updated name",
        "summary": "updated summary",
        "show_project_members": true,
        "auto_add_role_name": null
    }]`)
		assert.NoError(t, err)
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

func TestListEnvironmentsInProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/projects/12345/configurations" {
			t.Error("Bad path:", req.URL.Path)
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		_, err := io.WriteString(rw, `[{
		"id": "6789",
		"url": "https://cloud.skytap.com/v2/configurations/6789",
		"name": "tftest-vm-environment-1930588755153240024",
		"description": "This is an environment to support a vm skytap terraform provider acceptance test",
		"errors": [],
		"error_details": [],
		"runstate": "busy",
		"rate_limited": false,
		"last_run": "2021/05/14 10:46:33 +0100",
		"suspend_on_idle": null,
		"suspend_at_time": null,
		"owner_url": "https://cloud.skytap.com/v2/users/468583",
		"owner_name": "Gary Digby",
		"owner_id": "468583",
		"vm_count": 1,
		"licensed_vm_count": 0,
		"storage": 30720,
		"network_count": 2,
		"created_at": "2021/05/14 10:31:55 +0100",
		"region": "US-West",
		"region_backend": "skytap",
		"svms": 1,
		"can_save_as_template": true,
		"can_copy": true,
		"can_delete": true,
		"can_change_state": true,
		"can_share": true,
		"can_edit": true,
		"label_count": 0,
		"label_category_count": 0,
		"can_tag": true,
		"can_change_owner": true,
		"tag_list": "",
		"alerts": [],
		"vms": [
			{
				"id": "4859384",
				"url": "https://cloud.skytap.com/v2/vms/4859384",
				"name": "cassandra1",
				"runstate": "running",
				"rate_limited": false,
				"error": null,
				"status": "running",
				"hardware": {
					"cpus": 1,
					"ram": 1024,
					"svms": 1,
					"storage": 30720,
					"guestOS": "centos-64",
					"architecture": "x86"
				},
				"license_types": [],
				"region_backend": "skytap",
				"supports_suspend": true,
				"environment_locked": false
			}
		],
		"container_hosts_count": 0,
		"platform_errors": [],
		"svms_by_architecture": {
			"x86": 1,
			"power": 0
		},
		"all_vms_support_suspend": true,
		"shutdown_on_idle": null,
		"shutdown_at_time": null,
		"environment_locked": false,
		"can_lock": true,
		"environment_lock": null
		}]`)
		assert.NoError(t, err)
	}

	result, err := skytap.Projects.ListEnvironments(context.Background(), 12345)
	assert.NoError(t, err)

	assert.Len(t, result.Value, 1)
	assert.Equal(t, "6789", result.Value[0].ID)
}

func TestAddEnvironmentToProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/projects/12345/configurations/6789" {
			t.Error("Bad path:", req.URL.Path)
		}
		if req.Method != "POST" {
			t.Error("Bad method")
		}
		_, err := io.WriteString(rw, `{
		"id": "6789",
		"url": "https://cloud.skytap.com/v2/configurations/6789",
		"name": "tftest-vm-environment-1930588755153240024",
		"description": "This is an environment to support a vm skytap terraform provider acceptance test",
		"errors": [],
		"error_details": [],
		"runstate": "busy",
		"rate_limited": false,
		"last_run": "2021/05/14 10:46:33 +0100",
		"suspend_on_idle": null,
		"suspend_at_time": null,
		"owner_url": "https://cloud.skytap.com/v2/users/468583",
		"owner_name": "Gary Digby",
		"owner_id": "468583",
		"vm_count": 1,
		"licensed_vm_count": 0,
		"storage": 30720,
		"network_count": 2,
		"created_at": "2021/05/14 10:31:55 +0100",
		"region": "US-West",
		"region_backend": "skytap",
		"svms": 1,
		"can_save_as_template": true,
		"can_copy": true,
		"can_delete": true,
		"can_change_state": true,
		"can_share": true,
		"can_edit": true,
		"label_count": 0,
		"label_category_count": 0,
		"can_tag": true,
		"can_change_owner": true,
		"tag_list": "",
		"alerts": [],
		"vms": [
			{
				"id": "4859384",
				"url": "https://cloud.skytap.com/v2/vms/4859384",
				"name": "cassandra1",
				"runstate": "running",
				"rate_limited": false,
				"error": null,
				"status": "running",
				"hardware": {
					"cpus": 1,
					"ram": 1024,
					"svms": 1,
					"storage": 30720,
					"guestOS": "centos-64",
					"architecture": "x86"
				},
				"license_types": [],
				"region_backend": "skytap",
				"supports_suspend": true,
				"environment_locked": false
			}
		],
		"container_hosts_count": 0,
		"platform_errors": [],
		"svms_by_architecture": {
			"x86": 1,
			"power": 0
		},
		"all_vms_support_suspend": true,
		"shutdown_on_idle": null,
		"shutdown_at_time": null,
		"environment_locked": false,
		"can_lock": true,
		"environment_lock": null
	}`)
		assert.NoError(t, err)
	}

	env, err := skytap.Projects.AddEnvironment(context.Background(), 12345, "6789")
	assert.NoError(t, err)

	assert.Equal(t, "6789", env.ID)
}

func TestRemoveEnvironmentFromProject(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/projects/12345/configurations/6789" {
			t.Error("Bad path:", req.URL.Path)
		}
		if req.Method != "DELETE" {
			t.Error("Bad method")
		}
	}

	err := skytap.Projects.RemoveEnvironment(context.Background(), 12345, "6789")
	assert.NoError(t, err)
}
