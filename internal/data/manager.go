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

package data

import (
	"errors"

	appInterfaces "github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-record-replay/internal/interfaces"
	"github.com/edgexfoundry/app-record-replay/pkg/dtos"
	coreDtos "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
)

// dataManager implements interface that records and replays captured data
type dataManager struct {
	dataChan   chan []coreDtos.Event
	appSvc     appInterfaces.ApplicationService
	eventCount int
}

// NewManager is the factory function which instantiates a Data Manager
func NewManager(service appInterfaces.ApplicationService) interfaces.DataManager {
	return &dataManager{
		dataChan: make(chan []coreDtos.Event, 1),
		appSvc:   service,
	}
}

// StartRecording starts a recording session based on the values in the request.
// An error is returned if the request data is incomplete or a record or replay session is currently running.
func (m *dataManager) StartRecording(request *dtos.RecordRequest) error {
	//TODO implement me using TDD

	// TODO: Use new appSrv.AppContext.Done() to exit from long-running recording function

	return errors.New("not implemented")
}

// CancelRecording cancels the current recording session
func (m *dataManager) CancelRecording() error {
	//TODO implement me using TDD
	return errors.New("not implemented")
}

// RecordingStatus returns the status of the current recording session
func (m *dataManager) RecordingStatus() (*dtos.RecordStatus, error) {
	//TODO implement me using TDD
	return nil, errors.New("not implemented")
}

// StartReplay starts a replay session based on the values in the request
// An error is returned if the request data is incomplete or a record or replay session is currently running.
func (m *dataManager) StartReplay(request *dtos.ReplayRequest) error {
	//TODO implement me using TDD
	return errors.New("not implemented")
}

// CancelReplay cancels the current replay session
func (m *dataManager) CancelReplay() error {
	//TODO implement me using TDD
	return errors.New("not implemented")
}

// ReplayStatus returns the status of the current replay session
func (m *dataManager) ReplayStatus() (*dtos.ReplayStatus, error) {
	//TODO implement me using TDD
	return nil, errors.New("not implemented")
}

// ExportRecordedData returns the data for the last record session
// An error is returned if the no record session was run or a record session is currently running
func (m *dataManager) ExportRecordedData() (*dtos.RecordedData, error) {
	//TODO implement me using TDD
	return nil, errors.New("not implemented")
}

// ImportRecordedData imports data from a previously exported record session.
// An error is returned if a record or replay session is currently running or the data is incomplete
func (m *dataManager) ImportRecordedData(data *dtos.RecordedData) error {
	//TODO implement me using TDD
	return errors.New("not implemented")
}

// Pipeline functions

var noDataError = errors.New("CountEvents function received nil data")
var dataNotEvent = errors.New("CountEvents function received data that is not an Event")

// countEvents counts the number of Events recorded so far. Must be called after any filters and before the Batch function.
// This count is used when reporting Recording Status
func (m *dataManager) countEvents(_ appInterfaces.AppFunctionContext, data any) (bool, interface{}) {
	if data == nil {
		return false, noDataError
	}

	if _, ok := data.(coreDtos.Event); !ok {
		return false, dataNotEvent
	}

	m.eventCount++

	return true, data
}

// processBatchedData processes the batched data for the current recording session
func (m *dataManager) processBatchedData(_ appInterfaces.AppFunctionContext, data any) (bool, interface{}) {
	//TODO implement me using TDD
	return false, nil
}
