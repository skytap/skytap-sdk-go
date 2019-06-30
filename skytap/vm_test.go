package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVM(t *testing.T) {
	request := fmt.Sprintf(`{
		"template_id": "%d",
    		"vm_ids": [
        		"%d"
    	]
	}`, 42, 43)
	response := fmt.Sprintf(string(readTestFile(t, "createVMResponse.json")), 123, 123, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(readTestFile(t, "exampleEnvironment.json")))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodPut, req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, request, string(body), "Bad request body")

			_, err = io.WriteString(rw, response)
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(readTestFile(t, "exampleEnvironment.json")))
			assert.NoError(t, err)
		} else if requestCounter == 3 {
			assert.Equal(t, "/v2/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodPut, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(readTestFile(t, "exampleEnvironment.json")))
			assert.NoError(t, err)
		}
		requestCounter++
	}
	opts := &CreateVMRequest{
		TemplateID: "42",
		VMID:       "43",
	}

	createdVM, err := skytap.VMs.Create(context.Background(), "123", opts)
	assert.Nil(t, err, "Bad API method")

	var environment Environment
	err = json.Unmarshal([]byte(response), &environment)
	assert.NoError(t, err)
	assert.Equal(t, environment.VMs[1], *createdVM, "Bad VM")

	assert.Equal(t, 3, requestCounter)
}

func TestCreateVMError(t *testing.T) {
	request := fmt.Sprintf(`{
		"template_id": "%d",
    		"vm_ids": [
        		"%d"
    	]
	}`, 42, 43)
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(readTestFile(t, "exampleEnvironment.json")))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, request, string(body), "Bad request body")

			rw.WriteHeader(401)
		}
		requestCounter++
	}
	opts := &CreateVMRequest{
		TemplateID: "42",
		VMID:       "43",
	}

	createdVM, err := skytap.VMs.Create(context.Background(), "123", opts)
	assert.Nil(t, createdVM, "Bad API method")
	errorResponse := err.(*ErrorResponse)

	assert.Equal(t, http.StatusUnauthorized, errorResponse.Response.StatusCode)

	assert.Equal(t, 2, requestCounter)
}

func TestReadVM(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, response)
		assert.NoError(t, err)
		requestCounter++
	}

	vm, err := skytap.VMs.Get(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")

	var vmExpected VM
	err = json.Unmarshal([]byte(response), &vmExpected)
	assert.Equal(t, vmExpected, *vm, "Bad VM")

	assert.Equal(t, 1, requestCounter)
}

