package skytap

import "context"

// Default URL paths
const (
	interfacesBasePath = "/v2/configurations/"
	interfacesVMPath   = "/vms/"
	interfacesPath     = "/interfaces"
)

type interfacePathBuilder interface {
	Environment(string) interfacePathBuilder
	VM(string) interfacePathBuilder
	Adapter(string) interfacePathBuilder
	Build() string
}

type interfacePathBuilderImpl struct {
	environment string
	vm          string
	adapter     string
}

func (pb *interfacePathBuilderImpl) Environment(environment string) interfacePathBuilder {
	pb.environment = environment
	return pb
}

func (pb *interfacePathBuilderImpl) VM(vm string) interfacePathBuilder {
	pb.vm = vm
	return pb
}

func (pb *interfacePathBuilderImpl) Adapter(adapter string) interfacePathBuilder {
	pb.adapter = adapter
	return pb
}

func (pb *interfacePathBuilderImpl) Build() string {
	path := interfacesBasePath + pb.environment + interfacesVMPath + pb.vm + interfacesPath
	if pb.adapter != "" {
		return path + "/" + pb.adapter
	}
	return path
}

// AdaptersService is the contract for the services provided on the Skytap Interface resource
type AdaptersService interface {
	List(ctx context.Context, environmentID string, vmID string) (*AdapterListResult, error)
	Get(ctx context.Context, environmentID string, vmID string, id string) (*Interface, error)
	Create(ctx context.Context, environmentID string, vmID string, opts *CreateAdapterRequest) (*Interface, error)
	Update(ctx context.Context, environmentID string, vmID string, id string, adapter *UpdateAdapterRequest) (*Interface, error)
	Delete(ctx context.Context, environmentID string, vmID string, id string) error
}

// AdaptersServiceClient is the AdaptersService implementation
type AdaptersServiceClient struct {
	client *Client
}

// Interface describes the VM's virtual network interface configuration
type Interface struct {
	ID                  *string              `json:"id"`
	IP                  *string              `json:"ip"`
	Hostname            *string              `json:"hostname"`
	MAC                 *string              `json:"mac"`
	ServicesCount       *int                 `json:"services_count"`
	Services            []Service            `json:"services"`
	PublicIPsCount      *int                 `json:"public_ips_count"`
	PublicIPs           []map[string]string  `json:"public_ips"`
	VMID                *string              `json:"vm_id"`
	VMName              *string              `json:"vm_name"`
	Status              *string              `json:"status"`
	NetworkID           *string              `json:"network_id"`
	NetworkName         *string              `json:"network_name"`
	NetworkURL          *string              `json:"network_url"`
	NetworkType         *string              `json:"network_type"`
	NetworkSubnet       *string              `json:"network_subnet"`
	NICType             *NICType             `json:"nic_type"`
	SecondaryIPs        []SecondaryIP        `json:"secondary_ips"`
	PublicIPAttachments []PublicIPAttachment `json:"public_ip_attachments"`
}

// Service describes a service provided on the connected network
type Service struct {
	ID           *string `json:"id"`
	InternalPort *int    `json:"internal_port"`
	ExternalIP   *string `json:"external_ip"`
	ExternalPort *int    `json:"external_port"`
}

// SecondaryIP holds secondary IP address data
type SecondaryIP struct {
	ID      *string `json:"id"`
	Address *string `json:"address"`
}

// PublicIPAttachment describes the public IP address data
type PublicIPAttachment struct {
	ID                    *int    `json:"id"`
	PublicIPAttachmentKey *int    `json:"public_ip_attachment_key"`
	Address               *string `json:"address"`
	ConnectType           *int    `json:"connect_type"`
	Hostname              *string `json:"hostname"`
	DNSName               *string `json:"dns_name"`
	PublicIPKey           *string `json:"public_ip_key"`
}

// NICType describes the different Network Interface Card types
type NICType string

// A list of the different NIC types
const (
	NICTypeDefault NICType = "default"
	NICTypePCNet32 NICType = "pcnet32"
	NICTypeE1000   NICType = "e1000"
	NICTypeE1000E  NICType = "e1000e"
	NICTypeVMXNet  NICType = "vmxnet"
	NICTypeVMXNet3 NICType = "vmxnet3"
)

// CreateAdapterRequest describes the create the adapter data
type CreateAdapterRequest struct {
	NICType   *NICType `json:"nic_type"`
	NetworkID *string  `json:"network_id"`
	IP        *string  `json:"ip,omitempty"`
	Hostname  *string  `json:"hostname,omitempty"`
}

// UpdateAdapterRequest describes the update the adapter data
type UpdateAdapterRequest struct {
	IP       *string `json:"ip,omitempty"`
	Hostname *string `json:"hostname,omitempty"`
}

// AdapterListResult is the listing request specific struct
type AdapterListResult struct {
	Value []Interface
}

// List the adapters
func (s *AdaptersServiceClient) List(ctx context.Context, environmentID string, vmID string) (*AdapterListResult, error) {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Build()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	err = s.client.setRequestListParameters(req, nil)
	if err != nil {
		return nil, err
	}

	var adapterListResponse AdapterListResult
	_, err = s.client.do(ctx, req, &adapterListResponse.Value)
	if err != nil {
		return nil, err
	}

	return &adapterListResponse, nil
}

// Get a adapter
func (s *AdaptersServiceClient) Get(ctx context.Context, environmentID string, vmID string, id string) (*Interface, error) {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Adapter(id).Build()

	req, err := s.client.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var adapter Interface
	_, err = s.client.do(ctx, req, &adapter)
	if err != nil {
		return nil, err
	}

	return &adapter, nil
}

// Create a adapter
func (s *AdaptersServiceClient) Create(ctx context.Context, environmentID string, vmID string, opts *CreateAdapterRequest) (*Interface, error) {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Build()

	req, err := s.client.newRequest(ctx, "POST", path, opts)
	if err != nil {
		return nil, err
	}

	var createdAdapter Interface
	_, err = s.client.do(ctx, req, &createdAdapter)
	if err != nil {
		return nil, err
	}

	return &createdAdapter, nil
}

// Update a adapter
func (s *AdaptersServiceClient) Update(ctx context.Context, environmentID string, vmID string, id string, adapter *UpdateAdapterRequest) (*Interface, error) {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Adapter(id).Build()

	req, err := s.client.newRequest(ctx, "PUT", path, adapter)
	if err != nil {
		return nil, err
	}

	var updatedAdapter Interface
	_, err = s.client.do(ctx, req, &updatedAdapter)
	if err != nil {
		return nil, err
	}

	return &updatedAdapter, nil
}

// Delete a adapter
func (s *AdaptersServiceClient) Delete(ctx context.Context, environmentID string, vmID string, id string) error {
	var builder interfacePathBuilderImpl
	path := builder.Environment(environmentID).VM(vmID).Adapter(id).Build()

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
