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
	"testing"

	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"time"
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

func getEnvironment(t *testing.T, client SkytapClient, envId string) error {

	env, err := GetEnvironment(client, envId)

	if err != nil {
		t.Error(err)
	}
	if env.Id != envId {
		t.Error("Id didn't match")
	}

	return err
}

func TestCreateEnvironment(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, env.Id)
	require.NoError(t, err, "Error creating environment")

	err = getEnvironment(t, client, env.Id)
	require.NoError(t, err, "Error retrieving environment")

	env2, err := CreateNewEnvironmentWithVms(client, c.TemplateId, []string{c.VmId})
	env2, err = env2.WaitUntilReady(client)
	require.NoError(t, err, "Error creating environment with specific VMs")
	require.Equal(t, 1, len(env2.Vms), "Should only have 1 VM")
	require.Equal(t, env.Vms[0].Name, env2.Vms[0].Name, "Should match the requested VM name")
	err = DeleteEnvironment(client, env2.Id)
	require.NoError(t, err, "Error deleting environment")

	env3, err := CopyEnvironmentWithVms(client, env.Id, []string{env.Vms[0].Id})
	defer DeleteEnvironment(client, env3.Id)
	env3, err = env3.WaitUntilReady(client)
	require.NoError(t, err, "Error creating environment with specific VMs")
	require.Equal(t, 1, len(env3.Vms), "Should only have 1 VM")
	require.Equal(t, env.Vms[0].Name, env3.Vms[0].Name, "Should match the requested VM name")
}

func TestAddDeleteVirtualMachine(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	require.NoError(t, err, "Error creating environment")

	defer DeleteEnvironment(client, env.Id)

	// Add from template
	env2, err := env.AddVirtualMachine(client, c.VmId)
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, len(env.Vms)+1, len(env2.Vms))

	// Add from environment
	sourceEnv, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, sourceEnv.Id)
	require.NoError(t, err, "Error creating second environment")
	sourceEnv, err = sourceEnv.WaitUntilReady(client)
	require.NoError(t, err, "Error creating second environment")
	_, err = sourceEnv.Vms[0].WaitUntilReady(client)
	require.NoError(t, err, "Error creating second environment")

	env3, err := env2.AddVirtualMachine(client, sourceEnv.Vms[0].Id)
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, len(env2.Vms)+1, len(env3.Vms))

	// Delete last one
	DeleteVirtualMachine(client, env3.Vms[len(env3.Vms)-1].Id)
	env4, err := GetEnvironment(client, env.Id)
	require.NoError(t, err, "Error getting environment")
	require.Equal(t, len(env2.Vms), len(env4.Vms))

	// Add from same environment
	env5, err := env.AddVirtualMachine(client, env.Vms[0].Id)
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, len(env2.Vms)+1, len(env5.Vms))

	for _, value := range env5.Vms {
		require.Equal(t, "Ubuntu Server 14.04 - 64-bit", value.Name)
	}
}

func TestVmCredentials(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	require.NoError(t, err, "Error creating environment")

	defer DeleteEnvironment(client, env.Id)

	creds, err := env.Vms[0].GetCredentials(client)
	require.NoError(t, err, "Error getting VM credentials")

	user, err := creds[0].Username()
	require.NoError(t, err, "Error username")
	require.Equal(t, "root", user)

	pass, err := creds[0].Password()
	require.NoError(t, err, "Error getting password")
	require.Equal(t, "ChangeMe!", pass)

}

func TestManipulateVmRunstate(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, env.Id)
	require.NoError(t, err, "Error creating environment")

	vm, err := GetVirtualMachine(client, env.Vms[0].Id)
	require.NoError(t, err, "Error creating vm")

	stopped, err := vm.WaitUntilReady(client)
	require.NoError(t, err, "Error waiting on vm")
	require.Equal(t, RunStateStop, stopped.Runstate, "Should be stopped after waiting")

	started, err := stopped.Start(client)
	require.NoError(t, err, "Error starting VM")
	require.Equal(t, RunStateStart, started.Runstate, "Should be started")

	time.Sleep(10 * time.Second)

	// Can't get the VM to stop, waiting for a dialog
	stopped, err = started.Stop(client)
	require.NoError(t, err, "Error stopping VM")
	require.Equal(t, RunStateStop, stopped.Runstate, "Should be stopped")

	started, err = stopped.Start(client)
	require.NoError(t, err, "Error starting VM")
	require.Equal(t, RunStateStart, started.Runstate, "Should be started")

	killed, err := started.Kill(client)
	require.NoError(t, err, "Error stopping VM")
	require.Equal(t, RunStateStop, killed.Runstate, "Should be stopped/killed")

}

