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
	"net/http"
	"net/http/httptest"
	"testing"

	"strings"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"io/ioutil"
)

type testConfig struct {
	Username   string `json:"username"`
	ApiKey     string `json:"apiKey"`
	TemplateId string `json:"templateId"`
	VmId       string `json:"vmId"`
	VpnId      string `json:"vpnId"`
}

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func skytapClient(t *testing.T) SkytapClient {
	c := getTestConfig(t)
	fmt.Printf("c: %s, user: %s", c, c.Username)
	client := &http.Client{}
	return SkytapClient{
		HttpClient:  client,
		Credentials: SkytapCredentials{Username: c.Username, ApiKey: c.ApiKey},
	}
}

func getTestConfig(t *testing.T) *testConfig {
	configFile, err := os.Open("testdata/config.json")
	require.NoError(t, err, "Error reading config.json")

	jsonParser := json.NewDecoder(configFile)
	c := &testConfig{}
	err = jsonParser.Decode(c)
	require.NoError(t, err, "Error parsing config.json")
	return c
}

func getMockServer(client SkytapClient) *httptest.Server {
	return getMockServerForString(client, "")
}

func getMockServerForString(client SkytapClient, content string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		fmt.Fprintln(w, string(content))
	}))
	// Divert all API requests to this server
	baseUrlOveride = server.URL
	client.HttpClient = server.Client()
	return server
}

func readJson(t *testing.T, filename string) string {
	str, err := ioutil.ReadFile(filename)
	require.NoError(t, err, "Error reading " + filename)
	return string(str)
}

func TestCreateNewEnvironment(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")

	client := skytapClient(t)
	server := getMockServer(client)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"template_id":"2"}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, envJson)
	})

	env, err := CreateNewEnvironment(client, "2")
	require.NoError(t, err, "Error creating environment")
	require.Equal(t, "Environment 1", env.Name)
}

func TestCreateNewEnvironmentWithVms(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")

	client := skytapClient(t)
	server := getMockServer(client)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"template_id":"2","vm_ids":["1002"]}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, envJson)
	})

	env, err := CreateNewEnvironmentWithVms(client, "2", []string{"1002"})
	require.NoError(t, err, "Error creating environment")
	require.Equal(t, "Environment 1", env.Name)
}

func TestCopyEnvironmentWithVms(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")

	client := skytapClient(t)
	server := getMockServer(client)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"configuration_id":"1","vm_ids":["1001"]}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, envJson)
	})

	env, err := CopyEnvironmentWithVms(client, "1", []string{"1001"})
	require.NoError(t, err, "Error creating environment")
	require.Equal(t, "Environment 1", env.Name)
}

func TestAddVirtualMachineFromTemplate(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")
	templJson := readJson(t, "testdata/template-2.json")
	templateVmJson := readJson(t, "testdata/vm-1002.json")

	client := skytapClient(t)
	server := getMockServerForString(client, envJson)
	defer server.Close()

	env, err := GetEnvironment(client, "1")
	require.NoError(t, err, "Error getting environment")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpstr := ""
		if r.Method == "GET" {
			// handle requests for vm and its template
			if strings.Contains(r.URL.Path, "/vms") {
				// replace any hostnames in the vm json
				tmpstr = strings.Replace(templateVmJson, "https://cloud.skytap.com", server.URL, -1)
			} else {
				tmpstr = templJson
			}
			fmt.Fprintln(w, tmpstr)
		} else if r.Method == "PUT" {
			require.Equal(t, "/configurations/1.json", r.URL.Path)
			body, _ := ioutil.ReadAll(r.Body)
			require.Equal(t, `{"template_id":"2","vm_ids":["1002"]}`, strings.TrimSpace(string(body)))
			fmt.Fprintln(w, envJson)
		}
	})

	env, err = env.AddVirtualMachine(client, "1002")
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, "Environment 1", env.Name)
}

