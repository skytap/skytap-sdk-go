// Copyright 2016 Skytap Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"

	"github.com/dghubble/sling"

	log "github.com/Sirupsen/logrus"
)

const (
	NetworkPath          = "networks"
	InterfacePath        = "interfaces"
	VpnPath              = "vpns"
	PublishedServicePath = "services"
)

/*
 Network resource.
*/
type Network struct {
	Id                  string          `json:"id"`
	Url                 string          `json:"url"`
	Name                string          `json:"name"`
	Subnet              string          `json:"subnet"`
	Domain              string          `json:"domain"`
	Gateway             string          `json:"gateway"`
	NetworkType         string          `json:"network_type"`
	Tunnelable          bool            `json:"tunnelable"`
	Tunnels             interface{}     `json:"tunnels"`
	PrimaryNameserver   string          `json:"primary_nameserver"`
	SecondaryNameserver string          `json:"secondary_nameserver"`
	Region              string          `json:"region"`
	NatSubnet           string          `json:"nat_subnet"`
	NatPoolSize         int             `json:"nat_pool_size"`
	NatPoolRemaining    int             `json:"nat_pool_remaining"`
	VpnAttachments      []VpnAttachment `json:"vpn_attachments"`
}

/*
 Network interface inside a VM.
*/
type NetworkInterface struct {
	Id                string             `json:"id,omitempty"`
	Ip                string             `json:"ip,omitempty"`
	PublicIpsCount    int                `json:"public_ips_count,omitempty"`
	Hostname          string             `json:"hostname,omitempty"`
	PublicIps         []PublicIp         `json:"public_ips,omitempty"`
	NatAddresses      *NatAddresses      `json:"nat_addresses,omitempty"`
	Status            string             `json:"status,omitempty"`
	ExternalAddress   string             `json:"external_address,omitempty"`
	NicType           string             `json:"nic_type,omitempty"`
	NetworkId         string             `json:"network_id,omitempty"`
	PublishedServices []PublishedService `json:"services,omitempty"`
}

/*
 Nat addresses stored inside network interface.
*/
type NatAddresses struct {
	VpnNatAddresses     []VpnNatAddress     `json:"vpn_nat_addresses,omitempty"`
	NetworkNatAddresses []NetworkNatAddress `json:"network_nat_addresses,omitempty	"`
}

/*
 VPN based NAT address.
*/
type VpnNatAddress struct {
	IpAddress string `json:"ip_address"`
	VpnId     string `json:"vpn_id"`
	VpnName   string `json:"vpn_name"`
	VpnUrl    string `json:"vpn_url"`
}

/*
 Network based NAT address.
*/
type NetworkNatAddress struct {
	IpAddress        string `json:"ip_address"`
	NetworkId        string `json:"network_id"`
	NetworkName      string `json:"network_name"`
	NetworkUrl       string `json:"network_url"`
	ConfigurationId  string `json:"configuration_id"`
	ConfigurationUrl string `json:"configuration_url"`
}

/*
 IP type.
*/
type PublicIp struct {
	Id      string      `json:"id"`
	Address string      `json:"address"`
	Region  string      `json:"region"`
	Nics    interface{} `json:"nics"`
	VpnId   string      `json:"vpn_id"`
}

/*
 VPN attachments to network.
*/
type VpnAttachment struct {
	Id        string `json:"id"`
	Connected bool   `json:"connected"`
	Vpn       Vpn    `json:"vpn"`
}

type Vpn struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Enabled       bool   `json:"enabled"`
	NatEnabled    bool   `json:"nat_enabled"`
	RemoteSubnets string `json:"remote_subnets"`
	RemotePeerIp  string `json:"remote_subnets"`
	CanReconnect  bool   `json:"can_reconnect"`
}

/*
 Request body for VPN attach commands.
*/
type AttachVpnBody struct {
	VpnId string `json:"vpn_id"`
}

