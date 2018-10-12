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

const EXAMPLE_ENVIRONMENT = `{
    "id": "456",
    "url": "https://cloud.skytap.com/v2/configurations/456",
    "name": "No VM",
    "description": "test environment",
    "errors": [
        "error1"
    ],
    "error_details": [
        "error1 details"
    ],
    "runstate": "stopped",
    "rate_limited": false,
    "last_run": "2018/10/11 15:42:23 +0100",
    "suspend_on_idle": 1,
    "suspend_at_time": "2018/10/11 15:42:23 +0100",
    "owner_url": "https://cloud.skytap.com/v2/users/1",
    "owner_name": "Joe Bloggs",
    "owner_id": "1",
    "vm_count": 1,
    "storage": 30720,
    "network_count": 1,
    "created_at": "2018/10/11 15:42:23 +0100",
    "region": "US-West",
    "region_backend": "skytap",
    "svms": 1,
    "can_save_as_template": true,
    "can_copy": true,
    "can_delete": true,
    "can_change_state": true,
    "can_share": true,
    "can_edit": true,
    "label_count": 1,
    "label_category_count": 1,
    "can_tag": true,
    "tags": [
        {
            "id": "43894",
            "value": "tag1"
        },
        {
            "id": "43896",
            "value": "tag2"
        }
    ],
    "tag_list": "tag1,tag2",
    "alerts": [
        "alert1"
    ],
    "published_service_count": 0,
    "public_ip_count": 0,
    "auto_suspend_description": null,
    "stages": [
        {
            "delay_after_finish_seconds": 300,
            "index": 0,
            "vm_ids": [
                "123456",
                "123457"
            ]
        }
    ],
    "staged_execution": {
        "action_type": "suspend",
        "current_stage_delay_after_finish_seconds": 300,
        "current_stage_index": 1,
        "current_stage_finished_at": "2018/10/11 15:42:23 +0100",
        "vm_ids": [
            "123453",
            "123454"
        ]
    },
    "sequencing_enabled": false,
    "note_count": 1,
    "project_count_for_user": 0,
    "project_count": 0,
    "publish_set_count": 0,
    "schedule_count": 0,
    "vpn_count": 0,
    "outbound_traffic": false,
    "vms": [
        {
            "id": "36858580",
            "name": "CentOS 7 Server x64",
            "runstate": "stopped",
            "rate_limited": false,
            "hardware": {
                "cpus": 1,
                "supports_multicore": true,
                "cpus_per_socket": 1,
                "ram": 1024,
                "svms": 1,
                "guestOS": "centos-64",
                "max_cpus": 12,
                "min_ram": 256,
                "max_ram": 262144,
                "vnc_keymap": null,
                "uuid": null,
                "disks": [
                    {
                        "id": "disk-19861359-37668995-scsi-0-0",
                        "size": 30720,
                        "type": "SCSI",
                        "controller": "0",
                        "lun": "0"
                    }
                ],
                "storage": 30720,
                "upgradable": false,
                "instance_type": null,
                "time_sync_enabled": true,
                "rtc_start_time": null,
                "copy_paste_enabled": true,
                "nested_virtualization": false,
                "architecture": "x86"
            },
            "error": "vm error",
            "error_details": [
                "vm error details"
            ],
            "asset_id": "1",
            "hardware_version": 11,
            "max_hardware_version": 11,
            "interfaces": [
                {
                    "id": "nic-19861359-37668995-0",
                    "ip": "10.0.0.1",
                    "hostname": "centos7sx64",
                    "mac": "00:50:56:2B:87:F5",
                    "services_count": 0,
                    "services": [
                        {
                            "id": "3389",
                            "internal_port": 3389,
                            "external_ip": "76.191.118.29",
                            "external_port": 12345
                        }
                    ],
                    "public_ips_count": 0,
                    "public_ips":  [
                        {
                            "1.2.3.4": "5.6.7.8"
                        }
                    ],
                    "vm_id": "36858580",
                    "vm_name": "CentOS 7 Server x64",
                    "status": "Powered off",
                    "network_id": "23429874",
                    "network_name": "Default Network",
                    "network_url": "https://cloud.skytap.com/v2/configurations/456/networks/23429874",
                    "network_type": "automatic",
                    "network_subnet": "10.0.0.0/24",
                    "nic_type": "vmxnet3",
                    "secondary_ips": [
                        {
                            "id": "10.0.2.2",
                            "address": "10.0.2.2"
                        }
                    ],
                    "public_ip_attachments": [
                        {
                            "id": 1,
                            "public_ip_attachment_key": 2,
                            "address": "1.2.3.4",
                            "connect_type": 1,
                            "hostname": "host1",
                            "dns_name": "host.com",
                            "public_ip_key": "5.6.7.8"
                        }
                    ]
                }
            ],
            "notes": [
                {
                    "id": "5377708",
                    "user_id": 1,
                    "user": {
                        "id": "1",
                        "url": "https://cloud.skytap.com/v2/users/1",
                        "first_name": "Joe",
                        "last_name": "Bloggs",
                        "login_name": "Joe.Bloggs@opencredo",
                        "email": "Joe.Bloggs@opencredo.com",
                        "title": "",
                        "deleted": false
                    },
                    "created_at": "2018/10/11 15:27:45 +0100",
                    "updated_at": "2018/10/11 15:27:45 +0100",
                    "text": "a note"
                }
            ],
            "labels": [
                {
                    "id": "43892",
                    "value": "test vm",
                    "label_category": "test multi",
                    "label_category_id": "7704",
                    "label_category_single_value": false
                }
            ],
            "credentials": [
                {
                    "id": "35158632",
                    "text": "user/pass"
                }
            ],
            "desktop_resizable": true,
            "local_mouse_cursor": true,
            "maintenance_lock_engaged": false,
            "region_backend": "skytap",
            "created_at": "2018/10/11 15:42:26 +0100",
            "supports_suspend": true,
            "can_change_object_state": true,
            "containers": [
                {
                    "id": 1122,
                    "cid": "123456789abcdefghijk123456789abcdefghijk123456789abcdefghijk",
                    "name": "nginxtest1",
                    "image": "nginx:latest",
                    "created_at": "2016/06/16 11:58:50 -0700",
                    "last_run": "2016/06/16 11:58:51 -0700",
                    "can_change_state": true,
                    "can_delete": true,
                    "status": "running",
                    "privileged": false,
                    "vm_id": 111000,
                    "vm_name": "Docker VM1",
                    "vm_runstate": "running",
                    "configuration_id": 123456
                }
            ],
            "configuration_url": "https://cloud.skytap.com/v2/configurations/456"
        }
    ],
    "networks": [
        {
            "id": "1234567",
            "url": "https://cloud.skytap.com/configurations/1111111/networks/123467",
            "name": "Network 1",
            "network_type": "automatic",
            "subnet": "10.0.0.0/24",
            "subnet_addr": "10.0.0.0",
            "subnet_size": 24,
            "gateway": "10.0.0.254",
            "primary_nameserver": "8.8.8.8",
            "secondary_nameserver": "8.8.8.9",
            "region": "US-West",
            "domain": "sampledomain.com",
            "vpn_attachments": [
                {
                    "id": "111111-vpn-1234567",
                    "connected": false,
                    "network": {
                        "id": "1111111",
                        "subnet": "10.0.0.0/24",
                        "network_name": "Network 1",
                        "configuration_id": "1212121"
                    },
                    "vpn": {
                        "id": "vpn-1234567",
                        "name": "CorpNet",
                        "enabled": true,
                        "nat_enabled": true,
                        "remote_subnets": "10.10.0.0/24, 10.10.1.0/24, 10.10.2.0/24, 10.10.4.0/24",
                        "remote_peer_ip": "199.199.199.199",
                        "can_reconnect": true
                    }
                },
                {
                    "id": "111111-vpn-1234555",
                    "connected": false,
                    "network": {
                        "id": "1111111",
                        "subnet": "10.0.0.0/24",
                        "network_name": "Network 1",
                        "configuration_id": "1212121"
                    },
                    "vpn": {
                        "id": "vpn-1234555",
                        "name": "Offsite DC",
                        "enabled": true,
                        "nat_enabled": true,
                        "remote_subnets": "10.10.0.0/24, 10.10.1.0/24, 10.10.2.0/24, 10.10.4.0/24",
                        "remote_peer_ip": "188.188.188.188",
                        "can_reconnect": true
                    }
                }
            ],
            "tunnelable": false,
            "tunnels": [
                {
                    "id": "tunnel-123456-789011",
                    "status": "not_busy",
                    "error": null,
                    "source_network": {
                        "id": "000000",
                        "url": "https://cloud.skytap.com/configurations/249424/networks/0000000",
                        "name": "Network 1",
                        "network_type": "automatic",
                        "subnet": "10.0.0.0/24",
                        "subnet_addr": "10.0.0.0",
                        "subnet_size": 24,
                        "gateway": "10.0.0.254",
                        "primary_nameserver": null,
                        "secondary_nameserver": null,
                        "region": "US-West",
                        "domain": "skytap.example",
                        "vpn_attachments": []
                    },
                    "target_network": {
                        "id": "111111",
                        "url": "https://cloud.skytap.com/configurations/808216/networks/111111",
                        "name": "Network 2",
                        "network_type": "automatic",
                        "subnet": "10.0.2.0/24",
                        "subnet_addr": "10.0.2.0",
                        "subnet_size": 24,
                        "gateway": "10.0.2.254",
                        "primary_nameserver": null,
                        "secondary_nameserver": null,
                        "region": "US-West",
                        "domain": "test.net",
                        "vpn_attachments": []
                    }
                }
            ]
        }
    ],
    "containers_count": 0,
    "container_hosts_count": 0,
    "platform_errors": [
        "platform error1"
    ],
    "svms_by_architecture": {
        "x86": 1,
        "power": 0
    },
    "all_vms_support_suspend": true,
    "shutdown_on_idle": null,
    "shutdown_at_time": null,
    "auto_shutdown_description": "Shutting down!"
}`

