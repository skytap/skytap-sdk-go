package skytap

import (
	"context"
)

// A representation of a project, projects are an access permission model
type Project struct {
	Id      string
	Name    string
	Summary string
}

type createProjectRequest struct {
	Name string `json:"name"`
}

type createProjectResponse struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

type updateProjectRequest struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

type listProjectsRequest struct {
}

type deleteProjectRequest struct {
	Id string
}

type readProjectRequest struct {
	Id string
}

type updateProjectResponse = createProjectResponse

type readProjectResponse = createProjectResponse

type listProjectsResponse struct {
	Projects []listProjectItem
}

type listProjectItem struct {
	Id                 string
	Url                string
	Name               string
	Summary            string
	ShowProjectMembers bool   `json:"show_project_members"`
	AutoAddRoleName    string `json:"auto_add_role_name"`
}

func (c *Client) CreateProject(ctx context.Context, name string, summary string) (*Project, error) {
	request := createProjectRequest{
		Name: name,
	}

	req, err := c.newRequest("POST", "/projects", request)

	if err != nil {
		return nil, err
	}
	var createProjectResponse createProjectResponse
	_, err = c.do(req, &createProjectResponse)
	if err != nil {
		return nil, err
	}

	// update project after creation to establish the resource information.
	project, err := c.UpdateProject(ctx, &Project{
		createProjectResponse.Id,
		createProjectResponse.Name,
		summary})
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (c *Client) UpdateProject(ctx context.Context, project *Project) (*Project, error) {
	request := updateProjectRequest{
		Name:    project.Name,
		Summary: project.Summary,
	}

	req, err := c.newRequest("PUT", "/projects/"+project.Id, request)
	if err != nil {
		return nil, err
	}

	var updateProjectResponse updateProjectResponse
	_, err = c.do(req, &updateProjectResponse)
	if err != nil {
		return nil, err
	}

	return &Project{
		Id:      updateProjectResponse.Id,
		Name:    updateProjectResponse.Name,
		Summary: updateProjectResponse.Summary,
	}, nil
}

func (c *Client) ListProjects(ctx context.Context) (*[]listProjectItem, error) {
	request := listProjectsRequest{}

	req, err := c.newRequest("GET", "/projects", request)

	if err != nil {
		return nil, err
	}
	var listProjectsResponse listProjectsResponse
	_, err = c.do(req, &listProjectsResponse.Projects)
	if err != nil {
		return nil, err
	}

	return &listProjectsResponse.Projects, nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) error {
	request := deleteProjectRequest{}

	req, err := c.newRequest("DELETE", "/projects/"+id, request)

	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ReadProject(ctx context.Context, id string) (*Project, error) {
	request := readProjectRequest{}

	req, err := c.newRequest("GET", "/projects/"+id, request)

	if err != nil {
		return nil, err
	}
	var readProjectResponse readProjectResponse
	_, err = c.do(req, &readProjectResponse)
	if err != nil {
		return nil, err
	}

	return &Project{
		Id:      readProjectResponse.Id,
		Name:    readProjectResponse.Name,
		Summary: readProjectResponse.Summary,
	}, nil
}
