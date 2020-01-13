package skytap

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// Default URL paths
const (
	environmentLegacyBasePath = "/configurations"
	environmentBasePath       = "/v2/configurations"
)

// EnvironmentsService is the contract for the services provided on the Skytap Environment resource
type EnvironmentsService interface {
	List(ctx context.Context) (*EnvironmentListResult, error)
	Get(ctx context.Context, id string) (*Environment, error)
	Create(ctx context.Context, createEnvironmentRequest *CreateEnvironmentRequest) (*Environment, error)
	Update(ctx context.Context, id string, updateEnvironmentRequest *UpdateEnvironmentRequest) (*Environment, error)
	Delete(ctx context.Context, id string) error
	CreateTags(ctx context.Context, id string, createTagRequest []*CreateTagRequest) error
	DeleteTag(ctx context.Context, id string, tag string) error
	UpdateUserData(ctx context.Context, id string, userData *string) error
}

// EnvironmentsServiceClient is the EnvironmentsService implementation
type EnvironmentsServiceClient struct {
	client *Client
}

// Environment resource struct definitions.
type Environment struct {
	ID                      *string              `json:"id"`
	URL                     *string              `json:"url"`
	Name                    *string              `json:"name"`
	Description             *string              `json:"description"`
	Errors                  []string             `json:"errors"`
	ErrorDetails            []string             `json:"error_details"`
	Runstate                *EnvironmentRunstate `json:"runstate"`
	RateLimited             *bool                `json:"rate_limited"`
	LastRun                 *string              `json:"last_run"`
	SuspendOnIdle           *int                 `json:"suspend_on_idle"`
	SuspendAtTime           *string              `json:"suspend_at_time"`
	OwnerURL                *string              `json:"owner_url"`
	OwnerName               *string              `json:"owner_name"`
	OwnerID                 *string              `json:"owner_id"`
	VMCount                 *int                 `json:"vm_count"`
	Storage                 *int                 `json:"storage"`
	NetworkCount            *int                 `json:"network_count"`
	CreatedAt               *string              `json:"created_at"`
	Region                  *string              `json:"region"`
	RegionBackend           *string              `json:"region_backend"`
	SVMs                    *int                 `json:"svms"`
	CanSaveAsTemplate       *bool                `json:"can_save_as_template"`
	CanCopy                 *bool                `json:"can_copy"`
	CanDelete               *bool                `json:"can_delete"`
	CanChangeState          *bool                `json:"can_change_state"`
	CanShare                *bool                `json:"can_share"`
	CanEdit                 *bool                `json:"can_edit"`
	LabelCount              *int                 `json:"label_count"`
	LabelCategoryCount      *int                 `json:"label_category_count"`
	CanTag                  *bool                `json:"can_tag"`
	Tags                    []Tag                `json:"tags"`
	TagList                 *string              `json:"tag_list"`
	Alerts                  []Alert              `json:"alerts"`
	PublishedServiceCount   *int                 `json:"published_service_count"`
	PublicIPCount           *int                 `json:"public_ip_count"`
	AutoSuspendDescription  *string              `json:"auto_suspend_description"`
	Stages                  []Stage              `json:"stages"`
	StagedExecution         *StagedExecution     `json:"staged_execution"`
	SequencingEnabled       *bool                `json:"sequencing_enabled"`
	NoteCount               *int                 `json:"note_count"`
	ProjectCountForUser     *int                 `json:"project_count_for_user"`
	ProjectCount            *int                 `json:"project_count"`
	PublishSetCount         *int                 `json:"publish_set_count"`
	ScheduleCount           *int                 `json:"schedule_count"`
	VpnCount                *int                 `json:"vpn_count"`
	OutboundTraffic         *bool                `json:"outbound_traffic"`
	Routable                *bool                `json:"routable"`
	VMs                     []VM                 `json:"vms"`
	Networks                []Network            `json:"networks"`
	ContainersCount         *int                 `json:"containers_count"`
	ContainerHostsCount     *int                 `json:"container_hosts_count"`
	PlatformErrors          []string             `json:"platform_errors"`
	SVMsByArchitecture      *SVMsByArchitecture  `json:"svms_by_architecture"`
	AllVmsSupportSuspend    *bool                `json:"all_vms_support_suspend"`
	ShutdownOnIdle          *int                 `json:"shutdown_on_idle"`
	ShutdownAtTime          *string              `json:"shutdown_at_time"`
	AutoShutdownDescription *string              `json:"auto_shutdown_description"`
	UserData                *string              `json:"-"`
}

// Tag describes environment tag data
type Tag struct {
	ID    *string `json:"id"`
	Value *string `json:"value"`
}