func TestUpdateVMModifyDisks(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var vmCurrent VM
	err := json.Unmarshal([]byte(response), &vmCurrent)
	assert.NoError(t, err)
	bytesCurrent, err := json.Marshal(&vmCurrent)
	assert.Nil(t, err, "Bad vm")

	var vmOSSizeDiskAmendedDiskAdded VM
	err = json.Unmarshal([]byte(response), &vmOSSizeDiskAmendedDiskAdded)
	assert.NoError(t, err)
	vmOSSizeDiskAmendedDiskAdded.Hardware.Disks[0].Size = intToPtr(51201)
	vmOSSizeDiskAmendedDiskAdded.Hardware.Disks[1].Size = intToPtr(51202)
	vmOSSizeDiskAmendedDiskAdded.Hardware.Disks = append(vmOSSizeDiskAmendedDiskAdded.Hardware.Disks, *createDisk("disk-20142867-38186761-scsi-0-4", nil))
	bytesDisksModified, err := json.Marshal(&vmOSSizeDiskAmendedDiskAdded)
	assert.Nil(t, err, "Bad vm")

	requestCounter := 0
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			// add the hardware changes
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `
				{
					"name": "test vm", "runstate":"stopped", 
					"hardware" :{
						"disks": {
    						"new": [51200],
							"existing": {
								"disk-20142867-38186761-scsi-0-0" : {"id":"disk-20142867-38186761-scsi-0-0", "size": 51201},
								"disk-20142867-38186761-scsi-0-1" : {"id":"disk-20142867-38186761-scsi-0-1", "size": 51202}
							}
						}
					}
				}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytesDisksModified))
			assert.NoError(t, err)
		} else if requestCounter == 3 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesDisksModified))
			assert.NoError(t, err)
		} else if requestCounter == 4 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesDisksModified))
			assert.NoError(t, err)
		}
		requestCounter++
	}

	opts := createVMUpdateStructure()
	opts.Hardware.RAM = nil
	opts.Hardware.CPUs = nil
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, 5, requestCounter)

	vmOSSizeDiskAmendedDiskAdded.Hardware.Disks[1].Name = strToPtr("1")
	vmOSSizeDiskAmendedDiskAdded.Hardware.Disks[2].Name = strToPtr("2")
	vmOSSizeDiskAmendedDiskAdded.Hardware.Disks[3].Name = strToPtr("3")
	vmOSSizeDiskAmendedDiskAdded.Hardware.Disks[4].Name = strToPtr("4")
	assert.Equal(t, vmOSSizeDiskAmendedDiskAdded, *vmUpdate, "Bad vm")

	disks := vmUpdate.Hardware.Disks
	assert.Equal(t, 5, len(disks), "disk length")

	assert.Nil(t, disks[0].Name, "os")
}

func TestUpdateVMDeleteDisk(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var vmCurrent VM
	err := json.Unmarshal([]byte(response), &vmCurrent)
	assert.NoError(t, err)
	bytesCurrent, err := json.Marshal(&vmCurrent)
	assert.Nil(t, err, "Bad vm")

	var vmDisksRemoved VM
	err = json.Unmarshal([]byte(response), &vmDisksRemoved)
	vmDisksRemoved.Hardware.Disks = vmDisksRemoved.Hardware.Disks[0:2]
	bytesDisksModified, err := json.Marshal(&vmDisksRemoved)
	assert.Nil(t, err, "Bad vm")

	requestCounter := 0
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			// delete the disks
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `
				{
					"hardware" : {
						"disks": {
							"existing": {
								"disk-20142867-38186761-scsi-0-2" : {"id":"disk-20142867-38186761-scsi-0-2", "size": null},
								"disk-20142867-38186761-scsi-0-3" : {"id":"disk-20142867-38186761-scsi-0-3", "size": null}
							}
						}
					}
				}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytesDisksModified))
			assert.NoError(t, err)
		} else if requestCounter == 3 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesDisksModified))
			assert.NoError(t, err)
		} else if requestCounter == 4 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesDisksModified))
			assert.NoError(t, err)
		}
		requestCounter++
	}

	opts := createVMUpdateStructure()
	opts.Name = nil
	opts.Runstate = nil
	opts.Hardware.RAM = nil
	opts.Hardware.CPUs = nil
	opts.Hardware.UpdateDisks.NewDisks = nil
	opts.Hardware.UpdateDisks.OSSize = nil
	opts.Hardware.UpdateDisks.DiskIdentification = opts.Hardware.UpdateDisks.DiskIdentification[0:1]
	opts.Hardware.UpdateDisks.DiskIdentification[0].Size = intToPtr(51200)
	delete(opts.Hardware.UpdateDisks.ExistingDisks, "disk-20142867-38186761-scsi-0-1")
	delete(opts.Hardware.UpdateDisks.ExistingDisks, "disk-20142867-38186761-scsi-0-2")
	delete(opts.Hardware.UpdateDisks.ExistingDisks, "disk-20142867-38186761-scsi-0-3")
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, 5, requestCounter)

	vmDisksRemoved.Hardware.Disks[1].Name = strToPtr("1")
	assert.Equal(t, vmDisksRemoved, *vmUpdate, "Bad vm")

	disks := vmUpdate.Hardware.Disks
	assert.Equal(t, 2, len(disks), "disk length")
}

