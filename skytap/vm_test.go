package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

	first := true
	second := true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if first {
			assert.Equal(t, "/v2/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, exampleEnvironment)
			assert.NoError(t, err)
			first = false
		} else if second {
			assert.Equal(t, "/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, request, string(body), "Bad request body")

			_, err = io.WriteString(rw, response)
			assert.NoError(t, err)
			second = false
		}
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

	assert.False(t, first)
	assert.False(t, second)
}

//func TestCreateVM422(t *testing.T) {
//	request := fmt.Sprintf(`{
//		"template_id": "%d",
//    		"vm_ids": [
//        		"%d"
//    	]
//	}`, 42, 43)
//	response := fmt.Sprintf(string(readTestFile(t, "createVMResponse.json")), 123, 123, 456)
//	requestCounter := 0
//
//	skytap, hs, handler := createClient(t)
//	defer hs.Close()
//
//	first := true
//	second := true
//
//	*handler = func(rw http.ResponseWriter, req *http.Request) {
//		if first {
//			assert.Equal(t, "/v2/configurations/123", req.URL.Path, "Bad path")
//			assert.Equal(t, http.MethodGet, req.Method, "Bad method")
//
//			_, err := io.WriteString(rw, exampleEnvironment)
//			assert.NoError(t, err)
//			first = false
//		} else if second {
//			assert.Equal(t, "/configurations/123", req.URL.Path, "Bad path")
//			assert.Equal(t, "PUT", req.Method, "Bad method")
//
//			body, err := ioutil.ReadAll(req.Body)
//			assert.Nil(t, err, "Bad request body")
//			assert.JSONEq(t, request, string(body), "Bad request body")
//
//			if requestCounter == 0 {
//				rw.WriteHeader(http.StatusUnprocessableEntity)
//			} else {
//				_, err = io.WriteString(rw, response)
//				assert.NoError(t, err)
//			}
//			requestCounter++
//		}
//	}
//	opts := &CreateVMRequest{
//		TemplateID: "42",
//		VMID:       "43",
//	}
//
//	createdVM, err := skytap.VMs.Create(context.Background(), "123", opts)
//	assert.Nil(t, err, "Bad API method")
//
//	var environment Environment
//	err = json.Unmarshal([]byte(response), &environment)
//	assert.NoError(t, err)
//	assert.Equal(t, environment.VMs[1], *createdVM, "Bad VM")
//
//	assert.Equal(t, 2, requestCounter)
//
//	assert.False(t, first)
//	assert.False(t, second)
//}

func TestCreateVMError(t *testing.T) {
	request := fmt.Sprintf(`{
		"template_id": "%d",
    		"vm_ids": [
        		"%d"
    	]
	}`, 42, 43)
	requestCounter := 0

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	first := true
	second := true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if first {
			assert.Equal(t, "/v2/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, exampleEnvironment)
			assert.NoError(t, err)
			first = false
		} else if second {
			assert.Equal(t, "/configurations/123", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, request, string(body), "Bad request body")

			rw.WriteHeader(401)
			requestCounter++
			second = false
		}
	}
	opts := &CreateVMRequest{
		TemplateID: "42",
		VMID:       "43",
	}

	createdVM, err := skytap.VMs.Create(context.Background(), "123", opts)
	assert.Nil(t, createdVM, "Bad API method")
	errorResponse := err.(*ErrorResponse)

	assert.Nil(t, errorResponse.RetryAfter)
	assert.Equal(t, 1, requestCounter)
	assert.Equal(t, http.StatusUnauthorized, errorResponse.Response.StatusCode)

	assert.False(t, first)
	assert.False(t, second)
}

func TestReadVM(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, response)
		assert.NoError(t, err)
	}

	vm, err := skytap.VMs.Get(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")

	var vmExpected VM
	err = json.Unmarshal([]byte(response), &vmExpected)
	assert.Equal(t, vmExpected, *vm, "Bad VM")
}

