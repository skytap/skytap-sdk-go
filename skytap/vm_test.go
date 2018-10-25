package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	exampleVMRequest = `{
		"template_id": "%d",
    		"vm_ids": [
        		"%d"
    	]
	}`
	exampleCreateVMResponse = `{
    	"id": "%d",
    	"url": "https://cloud.skytap.com/configurations/%d",
    	"name": "base",
    	"error": "",
    	"runstate": "busy",
    	"rate_limited": false,
    	"description": "used for basic understanding",
    	"suspend_on_idle": null,
    	"suspend_at_time": null,
    	"routable": false,
    	"vms": [{
            "id": "%d",
            "name": "Red Hat Enterprise Linux 7 Server (with GUI) x64",
            "runstate": "busy",
            "rate_limited": false,
            "hardware": {
                "cpus": 1,
                "supports_multicore": true,
                "cpus_per_socket": 1,
                "ram": 2048,
                "svms": 2,
                "guestOS": "rhel7-64",
                "max_cpus": 12,
                "min_ram": 256,
                "max_ram": 262144,
                "vnc_keymap": null,
                "uuid": null,
                "disks": [
                    {
                        "id": "disk-20142867-38186761-scsi-0-0",
                        "size": 51200,
                        "type": "SCSI",
                        "controller": "0",
                        "lun": "0"
                    }
                ],
                "storage": 51200,
                "upgradable": false,
                "instance_type": null,
                "time_sync_enabled": true,
                "rtc_start_time": null,
                "copy_paste_enabled": true,
                "nested_virtualization": false,
                "architecture": "x86"
            },
            "error": false,
            "error_details": false,
            "asset_id": null,
            "hardware_version": 11,
            "max_hardware_version": 11,
            "interfaces": [
                {
                    "id": "nic-20142867-38186761-0",
                    "ip": "10.0.0.1",
                    "hostname": "rhel7sguix64",
                    "mac": "00:50:56:20:BB:95",
                    "services_count": 0,
                    "services": [],
                    "public_ips_count": 0,
                    "public_ips": [],
                    "vm_id": "37351858",
                    "vm_name": "Red Hat Enterprise Linux 7 Server (with GUI) x64",
                    "status": "Busy",
                    "network_id": "23788208",
                    "network_name": "Default Network",
                    "network_url": "https://cloud.skytap.com/configurations/39855984/networks/23788208",
                    "network_type": "automatic",
                    "network_subnet": "10.0.0.0/24",
                    "nic_type": "vmxnet3",
                    "secondary_ips": [],
                    "public_ip_attachments": []
                }
            ],
            "notes": [],
            "labels": [],
            "credentials": [],
            "desktop_resizable": true,
            "local_mouse_cursor": true,
            "maintenance_lock_engaged": false,
            "region_backend": "skytap",
            "created_at": "2018/10/25 12:57:02 +0100",
            "supports_suspend": true,
            "can_change_object_state": true,
            "containers": null,
            "configuration_url": "https://cloud.skytap.com/configurations/39855984"
        }],
    	"networks": [
       		{
            	"id": "23788208",
            	"url": "https://cloud.skytap.com/configurations/39855984/networks/23788208",
            	"name": "Default Network",
            	"network_type": "automatic",
            	"subnet": "10.0.0.0/24",
            	"subnet_addr": "10.0.0.0",
            	"subnet_size": 24,
            	"gateway": "10.0.0.254",
            	"primary_nameserver": null,
            	"secondary_nameserver": null,
            	"region": "US-West",
            	"domain": "skytap.example",
            	"vpn_attachments": [],
            	"tunnelable": false,
            	"tunnels": []
        	}
    	],
    	"lockversion": "32c107f17d7219f8965bf6b98dcffdfed458747d",
    	"use_smart_client": true,
    	"disable_internet": false,
    	"region": "US-West",
    	"region_backend": "skytap",
    	"owner": "https://cloud.skytap.com/users/372680",
    	"platform_errors": [],
    	"publish_sets": [],
    	"shutdown_on_idle": null,
    	"shutdown_at_time": null,
    	"containers_count": 0,
    	"container_hosts_count": 0
	}`

	exampleVMResponse = `{
        "id": "%d",
        "name": "test vm",
        "runstate": "stopped",
        "rate_limited": false,
        "hardware": {
            "cpus": 1,
            "supports_multicore": true,
            "cpus_per_socket": 1,
            "ram": 2048,
            "svms": 2,
            "guestOS": "rhel7-64",
            "max_cpus": 12,
            "min_ram": 256,
            "max_ram": 262144,
            "vnc_keymap": null,
            "uuid": null,
            "disks": [
                {
                    "id": "disk-20142867-38186761-scsi-0-0",
                    "size": 51200,
                    "type": "SCSI",
                    "controller": "0",
                    "lun": "0"
                }
            ],
            "storage": 51200,
            "upgradable": false,
            "instance_type": null,
            "time_sync_enabled": true,
            "rtc_start_time": null,
            "copy_paste_enabled": true,
            "nested_virtualization": false,
            "architecture": "x86"
        },
        "error": false,
        "error_details": false,
        "asset_id": null,
        "hardware_version": 11,
        "max_hardware_version": 11,
        "interfaces": [
            {
                "id": "nic-20142867-38186761-0",
                "ip": "10.0.0.1",
                "hostname": "rhel7sguix64",
                "mac": "00:50:56:20:BB:95",
                "services_count": 0,
                "services": [],
                "public_ips_count": 0,
                "public_ips": [],
                "vm_id": "37351858",
                "vm_name": "Red Hat Enterprise Linux 7 Server (with GUI) x64",
                "status": "Powered off",
                "network_id": "23788208",
                "network_name": "Default Network",
                "network_url": "https://cloud.skytap.com/v2/configurations/39855984/networks/23788208",
                "network_type": "automatic",
                "network_subnet": "10.0.0.0/24",
                "nic_type": "vmxnet3",
                "secondary_ips": [],
                "public_ip_attachments": []
            }
        ],
        "notes": [],
        "labels": [],
        "credentials": [],
        "desktop_resizable": true,
        "local_mouse_cursor": true,
        "maintenance_lock_engaged": false,
        "region_backend": "skytap",
        "created_at": "2018/10/25 12:57:02 +0100",
        "supports_suspend": true,
        "can_change_object_state": true,
        "containers": null,
        "configuration_url": "https://cloud.skytap.com/v2/configurations/39855984"
	}`
)