func TestAddVirtualMachineFromEnvironment(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, envJson)
	defer server.Close()

	env, err := GetEnvironment(client, "1")
	require.NoError(t, err, "Error getting environment")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpstr := ""
		if r.Method == "GET" {
			// handle requests for vm and its environment
			if strings.Contains(r.URL.Path, "/vms") {
				// replace any hostnames in the vm json
				tmpstr = strings.Replace(vmJson, "https://cloud.skytap.com", server.URL, -1)
			} else {
				tmpstr = envJson
			}
			fmt.Fprintln(w, tmpstr)
		} else if r.Method == "PUT" {
			require.Equal(t, "/configurations/1.json", r.URL.Path)
			body, _ := ioutil.ReadAll(r.Body)
			require.Equal(t, `{"merge_configuration":"1","vm_ids":["1001"]}`, strings.TrimSpace(string(body)))
			fmt.Fprintln(w, envJson)
		}
	})

	env, err = env.AddVirtualMachine(client, "1001")
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, "Environment 1", env.Name)
}

func TestDeleteVirtualMachine(t *testing.T) {
	client := skytapClient(t)
	server := getMockServer(client)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "DELETE", r.Method)
		require.Equal(t, "/vms/1001", r.URL.Path)
		fmt.Fprintln(w, "{}")
	})

	err := DeleteVirtualMachine(client, "1001")
	require.NoError(t, err, "Error deleting vm")
}

func TestVmCredentials(t *testing.T) {
	vmJson := readJson(t, "testdata/vm-1001.json")
	credJson := readJson(t, "testdata/credentials.json")

	client := skytapClient(t)
	server := getMockServerForString(client, vmJson)
	defer server.Close()

	vm, err := GetVirtualMachine(client, "1001")
	require.NoError(t, err, "Error creating vm")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, credJson)
	})

	creds, err := vm.GetCredentials(client)
	require.NoError(t, err, "Error getting VM credentials")

	user, err := creds[0].Username()
	require.NoError(t, err, "Error username")
	require.Equal(t, "root", user)

	pass, err := creds[0].Password()
	require.NoError(t, err, "Error getting password")
	require.Equal(t, "ChangeMe!", pass)
}

func TestVmWaitUntilReady(t *testing.T) {
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, strings.Replace(vmJson, "stopped", "busy", 1))
	defer server.Close()

	vm, err := GetVirtualMachine(client, "1001")
	require.NoError(t, err, "Error creating vm")
	require.Equal(t, RunStateBusy, vm.Runstate, "Should be busy")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/vms/1001", r.URL.Path)
		fmt.Fprintln(w, vmJson)
	})

	vm, err = vm.WaitUntilReady(client)
	require.NoError(t, err, "Error waiting for VM")
	require.Equal(t, RunStateStop, vm.Runstate, "Should be stopped")
}

func TestVmStart(t *testing.T) {
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, vmJson)
	defer server.Close()

	vm, err := GetVirtualMachine(client, "1001")
	require.NoError(t, err, "Error creating vm")
	require.Equal(t, RunStateStop, vm.Runstate, "Should be stopped")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/vms/1001", r.URL.Path)
		if r.Method == "PUT" {
			body, _ := ioutil.ReadAll(r.Body)
			require.Equal(t, `{"runstate":"running"}`, strings.TrimSpace(string(body)))
		}
		tmpstr := strings.Replace(vmJson, "stopped", "running", 1)
		fmt.Fprintln(w, tmpstr)
	})

	started, err := vm.Start(client)
	require.NoError(t, err, "Error starting VM")
	require.Equal(t, RunStateStart, started.Runstate, "Should be started")
}

func TestVmSuspend(t *testing.T) {
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, strings.Replace(vmJson, "stopped", "running", 1))
	defer server.Close()

	vm, err := GetVirtualMachine(client, "1001")
	require.NoError(t, err, "Error creating vm")
	require.Equal(t, RunStateStart, vm.Runstate, "Should be started")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/vms/1001", r.URL.Path)
		if r.Method == "PUT" {
			body, _ := ioutil.ReadAll(r.Body)
			require.Equal(t, `{"runstate":"suspended"}`, strings.TrimSpace(string(body)))
		}
		tmpstr := strings.Replace(vmJson, "stopped", "suspended", 1)
		fmt.Fprintln(w, tmpstr)
	})

	suspended, err := vm.Suspend(client)
	require.NoError(t, err, "Error suspending VM")
	require.Equal(t, RunStatePause, suspended.Runstate, "Should be suspended")
}