// Updating cpu and ram is possible on their own
// Testing stopping and starting.
func TestUpdateCPURAMVM(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var vmExisting VM
	err := json.Unmarshal([]byte(response), &vmExisting)
	assert.NoError(t, err)
	vmExisting.Runstate = vmRunStateToPtr(VMRunstateRunning)
	bytesExisting, err := json.Marshal(&vmExisting)
	assert.Nil(t, err, "Bad vm")

	var vmUpdated VM
	err = json.Unmarshal([]byte(response), &vmUpdated)
	assert.NoError(t, err)
	*vmUpdated.Name = "updated vm"
	*vmUpdated.Hardware.CPUs = 6
	*vmUpdated.Hardware.RAM = 4096
	vmUpdated.Runstate = vmRunStateToPtr(VMRunstateStopped)
	bytes, err := json.Marshal(&vmUpdated)
	assert.Nil(t, err, "Bad vm")

	vmUpdated.Runstate = vmRunStateToPtr(VMRunstateRunning)
	bytesRunning, err := json.Marshal(&vmUpdated)

	requestCounter := 0
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesExisting))
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesExisting))
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			// turn to stopped
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"runstate":"stopped"}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, response)
			assert.NoError(t, err)
		} else if requestCounter == 3 { // confirm vm in correct state, i.e. not busy
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
		} else if requestCounter == 4 { // confirm vm in correct state, i.e. not busy
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
		} else if requestCounter == 5 {
			// update
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"name": "updated vm", "hardware" : {"cpus": 6, "ram": 4096}}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
		} else if requestCounter == 6 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
		} else if requestCounter == 7 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
		} else if requestCounter == 8 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
		} else if requestCounter == 9 {
			// turn to running
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"runstate":"running"}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
		} else if requestCounter == 10 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err = io.WriteString(rw, string(bytesRunning))
			assert.NoError(t, err)
		} else if requestCounter == 11 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err = io.WriteString(rw, string(bytesRunning))
			assert.NoError(t, err)
		}
		requestCounter++
	}

	opts := createVMUpdateStructure()
	opts.Name = strToPtr("updated vm")
	opts.Runstate = nil
	opts.Hardware.RAM = intToPtr(4096)
	opts.Hardware.CPUs = intToPtr(6)
	opts.Hardware.UpdateDisks.NewDisks = nil
	opts.Hardware.UpdateDisks.ExistingDisks = nil
	opts.Hardware.UpdateDisks.OSSize = nil
	opts.Hardware.UpdateDisks.DiskIdentification[0].Size = intToPtr(51200)
	vm, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, 12, requestCounter)

	vmUpdated.Hardware.Disks[1].Name = strToPtr("1")
	vmUpdated.Hardware.Disks[2].Name = strToPtr("2")
	vmUpdated.Hardware.Disks[3].Name = strToPtr("3")
	assert.Equal(t, vmUpdated, *vm, "Bad vm")
	assert.Equal(t, VMRunstateRunning, *vm.Runstate, "running")
}

// Updating runstate can only be done on its own
func TestUpdateRunstateVM(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var vm VM
	err := json.Unmarshal([]byte(response), &vm)
	assert.NoError(t, err)
	*vm.Runstate = VMRunstateRunning

	bytes, err := json.Marshal(&vm)
	assert.Nil(t, err, "Bad vm")

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"runstate":"running"}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
		} else if requestCounter == 2 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err = io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
		}
		requestCounter++
	}

	opts := &UpdateVMRequest{
		Runstate: vmRunStateToPtr(*vm.Runstate),
	}
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, vm, *vmUpdate, "Bad vm")
	assert.Equal(t, VMRunstateRunning, *vmUpdate.Runstate, "running")

	assert.Equal(t, 3, requestCounter)
}

func TestDeleteVM(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		if requestCounter == 0 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
		} else if requestCounter == 1 {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "DELETE", req.Method, "Bad method")
		}
		requestCounter++
	}

	err := skytap.VMs.Delete(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, 2, requestCounter)
}

func TestListVMs(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	requestCounter := 0

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("Request: (%d)\n", requestCounter)
		assert.Equal(t, "/v2/configurations/123/vms", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, fmt.Sprintf(`[%+v]`, response))
		assert.NoError(t, err)
		requestCounter++
	}

	result, err := skytap.VMs.List(context.Background(), "123")
	assert.Nil(t, err, "Bad API method")

	var found = false
	for _, vm := range result.Value {
		if *vm.Name == "test vm" {
			found = true
			break
		}
	}
	assert.True(t, found, "VM not found")

	assert.Equal(t, 1, requestCounter)
}

