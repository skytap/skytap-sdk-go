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
	log "github.com/Sirupsen/logrus"
	"github.com/dghubble/sling"
	"strings"
)

const (
	VmPath = "vms"
)

/*
 Skytap VM resource.
*/
type VirtualMachine struct {
	Id             string              `json:"id"`
	Name           string              `json:"name"`
	Runstate       string              `json:"runstate"`
	Error          interface{}         `json:"error"`
	TemplateUrl    string              `json:"template_url,omitempty"`
	EnvironmentUrl string              `json:"configuration_url,omitempty"`
	Interfaces     []*NetworkInterface `json:"interfaces`
	Hardware       Hardware            `json:"hardware"`
}

type VmCredential struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

type NameUpdate struct {
	Hostname string `json:"hostname"`
}

type Hardware struct {
	Cpus          *int `json:"cpus,omitempty"`
	CpusPerSocket *int `json:"cpus_per_socket,omitempty"`
	Ram           *int `json:"ram,omitempty"`
}

type HardwareUpdate struct {
	Hardware Hardware `json:"hardware"`
}

// Paths for VMs.
func vmIdInEnvironmentPath(envId string, vmId string) string {
	return fmt.Sprintf("%s/%s/%s/%s.json", EnvironmentPath, envId, VmPath, vmId)
}
func vmIdInTemplatePath(templateId string, vmId string) string {
	return fmt.Sprintf("%s/%s/%s/%s.json", TemplatePath, templateId, VmPath, vmId)
}
func vmIdPath(vmId string) string         { return fmt.Sprintf("%s/%s", VmPath, vmId) }
func vmUpdatePath(vmId string) string     { return fmt.Sprintf("%s/%s.json", VmPath, vmId) }
func vmCredentialPath(vmId string) string { return fmt.Sprintf("%s/%s/credentials.json", VmPath, vmId) }
func networkInterfacePath(envId string, vmId string, interfaceId string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s/%s.json", EnvironmentPath, envId, VmPath, vmId, InterfacePath, interfaceId)
}

/*
 If VM is in a template, returns the template, otherwise nil.
*/
func (vm *VirtualMachine) GetTemplate(client SkytapClient) (*Template, error) {
	if vm.TemplateUrl == "" {
		return nil, nil
	}
	template := &Template{}
	_, err := GetSkytapResource(client, vm.TemplateUrl, template)
	return template, err
}

/*
 If a VM is in an environment, returns the environment, otherwise nil.
*/
func (vm *VirtualMachine) GetEnvironment(client SkytapClient) (*Environment, error) {
	if vm.EnvironmentUrl == "" {
		return nil, nil
	}
	env := &Environment{}
	_, err := GetSkytapResource(client, vm.EnvironmentUrl, env)
	return env, err
}

/*
 Fetch fresh representation.
*/
func (vm *VirtualMachine) Refresh(client SkytapClient) (RunstateAwareResource, error) {
	return GetVirtualMachine(client, vm.Id)
}

func (vm *VirtualMachine) RunstateStr() string { return vm.Runstate }

/*
 Waits until VM is either stopped or started.
*/
func (vm *VirtualMachine) WaitUntilReady(client SkytapClient) (*VirtualMachine, error) {
	return vm.WaitUntilInState(client, []string{RunStateStop, RunStateStart, RunStatePause}, false)
}

/*
  Wait until the VM is in one of the desired states.
*/
func (vm *VirtualMachine) WaitUntilInState(client SkytapClient, desiredStates []string, requireStateChange bool) (*VirtualMachine, error) {
	r, err := WaitUntilInState(client, desiredStates, vm, requireStateChange)
	v := r.(*VirtualMachine)
	return v, err
}

/*
 Suspends a VM.
*/
func (vm *VirtualMachine) Suspend(client SkytapClient) (*VirtualMachine, error) {
	log.WithFields(log.Fields{"vmId": vm.Id}).Info("Starting VM")

	return vm.ChangeRunstate(client, RunStateStart, RunStateStart)
}

/*
 Starts a VM.
*/
func (vm *VirtualMachine) Start(client SkytapClient) (*VirtualMachine, error) {
	log.WithFields(log.Fields{"vmId": vm.Id}).Info("Starting VM")

	return vm.ChangeRunstate(client, RunStateStart, RunStateStart)
}

/*
 Stops a VM. Note that some VMs may require user input and cannot be stopped with the method.
*/
func (vm *VirtualMachine) Stop(client SkytapClient) (*VirtualMachine, error) {
	log.WithFields(log.Fields{"vmId": vm.Id}).Info("Stopping VM")

	newVm, err := vm.ChangeRunstate(client, RunStateStop, RunStateStop, RunStateStart)
	if err != nil {
		return newVm, err
	}
	switch newVm.Error {
	case false:
		return newVm, err
	case "Shutdown cannot proceed. Please check your VM for open dialog windows.":
		log.WithFields(log.Fields{"vmId": vm.Id, "error": newVm.Error, "state": newVm.RunstateStr()}).Warn("Unable to gracefully stop VM, will attempt to kill VM forcefully")
		return vm.Kill(client)
	default:
		return nil, fmt.Errorf("Unknown error stopping VM %s, error: %+v", vm.Id, newVm.Error)
	}
}

/*
 Kills a VM forcefully.
*/
func (vm *VirtualMachine) Kill(client SkytapClient) (*VirtualMachine, error) {
	log.WithFields(log.Fields{"vmId": vm.Id}).Info("Killing VM")

	return vm.ChangeRunstate(client, RunStateKill, RunStateStop)
}

