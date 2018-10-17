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
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Username   string `json:"username"`
	ApiKey     string `json:"apiKey"`
	TemplateId string `json:"templateId"`
	VmId       string `json:"vmId"`
	VpnId      string `json:"vpnId"`
}

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func skytapClient(t *testing.T) SkytapClient {
	c := getTestConfig(t)
	fmt.Printf("c: %s, user: %s", c, c.Username)
	client := &http.Client{}
	return SkytapClient{
		HttpClient:  client,
		Credentials: SkytapCredentials{Username: c.Username, ApiKey: c.ApiKey},
	}
}

func getTestConfig(t *testing.T) *testConfig {
	configFile, err := os.Open("testdata/config.json")
	require.NoError(t, err, "Error reading config.json")

	jsonParser := json.NewDecoder(configFile)
	c := &testConfig{}
	err = jsonParser.Decode(c)
	require.NoError(t, err, "Error parsing config.json")
	return c
}

func getMockServer(client SkytapClient) *httptest.Server {
	return getMockServerForString(client, "")
}

func getMockServerForString(client SkytapClient, content string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(content))
	}))
	// Divert all API requests to this server
	baseUrlOveride = server.URL
	client.HttpClient = server.Client()
	return server
}

func readJson(t *testing.T, filename string) string {
	str, err := ioutil.ReadFile(filename)
	require.NoError(t, err, "Error reading "+filename)
	return string(str)
}