func TestCreateEnvironment(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var createPhase = true

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if createPhase {
			if req.URL.Path != "/configurations" {
				t.Error("Bad path")
			}
			if req.Method != "POST" {
				t.Error("Bad method")
			}
			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err)
			assert.JSONEq(t, fmt.Sprintf(`{"template_id":%q, "description":"test environment"}`, "12345"), string(body))
			io.WriteString(rw, `{"id": "456"}`)
			createPhase = false
		} else {
			if req.URL.Path != "/v2/configurations/456" {
				t.Error("Bad path")
			}
			if req.Method != "PUT" {
				t.Error("Bad method")
			}
			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err)
			assert.JSONEq(t, `{"description": "test environment"}`, string(body))

			io.WriteString(rw, EXAMPLE_ENVIRONMENT)
		}
	}

	opts := &CreateEnvironmentRequest{
		TemplateId:  StringPtr("12345"),
		Description: StringPtr("test environment"),
	}

	environment, err := skytap.Environments.Create(context.Background(), opts)

	assert.Nil(t, err)

	var environmentExpected Environment

	err = json.Unmarshal([]byte(EXAMPLE_ENVIRONMENT), &environmentExpected)

	assert.Equal(t, environmentExpected, *environment)
}

func TestReadEnvironment(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/configurations/456" {
			t.Error("Bad path")
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		io.WriteString(rw, EXAMPLE_ENVIRONMENT)
	}

	environment, err := skytap.Environments.Get(context.Background(), "456")

	assert.Nil(t, err)
	var environmentExpected Environment

	err = json.Unmarshal([]byte(EXAMPLE_ENVIRONMENT), &environmentExpected)

	assert.Equal(t, environmentExpected, *environment)
}

