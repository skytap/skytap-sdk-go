package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateNewEnvironment(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")

	client := skytapClient(t)
	server := getMockServer(client)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"template_id":"2"}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, envJson)
	})

	env, err := CreateNewEnvironment(client, "2")
	require.NoError(t, err, "Error creating environment")
	require.Equal(t, "Environment 1", env.Name)
}

func TestCreateNewEnvironmentWithVms(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")

	client := skytapClient(t)
	server := getMockServer(client)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"template_id":"2","vm_ids":["1002"]}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, envJson)
	})

	env, err := CreateNewEnvironmentWithVms(client, "2", []string{"1002"})
	require.NoError(t, err, "Error creating environment")
	require.Equal(t, "Environment 1", env.Name)
}

func TestCopyEnvironmentWithVms(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")

	client := skytapClient(t)
	server := getMockServer(client)
	defer server.Close()

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "/configurations.json", r.URL.Path)
		body, _ := ioutil.ReadAll(r.Body)
		require.Equal(t, `{"configuration_id":"1","vm_ids":["1001"]}`, strings.TrimSpace(string(body)))
		fmt.Fprintln(w, envJson)
	})

	env, err := CopyEnvironmentWithVms(client, "1", []string{"1001"})
	require.NoError(t, err, "Error creating environment")
	require.Equal(t, "Environment 1", env.Name)
}

func TestAddVirtualMachineFromTemplate(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")
	templJson := readJson(t, "testdata/template-2.json")
	templateVmJson := readJson(t, "testdata/vm-1002.json")

	client := skytapClient(t)
	server := getMockServerForString(client, envJson)
	defer server.Close()

	env, err := GetEnvironment(client, "1")
	require.NoError(t, err, "Error getting environment")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpstr := ""
		if r.Method == "GET" {
			// handle requests for vm and its template
			if strings.Contains(r.URL.Path, "/vms") {
				// replace any hostnames in the vm json
				tmpstr = strings.Replace(templateVmJson, "https://cloud.skytap.com", server.URL, -1)
			} else {
				tmpstr = templJson
			}
			fmt.Fprintln(w, tmpstr)
		} else if r.Method == "PUT" {
			require.Equal(t, "/configurations/1.json", r.URL.Path)
			body, _ := ioutil.ReadAll(r.Body)
			require.Equal(t, `{"template_id":"2","vm_ids":["1002"]}`, strings.TrimSpace(string(body)))
			fmt.Fprintln(w, envJson)
		}
	})

	env, err = env.AddVirtualMachine(client, "1002")
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, "Environment 1", env.Name)
}

func TestAddVirtualMachineFromEnvironment(t *testing.T) {
	envJson := readJson(t, "testdata/environment-1.json")
	vmJson := readJson(t, "testdata/vm-1001.json")

	client := skytapClient(t)
	server := getMockServerForString(client, envJson)
	defer server.Close()

	env, err := GetEnvironment(client, "1")
	require.NoError(t, err, "Error getting environment")

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpstr := ""
		if r.Method == "GET" {
			// handle requests for vm and its environment
			if strings.Contains(r.URL.Path, "/vms") {
				// replace any hostnames in the vm json
				tmpstr = strings.Replace(vmJson, "https://cloud.skytap.com", server.URL, -1)
			} else {
				tmpstr = envJson
			}
			fmt.Fprintln(w, tmpstr)
		} else if r.Method == "PUT" {
			require.Equal(t, "/configurations/1.json", r.URL.Path)
			body, _ := ioutil.ReadAll(r.Body)
			require.Equal(t, `{"merge_configuration":"1","vm_ids":["1001"]}`, strings.TrimSpace(string(body)))
			fmt.Fprintln(w, envJson)
		}
	})

	env, err = env.AddVirtualMachine(client, "1001")
	require.NoError(t, err, "Error adding vm from template")
	require.Equal(t, "Environment 1", env.Name)
}
