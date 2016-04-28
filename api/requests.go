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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/dghubble/sling"
)

const (
	AcceptHeaderV2 = "application/vnd.skytap.api.v2+json"
	AcceptHeaderV1 = "application/json"
	BaseUriV1      = "https://cloud.skytap.com"
	BaseUriV2      = "https://cloud.skytap.com/v2"

	maxRetries = 6
)

/*
 General skytap json error response.
*/
type SkytapApiError struct {
	Error string `json:error`
}

/*
 Credentials for accessing skytap REST API.
*/
type SkytapCredentials struct {
	Username string
	ApiKey   string
}

/*
 Skytap client object, needed for all REST calls.
*/
type SkytapClient struct {
	HttpClient  *http.Client
	Credentials SkytapCredentials
}

/*
 Create a new client from username and key.
*/
func NewSkytapClient(username string, apiKey string) *SkytapClient {
	return NewSkytapClientFromCredentials(SkytapCredentials{username, apiKey})
}

/*
 Create a new client from credentials
*/
func NewSkytapClientFromCredentials(credentials SkytapCredentials) *SkytapClient {
	return &SkytapClient{&http.Client{}, credentials}
}

/*
 Request methods use this to create/customize the requests.
*/
type SlingDecorator func(*sling.Sling) *sling.Sling

/*
 Some skytap resources have a runstate in response, use this for monitoring.
*/
type RunstateBody struct {
	Runstate string `json:"runstate"`
}

/*
 A runstate aware resource has a runstate in its representation, which can be used when waiting for a specific state.
*/
type RunstateAwareResource interface {
	//
	RunstateStr() string
	// Should fetch a fresh representation of the resource and return the current runstate, or error
	Refresh(client SkytapClient) (RunstateAwareResource, error)
}

/*
 Wait until the given resource is in one of the desired states.

 If the resource reaches the desired state, a recently fetched representation is returned. Otherwise an error is returned, along
 with the result of the last attempt.

 If requireStateChange is set, a transition must occur. The function will wait until the state changes or timeout.
*/
func WaitUntilInState(client SkytapClient, desiredStates []string, r RunstateAwareResource, requireStateChange bool) (RunstateAwareResource, error) {
	log.WithFields(log.Fields{"desiredStates": desiredStates, "resource": r}).Info("Waiting until resource is in desired state")
	start := time.Now()

	current, err := r.Refresh(client)
	if err != nil {
		return current, err
	}

	hasChanged := !requireStateChange || current.RunstateStr() != r.RunstateStr()

	maxBusyWaitPeriods := 20
	waitPeriod := 10 * time.Second
	for i := 0; i < maxBusyWaitPeriods && !(hasChanged && stringInSlice(current.RunstateStr(), desiredStates)); i++ {
		time.Sleep(waitPeriod)
		current, err = r.Refresh(client)
		if err != nil {
			return current, err
		}
		hasChanged = hasChanged || current.RunstateStr() != r.RunstateStr()
	}
	if !stringInSlice(current.RunstateStr(), desiredStates) {
		return current, errors.New(fmt.Sprintf("Didn't achieve any desired runstate in %s after %d seconds, resource is in runstate %s", desiredStates, time.Now().Unix()-start.Unix(), current.RunstateStr()))
	}
	return current, err
}

/*
 Runs an initial skytap API request attempt, with retries.

 Returns the resulting response, or error. If no error occurs, the response json will be present in respJson.

 useV2 - If true the request should use V2 API path.
 respJson - Interface to fill with response JSON.
 slingDecorator - Decorate request with specifics, set request path relative to root, add body, etc.
*/
func RunSkytapRequest(client SkytapClient, useV2 bool, respJson interface{}, slingDecorator SlingDecorator) (*http.Response, error) {
	return runSkytapRequestWithRetry(client, useV2, respJson, slingDecorator, 0)
}

/*
 Return a skytap resource specified as complete GET based URL.
*/
func GetSkytapResource(client SkytapClient, url string, respObj interface{}) (*http.Response, error) {
	fromUrl := func(s *sling.Sling) *sling.Sling {
		return s.New().Base(url)
	}
	return RunSkytapRequest(client, false, respObj, fromUrl)
}

/*
  Runs a skytap API request attempt, retry number as specified by retryNum.

*/
func runSkytapRequestWithRetry(client SkytapClient, useV2 bool, respObj interface{}, slingDecorator SlingDecorator, retryNum int) (*http.Response, error) {
	baseUrl := BaseUriV1
	if useV2 {
		baseUrl = BaseUriV2
	}

	base := sling.New().Base(baseUrl + "/").Client(client.HttpClient)
	s := slingDecorator(base)
	skytapError := &SkytapApiError{}
	req, err := s.Request()
	req.SetBasicAuth(client.Credentials.Username, client.Credentials.ApiKey)
	acceptHeader := AcceptHeaderV1
	if useV2 {
		acceptHeader = AcceptHeaderV2
	}
	req.Header.Set("Accept", acceptHeader)
	var resp *http.Response
	resp, err = s.Do(req, respObj, skytapError)

	returnError := err
	logRequestResponse(req, resp, respObj, returnError)

	if !isOkStatus(resp.StatusCode) {
		if isBusy(resp.StatusCode) {
			retrySecs := 10
			after := resp.Header.Get("Retry-After")
			if after != "" {
				parsedSecs, parseErr := strconv.ParseInt(after, 10, 32)
				if parseErr != nil {
					log.Warnf("Couldn't parse Retry-After (%s)", after)
				} else {
					retrySecs = int(parsedSecs)
				}
			} else if retryNum <= maxRetries {
				log.WithFields(log.Fields{
					"method":         req.Method,
					"url":            req.URL,
					"retryNum":       retryNum,
					"retryAfterSecs": retrySecs,
				}).Info("Got resource busy response, retrying")
				time.Sleep(time.Duration(retrySecs) * time.Second)
				runSkytapRequestWithRetry(client, useV2, respObj, slingDecorator, retryNum+1)
			} else {
				log.WithFields(log.Fields{"url": req.URL, "maxRetries": maxRetries, "error": err}).Error("Maximum retries reached")
				returnError = errors.New(fmt.Sprintf("Maximum retries (%d) reached calling %s(%s), resource is still busy", maxRetries, req.Method, req.URL))
			}
		} else {
			if skytapError.Error != "" {
				returnError = errors.New(skytapError.Error)
			} else if err == nil {
				returnError = fmt.Errorf("Received error status code calling SkyTap API, but no additional error info: %s", resp.Status)
				logRequestResponse(req, resp, respObj, returnError)
			}
		}
	}
	return resp, returnError
}

func logRequestResponse(req *http.Request, resp *http.Response, respObj interface{}, err error) {

	jsonStr, err := json.Marshal(respObj)
	if err != nil {
		log.Errorf("Couldn't marshall response: %s", err)
	}

	entry := log.WithFields(log.Fields{
		"method":         req.Method,
		"url":            req.URL,
		"status":         resp.Status,
		"marshallError":  err,
		"responseObject": string(jsonStr),
	})

	if err != nil {
		entry.Error("Request caused error")
	} else {
		entry.Debug("Made request")
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