// Not testing stopping and starting.
func TestUpdateVM(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var vmCurrent VM
	err := json.Unmarshal([]byte(response), &vmCurrent)
	assert.NoError(t, err)
	vmCurrent.Hardware.Disks = vmCurrent.Hardware.Disks[0:3]
	bytesCurrent, err := json.Marshal(&vmCurrent)
	assert.Nil(t, err, "Bad vm")

	var vmDisksAdded VM
	err = json.Unmarshal([]byte(response), &vmDisksAdded)
	assert.NoError(t, err)
	vmDisksAdded.Hardware.Disks = vmDisksAdded.Hardware.Disks[0:3]
	vmDisksAdded.Hardware.Disks[0].Size = intToPtr(10000)
	vmDisksAdded.Hardware.Disks = append(vmDisksAdded.Hardware.Disks, *createDisk("disk-20142867-38186761-scsi-0-3", nil))
	bytesDisksAdded, err := json.Marshal(&vmDisksAdded)
	assert.Nil(t, err, "Bad vm")

	var vmNew VM
	err = json.Unmarshal([]byte(response), &vmNew)
	vmNew.Hardware.Disks = vmNew.Hardware.Disks[0:2]
	vmNew.Hardware.Disks[1].Name = strToPtr("1")
	vmNew.Hardware.Disks = append(vmNew.Hardware.Disks, *createDisk("disk-20142867-38186761-scsi-0-3", strToPtr("3")))
	bytesNew, err := json.Marshal(&vmNew)
	assert.Nil(t, err, "Bad vm")

	zero := true
	first := true
	second := true
	third := true
	fourth := true
	fifth := true
	sixth := true
	seventh := true
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if zero {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
			zero = false
		} else if first {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
			first = false
		} else if second {
			// add the hardware changes
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"name": "updated vm", "hardware" : {"cpus": 12, "ram": 8192, "disks": {"new": [51200], "existing": {"disk-20142867-38186761-scsi-0-0" : {"id":"disk-20142867-38186761-scsi-0-0", "size": 10000}}}}}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
			second = false
		} else if third {
			// wait until not busy
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
			third = false
		} else if fourth {
			// the last retry - gives the expected count (6 currently)
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesDisksAdded))
			assert.NoError(t, err)
			fourth = false
		} else if fifth {
			// delete the disks
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"name": "updated vm", "hardware" : {"disks": {"existing": {"disk-20142867-38186761-scsi-0-2" : {"id":"disk-20142867-38186761-scsi-0-2", "size": null}}}}}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytesDisksAdded))
			assert.NoError(t, err)
			fifth = false
		} else if sixth {
			// wait until not busy
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesDisksAdded))
			assert.NoError(t, err)
			sixth = false
		} else if seventh {
			// the last retry - gives the expected count (4 currently)
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesNew))
			assert.NoError(t, err)
			fourth = false
		} else {
			seventh = false
		}
	}

	diskIdentification := make([]DiskIdentification, 2)
	diskIdentification[0] = DiskIdentification{ID: strToPtr("disk-20142867-38186761-scsi-0-1"),
		Size: intToPtr(51200),
		Name: strToPtr("1")}
	diskIdentification[1] = DiskIdentification{ID: nil,
		Size: intToPtr(51200),
		Name: strToPtr("3")}

	opts := &UpdateVMRequest{
		Name: strToPtr("updated vm"),
		Hardware: &UpdateHardware{
			CPUs: intToPtr(12),
			RAM:  intToPtr(8192),
			UpdateDisks: &UpdateDisks{
				NewDisks:           []int{51200},
				DiskIdentification: diskIdentification,
				OSSize:             intToPtr(10000),
			},
		},
	}
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.False(t, zero)
	assert.False(t, first)
	assert.False(t, second)
	assert.False(t, third)
	assert.False(t, fourth)
	assert.False(t, fifth)
	assert.False(t, sixth)
	assert.True(t, seventh)

	assert.Equal(t, vmNew, *vmUpdate, "Bad vm")

	disks := vmUpdate.Hardware.Disks
	assert.Equal(t, 3, len(disks), "disk length")

	assert.Nil(t, disks[0].Name, "os")
	assert.Equal(t, "1", *disks[1].Name, "disk name")
	assert.Equal(t, "3", *disks[2].Name, "disk name")

}

