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
	NetworkPath   = "networks"
	InterfacePath = "interfaces"
	VpnPath       = "vpns"
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
	Id              string       `json:"id"`
	Ip              string       `json:"ip"`
	PublicIpsCount  int          `json:"public_ips_count"`
	Hostname        string       `json:"hostname"`
	PublicIps       []PublicIp   `json:"public_ips"`
	NatAddresses    NatAddresses `json:"nat_addresses"`
	Status          string       `json:"status"`
	ExternalAddress string       `json:"external_address"`
}

/*
 Nat addresses stored inside network interface.
*/
type NatAddresses struct {
	VpnNatAddresses     []VpnNatAddress     `json:"vpn_nat_addresses"`
	NetworkNatAddresses []NetworkNatAddress `json:"network_nat_addresses"`
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
