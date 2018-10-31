package skytap

import "context"

// Default URL paths
const (
	servicesBasePath       = "/v2/configurations/"
	servicesVMPath         = "/vms/"
	servicesInterfacesPath = "/interfaces/"
	servicesPath           = "/services"
)

type servicePathBuilder interface {
	Environment(string) servicePathBuilder
	VM(string) servicePathBuilder
	Interface(string) servicePathBuilder
	Service(string) servicePathBuilder
	Build() string
}

type servicePathBuilderImpl struct {
	environment      string
	vm               string
	networkInterface string
	service          string
}

func (pb *servicePathBuilderImpl) Environment(environment string) servicePathBuilder {
	pb.environment = environment
	return pb
}

func (pb *servicePathBuilderImpl) VM(vm string) servicePathBuilder {
	pb.vm = vm
	return pb
}

func (pb *servicePathBuilderImpl) Interface(networkInterface string) servicePathBuilder {
	pb.networkInterface = networkInterface
	return pb
}

func (pb *servicePathBuilderImpl) Service(service string) servicePathBuilder {
	pb.service = service
	return pb
}

func (pb *servicePathBuilderImpl) Build() string {
	path := servicesBasePath + pb.environment + servicesVMPath + pb.vm + servicesInterfacesPath + pb.networkInterface + servicesPath
	if pb.service != "" {
		return path + "/" + pb.service
	}
	return path
}

// ServicesService is the contract for the services provided on the Skytap Services resource
type ServicesService interface {
	List(ctx context.Context, environmentID string, vmID string, nicID string) (*ServiceListResult, error)
	Get(ctx context.Context, environmentID string, vmID string, nicID string, id string) (*Service, error)
	Create(ctx context.Context, environmentID string, vmID string, nicID string, internalPort *CreateServiceRequest) (*Service, error)
	Update(ctx context.Context, environmentID string, vmID string, nicID string, id string, internalPort *UpdateServiceRequest) (*Service, error)
	Delete(ctx context.Context, environmentID string, vmID string, nicID string, id string) error
}

// ServicesServiceClient is the ServicesService implementation
type ServicesServiceClient struct {
	client *Client
}

// Service describes a service provided on the connected network
type Service struct {
	ID           *string `json:"id"`
	InternalPort *int    `json:"internal_port"`
	ExternalIP   *string `json:"external_ip"`
	ExternalPort *int    `json:"external_port"`
}

// CreateServiceRequest describes the create the service data
type CreateServiceRequest struct {
	InternalPort *int `json:"internal_port"`
}

// UpdateServiceRequest describes the update the service data
type UpdateServiceRequest struct {
	CreateServiceRequest
}

// ServiceListResult is the listing request specific struct
type ServiceListResult struct {
	Value []Service
}

// List the services
func (s *ServicesServiceClient) List(ctx context.Context, environmentID string, vmID string, nicID string) (*ServiceListResult, error) {
	var builder servicePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(nicID).Build()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var serviceListResponse ServiceListResult
	_, err = s.client.do(ctx, req, &serviceListResponse.Value)
	if err != nil {
		return nil, err
	}

	return &serviceListResponse, nil
}

// Get a service
func (s *ServicesServiceClient) Get(ctx context.Context, environmentID string, vmID string, nicID string, id string) (*Service, error) {
	var builder servicePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(nicID).Service(id).Build()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var service Service
	_, err = s.client.do(ctx, req, &service)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

// Create a service
func (s *ServicesServiceClient) Create(ctx context.Context, environmentID string, vmID string, nicID string, internalPort *CreateServiceRequest) (*Service, error) {
	var builder servicePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(nicID).Build()

	req, err := s.client.newRequest(ctx, "POST", path, internalPort)
	if err != nil {
		return nil, err
	}

	var createdService Service
	_, err = s.client.do(ctx, req, &createdService)
	if err != nil {
		return nil, err
	}

	return &createdService, nil
}

// Update a service
func (s *ServicesServiceClient) Update(ctx context.Context, environmentID string, vmID string, nicID string, id string, internalPort *UpdateServiceRequest) (*Service, error) {
	err := s.Delete(ctx, environmentID, vmID, nicID, id)
	if err != nil {
		return nil, err
	}
	return s.Create(ctx, environmentID, vmID, nicID, &internalPort.CreateServiceRequest)
}

// Delete a service
func (s *ServicesServiceClient) Delete(ctx context.Context, environmentID string, vmID string, nicID string, id string) error {
	var builder servicePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Interface(nicID).Service(id).Build()

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