func TestUpdateEnvironment(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var environment Environment
	json.Unmarshal([]byte(EXAMPLE_ENVIRONMENT), &environment)
	*environment.Description = "updated environment"

	bytes, err := json.Marshal(&environment)
	assert.Nil(t, err)

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/configurations/456" {
			t.Error("Bad path")
		}
		if req.Method != "PUT" {
			t.Error("Bad method")
		}
		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		assert.JSONEq(t, `{"description": "updated environment"}`, string(body))

		io.WriteString(rw, string(bytes))
	}

	opts := &UpdateEnvironmentRequest{
		Description: StringPtr(*environment.Description),
	}

	environmentUpdate, err := skytap.Environments.Update(context.Background(), "456", opts)

	assert.Nil(t, err)
	assert.Equal(t, environment, *environmentUpdate)
}

func TestDeleteEnvironment(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/configurations/456" {
			t.Error("Bad path")
		}
		if req.Method != "DELETE" {
			t.Error("Bad method")
		}
	}

	err := skytap.Environments.Delete(context.Background(), "456")
	assert.Nil(t, err)
}

func TestListEnvironments(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/configurations" {
			t.Error("Bad path")
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		io.WriteString(rw, fmt.Sprintf(`[%+v]`, EXAMPLE_ENVIRONMENT))
	}

	result, err := skytap.Environments.List(context.Background())

	assert.Nil(t, err)

	var found = false
	for _, environment := range result.Value {
		if *environment.Description == "test environment" {
			found = true
			break
		}
	}

	assert.True(t, found)
}
