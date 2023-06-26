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

package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appMocks "github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/app-record-replay/internal/interfaces/mocks"
	"github.com/edgexfoundry/app-record-replay/pkg/dtos"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	target := New(&mocks.DataManager{}, &appMocks.ApplicationService{}, &loggerMocks.LoggingClient{})
	require.NotNil(t, target)
	c := target.(*httpController)
	require.NotNil(t, c)
	assert.NotNil(t, c.appSdk)
	assert.NotNil(t, c.lc)
	assert.NotNil(t, c.dataManager)
}

func TestHttpController_AddRoutes_Success(t *testing.T) {
	mockSdk := &appMocks.ApplicationService{}
	mockSdk.On("AddRoute", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	target := New(nil, mockSdk, &loggerMocks.LoggingClient{})

	err := target.AddRoutes()
	require.NoError(t, err)
}

func TestHttpController_AddRoutes_Error(t *testing.T) {
	tests := []struct {
		Name   string
		Route  string
		Method string
	}{
		{"Start Recording", recordRoute, http.MethodPost},
		{"Cancel Recording", recordRoute, http.MethodDelete},
		{"Recording Status", recordRoute, http.MethodGet},

		{"Start Replay", replayRoute, http.MethodPost},
		{"Cancel Replay", replayRoute, http.MethodDelete},
		{"Replay Status", replayRoute, http.MethodGet},

		{"Export", dataRoute, http.MethodGet},
		{"Import", dataRoute, http.MethodPost},
	}

	expectedError := errors.New("AddRoute error")
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockSdk := &appMocks.ApplicationService{}
			mockSdk.On("AddRoute", test.Route, mock.Anything, test.Method).Return(expectedError)
			mockSdk.On("AddRoute", mock.Anything, mock.Anything, mock.Anything).Return(nil)

			target := New(nil, mockSdk, &loggerMocks.LoggingClient{})

			err := target.AddRoutes()
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.Route)
			assert.Contains(t, err.Error(), test.Method)
			assert.Contains(t, err.Error(), expectedError.Error())
		})
	}
}

func TestHttpController_StartRecording(t *testing.T) {
	// TODO: Complete Implementation using TDD
	mockDataManager := &mocks.DataManager{}
	mockSdk := &appMocks.ApplicationService{}
	target := New(mockDataManager, mockSdk, &loggerMocks.LoggingClient{}).(*httpController)

	handler := http.HandlerFunc(target.startRecording)

	emptyRequestDTO := dtos.RecordRequest{}
	emptyRequestDTOJson, err := json.Marshal(emptyRequestDTO)
	require.NoError(t, err)

	validRequestDTO := dtos.RecordRequest{
		Duration:              1 * time.Minute,
		EventLimit:            10,
		IncludeDeviceProfiles: nil,
		IncludeDevices:        nil,
		IncludeSources:        nil,
		ExcludeDeviceProfiles: nil,
		ExcludeDevices:        nil,
		ExcludeSources:        nil,
	}
	validRequestDTOJson, err := json.Marshal(validRequestDTO)
	require.NoError(t, err)

	tests := []struct {
		Name                         string
		Input                        []byte
		MockDataManagerStartResponse error
		ExpectedStatus               int
		ExpectedMessage              string
	}{
		{"Success", validRequestDTOJson, nil, http.StatusAccepted, ""},
		{"Recording failed", validRequestDTOJson, errors.New("recording failed"), http.StatusInternalServerError, failedRecording},
		{"No Input", nil, nil, http.StatusBadRequest, failedRequestJSON},
		{"Bad JSON Input", []byte("bad input"), nil, http.StatusBadRequest, failedRequestJSON},
		{"Empty DTO Input", emptyRequestDTOJson, nil, http.StatusBadRequest, failedRecordRequestValidate},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockDataManager.On("StartRecording", validRequestDTO).Return(test.MockDataManagerStartResponse).Once()
			req, err := http.NewRequest(http.MethodPost, recordRoute, bytes.NewReader(test.Input))
			require.NoError(t, err)

			testRecorder := httptest.NewRecorder()
			handler.ServeHTTP(testRecorder, req)

			require.Equal(t, test.ExpectedStatus, testRecorder.Code)
			assert.Contains(t, testRecorder.Body.String(), test.ExpectedMessage)
		})
	}
}

func TestHttpController_RecordingStatus(t *testing.T) {
	// TODO: Implement using TDD
}

func TestHttpController_CancelRecording(t *testing.T) {
	// TODO: Implement using TDD
}

func TestHttpController_StartReplay(t *testing.T) {
	// TODO: Implement using TDD
}

func TestHttpController_ReplayStatus(t *testing.T) {
	// TODO: Implement using TDD
}

func TestHttpController_CancelReplay(t *testing.T) {
	// TODO: Implement using TDD
}

func TestHttpController_ExportRecordedData(t *testing.T) {
	// TODO: Implement using TDD
}

func TestHttpController_ImportRecordedData(t *testing.T) {
	// TODO: Implement using TDD
}
