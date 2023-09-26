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
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strconv"

	appInterfaces "github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
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
	failedReplayRateValidate       = "Replay request failed validation: Replay Rate must be greater than 0"
	failedRepeatCountValidate      = "Replay request failed validation: Repeat Count must be equal or greater than 0"
	failedReplay                   = "Replay failed"
	failedDataCompression          = "failed to compress recorded data of type"
	failedToUncompressData         = "failed to uncompress data"
	failedImportingData            = "Import data failed"
	noDataFound                    = "no recorded data found"

	noCompression       = ""
	zlibCompression     = "zlib"
	gzipCompression     = "gzip"
	contentEncodingGzip = "gzip"
	contentEncodingZlib = "deflate" // standard value used for zlib is deflate
)

type httpController struct {
	lc          logger.LoggingClient
	dataManager interfaces.DataManager
	appSdk      appInterfaces.ApplicationService
}

// New is the factory function which instantiates a new HTTP Controller
func New(dataManager interfaces.DataManager, appSdk appInterfaces.ApplicationService) interfaces.HttpController {
	return &httpController{
		lc:          appSdk.LoggingClient(),
		dataManager: dataManager,
		appSdk:      appSdk,
	}
}

func (c *httpController) AddRoutes() error {

	if err := c.appSdk.AddCustomRoute(recordRoute, false, c.startRecording, http.MethodPost); err != nil {
		return fmt.Errorf(failedRouteMessage, recordRoute, http.MethodPost, err)
	}
	if err := c.appSdk.AddCustomRoute(recordRoute, false, c.recordingStatus, http.MethodGet); err != nil {
		return fmt.Errorf(failedRouteMessage, recordRoute, http.MethodGet, err)
	}
	if err := c.appSdk.AddCustomRoute(recordRoute, false, c.cancelRecording, http.MethodDelete); err != nil {
		return fmt.Errorf(failedRouteMessage, recordRoute, http.MethodDelete, err)
	}

	if err := c.appSdk.AddCustomRoute(replayRoute, false, c.startReplay, http.MethodPost); err != nil {
		return fmt.Errorf(failedRouteMessage, replayRoute, http.MethodPost, err)
	}
	if err := c.appSdk.AddCustomRoute(replayRoute, false, c.replayStatus, http.MethodGet); err != nil {
		return fmt.Errorf(failedRouteMessage, replayRoute, http.MethodGet, err)
	}
	if err := c.appSdk.AddCustomRoute(replayRoute, false, c.cancelReplay, http.MethodDelete); err != nil {
		return fmt.Errorf(failedRouteMessage, replayRoute, http.MethodDelete, err)
	}

	if err := c.appSdk.AddCustomRoute(dataRoute, false, c.exportRecordedData, http.MethodGet); err != nil {
		return fmt.Errorf(failedRouteMessage, dataRoute, http.MethodGet, err)
	}
	if err := c.appSdk.AddCustomRoute(dataRoute, false, c.importRecordedData, http.MethodPost); err != nil {
		return fmt.Errorf(failedRouteMessage, dataRoute, http.MethodPost, err)
	}

	c.lc.Info("Add Record & Replay routes")

	return nil
}

// StartRecording starts a recording session based on the values in the ctx.Request().
// An error is returned if the request data is incomplete.
func (c *httpController) startRecording(ctx echo.Context) error {
	startRequest := &dtos.RecordRequest{}

	if err := json.NewDecoder(ctx.Request().Body).Decode(startRequest); err != nil {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s: %v", failedRequestJSON, err))
	}

	if startRequest.Duration == 0 && startRequest.EventLimit == 0 {
		return ctx.String(http.StatusBadRequest, failedRecordRequestValidate)
	}

	if startRequest.Duration < 0 {
		return ctx.String(http.StatusBadRequest, failedRecordDurationValidate)
	}

	if startRequest.EventLimit < 0 {
		return ctx.String(http.StatusBadRequest, failedRecordEventLimitValidate)
	}

	if err := c.dataManager.StartRecording(*startRequest); err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("%s: %v", failedRecording, err))
	}

	ctx.Response().WriteHeader(http.StatusAccepted)
	return nil
}

// CancelRecording cancels the current recording session
func (c *httpController) cancelRecording(ctx echo.Context) error {
	if err := c.dataManager.CancelRecording(); err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to cancel recording: %v", err))
	}

	ctx.Response().WriteHeader(http.StatusAccepted)
	return nil
}

// recordingStatus return nils the status of the current recording session
func (c *httpController) recordingStatus(ctx echo.Context) error {
	recordingStatus := c.dataManager.RecordingStatus()

	jsonResponse, err := json.Marshal(recordingStatus)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to marshal recording status: %s", err))
	}

	ctx.Response().WriteHeader(http.StatusOK)
	_, _ = ctx.Response().Write(jsonResponse)
	return nil
}

// startReplay starts a replay session based on the values in the request
// An error is return niled if the request data is incomplete or a record or replay session is currently running.
func (c *httpController) startReplay(ctx echo.Context) error {
	startRequest := &dtos.ReplayRequest{}

	if err := json.NewDecoder(ctx.Request().Body).Decode(startRequest); err != nil {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s: %v", failedRequestJSON, err))
	}

	if startRequest.ReplayRate <= 0 {
		return ctx.String(http.StatusBadRequest, failedReplayRateValidate)
	}

	if startRequest.RepeatCount < 0 {
		return ctx.String(http.StatusBadRequest, failedRepeatCountValidate)
	}

	if err := c.dataManager.StartReplay(*startRequest); err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("%s: %v", failedReplay, err))
	}

	ctx.Response().WriteHeader(http.StatusAccepted)
	return nil
}

