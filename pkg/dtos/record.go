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

import (
	"time"

	coreDtos "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
)

// RecordRequest DTO specifies the record parameters to start a recording session
type RecordRequest struct {
	// Duration is the amount of time to record. Required if EventLimit is 0.
	Duration time.Duration `json:"duration"`
	// EventLimit is the maximum number of Events to record. Required if Duration is 0.
	EventLimit int `json:"eventLimit"`

	// IncludeDeviceProfiles is a list of Device Profile names to Filter For.
	IncludeDeviceProfiles []string `json:"includeDeviceProfiles"`
	// IncludeDevices is a list of Device names to Filter For.
	IncludeDevices []string `json:"includeDevices"`
	// IncludeSources is a list of Source names to Filter For.
	IncludeSources []string `json:"includeSources"`

	// ExcludeDeviceProfiles is a list of Device Profile names to Filter Out.
	ExcludeDeviceProfiles []string `json:"excludeDeviceProfiles"`
	// ExcludeDevices is a list of Device names to Filter Out.
	ExcludeDevices []string `json:"excludeDevices"`
	// ExcludeSources is a list of Source names to Filter Out.
	ExcludeSources []string `json:"excludeSources"`
}

// RecordStatus DTO contains the data describing the status of a recording session
type RecordStatus struct {
	// InProgress indicates if the recording is currently in progress or not
	InProgress bool `json:"inProgress"`
	// EventCount is the count of Events batched so far (In Progress) or recorded (completed)
	EventCount int `json:"eventCount"`
	// Duration is the amount of time recording so far (In Progress) or recording took (completed)
	Duration time.Duration `json:"duration"`
}

// RecordedData DTO contains the data from a completed or imported recording
type RecordedData struct {
	// RecordedEvents is the list of Events that were recorded
	RecordedEvents []coreDtos.Event `json:"recordedEvents"`
	// Profiles is the list of Device Profiles that recorded Events referenced
	Profiles []coreDtos.DeviceProfile `json:"profiles"`
	// Devices is the list of Devices that that recorded Events referenced
	Devices []coreDtos.Device `json:"devices"`
}