/*
 Changes the runstate of the VM to the specified state and waits until the VM is in the desired state.
*/
func (vm *VirtualMachine) ChangeRunstate(client SkytapClient, runstate string, desiredRunstates ...string) (*VirtualMachine, error) {
	log.WithFields(log.Fields{"changeState": runstate, "targetState": desiredRunstates, "vmId": vm.Id}).Info("Changing VM runstate")

	ready, err := vm.WaitUntilReady(client)
	if err != nil {
		return ready, err
	}
	changeState := func(s *sling.Sling) *sling.Sling {
		return s.Put(vmIdPath(vm.Id)).BodyJSON(&RunstateBody{Runstate: runstate})
	}
	_, err = RunSkytapRequest(client, false, nil, changeState)

	if err != nil {
		return vm, err
	}
	return vm.WaitUntilInState(client, desiredRunstates, true)
}

func (vm *VirtualMachine) GetCredentials(client SkytapClient) ([]VmCredential, error) {
	credentialReq := func(s *sling.Sling) *sling.Sling {
		return s.Get(vmCredentialPath(vm.Id))
	}

	credentials := &[]VmCredential{}

	_, err := RunSkytapRequest(client, false, credentials, credentialReq)
	return *credentials, err
}

func (vm *VirtualMachine) RenameNetworkInterface(client SkytapClient, envId string, interfaceId string, name string) (*NetworkInterface, error) {
	nameReq := func(s *sling.Sling) *sling.Sling {
		return s.Put(networkInterfacePath(envId, vm.Id, interfaceId)).BodyJSON(&NameUpdate{Hostname: name})
	}

	interfaceResp := &NetworkInterface{}

	log.WithFields(log.Fields{"newName": name, "interfaceId": interfaceId, "envId": envId, "vmId": vm.Id}).Infof("Renaming interface")
	_, err := RunSkytapRequest(client, false, interfaceResp, nameReq)
	return interfaceResp, err
}

func (vm *VirtualMachine) UpdateHardware(client SkytapClient, hardware Hardware, restartVm bool) (*VirtualMachine, error) {
	if vm.Runstate != RunStateStop {
		vm, err := vm.Stop(client)
		if err != nil {
			return vm, err
		}
	}

	hardwareReq := func(s *sling.Sling) *sling.Sling {
		return s.Put(vmUpdatePath(vm.Id)).BodyJSON(&HardwareUpdate{Hardware: hardware})
	}

	newVm := &VirtualMachine{}

	log.WithFields(log.Fields{"vmId": vm.Id}).Infof("Updating VM hardware: %+v", hardware)
	_, err := RunSkytapRequest(client, false, newVm, hardwareReq)

	if err != nil {
		return newVm, err
	}
	if restartVm {
		newVm, err = newVm.Start(client)
	}

	return newVm, err
}

func (vm *VirtualMachine) ChangeAttribute(client SkytapClient, queryStruct interface{}) (*VirtualMachine, error) {
	changeReq := func(s *sling.Sling) *sling.Sling {
		return s.Put(vmUpdatePath(vm.Id)).QueryStruct(queryStruct)
	}

	newVm := &VirtualMachine{}

	log.WithFields(log.Fields{"vmId": vm.Id}).Infof("Updating VM attribute: %+v", queryStruct)
	_, err := RunSkytapRequest(client, false, newVm, changeReq)

	return newVm, err
}

type NameQuery struct {
	Name string `url:"name"`
}

func (vm *VirtualMachine) SetName(client SkytapClient, name string) (*VirtualMachine, error) {
	return vm.ChangeAttribute(client, &NameQuery{name})
}

func (c *VmCredential) Username() (string, error) {
	parts := strings.Split(c.Text, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("Incorrect parts in credential string '%s'", c.Text)
	}
	return strings.TrimSpace(parts[0]), nil
}

func (c *VmCredential) Password() (string, error) {
	parts := strings.Split(c.Text, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("Incorrect parts in credential string '%s'", c.Text)
	}
	return strings.TrimSpace(parts[1]), nil
}

/*
 Get a VM from an existing environment.
*/
func GetVirtualMachineInEnvironment(client SkytapClient, envId string, vmId string) (*VirtualMachine, error) {
	vm := &VirtualMachine{}

	getVm := func(s *sling.Sling) *sling.Sling {
		return s.Get(vmIdInEnvironmentPath(envId, vmId))
	}

	_, err := RunSkytapRequest(client, true, vm, getVm)
	return vm, err
}

/*
 Get a VM from an existing template.
*/
func GetVirtualMachineInTemplate(client SkytapClient, templateId string, vmId string) (*VirtualMachine, error) {
	vm := &VirtualMachine{}

	getVm := func(s *sling.Sling) *sling.Sling {
		return s.Get(vmIdInTemplatePath(templateId, vmId))
	}

	_, err := RunSkytapRequest(client, true, vm, getVm)
	return vm, err
}

/*
 Get a VM without reference to environment or template. The result object should contain information on its source.
*/
func GetVirtualMachine(client SkytapClient, vmId string) (*VirtualMachine, error) {
	vm := &VirtualMachine{}

	getVm := func(s *sling.Sling) *sling.Sling {
		return s.Get(vmIdPath(vmId))
	}

	_, err := RunSkytapRequest(client, false, vm, getVm)
	return vm, err
}

/*
 Delete a VM.
*/
func DeleteVirtualMachine(client SkytapClient, vmId string) error {
	log.WithFields(log.Fields{"vmId": vmId}).Info("Deleting VM")

	deleteVm := func(s *sling.Sling) *sling.Sling { return s.Delete(vmIdPath(vmId)) }
	_, err := RunSkytapRequest(client, false, nil, deleteVm)
	return err
}