// cancelReplay cancels the current replay session
func (c *httpController) cancelReplay(ctx echo.Context) error {
	if err := c.dataManager.CancelReplay(); err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to cancel replay: %v", err))
	}

	ctx.Response().WriteHeader(http.StatusAccepted)
	return nil
}

// replayStatus return nils the status of the current replay session
func (c *httpController) replayStatus(ctx echo.Context) error {
	replayStatus := c.dataManager.ReplayStatus()

	jsonResponse, err := json.Marshal(replayStatus)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to marshal replay status: %s", err))
	}

	ctx.Response().WriteHeader(http.StatusOK)
	_, _ = ctx.Response().Write(jsonResponse)
	return nil
}

// exportRecordedData return nils the data for the last record session
// An error is return niled if the no record session was run or a record session is currently running
func (c *httpController) exportRecordedData(ctx echo.Context) error {
	recordedData, err := c.dataManager.ExportRecordedData()
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to export recorded data: %v", err))
	}

	compression := ctx.Request().URL.Query().Get("compression")
	switch compression {
	case noCompression:
		c.appSdk.LoggingClient().Debug("ARR Export - Exporting as JSON w/o compression")
		jsonResponse, err := json.Marshal(recordedData)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, "failed to marshal recorded data")
		}
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(http.StatusOK)
		_, _ = ctx.Response().Write(jsonResponse)

	case zlibCompression:
		c.appSdk.LoggingClient().Debug("ARR Export - Exporting as JSON using ZLIB compression")
		ctx.Response().Header().Set("Content-Encoding", contentEncodingZlib)
		ctx.Response().Header().Set("Content-Type", "application/json")
		zlibWriter := zlib.NewWriter(ctx.Response().Writer)
		defer zlibWriter.Close()
		err = json.NewEncoder(zlibWriter).Encode(&recordedData)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, fmt.Sprintf("%s %s: %s", failedDataCompression, zlibCompression, err))
		}

	case gzipCompression:
		c.appSdk.LoggingClient().Debug("ARR Export - Exporting as JSON using GZIP compression")
		ctx.Response().Header().Set("Content-Encoding", contentEncodingGzip)
		ctx.Response().Header().Set("Content-Type", "application/json")
		gZipWriter := gzip.NewWriter(ctx.Response().Writer)
		defer gZipWriter.Close()
		err = json.NewEncoder(gZipWriter).Encode(&recordedData)
		if err != nil {
			return ctx.String(http.StatusInternalServerError, fmt.Sprintf("%s %s: %s", failedDataCompression, gzipCompression, err))
		}

	default:
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("compression format not available: %s", compression))
	}

	return nil
}

// importRecordedData imports data from a previously exported record session.
// An error is return niled if a record or replay session is currently running or the data is incomplete
func (c *httpController) importRecordedData(ctx echo.Context) error {
	importedRecordedData := &dtos.RecordedData{}
	var reader io.ReadCloser
	var err error
	var overWriteProfilesDevices bool

	contentType := ctx.Request().Header.Get(common.ContentType)
	if contentType != common.ContentTypeJSON {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("Invalid content type '%s'. Must be application/json", contentType))
	}

	queryParam := ctx.Request().URL.Query().Get("overwrite")
	if len(queryParam) == 0 {
		overWriteProfilesDevices = true
	} else {
		overWriteProfilesDevices, err = strconv.ParseBool(queryParam)
		if err != nil {
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("failed to parse overwrite parameter: %v", err))
		}
	}

	compression := ctx.Request().Header.Get("Content-Encoding")
	switch compression {

	case noCompression:
		c.appSdk.LoggingClient().Debug("ARR Import - Importing as JSON w/o compression")
		reader = ctx.Request().Body

	case contentEncodingGzip:
		c.appSdk.LoggingClient().Debug("ARR Import - Importing as JSON using GZIP compression")
		reader, err = gzip.NewReader(ctx.Request().Body)
		if err != nil {
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s: %s", failedToUncompressData, err))
		}

	case contentEncodingZlib:
		c.appSdk.LoggingClient().Debug("ARR Import - Importing as JSON using ZLIB compression")
		reader, err = zlib.NewReader(ctx.Request().Body)
		if err != nil {
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s: %s", failedToUncompressData, err))
		}
	default:
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("compression format %s not supported", compression))

	}
	defer reader.Close()
	err = json.NewDecoder(reader).Decode(&importedRecordedData)
	if err != nil {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s: %v", failedRequestJSON, err))
	}

	if len(importedRecordedData.RecordedEvents) < 1 {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s: no recorded events", noDataFound))
	}

	if len(importedRecordedData.Devices) < 1 {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s: no devices", noDataFound))
	}

	if len(importedRecordedData.Profiles) < 1 {
		return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s: no profiles", noDataFound))
	}

	if err := c.dataManager.ImportRecordedData(importedRecordedData, overWriteProfilesDevices); err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("%s: %v", failedImportingData, err))
	}

	ctx.Response().WriteHeader(http.StatusAccepted)
	return nil
}