// Updating cpu and ram is possible on their own
// Testing stopping and starting.
func TestUpdateCPURAMVM(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var vmRunning VM
	err := json.Unmarshal([]byte(response), &vmRunning)
	assert.NoError(t, err)
	vmRunning.Runstate = vmRunStateToPtr(VMRunstateRunning)
	bytesRunning, err := json.Marshal(&vmRunning)
	assert.Nil(t, err, "Bad vm")

	var vm VM
	err = json.Unmarshal([]byte(response), &vm)
	assert.NoError(t, err)
	*vm.Name = "updated vm"
	*vm.Hardware.CPUs = 12
	*vm.Hardware.RAM = 8192
	bytes, err := json.Marshal(&vm)
	assert.Nil(t, err, "Bad vm")

	var vmEnriched VM
	err = json.Unmarshal([]byte(response), &vmEnriched)
	assert.NoError(t, err)
	*vmEnriched.Name = "updated vm"
	*vmEnriched.Hardware.CPUs = 12
	*vmEnriched.Hardware.RAM = 8192
	vmEnriched.Hardware.Disks[1].Name = strToPtr("1")
	vmEnriched.Hardware.Disks[2].Name = strToPtr("2")
	vmEnriched.Hardware.Disks[3].Name = strToPtr("3")
	vmEnriched.Runstate = vmRunStateToPtr(VMRunstateRunning)
	bytesEnriched, err := json.Marshal(&vmEnriched)
	assert.Nil(t, err, "Bad vm")

	zero := true
	first := true
	second := true
	secondHalf := true
	third := true
	fourth := true
	fifth := true
	sixth := true
	sixHalf := true
	seventh := true
	eighth := true
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if zero {
			// get vm
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesRunning))
			assert.NoError(t, err)
			zero = false
		} else if first {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesRunning))
			assert.NoError(t, err)
			first = false
		} else if second {
			// turn to stopped
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"runstate":"stopped"}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, response)
			assert.NoError(t, err)
			second = false
		} else if secondHalf {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
			secondHalf = false
		} else if third {
			// get vm to confirm it is stopped
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, response)
			assert.NoError(t, err)
			third = false
		} else if fourth {
			// update
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"name": "updated vm", "hardware" : {"cpus": 12, "ram": 8192}}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
			fourth = false
		} else if fifth {
			// wait until not busy
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
			fifth = false
		} else if sixth {
			// get updated vm
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesEnriched))
			assert.NoError(t, err)
			sixth = false
		} else if sixHalf {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesEnriched))
			assert.NoError(t, err)
			sixHalf = false
		} else if seventh {
			// switch back to running
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"runstate":"running"}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytesRunning))
			assert.NoError(t, err)
			seventh = false
		} else {
			// dont bother waiting for vm to be running
			eighth = false
		}
	}

	diskIDs := []DiskIdentification{
		{strToPtr("disk-20142867-38186761-scsi-0-1"), intToPtr(51200), strToPtr("1")},
		{strToPtr("disk-20142867-38186761-scsi-0-2"), intToPtr(51200), strToPtr("2")},
		{strToPtr("disk-20142867-38186761-scsi-0-3"), intToPtr(51200), strToPtr("3")},
	}

	opts := &UpdateVMRequest{
		Name: strToPtr(*vm.Name),
		Hardware: &UpdateHardware{
			CPUs: intToPtr(*vm.Hardware.CPUs),
			RAM:  intToPtr(*vm.Hardware.RAM),
			UpdateDisks: &UpdateDisks{
				DiskIdentification: diskIDs,
			},
		},
	}
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.False(t, zero)
	assert.False(t, first)
	assert.False(t, second)
	assert.False(t, secondHalf)
	assert.False(t, third)
	assert.False(t, fourth)
	assert.False(t, fifth)
	assert.False(t, sixth)
	assert.False(t, sixHalf)
	assert.False(t, seventh)
	assert.True(t, eighth)

	assert.Equal(t, vmEnriched, *vmUpdate, "Bad vm")
	assert.Equal(t, VMRunstateRunning, *vmUpdate.Runstate, "running")
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

	first := true
	second := true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if first {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, http.MethodGet, req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
			first = false
		} else if second {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"runstate":"running"}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytes))
			assert.NoError(t, err)
			second = false
		}
	}

	opts := &UpdateVMRequest{
		Runstate: vmRunStateToPtr(*vm.Runstate),
	}
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, vm, *vmUpdate, "Bad vm")
	assert.Equal(t, VMRunstateRunning, *vmUpdate.Runstate, "running")

	assert.False(t, first)
	assert.False(t, second)
}

func TestDeleteVM(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
		assert.Equal(t, "DELETE", req.Method, "Bad method")
	}

	err := skytap.VMs.Delete(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")
}

func TestListVMs(t *testing.T) {
	response := fmt.Sprintf(string(readTestFile(t, "VMResponse.json")), 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		_, err := io.WriteString(rw, fmt.Sprintf(`[%+v]`, response))
		assert.NoError(t, err)
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

	addOSDiskResize(intToPtr(10000), &vm, updates)
	assert.Equal(t, 2, len(updates))
	assert.Equal(t, ExistingDisk{ID: strToPtr("disk-20142867-38186761-scsi-0-0"),
		Size: intToPtr(10000)}, updates["disk-20142867-38186761-scsi-0-0"])
}
