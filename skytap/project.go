package skytap

import (
	"context"
	"fmt"
)

const (
	projectsBasePath = "/projects"
)

type ProjectsService interface {
	List(ctx context.Context) (*ProjectListResult, error)
	Get(ctx context.Context, id string) (*Project, error)
	Create(ctx context.Context, project *Project) (*Project, error)
	Update(ctx context.Context, id string, project *Project) (*Project, error)
	Delete(ctx context.Context, id string) error
}

// Project service implementation
type ProjectsServiceClient struct {
	client *Client
}

// Project resource struct definitions
type Project struct {
	Id      *string `json:"id,omitempty"`
	Name    *string `json:"name,omitempty"`
	Summary *string `json:"summary,omitempty"`
}

// Request specific structs
type ProjectListResult struct {
	Value []Project
}

func (s *ProjectsServiceClient) List(ctx context.Context) (*ProjectListResult, error) {
	req, err := s.client.newRequest(ctx, "GET", projectsBasePath, nil)
	if err != nil {
		return nil, err
	}

	var projectListResponse ProjectListResult
	_, err = s.client.do(ctx, req, &projectListResponse.Value)
	if err != nil {
		return nil, err
	}

	return &projectListResponse, nil
}

func (s *ProjectsServiceClient) Get(ctx context.Context, id string) (*Project, error) {
	path := fmt.Sprintf("%s/%s", projectsBasePath, id)

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var project Project
	_, err = s.client.do(ctx, req, &project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (s *ProjectsServiceClient) Create(ctx context.Context, project *Project) (*Project, error) {
	req, err := s.client.newRequest(ctx, "POST", projectsBasePath, project)
	if err != nil {
		return nil, err
	}

	var createdProject Project
	_, err = s.client.do(ctx, req, &createdProject)
	if err != nil {
		return nil, err
	}

	createdProject.Summary = project.Summary

	// update project after creation to establish the resource information.
	updatedProject, err := s.Update(ctx, *createdProject.Id, &createdProject)
	if err != nil {
		return nil, err
	}

	return updatedProject, nil
}

func (s *ProjectsServiceClient) Update(ctx context.Context, id string, project *Project) (*Project, error) {
	path := fmt.Sprintf("%s/%s", projectsBasePath, id)

	req, err := s.client.newRequest(ctx, "PUT", path, project)
	if err != nil {
		return nil, err
	}

	var updatedProject Project
	_, err = s.client.do(ctx, req, &updatedProject)
	if err != nil {
		return nil, err
	}

	return &updatedProject, nil
}

func (s *ProjectsServiceClient) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", projectsBasePath, id)

	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
