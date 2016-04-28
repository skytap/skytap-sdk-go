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
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/dghubble/sling"
)

const (
	EnvironmentPath = "configurations"
)

/**
Skytap Environment resource.
*/
type Environment struct {
	Id          string            `json:"id"`
	Url         string            `json:"url"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Error       []string          `json:"errors"`
	Runstate    string            `json:"runstate"`
	Vms         []*VirtualMachine `json:"vms"`
	Networks    []Network         `json:"networks"`
}

/*
 Request body for create commands.
*/
type CreateEnvironmentBody struct {
	TemplateId string `json:"template_id"`
}

/*
 Request body for merge commands.
*/
type MergeTemplateBody struct {
	TemplateId string   `json:"template_id"`
	VmIds      []string `json:"vm_ids"`
}

/*
 Request body for copy environment commands.
*/
type CopyEnvironmentBody struct {
	EnvironmentId string   `json:"configuration_id"`
	VmIds         []string `json:"vm_ids"`
}

/*
 Request body for merge commands.
*/
type MergeEnvironmentBody struct {
	EnvironmentId string   `json:"merge_configuration"`
	VmIds         []string `json:"vm_ids"`
}

func environmentIdV1Path(envId string) string { return EnvironmentPath + "/" + envId }
func environmentIdPath(envId string) string   { return EnvironmentPath + "/" + envId + ".json" }

/*
 Adds a VM to an existing environment.
*/
func (e *Environment) AddVirtualMachine(client SkytapClient, vmId string) (*Environment, error) {
	log.WithFields(log.Fields{"vmId": vmId, "envId": e.Id}).Info("Adding virtual machine")

	vm, err := GetVirtualMachine(client, vmId)
	if err != nil {
		return e, err
	}

	template, err := vm.GetTemplate(client)
	if err != nil {
		return e, err
	}
	if template != nil {
		return e.MergeTemplateVirtualMachine(client, template.Id, vmId)
	}

	sourceEnv, err := vm.GetEnvironment(client)
	if err != nil {
		return e, err
	}
	if sourceEnv != nil {
		return e.MergeEnvironmentVirtualMachine(client, sourceEnv.Id, vmId)
	}

	return e, errors.New("Unable to determine source of VM, no environment or template url found")
}

func (e *Environment) RunstateStr() string { return e.Runstate }

func (e *Environment) Refresh(client SkytapClient) (RunstateAwareResource, error) {
	return GetEnvironment(client, e.Id)
}

func (e *Environment) WaitUntilInState(client SkytapClient, desiredStates []string, requireStateChange bool) (*Environment, error) {
	r, err := WaitUntilInState(client, desiredStates, e, requireStateChange)
	newEnv := r.(*Environment)
	return newEnv, err
}

func (e *Environment) WaitUntilReady(client SkytapClient) (*Environment, error) {
	return e.WaitUntilInState(client, []string{RunStateStop, RunStateStart, RunStatePause}, false)
}

/*
 Merge an environment based VM into this environment (the VM must be in an existing environment).
*/
func (e *Environment) MergeEnvironmentVirtualMachine(client SkytapClient, envId string, vmId string) (*Environment, error) {
	return e.MergeVirtualMachine(client, &MergeEnvironmentBody{EnvironmentId: envId, VmIds: []string{vmId}})
}

/*
 Merge a template based VM into this environment (the VM must be in an existing template).
*/
func (e *Environment) MergeTemplateVirtualMachine(client SkytapClient, templateId string, vmId string) (*Environment, error) {
	return e.MergeVirtualMachine(client, &MergeTemplateBody{TemplateId: templateId, VmIds: []string{vmId}})
}

/*
 Merge arbitrary VM into this environment.

 mergeBody - The correct representation of the request body, see the MergeEnvironmentVirtualMachine and MergeTemplateVirtualMachine methods.
*/
func (e *Environment) MergeVirtualMachine(client SkytapClient, mergeBody interface{}) (*Environment, error) {

	log.WithFields(log.Fields{"mergeBody": mergeBody, "envId": e.Id}).Info("Merging a VM into environment")

	merge := func(s *sling.Sling) *sling.Sling {
		return s.Put(environmentIdPath(e.Id)).BodyJSON(mergeBody)
	}

	newEnv := &Environment{}
	_, err := RunSkytapRequest(client, false, newEnv, merge)
	if err != nil {
		log.Errorf("Unable to add VM to environment (%s), requestBody: %+v, cause: %s", e.Id, mergeBody, err)
		return nil, err
	}
	return newEnv, nil
}

/*
 Starts an environment.
*/
func (e *Environment) Start(client SkytapClient) (*Environment, error) {
	log.WithFields(log.Fields{"envId": e.Id}).Info("Starting Environment")

	return e.ChangeRunstate(client, RunStateStart, RunStateStart)
}

/*
 Suspends an environment.
*/
func (e *Environment) Suspend(client SkytapClient) (*Environment, error) {
	log.WithFields(log.Fields{"envId": e.Id}).Info("Stopping Environment")

	return e.ChangeRunstate(client, RunStatePause, RunStatePause)
}

/*
 Changes the runstate of the Environment to the specified state and waits until the Environment is in the desired state.
*/
func (e *Environment) ChangeRunstate(client SkytapClient, runstate string, desiredRunstate string) (*Environment, error) {
	log.WithFields(log.Fields{"changeState": runstate, "targetState": desiredRunstate, "envId": e.Id}).Info("Changing VM runstate")

	ready, err := e.WaitUntilReady(client)
	if err != nil {
		return ready, err
	}
	changeState := func(s *sling.Sling) *sling.Sling {
		return s.Put(environmentIdPath(e.Id)).BodyJSON(&RunstateBody{Runstate: runstate})
	}
	_, err = RunSkytapRequest(client, false, nil, changeState)

	if err != nil {
		return e, err
	}
	return e.WaitUntilInState(client, []string{desiredRunstate}, true)
}

/*
 Return an existing environment by id.
*/
func GetEnvironment(client SkytapClient, envId string) (*Environment, error) {
	env := &Environment{}

	getEnv := func(s *sling.Sling) *sling.Sling {
		return s.Get(environmentIdPath(envId))
	}

	_, err := RunSkytapRequest(client, true, env, getEnv)
	return env, err
}

/*
 Create a new environment from a template.
*/
func CreateNewEnvironment(client SkytapClient, templateId string) (*Environment, error) {
	log.WithFields(log.Fields{"templateId": templateId}).Info("Creating environment from template")

	env := &Environment{}

	createEnv := func(s *sling.Sling) *sling.Sling {
		return s.Post(EnvironmentPath + ".json").BodyJSON(&CreateEnvironmentBody{TemplateId: templateId})
	}

	_, err := RunSkytapRequest(client, false, env, createEnv)
	return env, err
}

/*
 Create a new environment from a source template, including only specific VMs, which must be a part of the template.
*/
func CreateNewEnvironmentWithVms(client SkytapClient, templateId string, vmIds []string) (*Environment, error) {
	log.WithFields(log.Fields{"templateId": templateId}).Info("Creating environment from template")

	env := &Environment{}

	createEnvWithVM := func(s *sling.Sling) *sling.Sling {
		return s.Post(EnvironmentPath + ".json").BodyJSON(&MergeTemplateBody{TemplateId: templateId, VmIds: vmIds})
	}

	_, err := RunSkytapRequest(client, false, env, createEnvWithVM)
	return env, err
}

/*
 Create a new environment from a source environment, including only specific VMs, which must be a part of the template.
*/
func CopyEnvironmentWithVms(client SkytapClient, sourceEnvId string, vmIds []string) (*Environment, error) {
	log.WithFields(log.Fields{"sourceEnvId": sourceEnvId}).Info("Copying environment from existing")

	env := &Environment{}

	createEnvWithVM := func(s *sling.Sling) *sling.Sling {
		return s.Post(EnvironmentPath + ".json").BodyJSON(&CopyEnvironmentBody{EnvironmentId: sourceEnvId, VmIds: vmIds})
	}

	_, err := RunSkytapRequest(client, false, env, createEnvWithVM)
	return env, err
}

/*
 Delete an environment by id.
*/
func DeleteEnvironment(client SkytapClient, envId string) error {
	log.WithFields(log.Fields{"envId": envId}).Info("Deleting environment")

	deleteEnv := func(s *sling.Sling) *sling.Sling {
		return s.Delete(EnvironmentPath + "/" + envId)
	}

	_, err := RunSkytapRequest(client, false, nil, deleteEnv)
	return err
}