// Alert describes an environment alert
type Alert struct {
	ID                   string `json:"id"`
	DisplayType          string `json:"display_type"`
	Dismissable          bool   `json:"dismissable"`
	Message              string `json:"message"`
	DisplayOnGeneral     bool   `json:"display_on_general"`
	DisplayOnLogin       bool   `json:"display_on_login"`
	DisplayOnSmartclient bool   `json:"display_on_smartclient"`
}

// Stage describes the VM stage sequence
type Stage struct {
	DelayAfterFinishSeconds *int     `json:"delay_after_finish_seconds"`
	Index                   *int     `json:"index"`
	VMIDs                   []string `json:"vm_ids"`
}

// StagedExecution describes the status of a running VM sequence
type StagedExecution struct {
	ActionType                          *string  `json:"action_type"`
	CurrentStageDelayAfterFinishSeconds *int     `json:"current_stage_delay_after_finish_seconds"`
	CurrentStageIndex                   *int     `json:"current_stage_index"`
	CurrentStageFinishedAt              *string  `json:"current_stage_finished_at"`
	VMIDs                               []string `json:"vm_ids"`
}

// SVMsByArchitecture lists the number of x86 and power SVMs consumed by VMs in the environment
type SVMsByArchitecture struct {
	X86   *int `json:"x86"`
	Power *int `json:"power"`
}

// EnvironmentRunstate enumerates the possible environment running states
type EnvironmentRunstate string

// The environment running states
const (
	EnvironmentRunstateStopped   EnvironmentRunstate = "stopped"
	EnvironmentRunstateSuspended EnvironmentRunstate = "suspended"
	EnvironmentRunstateRunning   EnvironmentRunstate = "running"
	EnvironmentRunstateHalted    EnvironmentRunstate = "halted"
	EnvironmentRunstateBusy      EnvironmentRunstate = "busy"
)

// EnvironmentListResult is the list request specific struct
type EnvironmentListResult struct {
	Value []Environment
}

// CreateTagRequest describe the creation of a tag
type CreateTagRequest struct {
	Tag string `json:"value,omitempty"`
}

type userData struct {
	Contents *string `json:"contents,omitempty"`
}

// CreateEnvironmentRequest describes the create the environment data
type CreateEnvironmentRequest struct {
	TemplateID      *string             `json:"template_id,omitempty"`
	ProjectID       *int                `json:"project_id,omitempty"`
	Name            *string             `json:"name,omitempty"`
	Description     *string             `json:"description,omitempty"`
	Owner           *string             `json:"owner,omitempty"`
	OutboundTraffic *bool               `json:"outbound_traffic,omitempty"`
	Routable        *bool               `json:"routable,omitempty"`
	SuspendOnIdle   *int                `json:"suspend_on_idle,omitempty"`
	SuspendAtTime   *string             `json:"suspend_at_time,omitempty"`
	ShutdownOnIdle  *int                `json:"shutdown_on_idle,omitempty"`
	ShutdownAtTime  *string             `json:"shutdown_at_time,omitempty"`
	Tags            []*CreateTagRequest `json:"-"`
	UserData        *string             `json:"-"`
}

// UpdateEnvironmentRequest describes the update the environment data
type UpdateEnvironmentRequest struct {
	Name            *string              `json:"name,omitempty"`
	Description     *string              `json:"description,omitempty"`
	Owner           *string              `json:"owner,omitempty"`
	OutboundTraffic *bool                `json:"outbound_traffic,omitempty"`
	Routable        *bool                `json:"routable,omitempty"`
	SuspendOnIdle   *int                 `json:"suspend_on_idle,omitempty"`
	SuspendAtTime   *string              `json:"suspend_at_time,omitempty"`
	ShutdownOnIdle  *int                 `json:"shutdown_on_idle,omitempty"`
	ShutdownAtTime  *string              `json:"shutdown_at_time,omitempty"`
	Runstate        *EnvironmentRunstate `json:"runstate,omitempty"`
}

