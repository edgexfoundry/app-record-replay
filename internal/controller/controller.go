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

package controller

import (
	"net/http"

	interfaces2 "github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-record-replay/internal/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

type httpController struct {
	lc          logger.LoggingClient
	dataManager interfaces.DataManager
	appSdk      interfaces2.ApplicationService
}

// New is the factory function which instantiates a new HTTP Controller
func New(dataManager interfaces.DataManager, appSdk interfaces2.ApplicationService, lc logger.LoggingClient) interfaces.HttpController {
	return &httpController{
		lc:          lc,
		dataManager: dataManager,
		appSdk:      appSdk,
	}
}

func (c *httpController) AddRoutes() {
	//TODO implement me using TDD
	panic("implement me")
}

// StartRecording starts a recording session based on the values in the request.
// An error is returned if the request data is incomplete or a record or replay session is currently running.
func (c *httpController) StartRecording(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	panic("implement me")
}

// CancelRecording cancels the current recording session
func (c *httpController) CancelRecording(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	panic("implement me")
}

// RecordingStatus returns the status of the current recording session
func (c *httpController) RecordingStatus(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	panic("implement me")
}

// StartReplay starts a replay session based on the values in the request
// An error is returned if the request data is incomplete or a record or replay session is currently running.
func (c *httpController) StartReplay(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	panic("implement me")
}

// CancelReplay cancels the current replay session
func (c *httpController) CancelReplay(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	panic("implement me")
}

// ReplayStatus returns the status of the current replay session
func (c *httpController) ReplayStatus(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	panic("implement me")
}

// ExportRecordedData returns the data for the last record session
// An error is returned if the no record session was run or a record session is currently running
func (c *httpController) ExportRecordedData(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	panic("implement me")
}

// ImportRecordedData imports data from a previously exported record session.
// An error is returned if a record or replay session is currently running or the data is incomplete
func (c *httpController) ImportRecordedData(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	panic("implement me")
}