func readTestFile(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func TestBuildListOfDiskSizes(t *testing.T) {
	removes := make(map[string]ExistingDisk)
	removes["1"] = ExistingDisk{
		ID: strToPtr("1"),
	}
	removes["2"] = ExistingDisk{
		ID: strToPtr("2"),
	}
	nameSizes := []DiskIdentification{
		{nil, intToPtr(51200), strToPtr("1")},
		{nil, intToPtr(51201), strToPtr("2")},
		{nil, intToPtr(51200), strToPtr("3")},
	}
	opts := &UpdateVMRequest{
		Hardware: &UpdateHardware{
			UpdateDisks: &UpdateDisks{
				NewDisks:           []int{51200, 51201, 51200},
				DiskIdentification: nameSizes,
				ExistingDisks:      removes,
			},
		},
	}

	sizes := opts.Hardware.UpdateDisks.NewDisks
	assert.Equal(t, []int{51200, 51201, 51200}, sizes)

	opts.Hardware.UpdateDisks.NewDisks = []int{51200, 51201, 51200}
	opts.Hardware.UpdateDisks.DiskIdentification = nil

	sizes = opts.Hardware.UpdateDisks.NewDisks
	assert.Equal(t, []int{51200, 51201, 51200}, sizes)
}

func TestMatchUpNamesWithExistingDisks(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	var vm VM
	err := json.Unmarshal([]byte(response), &vm)
	assert.NoError(t, err)
	currentDisks := vm.Hardware.Disks

	nameSizes := []DiskIdentification{
		{strToPtr("disk-20142867-38186761-scsi-0-1"), intToPtr(51200), strToPtr("1")},
		{strToPtr("disk-20142867-38186761-scsi-0-2"), intToPtr(51201), strToPtr("2")},
		{nil, intToPtr(51200), strToPtr("4")},
		{strToPtr("disk-20142867-38186761-scsi-0-3"), intToPtr(51200), strToPtr("3")},
		{nil, intToPtr(51200), strToPtr("5")},
		{nil, intToPtr(51200), strToPtr("6")},
	}
	opts := &UpdateVMRequest{
		Hardware: &UpdateHardware{
			UpdateDisks: &UpdateDisks{
				DiskIdentification: nameSizes,
			},
		},
	}

	matchUpExistingDisks(&vm, opts.Hardware.UpdateDisks.DiskIdentification, nil)

	assert.Nil(t, currentDisks[0].Name, "os")
	assert.Equal(t, "1", *currentDisks[1].Name, "disk name")
	assert.Equal(t, "2", *currentDisks[2].Name, "disk name")
	assert.Equal(t, "3", *currentDisks[3].Name, "disk name")
}

func TestMatchUpNamesWithNewDisks(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	var vm VM
	err := json.Unmarshal([]byte(response), &vm)
	assert.NoError(t, err)
	// add three more
	vm.Hardware.Disks = append(vm.Hardware.Disks, *createDisk("disk-20142867-38186761-scsi-0-4", strToPtr("4")))
	vm.Hardware.Disks = append(vm.Hardware.Disks, *createDisk("disk-20142867-38186761-scsi-0-5", strToPtr("5")))
	vm.Hardware.Disks = append(vm.Hardware.Disks, *createDisk("disk-20142867-38186761-scsi-0-6", strToPtr("6")))
	allDisks := vm.Hardware.Disks

	nameSizes := []DiskIdentification{
		{strToPtr("disk-20142867-38186761-scsi-0-1"), intToPtr(51200), strToPtr("1")},
		{strToPtr("disk-20142867-38186761-scsi-0-2"), intToPtr(51201), strToPtr("2")},
		{nil, intToPtr(51200), strToPtr("4")},
		{strToPtr("disk-20142867-38186761-scsi-0-3"), intToPtr(51200), strToPtr("3")},
		{nil, intToPtr(51200), strToPtr("5")},
		{nil, intToPtr(51200), strToPtr("6")},
	}
	opts := &UpdateVMRequest{
		Hardware: &UpdateHardware{
			UpdateDisks: &UpdateDisks{
				DiskIdentification: nameSizes,
			},
		},
	}

	matchUpExistingDisks(&vm, opts.Hardware.UpdateDisks.DiskIdentification, nil)

	matchUpNewDisks(&vm, opts.Hardware.UpdateDisks.DiskIdentification, nil)

	assert.Equal(t, "4", *allDisks[4].Name, "disk name")
	assert.Equal(t, "5", *allDisks[5].Name, "disk name")
	assert.Equal(t, "6", *allDisks[6].Name, "disk name")
}

func createDisk(id string, name *string) *Disk {
	return &Disk{
		ID:         strToPtr(id),
		Size:       intToPtr(51200),
		Type:       strToPtr("SCSI"),
		Controller: strToPtr("0"),
		LUN:        strToPtr("-1"),
		Name:       name,
	}
}

func TestBuildRemoveList(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	var vm VM
	err := json.Unmarshal([]byte(response), &vm)
	assert.NoError(t, err)

	// expecting to delete disk-20142867-38186761-scsi-0-3
	nameSizes := []DiskIdentification{
		{strToPtr("disk-20142867-38186761-scsi-0-1"), intToPtr(51200), strToPtr("old1")},
		{strToPtr("disk-20142867-38186761-scsi-0-2"), intToPtr(51200), strToPtr("old2")},
		{nil, intToPtr(51200), strToPtr("new1")},
	}

	removes := buildRemoveList(&vm, nameSizes)

	assert.Equal(t, ExistingDisk{ID: strToPtr("disk-20142867-38186761-scsi-0-3")}, removes["disk-20142867-38186761-scsi-0-3"])
}

func TestBuildUpdateList(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	var vm VM
	err := json.Unmarshal([]byte(response), &vm)
	assert.NoError(t, err)

	// expecting to delete disk-20142867-38186761-scsi-0-3
	nameSizes := []DiskIdentification{
		{strToPtr("disk-20142867-38186761-scsi-0-1"), intToPtr(51199), strToPtr("old1")},
		{strToPtr("disk-20142867-38186761-scsi-0-2"), intToPtr(51201), strToPtr("old2")},
		{nil, intToPtr(51200), strToPtr("new1")},
	}

	updates := buildUpdateList(&vm, nameSizes)

	assert.Equal(t, ExistingDisk{ID: strToPtr("disk-20142867-38186761-scsi-0-2"),
		Size: intToPtr(51201)}, updates["disk-20142867-38186761-scsi-0-2"])
}

func TestOSDiskResize(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)
	var vm VM
	err := json.Unmarshal([]byte(response), &vm)
	assert.NoError(t, err)

	nameSizes := []DiskIdentification{
		{strToPtr("disk-20142867-38186761-scsi-0-1"), intToPtr(51199), strToPtr("old1")},
		{strToPtr("disk-20142867-38186761-scsi-0-2"), intToPtr(51201), strToPtr("old2")},
		{nil, intToPtr(51200), strToPtr("new1")},
	}
	updates := buildUpdateList(&vm, nameSizes)

	addOSDiskResize(nil, &vm, updates)
	assert.Equal(t, 1, len(updates))

	addOSDiskResize(intToPtr(51203), &vm, updates)
	assert.Equal(t, 2, len(updates))
	assert.Equal(t, ExistingDisk{ID: strToPtr("disk-20142867-38186761-scsi-0-0"),
		Size: intToPtr(51203)}, updates["disk-20142867-38186761-scsi-0-0"])
}