// List the environments
func (s *EnvironmentsServiceClient) List(ctx context.Context) (*EnvironmentListResult, error) {
	req, err := s.client.newRequest(ctx, "GET", environmentBasePath, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var environmentsListResponse EnvironmentListResult
	_, err = s.client.do(ctx, req, &environmentsListResponse.Value, nil, nil)
	if err != nil {
		return nil, err
	}

	return &environmentsListResponse, nil
}

// Get an environment
func (s *EnvironmentsServiceClient) Get(ctx context.Context, id string) (*Environment, error) {
	path := fmt.Sprintf("%s/%s", environmentBasePath, id)

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var environment Environment
	_, err = s.client.do(ctx, req, &environment, nil, nil)
	if err != nil {
		return nil, err
	}

	// Read user data for the environment with the full configuration, this is done here instead in an different
	// exported method to maintain similarity with other sub resources as
	// tags and labels.
	userDataPath := fmt.Sprintf("%s/%s/user_data.json", environmentBasePath, id)
	req, err = s.client.newRequest(ctx, "GET", userDataPath, nil)
	if err != nil {
		return nil, err
	}
	var userDataResp userData
	_, err = s.client.do(ctx, req, &userDataResp, nil, nil)
	if err != nil {
		return nil, err
	}
	environment.UserData = userDataResp.Contents
	return &environment, nil
}

// Create an environment
func (s *EnvironmentsServiceClient) Create(ctx context.Context, opts *CreateEnvironmentRequest) (*Environment, error) {
	req, err := s.client.newRequest(ctx, "POST", environmentLegacyBasePath, opts)
	if err != nil {
		return nil, err
	}

	var createdEnvironment Environment
	_, err = s.client.do(ctx, req, &createdEnvironment, nil, opts)
	if err != nil {
		return nil, err
	}

	env, err := s.Get(ctx, *createdEnvironment.ID)
	if err != nil {
		return nil, err
	}

	var runstate *EnvironmentRunstate
	if *env.VMCount > 0 {
		runstate = environmentRunStateToPtr(EnvironmentRunstateRunning)
	}

	updateOpts := &UpdateEnvironmentRequest{
		Name:            opts.Name,
		Description:     opts.Description,
		Owner:           opts.Owner,
		OutboundTraffic: opts.OutboundTraffic,
		Routable:        opts.Routable,
		SuspendOnIdle:   opts.SuspendOnIdle,
		SuspendAtTime:   opts.SuspendAtTime,
		ShutdownOnIdle:  opts.ShutdownOnIdle,
		ShutdownAtTime:  opts.ShutdownAtTime,
		Runstate:        runstate, // we are expecting the environment to start its VMs after creation
	}

	// update environment after creation to establish the resource information.
	environment, err := s.Update(ctx, ptrToStr(createdEnvironment.ID), updateOpts)
	if err != nil {
		return nil, err
	}

	// update user data, before the update on the runtime, to avoid vms to start without the metadata update
	if err := s.UpdateUserData(ctx, ptrToStr(createdEnvironment.ID), opts.UserData); err != nil {
		return nil, err
	}

	// update tags
	if err := s.CreateTags(ctx, ptrToStr(createdEnvironment.ID), opts.Tags); err != nil {
		return nil, err
	}

	return environment, nil
}

// Update an environment
func (s *EnvironmentsServiceClient) Update(ctx context.Context, id string, updateEnvironment *UpdateEnvironmentRequest) (*Environment, error) {
	path := fmt.Sprintf("%s/%s", environmentBasePath, id)

	req, err := s.client.newRequest(ctx, "PUT", path, updateEnvironment)
	if err != nil {
		return nil, err
	}

	var environment Environment
	_, err = s.client.do(ctx, req, &environment, envRunStateNotBusy(id), updateEnvironment)
	if err != nil {
		return nil, err
	}

	return &environment, nil
}

// Delete an environment
func (s *EnvironmentsServiceClient) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", environmentLegacyBasePath, id)

	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	_, err = s.client.do(ctx, req, nil, envRunStateNotBusy(id), nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateTags add tags to the specific environment
func (s *EnvironmentsServiceClient) CreateTags(ctx context.Context, id string, createTagRequest []*CreateTagRequest) error {
	if createTagRequest == nil || len(createTagRequest) == 0 {
		return nil
	}

	tagPath := fmt.Sprintf("%s/%s/tags.json", environmentBasePath, id)
	req, err := s.client.newRequest(ctx, "PUT", tagPath, createTagRequest)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// DeleteTag remove tags given the tag and the environment
func (s *EnvironmentsServiceClient) DeleteTag(ctx context.Context, id string, tagID string) error {
	path := fmt.Sprintf("%s/%s/tags/%s.json", environmentBasePath, id, tagID)

	req, err := s.client.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	var tags []*Tag
	_, err = s.client.do(ctx, req, tags, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

//UpdateUserData changes enviroment user data
func (s *EnvironmentsServiceClient) UpdateUserData(ctx context.Context, id string, userDataUpdate *string) error {
	if userDataUpdate == nil {
		return nil
	}
	tagPath := fmt.Sprintf("%s/%s/user_data.json", environmentBasePath, id)
	userDataRequest := userData{
		Contents: userDataUpdate,
	}
	req, err := s.client.newRequest(ctx, "PUT", tagPath, userDataRequest)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (payload *CreateEnvironmentRequest) compareResponse(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool) {
	if envOriginal, ok := v.(*Environment); ok {
		env, err := c.Environments.Get(ctx, *envOriginal.ID)
		if err != nil {
			return requestNotAsExpected, false
		}
		logEnvironmentStatus(env)
		log.Printf("[DEBUG] SDK environment runstate after create (%s)\n", *env.Runstate)
		if *env.Runstate != EnvironmentRunstateBusy {
			return "", true
		}
		return "environment not ready", false
	}
	log.Printf("[ERROR] SDK environment comparison not possible on (%v)\n", v)
	return requestNotAsExpected, false
}

func (payload *UpdateEnvironmentRequest) compareResponse(ctx context.Context, c *Client, v interface{}, state *environmentVMRunState) (string, bool) {
	if envOriginal, ok := v.(*Environment); ok {
		env, err := c.Environments.Get(ctx, *envOriginal.ID)
		if err != nil {
			return requestNotAsExpected, false
		}
		logEnvironmentStatus(env)
		actual := payload.buildComparison(env)
		if payload.string() == actual.string() {
			return "", true
		}
		return "environment not ready", false
	}
	log.Printf("[ERROR] SDK environment comparison not possible on (%v)\n", v)
	return requestNotAsExpected, false
}

func (payload *UpdateEnvironmentRequest) buildComparison(env *Environment) UpdateEnvironmentRequest {
	actual := UpdateEnvironmentRequest{}
	if payload.Name != nil {
		actual.Name = env.Name
	}
	if payload.Description != nil {
		actual.Description = env.Description
	}
	if payload.Owner != nil {
		actual.Owner = env.OwnerName
	}
	if payload.OutboundTraffic != nil {
		actual.OutboundTraffic = env.OutboundTraffic
	}
	if payload.Routable != nil {
		actual.Routable = env.Routable
	}
	if payload.SuspendOnIdle != nil {
		actual.SuspendOnIdle = env.SuspendOnIdle
	}
	if payload.SuspendAtTime != nil {
		actual.SuspendAtTime = env.SuspendAtTime
	}
	if payload.ShutdownOnIdle != nil {
		actual.ShutdownOnIdle = env.ShutdownOnIdle
	}
	if payload.ShutdownAtTime != nil {
		actual.ShutdownAtTime = env.ShutdownAtTime
	}
	if payload.Runstate != nil {
		actual.Runstate = env.Runstate
	}
	return actual
}

func (payload *UpdateEnvironmentRequest) string() string {
	name := ""
	description := ""
	owner := ""
	outboundTraffic := ""
	routable := "false"
	suspendOnIdle := ""
	suspendAtTime := ""
	shutdownOnIdle := ""
	shutdownAtTime := ""
	runstate := ""

	if payload.Name != nil {
		name = *payload.Name
	}
	if payload.Description != nil {
		description = *payload.Description
	}
	if payload.Owner != nil {
		owner = *payload.Owner
	}
	if payload.OutboundTraffic != nil {
		outboundTraffic = fmt.Sprintf("%t", *payload.OutboundTraffic)
	}
	if payload.Routable != nil {
		routable = fmt.Sprintf("%t", *payload.Routable)
	}
	if payload.SuspendOnIdle != nil {
		suspendOnIdle = fmt.Sprintf("%d", *payload.SuspendOnIdle)
	}
	if payload.SuspendAtTime != nil {
		suspendAtTime = *payload.SuspendAtTime
	}
	if payload.ShutdownOnIdle != nil {
		shutdownOnIdle = fmt.Sprintf("%d", *payload.ShutdownOnIdle)
	}
	if payload.ShutdownAtTime != nil {
		shutdownAtTime = *payload.ShutdownAtTime
	}
	if payload.Runstate != nil {
		runstate = string(*payload.Runstate)
	}
	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString(description)
	sb.WriteString(owner)
	sb.WriteString(outboundTraffic)
	sb.WriteString(routable)
	sb.WriteString(suspendOnIdle)
	sb.WriteString(suspendAtTime)
	sb.WriteString(shutdownOnIdle)
	sb.WriteString(shutdownAtTime)
	sb.WriteString(runstate)
	log.Printf("[DEBUG] SDK environment payload (%s)\n", sb.String())
	return sb.String()
}

func logEnvironmentStatus(env *Environment) {
	if env.RateLimited != nil && *env.RateLimited {
		log.Printf("[INFO] SDK environment rate limiting detected\n")
	}
	if len(env.Errors) > 0 {
		log.Printf("[INFO] SDK environment errors detected: (%s)\n",
			strings.Join(env.Errors, ", "))
	}
	if len(env.ErrorDetails) > 0 {
		log.Printf("[INFO] SDK environment errors detected: (%s)\n",
			strings.Join(env.ErrorDetails, ", "))
	}
}
