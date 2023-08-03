//
// Copyright (c) 2023 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package dtos

import "time"

// ReplayRequest DTO specifies the replay parameters to start a replay session
type ReplayRequest struct {
	// ReplayRate is the rate at which to replay the data compared to the rate the data was recorded.
	// Values must be greater than 0 where 1 is the same rate, less than 1 is slower rate and greater than 1 is
	// faster rate than the rate the data was recorded.
	ReplayRate float32 `json:"replayRate"`

	// RepeatCount is the count of number of times to repeat the replay. Optional, defaults to 1 if value is less than 1.
	RepeatCount int `json:"repeatCount"`
}

// ReplayStatus DTO contains the data describing the status of a replay session
type ReplayStatus struct {
	// Running indicates if the Replay is currently running or not
	Running bool `json:"running"`
	// EventCount is the number of Events replayed
	EventCount int `json:"eventCount"`
	// Duration is the time the replay has been running or ran.
	Duration time.Duration `json:"duration"`
	// RepeatCount is the number of times the replay of the recorded data has been completed.
	RepeatCount int `json:"repeatCount"`
	// ErrorMessage contains the error message if an error occurred during replay
	ErrorMessage string
}
