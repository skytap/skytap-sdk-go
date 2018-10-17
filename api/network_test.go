package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAutomaticNetwork(t *testing.T) {
	netJson := readJson(t, "testdata/network-1.json")

	client := skytapClient(t)
	server := getMockServerForString(client, netJson)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations/1/networks.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"name":"Default Network","network_type":"automatic","subnet":"10.0.0.0/24","domain":"skytap.example"}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, netJson)
	})

	net, err := CreateAutomaticNetwork(client, "1", "Default Network", "10.0.0.0/24", "skytap.example")
	require.NoError(t, err, "Error creating automatic network")
	require.Equal(t, "Default Network", net.Name)
	require.Equal(t, "10.0.0.0/24", net.Subnet)
	require.Equal(t, "skytap.example", net.Domain)
}

func TestCreateManualNetwork(t *testing.T) {
	netJson := readJson(t, "testdata/network-2.json")

	client := skytapClient(t)
	server := getMockServerForString(client, netJson)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations/1/networks.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"name":"API Network","network_type":"manual","subnet":"10.0.1.0/24","gateway":"10.0.1.254"}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, netJson)
	})

	net, err := CreateManualNetwork(client, "1", "API Network", "10.0.1.0/24", "10.0.1.254")
	require.NoError(t, err, "Error creating automatic network")
	require.Equal(t, "API Network", net.Name)
	require.Equal(t, "10.0.1.0/24", net.Subnet)
	require.Equal(t, "10.0.1.254", net.Gateway)
}

func TestDeleteNetwork(t *testing.T) {

	client := skytapClient(t)
	server := getMockServerForString(client, "")
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "DELETE", r.Method)
		require.Equal(t, "/configurations/1/networks/99", r.URL.Path)
	})

	err := DeleteNetwork(client, "1", "99")
	require.NoError(t, err, "Error deleting network")

}

func TestAttachVpn(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")
	attachVpnJson := readJson(t, "testdata/attach-vpn-1.json")

	client := skytapClient(t)
	server := getMockServerForString(client, envJson)
	defer server.Close()

	env, err := GetEnvironment(client, "1")
	require.NoError(t, err, "Error getting environment")
	vm := env.Vms[0]
	network := env.Networks[0]

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations/1/networks/99/vpns.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"vpn_id":"vpn-1"}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, attachVpnJson)
	})

	result, err := network.AttachToVpn(client, env.Id, "vpn-1")
	require.NoError(t, err, "Error attaching VPN")
	require.Equal(t, false, result.Connected)

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PUT", r.Method)
		require.Equal(t, "/configurations/1/networks/99/vpns/vpn-1", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"connected":true}`, strings.TrimSpace(string(body)))

		tmpstr := strings.Replace(attachVpnJson, `"connected": false`, `"connected": true`, 1)
		fmt.Fprintln(w, tmpstr)
	})

	err = network.ConnectToVpn(client, env.Id, "vpn-1")
	require.NoError(t, err, "Error connecting VPN")

	// Verify parsing of NatAddresses
	require.Equal(t, "vpn-1", vm.Interfaces[0].NatAddresses.VpnNatAddresses[0].VpnId, "Should have correct VPN id")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PUT", r.Method)
		require.Equal(t, "/configurations/1/networks/99/vpns/vpn-1", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"connected":false}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, attachVpnJson)
	})

	err = network.DisconnectFromVpn(client, env.Id, "vpn-1")
	require.NoError(t, err, "Error disconnecting VPN")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "DELETE", r.Method)
		require.Equal(t, "/configurations/1/networks/99/vpns/vpn-1", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, "", strings.TrimSpace(string(body)))
		fmt.Fprintln(w, attachVpnJson)
	})

	err = network.DetachFromVpn(client, env.Id, "vpn-1")
	require.NoError(t, err, "Error detaching VPN")
}

func TestAddPublishedService(t *testing.T) {
	netJson := readJson(t, "testdata/network-1.json")

	client := skytapClient(t)
	server := getMockServerForString(client, netJson)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

}
