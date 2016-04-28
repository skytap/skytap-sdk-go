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

import ()

const (
	RunStateStart = "running"
	RunStateStop  = "stopped"
	RunStatePause = "suspended"
	RunStateKill  = "halted"
	RunStateBusy  = "busy"
	RunStateReset = "reset"
)

func isOkStatus(code int) bool {
	codes := map[int]bool{
		200: true,
		201: true,
		204: true,
		401: false,
		404: false,
		409: false,
		422: false,
		423: false,
		429: false,
		500: false,
	}

	return codes[code]
}

func isBusy(code int) bool {
	return code == 423
}