func TestChangeNetworkHostname(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, env.Id)
	require.NoError(t, err, "Error creating environment")

	vm, err := GetVirtualMachine(client, env.Vms[0].Id)
	require.NoError(t, err, "Error creating vm")

	testName := "newname1234"
	renamed, err := vm.RenameNetworkInterface(client, env.Id, vm.Interfaces[0].Id, testName)
	require.NoError(t, err, "Error renaming interface")
	require.Equal(t, testName, renamed.Hostname)

	vm, err = GetVirtualMachine(client, vm.Id)
	require.NoError(t, err, "Error refreshing VM")
	require.Equal(t, testName, vm.Interfaces[0].Hostname)

}

func TestUpdateHardware(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, env.Id)
	require.NoError(t, err, "Error creating environment")

	vm, err := GetVirtualMachine(client, env.Vms[0].Id)
	require.NoError(t, err, "Error creating vm")

	cpus := 4
	persock := 2
	hardware := Hardware{&cpus, &persock, nil}

	updated, err := vm.UpdateHardware(client, hardware, false)
	require.NoError(t, err, "Error updating hardware")

	require.Equal(t, hardware.Cpus, updated.Hardware.Cpus)
	require.Equal(t, hardware.CpusPerSocket, updated.Hardware.CpusPerSocket)
	require.Equal(t, vm.Hardware.Ram, updated.Hardware.Ram)

	ram := 4096
	updateRam := Hardware{Ram: &ram}

	updated, err = vm.WaitUntilReady(client)
	require.NoError(t, err, "Error waiting on vm")

	updated, err = vm.UpdateHardware(client, updateRam, false)
	require.NoError(t, err, "Error updating ram")
	require.Equal(t, hardware.Cpus, updated.Hardware.Cpus)
	require.Equal(t, hardware.CpusPerSocket, updated.Hardware.CpusPerSocket)
	require.Equal(t, updateRam.Ram, updated.Hardware.Ram)
}

func TestChangeName(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, env.Id)
	require.NoError(t, err, "Error creating environment")

	vm, err := GetVirtualMachine(client, env.Vms[0].Id)
	require.NoError(t, err, "Error creating vm")

	name := "foo"
	vm, err = vm.SetName(client, name)
	require.NoError(t, err, "Error setting name")
	require.Equal(t, name, vm.Name)

}

func TestAttachVpn(t *testing.T) {
	client := skytapClient(t)
	c := getTestConfig(t)

	env, err := CreateNewEnvironment(client, c.TemplateId)
	defer DeleteEnvironment(client, env.Id)
	require.NoError(t, err, "Error creating environment")

	result, err := env.Networks[0].AttachToVpn(client, env.Id, c.VpnId)
	require.NoError(t, err, "Error attaching VPN")
	log.Println(result)

	err = env.Networks[0].ConnectToVpn(client, env.Id, c.VpnId)
	require.NoError(t, err, "Error connecting VPN")

	// Now start VM and make sure it gets a good address from VPN NAT
	started, err := env.Vms[0].Start(client)
	require.NoError(t, err, "Error starting VM")
	require.Equal(t, c.VpnId, started.Interfaces[0].NatAddresses.VpnNatAddresses[0].VpnId, "Should have correct VPN id")

	err = env.Networks[0].DisconnectFromVpn(client, env.Id, c.VpnId)
	require.NoError(t, err, "Error disconnecting VPN")

	err = env.Networks[0].DetachFromVpn(client, env.Id, c.VpnId)
	require.NoError(t, err, "Error detaching VPN")

	env, err = env.WaitUntilReady(client)
	require.NoError(t, err, "Error waiting for environment")
	log.Println(env.Networks[0].VpnAttachments)
}

func TestMergeBody(t *testing.T) {
	vmId := "9285760"
	c := getTestConfig(t)

	b := &MergeTemplateBody{TemplateId: c.TemplateId, VmIds: []string{vmId}}
	j, _ := json.Marshal(b)
	println(string(j))

	j, _ = json.Marshal(&MergeEnvironmentBody{EnvironmentId: c.TemplateId, VmIds: []string{vmId}})
	println(string(j))

}
