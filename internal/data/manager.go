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
	"context"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-record-replay/pkg/dtos"
	dtos2 "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
)

// CreateDefaultDataManger is the factory function which instantiates the default Data Manager
func CreateDefaultDataManger(ctx context.Context, service interfaces.ApplicationService) RecordManager {
	return &defaultDataManager{
		dataChan: make(chan []dtos2.Event, 1),
		ctx:      ctx,
		appSvc:   service,
	}
}

// defaultDataManager implements interface that records and replays captured data
type defaultDataManager struct {
	dataChan chan []dtos2.Event
	ctx      context.Context
	appSvc   interfaces.ApplicationService
}

// StartRecording starts a recording session based on the values in the request.
// An error is returned if the request data is incomplete or a record or replay session is currently running.
func (m *defaultDataManager) StartRecording(request dtos.RecordRequest) error {
	//TODO implement me using TDD
	panic("implement me")
}

// CancelRecording cancels the current recording session
func (m *defaultDataManager) CancelRecording() {
	//TODO implement me using TDD
	panic("implement me")
}

// RecordingStatus returns the status of the current recording session
func (m *defaultDataManager) RecordingStatus() dtos.RecordStatus {
	//TODO implement me using TDD
	panic("implement me")
}

// StartReplay starts a replay session based on the values in the request
// An error is returned if the request data is incomplete or a record or replay session is currently running.
func (m *defaultDataManager) StartReplay(request dtos.ReplayRequest) error {
	//TODO implement me using TDD
	panic("implement me")
}

// CancelReplay cancels the current replay session
func (m *defaultDataManager) CancelReplay() {
	//TODO implement me using TDD
	panic("implement me")
}

// ReplayStatus returns the status of the current replay session
func (m *defaultDataManager) ReplayStatus() dtos.ReplayStatus {
	//TODO implement me using TDD
	panic("implement me")
}

// ExportRecordedData returns the data for the last record session
// An error is returned if the no record session was run or a record session is currently running
func (m *defaultDataManager) ExportRecordedData() (dtos.RecordedData, error) {
	//TODO implement me using TDD
	panic("implement me")
}

// ImportRecordedData imports data from a previously exported record session.
// An error is returned if a record or replay session is currently running or the data is incomplete
func (m *defaultDataManager) ImportRecordedData(data dtos.RecordedData) error {
	//TODO implement me using TDD
	panic("implement me")
}

// Pipeline functions

// countEvents counts the number of Events the function receives.
func (m *defaultDataManager) countEvents(_ interfaces.AppFunctionContext, data any) (bool, interface{}) {
	//TODO implement me using TDD
	return false, nil
}

// processBatchedData processes the batched data for the current recording session
func (m *defaultDataManager) processBatchedData(_ interfaces.AppFunctionContext, data any) (bool, interface{}) {
	//TODO implement me using TDD
	return false, nil
}
