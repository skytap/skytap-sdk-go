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

type updateProjectResponse = createProjectResponse

func (c *Client) CreateProject(ctx context.Context, name string, summary string) (*Project, error) {
	request := createProjectRequest{
		Name: name,
	}

	req, err := c.newRequest("POST", "/projects.json", request)

	if err != nil {
		return nil, err
	}
	var createProjectResponse createProjectResponse
	_, err = c.do(req, &createProjectResponse)
	if err != nil {
		return nil, err
	}

	// update project after creation to establish the resource information.
	project, err := c.UpdateProject(ctx, createProjectResponse.Id, name, summary)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (c *Client) UpdateProject(ctx context.Context, id string, name string, summary string) (*Project, error) {
	request := updateProjectRequest{
		Name:    name,
		Summary: summary,
	}

	req, err := c.newRequest("PUT", "/projects/"+id, request)
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
