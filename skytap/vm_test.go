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

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/configurations/123", req.URL.Path, "Bad path")
		assert.Equal(t, "PUT", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, request, string(body), "Bad request body")

		_, err = io.WriteString(rw, response)
		assert.NoError(t, err)
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

	first := true
	second := true
	third := true
	fourth := true
	fifth := true
	sixth := true
	seventh := true
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if first {
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

			_, err := io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
			first = false
		} else if second {
			// add the hardware changes
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"name": "updated vm", "hardware" : {"cpus": 12, "ram": 8192, "disks": {"new": [51200]}}}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytesCurrent))
			assert.NoError(t, err)
			second = false
		} else if third {
			// does it need a retry block?
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
			// first retry - still not as expected
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
			},
		},
	}
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

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

	first := true
	second := true
	third := true
	fourth := true
	fifth := true
	sixth := true
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if first {
			// get vm
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "GET", req.Method, "Bad method")

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
			// switch back to running
			assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
			assert.Equal(t, "PUT", req.Method, "Bad method")

			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err, "Bad request body")
			assert.JSONEq(t, `{"runstate":"running"}`, string(body), "Bad request body")

			_, err = io.WriteString(rw, string(bytesRunning))
			assert.NoError(t, err)
			fifth = false
		} else {
			// dont bother waiting for vm to be running
			sixth = false
		}
	}

	opts := &UpdateVMRequest{
		Name: strToPtr(*vm.Name),
		Hardware: &UpdateHardware{
			CPUs: intToPtr(*vm.Hardware.CPUs),
			RAM:  intToPtr(*vm.Hardware.RAM),
		},
	}
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.False(t, first)
	assert.False(t, second)
	assert.False(t, third)
	assert.False(t, fourth)
	assert.False(t, fifth)
	assert.True(t, sixth)

	assert.Equal(t, vm, *vmUpdate, "Bad vm")
	assert.Equal(t, VMRunstateStopped, *vmUpdate.Runstate, "still stopped")
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

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
		assert.Equal(t, "PUT", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, `{"runstate":"running"}`, string(body), "Bad request body")

		_, err = io.WriteString(rw, string(bytes))
		assert.NoError(t, err)
	}

	opts := &UpdateVMRequest{
		Runstate: vmRunStateToPtr(*vm.Runstate),
	}
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, vm, *vmUpdate, "Bad vm")
	assert.Equal(t, VMRunstateRunning, *vmUpdate.Runstate, "running")
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