func TestCompareVMCreateTrue(t *testing.T) {
	exampleVM := fmt.Sprintf(string(readTestFile(t, "createVMResponse.json")), 123, 123, 456)

	var env Environment
	err := json.Unmarshal([]byte(exampleVM), &env)
	env.Runstate = environmentRunStateToPtr(EnvironmentRunstateStopped)
	assert.NoError(t, err)
	opts := CreateVMRequest{
		TemplateID: "42",
		VMID:       "43",
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		bytes, err := json.Marshal(&env)
		assert.Nil(t, err, "Bad vm")
		_, err = io.WriteString(rw, string(bytes))
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.diskIdentification = buildDiskidentification()
	message, ok := opts.compare(context.Background(), skytap, &env, state)
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestCompareVMCreateFalse(t *testing.T) {
	exampleVM := fmt.Sprintf(string(readTestFile(t, "createVMResponse.json")), 123, 123, 456)

	var env Environment
	err := json.Unmarshal([]byte(exampleVM), &env)
	assert.NoError(t, err)
	opts := CreateVMRequest{
		TemplateID: "42",
		VMID:       "43",
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, exampleVM)
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.diskIdentification = buildDiskidentification()
	message, ok := opts.compare(context.Background(), skytap, &env, state)
	assert.False(t, ok)
	assert.Equal(t, "VM environment not ready", message)
}

func TestCompareVMUpdateRunStateTrue(t *testing.T) {
	exampleVM := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	var vm VM
	err := json.Unmarshal([]byte(exampleVM), &vm)
	assert.NoError(t, err)
	opts := &UpdateVMRequest{
		Runstate: vmRunStateToPtr(VMRunstateStopped),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, exampleVM)
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.diskIdentification = buildDiskidentification()
	message, ok := opts.compare(context.Background(), skytap, &vm, state)
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestCompareVMUpdateRunStateFalse(t *testing.T) {
	exampleVM := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	var vm VM
	err := json.Unmarshal([]byte(exampleVM), &vm)
	vm.Runstate = vmRunStateToPtr(VMRunstateBusy)
	assert.NoError(t, err)
	opts := &UpdateVMRequest{
		Runstate: vmRunStateToPtr(VMRunstateStopped),
	}
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		bytes, err := json.Marshal(&vm)
		assert.Nil(t, err, "Bad vm")
		_, err = io.WriteString(rw, string(bytes))
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.diskIdentification = buildDiskidentification()
	message, ok := opts.compare(context.Background(), skytap, &vm, state)
	assert.False(t, ok)
	assert.Equal(t, "VM not ready", message)
}

func TestCompareVMUpdateTrue(t *testing.T) {
	exampleVM := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	var vm VM
	err := json.Unmarshal([]byte(exampleVM), &vm)
	assert.NoError(t, err)
	vm.Hardware.Disks = append(vm.Hardware.Disks, Disk{
		ID:   strToPtr("disk-20142867-38186761-scsi-0-4"),
		Size: intToPtr(51200),
	})
	vm.Hardware.Disks[0].Size = intToPtr(51200)

	existingDisks := make(map[string]ExistingDisk)
	existingDisks["disk-20142867-38186761-scsi-0-1"] = ExistingDisk{
		ID:   strToPtr("disk-20142867-38186761-scsi-0-1"),
		Size: intToPtr(51200),
	}
	existingDisks["disk-20142867-38186761-scsi-0-2"] = ExistingDisk{
		ID:   strToPtr("disk-20142867-38186761-scsi-0-2"),
		Size: intToPtr(51200),
	}
	existingDisks["disk-20142867-38186761-scsi-0-3"] = ExistingDisk{
		ID:   strToPtr("disk-20142867-38186761-scsi-0-3"),
		Size: intToPtr(51200),
	}
	opts := &UpdateVMRequest{
		Name:     strToPtr("test vm"),
		Runstate: vmRunStateToPtr(VMRunstateStopped),
		Hardware: &UpdateHardware{
			CPUs: intToPtr(12),
			RAM:  intToPtr(8192),
			UpdateDisks: &UpdateDisks{
				NewDisks:      []int{51200},
				ExistingDisks: existingDisks,
			},
		},
	}

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		bytes, err := json.Marshal(&vm)
		assert.Nil(t, err, "Bad vm")
		_, err = io.WriteString(rw, string(bytes))
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.diskIdentification = buildDiskidentification()
	message, ok := opts.compare(context.Background(), skytap, &vm, state)
	assert.True(t, ok)
	assert.Equal(t, "", message)
}

func TestCompareVMUpdateFalse(t *testing.T) {
	exampleVM := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	var vm VM
	vm.Runstate = vmRunStateToPtr(VMRunstateBusy)
	err := json.Unmarshal([]byte(exampleVM), &vm)
	assert.NoError(t, err)
	opts := createVMUpdateStructure()
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		bytes, err := json.Marshal(&vm)
		assert.Nil(t, err, "Bad vm")
		_, err = io.WriteString(rw, string(bytes))
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.diskIdentification = buildDiskidentification()
	message, ok := opts.compare(context.Background(), skytap, &vm, state)
	assert.False(t, ok)
	assert.Equal(t, "VM not ready", message)
}

func TestCompareDiskStructureNoDisks(t *testing.T) {
	exampleEnvironment := fmt.Sprintf(string(readTestFile(t, "createVMResponse.json")), 123, 123, 456)
	exampleVM := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	var env Environment
	err := json.Unmarshal([]byte(exampleEnvironment), &env)
	assert.NoError(t, err)
	vm := env.VMs[1]
	vm.Hardware.Disks = nil
	opts := createVMUpdateStructure()
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(rw, exampleVM)
		assert.NoError(t, err)
	}
	state := vmRunStateNotBusy("123", "456")
	state.diskIdentification = buildDiskidentification()
	message, ok := opts.compare(context.Background(), skytap, &vm, state)
	assert.False(t, ok)
	assert.Equal(t, "VM not ready", message)
}

func createVMUpdateStructure() *UpdateVMRequest {
	diskIdentification := buildDiskidentification()
	diskIdentification[0] = DiskIdentification{ID: strToPtr("disk-20142867-38186761-scsi-0-1"),
		Size: intToPtr(51202),
		Name: strToPtr("1")}

	existingDisks := make(map[string]ExistingDisk)
	existingDisks["disk-20142867-38186761-scsi-0-1"] = ExistingDisk{
		ID:   strToPtr("disk-20142867-38186761-scsi-0-1"),
		Size: intToPtr(51200),
	}
	existingDisks["disk-20142867-38186761-scsi-0-2"] = ExistingDisk{
		ID:   strToPtr("disk-20142867-38186761-scsi-0-2"),
		Size: intToPtr(51200),
	}
	existingDisks["disk-20142867-38186761-scsi-0-3"] = ExistingDisk{
		ID:   strToPtr("disk-20142867-38186761-scsi-0-3"),
		Size: intToPtr(51200),
	}
	opts := &UpdateVMRequest{
		Name:     strToPtr("test vm"),
		Runstate: vmRunStateToPtr(VMRunstateStopped),
		Hardware: &UpdateHardware{
			CPUs: intToPtr(12),
			RAM:  intToPtr(8192),
			UpdateDisks: &UpdateDisks{
				NewDisks:           []int{51200},
				ExistingDisks:      existingDisks,
				DiskIdentification: diskIdentification,
				OSSize:             intToPtr(51201),
			},
		},
	}
	return opts
}

func buildDiskidentification() []DiskIdentification {
	diskIdentification := make([]DiskIdentification, 4)
	diskIdentification[0] = DiskIdentification{ID: strToPtr("disk-20142867-38186761-scsi-0-1"),
		Size: intToPtr(51200),
		Name: strToPtr("1")}
	diskIdentification[1] = DiskIdentification{ID: strToPtr("disk-20142867-38186761-scsi-0-2"),
		Size: intToPtr(51200),
		Name: strToPtr("2")}
	diskIdentification[2] = DiskIdentification{ID: strToPtr("disk-20142867-38186761-scsi-0-3"),
		Size: intToPtr(51200),
		Name: strToPtr("3")}
	diskIdentification[3] = DiskIdentification{ID: nil,
		Size: intToPtr(51200),
		Name: strToPtr("4")}
	return diskIdentification
}