/*
 Response body for VPN attach commands.
*/
type AttachVpnResult struct {
	Id        string           `json:"id"`
	Connected bool             `json:"connected"`
	Network   NetworkInterface `json:"network"`
	Vpn       interface{}      `json:"vpn"`
}

/*
 Request body for VPN connect commands.
*/
type ConnectVpnBody struct {
	Connected bool `json:"connected"`
}

type PublishedService struct {
	Id           string `json:"id,omitempty"`
	InternalPort int    `json:"internal_port,omitempty"`
	ExternalIp   string `json:"external_ip,omitempty"`
	ExternalPort int    `json:"external_port,omitempty"`
}

// CreateAutomaticNetwork - create a new network in an Environment
func CreateAutomaticNetwork(
	client SkytapClient,
	envId string,
	name string,
	subnet string,
	domain string) (*Network, error) {

	log.WithFields(log.Fields{"envId": envId, "network_name": name}).Info("Adding network to environment")

	createAutoNetwork := func(s *sling.Sling) *sling.Sling {
		network := struct {
			Name        string `json:"name"`
			NetworkType string `json:"network_type"`
			Subnet      string `json:"subnet"`
			Domain      string `json:"domain"`
		}{
			Name:        name,
			NetworkType: "automatic",
			Subnet:      subnet,
			Domain:      domain,
		}

		return s.Post(EnvironmentPath + "/" + envId + "/" + NetworkPath + ".json").BodyJSON(network)
	}

	network := new(Network)
	_, err := RunSkytapRequest(client, false, network, createAutoNetwork)

	return network, err
}

func CreateManualNetwork(
	client SkytapClient,
	envId string,
	name string,
	subnet string,
	gateway string) (*Network, error) {
	log.WithFields(log.Fields{"envId": envId, "network_name": name}).Info("Adding network to environment")

	createAutoNetwork := func(s *sling.Sling) *sling.Sling {
		network := struct {
			Name        string `json:"name"`
			NetworkType string `json:"network_type"`
			Subnet      string `json:"subnet"`
			Gateway     string `json:"gateway"`
		}{
			Name:        name,
			NetworkType: "manual",
			Subnet:      subnet,
			Gateway:     gateway,
		}
		return s.Post(EnvironmentPath + "/" + envId + "/" + NetworkPath + ".json").BodyJSON(network)

	}
	network := new(Network)
	_, err := RunSkytapRequest(client, false, network, createAutoNetwork)
	return network, err

}

// DeleteNetwork - delete a network from an environment
func DeleteNetwork(client SkytapClient, envId string, netId string) error {
	log.WithFields(log.Fields{"envId": envId, "netId": netId}).Info("Deleting network in environment")

	deleteNet := func(s *sling.Sling) *sling.Sling {
		return s.Delete(fmt.Sprintf("%s/%s/%s/%s", EnvironmentPath, envId, NetworkPath, netId))
	}

	_, err := RunSkytapRequest(client, false, nil, deleteNet)
	return err
}

/*
 Path for all VPNs in a network and environment.
*/
func vpnsForNetworkInEnvironmentPath(netId string, envId string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s.json", EnvironmentPath, envId, NetworkPath, netId, VpnPath)
}

/*
 Path for a single VPN in a network and environment.
*/
func vpnForNetworkInEnvironmentPath(netId string, envId string, vpnId string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s/%s", EnvironmentPath, envId, NetworkPath, netId, VpnPath, vpnId)
}