func TestCreateVM(t *testing.T) {
	response := fmt.Sprintf(exampleCreateVMResponse, 123, 123, 456)
	request := fmt.Sprintf(exampleVMRequest, 42, 43)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/configurations/123", req.URL.Path, "Bad path")
		assert.Equal(t, "PUT", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, request, string(body), "Bad request body")

		io.WriteString(rw, response)
	}
	opts := &CreateVMRequest{
		TemplateID: "42",
		VMID:       []string{"43"},
	}

	vm, err := skytap.VMs.Create(context.Background(), "123", opts)
	assert.Nil(t, err, "Bad API method")

	var vmExpected VM
	err = json.Unmarshal([]byte(response), &vmExpected)
	assert.Equal(t, vmExpected, *vm, "Bad vm")
}

func TestReadVM(t *testing.T) {
	response := fmt.Sprintf(exampleVMResponse, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, response)
	}

	vm, err := skytap.VMs.Get(context.Background(), "123", "456")
	assert.Nil(t, err, "Bad API method")

	var vmExpected VM
	err = json.Unmarshal([]byte(response), &vmExpected)
	assert.Equal(t, vmExpected, *vm, "Bad VM")
}

func TestUpdateVM(t *testing.T) {
	response := fmt.Sprintf(exampleVMResponse, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var vm VM
	json.Unmarshal([]byte(response), &vm)
	*vm.Name = "updated vm"

	bytes, err := json.Marshal(&vm)
	assert.Nil(t, err, "Bad vm")

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms/456", req.URL.Path, "Bad path")
		assert.Equal(t, "PUT", req.Method, "Bad method")

		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err, "Bad request body")
		assert.JSONEq(t, `{"name": "updated vm"}`, string(body), "Bad request body")

		io.WriteString(rw, string(bytes))
	}

	opts := &UpdateVMRequest{
		Name: strToPtr(*vm.Name),
	}
	vmUpdate, err := skytap.VMs.Update(context.Background(), "123", "456", opts)
	assert.Nil(t, err, "Bad API method")

	assert.Equal(t, vm, *vmUpdate, "Bad vm")
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
	response := fmt.Sprintf(exampleVMResponse, 456)

	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/v2/configurations/123/vms", req.URL.Path, "Bad path")
		assert.Equal(t, "GET", req.Method, "Bad method")

		io.WriteString(rw, fmt.Sprintf(`[%+v]`, response))
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
