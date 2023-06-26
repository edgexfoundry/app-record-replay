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
	"encoding/json"
	"fmt"
	"net/http"

	interfaces2 "github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-record-replay/internal/interfaces"
	"github.com/edgexfoundry/app-record-replay/pkg/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
)

const (
	recordRoute = common.ApiBase + "/record"
	replayRoute = common.ApiBase + "/replay"
	dataRoute   = common.ApiBase + "/data"

	failedRouteMessage = "failed to added %s route for %s method: %v"

	failedRequestJSON              = "Unable to process request JSON"
	failedRecordRequestValidate    = "Record request failed validation: Duration and/or EventLimit must be set"
	failedRecordDurationValidate   = "Record request failed validation: Duration must be > 0 when set"
	failedRecordEventLimitValidate = "Record request failed validation: Event Limit must be > 0 when set"
	failedRecording                = "Recording failed"
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

func (c *httpController) AddRoutes() error {
	if err := c.appSdk.AddRoute(recordRoute, c.startRecording, http.MethodPost); err != nil {
		return fmt.Errorf(failedRouteMessage, recordRoute, http.MethodPost, err)
	}
	if err := c.appSdk.AddRoute(recordRoute, c.recordingStatus, http.MethodGet); err != nil {
		return fmt.Errorf(failedRouteMessage, recordRoute, http.MethodGet, err)
	}
	if err := c.appSdk.AddRoute(recordRoute, c.cancelRecording, http.MethodDelete); err != nil {
		return fmt.Errorf(failedRouteMessage, recordRoute, http.MethodDelete, err)
	}

	if err := c.appSdk.AddRoute(replayRoute, c.startReplay, http.MethodPost); err != nil {
		return fmt.Errorf(failedRouteMessage, replayRoute, http.MethodPost, err)
	}
	if err := c.appSdk.AddRoute(replayRoute, c.replayStatus, http.MethodGet); err != nil {
		return fmt.Errorf(failedRouteMessage, replayRoute, http.MethodGet, err)
	}
	if err := c.appSdk.AddRoute(replayRoute, c.cancelReplay, http.MethodDelete); err != nil {
		return fmt.Errorf(failedRouteMessage, replayRoute, http.MethodDelete, err)
	}

	if err := c.appSdk.AddRoute(dataRoute, c.exportRecordedData, http.MethodGet); err != nil {
		return fmt.Errorf(failedRouteMessage, dataRoute, http.MethodGet, err)
	}
	if err := c.appSdk.AddRoute(dataRoute, c.importRecordedData, http.MethodPost); err != nil {
		return fmt.Errorf(failedRouteMessage, dataRoute, http.MethodPost, err)
	}

	return nil
}

// StartRecording starts a recording session based on the values in the request.
// An error is returned if the request data is incomplete.
func (c *httpController) startRecording(writer http.ResponseWriter, request *http.Request) {
	startRequest := dtos.RecordRequest{}

	if err := json.NewDecoder(request.Body).Decode(&startRequest); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(fmt.Sprintf("%s: %v", failedRequestJSON, err)))
		return
	}

	if startRequest.Duration == 0 && startRequest.EventLimit == 0 {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(failedRecordRequestValidate))
		return
	}

	if startRequest.Duration < 0 {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(failedRecordDurationValidate))
		return
	}

	if startRequest.EventLimit < 0 {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(failedRecordEventLimitValidate))
		return
	}

	if err := c.dataManager.StartRecording(startRequest); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(fmt.Sprintf("%s: %v", failedRecording, err)))
		return
	}

	writer.WriteHeader(http.StatusAccepted)
}

// CancelRecording cancels the current recording session
func (c *httpController) cancelRecording(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	writer.WriteHeader(http.StatusNotImplemented)
}

// recordingStatus returns the status of the current recording session
func (c *httpController) recordingStatus(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	writer.WriteHeader(http.StatusNotImplemented)
}

// startReplay starts a replay session based on the values in the request
// An error is returned if the request data is incomplete or a record or replay session is currently running.
func (c *httpController) startReplay(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	writer.WriteHeader(http.StatusNotImplemented)
}

// cancelReplay cancels the current replay session
func (c *httpController) cancelReplay(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	writer.WriteHeader(http.StatusNotImplemented)
}

// replayStatus returns the status of the current replay session
func (c *httpController) replayStatus(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	writer.WriteHeader(http.StatusNotImplemented)
}

// exportRecordedData returns the data for the last record session
// An error is returned if the no record session was run or a record session is currently running
func (c *httpController) exportRecordedData(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	writer.WriteHeader(http.StatusNotImplemented)
}

// importRecordedData imports data from a previously exported record session.
// An error is returned if a record or replay session is currently running or the data is incomplete
func (c *httpController) importRecordedData(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me using TDD
	writer.WriteHeader(http.StatusNotImplemented)
}