/*
 Attach a network to a VPN, in the context of the given environment.
*/
func (n *Network) AttachToVpn(client SkytapClient, envId string, vpnId string) (*AttachVpnResult, error) {
	log.WithFields(log.Fields{"netId": n.Id, "vpnId": vpnId, "envId": envId}).Info("Attach network to VPN")

	attachBody := &AttachVpnBody{vpnId}
	attach := func(s *sling.Sling) *sling.Sling {
		return s.Post(vpnsForNetworkInEnvironmentPath(n.Id, envId)).BodyJSON(attachBody)
	}

	result := &AttachVpnResult{}
	_, err := RunSkytapRequest(client, false, result, attach)
	if err != nil {
		log.WithFields(log.Fields{"envId": envId, "vpnId": vpnId, "networkId": n.Id, "requestBody": attachBody, "error": err}).Errorf("Unable to attach VPN to environment.")
		return result, err
	}
	return result, nil
}

/*
 Connect to a given VPN in the context of a given environment.
*/
func (n *Network) ConnectToVpn(client SkytapClient, envId string, vpnId string) error {
	return n.ChangeConnectionToVpn(client, envId, vpnId, true)
}

/*
 Disconnect an environment's network from a VPN.
*/
func (n *Network) DisconnectFromVpn(client SkytapClient, envId string, vpnId string) error {
	return n.ChangeConnectionToVpn(client, envId, vpnId, false)
}

/*
 General method for manipulating VPN connection state.
*/
func (n *Network) ChangeConnectionToVpn(client SkytapClient, envId string, vpnId string, connected bool) error {
	log.WithFields(log.Fields{"netId": n.Id, "vpnId": vpnId, "envId": envId, "connected": connected}).Info("Change network VPN connection")

	connectBody := &ConnectVpnBody{connected}

	connect := func(s *sling.Sling) *sling.Sling {
		return s.Put(vpnForNetworkInEnvironmentPath(n.Id, envId, vpnId)).BodyJSON(connectBody)
	}

	_, err := RunSkytapRequest(client, false, nil, connect)
	if err != nil {
		log.WithFields(log.Fields{"envId": envId, "vpnId": vpnId, "networkId": n.Id, "requestBody": connectBody, "error": err}).Errorf("Unable to attach VPN to environment.")
	}
	return err
}

/*
 Detach a network from a VPN in the context of the given environment.
*/
func (n *Network) DetachFromVpn(client SkytapClient, envId string, vpnId string) error {
	log.WithFields(log.Fields{"netId": n.Id, "vpnId": vpnId, "envId": envId}).Info("Detach network from VPN")

	detach := func(s *sling.Sling) *sling.Sling {
		return s.Delete(vpnForNetworkInEnvironmentPath(n.Id, envId, vpnId))
	}

	_, err := RunSkytapRequest(client, false, nil, detach)
	if err != nil {
		log.WithFields(log.Fields{"envId": envId, "vpnId": vpnId, "networkId": n.Id, "error": err}).Errorf("Unable to detach VPN from environment.")
	}
	return err
}

func vpnIdPath(vpnId string) string { return VpnPath + "/" + vpnId + ".json" }

/*
 Return an existing VPN by id.
*/
func GetVpn(client SkytapClient, vpnId string) (*Vpn, error) {
	vpn := &Vpn{}

	getVpn := func(s *sling.Sling) *sling.Sling {
		return s.Get(vpnIdPath(vpnId))
	}

	_, err := RunSkytapRequest(client, true, vpn, getVpn)
	return vpn, err
}

func (nic *NetworkInterface) AddPublishedService(client SkytapClient, port int, envId, vmId string) (*NetworkInterface, error) {

	log.WithFields(log.Fields{"envId": envId, "vmId": vmId, "interfaceId": nic.Id}).Infof("Adding service")

	service := PublishedService{InternalPort: port}

	addReq := func(s *sling.Sling) *sling.Sling {
		path := fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s.json", EnvironmentPath, envId, VmPath, vmId, InterfacePath, nic.Id, PublishedServicePath)
		return s.Post(path).BodyJSON(service)
	}

	_, err := RunSkytapRequest(client, true, service, addReq)
	if err != nil {
		return nic, err
	}

	nic.PublishedServices = append(nic.PublishedServices, service)

	log.WithField("publishedService", service).Info("Service Added")

	return nic, err
}