func TestVmKill(t *testing.T) {
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, strings.Replace(vmJson, "stopped", "running", 1))
	defer server.Close()

	vm, err := GetVirtualMachine(client, "1001")
	require.NoError(t, err, "Error creating vm")
	require.Equal(t, RunStateStart, vm.Runstate, "Should be started")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/vms/1001", r.URL.Path)
		if r.Method == "PUT" {
			body, _ := ioutil.ReadAll(r.Body)
			require.Equal(t, `{"runstate":"halted"}`, strings.TrimSpace(string(body)))
		}
		fmt.Fprintln(w, vmJson)
	})

	killed, err := vm.Kill(client)
	require.NoError(t, err, "Error stopping VM")
	require.Equal(t, RunStateStop, killed.Runstate, "Should be stopped/killed")
}

func TestChangeNetworkHostname(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, envJson)
	defer server.Close()

	env, err := GetEnvironment(client, "1")
	require.NoError(t, err, "Error getting environment")
	vm := env.Vms[0]

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PUT", r.Method)
		require.Equal(t, "/configurations/1/vms/1001/interfaces/nic-5971736-13548234-0.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"hostname":"newname1234"}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, vmJson)
	})

	_, err = vm.RenameNetworkInterface(client, env.Id, vm.Interfaces[0].Id, "newname1234")
	require.NoError(t, err, "Error renaming interface")
}

func TestUpdateHardware(t *testing.T) {
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, vmJson)
	defer server.Close()

	vm, err := GetVirtualMachine(client, "1001")
	require.NoError(t, err, "Error creating vm")

	cpus := 4
	persock := 2
	hardware := Hardware{&cpus, &persock, nil}

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/vms/1001.json", r.URL.Path)
		require.Equal(t, "PUT", r.Method)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"hardware":{"cpus":4,"cpus_per_socket":2}}`, strings.TrimSpace(string(body)))

		tmpstr := strings.Replace(vmJson, `"cpus": 1`, `"cpus": 4`, 1)
		tmpstr = strings.Replace(tmpstr, `"cpus_per_socket": 1`, `"cpus_per_socket": 2`, 1)
		fmt.Fprintln(w, tmpstr)
	})

	updated, err := vm.UpdateHardware(client, hardware, false)
	require.NoError(t, err, "Error updating hardware")

	require.Equal(t, hardware.Cpus, updated.Hardware.Cpus)
	require.Equal(t, hardware.CpusPerSocket, updated.Hardware.CpusPerSocket)
	require.Equal(t, vm.Hardware.Ram, updated.Hardware.Ram)

	ram := 4096
	updateRam := Hardware{Ram: &ram}

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/vms/1001.json", r.URL.Path)
		require.Equal(t, "PUT", r.Method)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"hardware":{"ram":4096}}`, strings.TrimSpace(string(body)))

		tmpstr := strings.Replace(vmJson, `"cpus": 1`, `"cpus": 4`, 1)
		tmpstr = strings.Replace(tmpstr, `"cpus_per_socket": 1`, `"cpus_per_socket": 2`, 1)
		tmpstr = strings.Replace(tmpstr, `"ram": 1024`, `"ram": 4096`, 1)
		fmt.Fprintln(w, tmpstr)
	})

	updated, err = vm.UpdateHardware(client, updateRam, false)
	require.NoError(t, err, "Error updating ram")
	require.Equal(t, hardware.Cpus, updated.Hardware.Cpus)
	require.Equal(t, hardware.CpusPerSocket, updated.Hardware.CpusPerSocket)
	require.Equal(t, updateRam.Ram, updated.Hardware.Ram)
}

func TestChangeName(t *testing.T) {
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, vmJson)
	defer server.Close()

	vm, err := GetVirtualMachine(client, "1001")
	require.NoError(t, err, "Error creating vm")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PUT", r.Method)
		require.Equal(t, "/vms/1001.json", r.URL.Path)
		require.Equal(t, "name=foo", r.URL.RawQuery)
		fmt.Fprintln(w, vmJson)
	})

	_, err = vm.SetName(client, "foo")
	require.NoError(t, err, "Error updating name")
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
