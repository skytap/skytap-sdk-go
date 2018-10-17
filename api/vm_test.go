package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	hardware := Hardware{&cpus, &persock, nil, nil}

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

func TestAddNetworkInterface(t *testing.T) {
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, vmJson)
	defer server.Close()

	vm, err := GetVirtualMachine(client, "1001")
	require.NoError(t, err, "Error creating vm")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations/1/vms/1001/interfaces.json", r.URL.Path)
		fmt.Fprintln(w, vmJson)
	})

	nic, err := vm.AddNetworkInterface(client, "1", "10.0.0.1", "host-1", "vxnet3", false)
	require.NoError(t, err, "Error adding network interface")
	require.Equal(t, "10.0.0.1", nic.Ip)
	require.Equal(t, "host-1", nic.Hostname)
	require.Equal(t, "vxnet3", nic.NicType)
}
func TestDeleteNetworkInterface(t *testing.T) {
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, vmJson)
	defer server.Close()

	vm, err := GetVirtualMachine(client, "1001")
	require.NoError(t, err, "Error creating vm")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "DELETE", r.Method)
		require.Equal(t, "/configurations/1/vms/1001/interfaces/nic-5971736-13548234-0.json", r.URL.Path)
	})

	err = vm.RemoveNetworkInterface(client, "1", "nic-5971736-13548234-0")
	require.NoError(t, err, "Error removing network interface")

}

func TestAddDisk(t *testing.T) {

}
