package skytap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const exampleTemplate = `{
    "id": "1448141",
    "url": "https://cloud.skytap.com/v2/templates/1448141",
    "name": "CentOS 6.10 Desktop Firstboot",
    "errors": [
        "error1"
    ],
    "busy": null,
    "public": true,
    "description": "test template",
    "vm_count": 1,
    "storage": 30720,
    "network_count": 1,
    "created_at": "2018/10/05 00:46:20 +0100",
    "region": "US-West",
    "region_backend": "skytap",
    "svms": 2,
    "last_installed": "2018/10/19 16:19:41 +0100",
    "can_copy": true,
    "can_delete": false,
    "can_share": true,
    "label_count": 0,
    "label_category_count": 0,
    "can_tag": false,
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
    "project_count_for_user": 0,
    "project_count": 1,
    "containers_count": 0,
    "container_hosts_count": 0,
    "vms": [
        {
            "id": "36628546",
            "name": "CentOS 6 Desktop x64",
            "runstate": "stopped",
            "rate_limited": false,
            "hardware": {
                "cpus": 1,
                "supports_multicore": true,
                "cpus_per_socket": 1,
                "ram": 2048,
                "svms": 2,
                "guestOS": "centos-64",
                "max_cpus": 12,
                "min_ram": 256,
                "max_ram": 262144,
                "vnc_keymap": null,
                "uuid": null,
                "disks": [
                    {
                        "id": "disk-19728485-37432655-scsi-0-0",
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
            "error": false,
            "error_details": false,
            "asset_id": null,
            "hardware_version": 11,
            "max_hardware_version": 11,
            "interfaces": [
                {
                    "id": "nic-19728485-37432655-0",
                    "ip": "10.0.0.1",
                    "hostname": "centos6dx64",
                    "mac": "00:50:56:0C:C0:C2",
                    "services_count": 0,
                    "services": [],
                    "public_ips_count": 0,
                    "public_ips": [],
                    "vm_id": "36628546",
                    "vm_name": "CentOS 6 Desktop x64",
                    "status": "Powered off",
                    "network_id": "23263018",
                    "network_name": "Default Network",
                    "network_url": "https://cloud.skytap.com/v2/templates/1448141/networks/23263018",
                    "network_type": "automatic",
                    "network_subnet": "10.0.0.0/24",
                    "nic_type": "vmxnet3",
                    "secondary_ips": []
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
            "created_at": "2018/10/05 00:46:22 +0100",
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
            "template_url": "https://cloud.skytap.com/v2/templates/1448141"
        }
    ],
    "networks": [
        {
            "id": "23263018",
            "url": "https://cloud.skytap.com/v2/configurations/1448141/networks/23263018",
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
    "svms_by_architecture": {
        "x86": 2,
        "power": 0
    }
}`

func TestReadTemplate(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/templates/456" {
			t.Error("Bad path")
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		io.WriteString(rw, exampleTemplate)
	}

	template, err := skytap.Templates.Get(context.Background(), "456")

	assert.Nil(t, err)
	var templateExpected Template

	err = json.Unmarshal([]byte(exampleTemplate), &templateExpected)

	assert.Equal(t, templateExpected, *template)
}

func TestListTemplates(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v2/templates" {
			t.Error("Bad path")
		}
		if req.Method != "GET" {
			t.Error("Bad method")
		}
		io.WriteString(rw, fmt.Sprintf(`[%+v]`, exampleTemplate))
	}

	result, err := skytap.Templates.List(context.Background())

	assert.Nil(t, err)

	var found = false
	for _, template := range result.Value {
		if *template.Description == "test template" {
			found = true
			break
		}
	}

	assert.True(t, found)
}
