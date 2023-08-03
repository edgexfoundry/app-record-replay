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

package interfaces

import "github.com/edgexfoundry/app-record-replay/pkg/dtos"

// DataManager defines the interface for implementations that records and replays captured data
type DataManager interface {
	// StartRecording starts a recording session based on the values in the request.
	// An error is returned if the request data is incomplete or a record or replay session is currently running.
	StartRecording(request dtos.RecordRequest) error
	// CancelRecording cancels the current recording session
	CancelRecording() error
	// RecordingStatus returns the status of the current recording session
	RecordingStatus() dtos.RecordStatus
	// StartReplay starts a replay session based on the values in the request
	// An error is returned if the request data is incomplete or a record or replay session is currently running.
	StartReplay(request dtos.ReplayRequest) error
	// CancelReplay cancels the current replay session
	CancelReplay() error
	// ReplayStatus returns the status of the current replay session
	ReplayStatus() dtos.ReplayStatus
	// ExportRecordedData returns the data for the last record session
	// An error is returned if the no record session was run or a record session is currently running
	ExportRecordedData() (*dtos.RecordedData, error)
	// ImportRecordedData imports data from a previously exported record session.
	// An error is returned if a record or replay session is currently running or the data is incomplete
	ImportRecordedData(data *dtos.RecordedData) error
}
