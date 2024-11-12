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
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	appMocks "github.com/edgexfoundry/app-functions-sdk-go/v4/pkg/interfaces/mocks"
	"github.com/edgexfoundry/app-record-replay/internal/interfaces/mocks"
	"github.com/edgexfoundry/app-record-replay/pkg/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	coreDtos "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	target, _, _ := createTargetAndMocks()

	require.NotNil(t, target)
	require.NotNil(t, target)
	assert.NotNil(t, target.appSdk)
	assert.NotNil(t, target.lc)
	assert.NotNil(t, target.dataManager)
}

func TestHttpController_AddCustomRoutes_Success(t *testing.T) {
	target, _, mockSdk := createTargetAndMocks()
	mockSdk.On("AddCustomRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

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

	expectedError := errors.New("AddRoutes error")
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockSdk := &appMocks.ApplicationService{}
			mockSdk.On("AddCustomRoute", test.Route, mock.Anything, mock.Anything, test.Method).Return(expectedError)
			mockSdk.On("AddCustomRoute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockSdk.On("LoggingClient").Return(logger.NewMockClient())

			target := New(nil, mockSdk)

			err := target.AddRoutes()
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.Route)
			assert.Contains(t, err.Error(), test.Method)
			assert.Contains(t, err.Error(), expectedError.Error())
		})
	}
}

func TestHttpController_StartRecording(t *testing.T) {
	target, mockDataManager, _ := createTargetAndMocks()

	handler := http.HandlerFunc(WrapEchoHandler(t, target.startRecording))

	emptyRequestDTO := dtos.RecordRequest{}

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
	badDurationRequestDTO := dtos.RecordRequest{
		Duration:              -99,
		EventLimit:            10,
		IncludeDeviceProfiles: nil,
		IncludeDevices:        nil,
		IncludeSources:        nil,
		ExcludeDeviceProfiles: nil,
		ExcludeDevices:        nil,
		ExcludeSources:        nil,
	}

	badEventLimitRequestDTO := dtos.RecordRequest{
		Duration:              0,
		EventLimit:            -99,
		IncludeDeviceProfiles: nil,
		IncludeDevices:        nil,
		IncludeSources:        nil,
		ExcludeDeviceProfiles: nil,
		ExcludeDevices:        nil,
		ExcludeSources:        nil,
	}

	tests := []struct {
		Name                         string
		Input                        []byte
		MockDataManagerStartResponse error
		ExpectedStatus               int
		ExpectedMessage              string
	}{
		{"Success", marshal(t, validRequestDTO), nil, http.StatusAccepted, ""},
		{"Recording failed", marshal(t, validRequestDTO), errors.New("recording failed"), http.StatusInternalServerError, failedRecording},
		{"No Input", nil, nil, http.StatusBadRequest, failedRequestJSON},
		{"Bad JSON Input", []byte("bad input"), nil, http.StatusBadRequest, failedRequestJSON},
		{"Empty DTO Input", marshal(t, emptyRequestDTO), nil, http.StatusBadRequest, failedRecordRequestValidate},
		{"Bad Duration", marshal(t, badDurationRequestDTO), nil, http.StatusBadRequest, failedRecordDurationValidate},
		{"Bad Event Limit", marshal(t, badEventLimitRequestDTO), nil, http.StatusBadRequest, failedRecordEventLimitValidate},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockDataManager.On("StartRecording", mock.Anything).Return(test.MockDataManagerStartResponse).Once()
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
	target, mockDataManager, _ := createTargetAndMocks()

	handler := http.HandlerFunc(WrapEchoHandler(t, target.recordingStatus))

	inProgressRecordStatus := dtos.RecordStatus{
		InProgress: true,
		EventCount: 10,
		Duration:   time.Second,
	}
	tests := []struct {
		Name             string
		ExpectedResponse dtos.RecordStatus
		ExpectedStatus   int
		ExpectedError    error
	}{
		{
			Name:             "Valid in progress status test",
			ExpectedResponse: inProgressRecordStatus,
			ExpectedStatus:   http.StatusOK,
			ExpectedError:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockDataManager.On("RecordingStatus").Return(test.ExpectedResponse, test.ExpectedError).Once()
			req, err := http.NewRequest(http.MethodGet, recordRoute, nil)
			require.NoError(t, err)

			testRecorder := httptest.NewRecorder()
			handler.ServeHTTP(testRecorder, req)

			require.Equal(t, test.ExpectedStatus, testRecorder.Code)
			if test.ExpectedStatus != http.StatusOK {
				return
			}

			require.NotNil(t, testRecorder.Body)
			actualResponse := dtos.RecordStatus{}
			err = json.Unmarshal(testRecorder.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			require.Equal(t, test.ExpectedResponse, actualResponse)
		})
	}
}

func TestHttpController_CancelRecording(t *testing.T) {
	target, mockDataManager, _ := createTargetAndMocks()

	handler := http.HandlerFunc(WrapEchoHandler(t, target.cancelRecording))

	tests := []struct {
		Name           string
		ExpectedStatus int
		ExpectedError  error
	}{
		{"Valid", http.StatusAccepted, nil},
		{"Error", http.StatusInternalServerError, errors.New("failed")},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockDataManager.On("CancelRecording").Return(test.ExpectedError).Once()

			req, err := http.NewRequest(http.MethodGet, recordRoute, nil)
			require.NoError(t, err)

			testRecorder := httptest.NewRecorder()
			handler.ServeHTTP(testRecorder, req)

			require.Equal(t, test.ExpectedStatus, testRecorder.Code)
		})
	}
}

func TestHttpController_StartReplay(t *testing.T) {
	target, mockDataManager, _ := createTargetAndMocks()

	handler := http.HandlerFunc(WrapEchoHandler(t, target.startReplay))

	validRequestDTO := dtos.ReplayRequest{
		ReplayRate: 1,
	}

	invalidEmptyRequestDTO := dtos.ReplayRequest{}

	invalidCountRequestDTO := dtos.ReplayRequest{
		ReplayRate:  1.5,
		RepeatCount: -1,
	}

	invalidRateRequestDTO := dtos.ReplayRequest{
		ReplayRate:  0.0,
		RepeatCount: 0,
	}

	tests := []struct {
		Name                         string
		Input                        []byte
		MockDataManagerStartResponse error
		ExpectedStatus               int
		ExpectedMessage              string
	}{
		{"Success", marshal(t, validRequestDTO), nil, http.StatusAccepted, ""},
		{"Recording failed", marshal(t, validRequestDTO), errors.New("replay failed"), http.StatusInternalServerError, failedReplay},
		{"No Input", nil, nil, http.StatusBadRequest, failedRequestJSON},
		{"Bad JSON Input", []byte("bad input"), nil, http.StatusBadRequest, failedRequestJSON},
		{"Empty DTO Input", marshal(t, invalidEmptyRequestDTO), nil, http.StatusBadRequest, failedReplayRateValidate},
		{"Bad Rate", marshal(t, invalidRateRequestDTO), nil, http.StatusBadRequest, failedReplayRateValidate},
		{"Bad Count", marshal(t, invalidCountRequestDTO), nil, http.StatusBadRequest, failedRepeatCountValidate},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockDataManager.On("StartReplay", mock.Anything).Return(test.MockDataManagerStartResponse).Once()
			req, err := http.NewRequest(http.MethodPost, recordRoute, bytes.NewReader(test.Input))
			require.NoError(t, err)

			testRecorder := httptest.NewRecorder()
			handler.ServeHTTP(testRecorder, req)

			require.Equal(t, test.ExpectedStatus, testRecorder.Code)
			assert.Contains(t, testRecorder.Body.String(), test.ExpectedMessage)
		})
	}
}

func TestHttpController_ReplayStatus(t *testing.T) {
	target, mockDataManager, _ := createTargetAndMocks()

	handler := http.HandlerFunc(WrapEchoHandler(t, target.replayStatus))

	notRunningReplayStatus := dtos.ReplayStatus{
		Running:     false,
		EventCount:  0,
		Duration:    0,
		RepeatCount: 0,
	}

	inProgressReplayStatus := dtos.ReplayStatus{
		Running:     true,
		EventCount:  4,
		Duration:    time.Second,
		RepeatCount: 1,
	}
	tests := []struct {
		Name             string
		ExpectedResponse dtos.ReplayStatus
		ExpectedStatus   int
		ExpectedError    error
	}{
		{
			Name:             "Valid in progress status test",
			ExpectedResponse: inProgressReplayStatus,
			ExpectedStatus:   http.StatusOK,
			ExpectedError:    nil,
		},
		{
			Name:             "Valid not running replay status",
			ExpectedResponse: notRunningReplayStatus,
			ExpectedStatus:   http.StatusOK,
			ExpectedError:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockDataManager.On("ReplayStatus").Return(test.ExpectedResponse, test.ExpectedError).Once()
			req, err := http.NewRequest(http.MethodGet, replayRoute, nil)
			require.NoError(t, err)

			testRecorder := httptest.NewRecorder()
			handler.ServeHTTP(testRecorder, req)

			require.Equal(t, test.ExpectedStatus, testRecorder.Code)
			if test.ExpectedStatus != http.StatusOK {
				return
			}

			require.NotNil(t, testRecorder.Body)
			actualResponse := dtos.ReplayStatus{}
			err = json.Unmarshal(testRecorder.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			require.Equal(t, test.ExpectedResponse, actualResponse)
		})
	}
}
func TestHttpController_CancelReplay(t *testing.T) {
	target, mockDataManager, _ := createTargetAndMocks()

	handler := http.HandlerFunc(WrapEchoHandler(t, target.cancelReplay))

	tests := []struct {
		Name           string
		ExpectedStatus int
		ExpectedError  error
	}{
		{"Valid", http.StatusAccepted, nil},
		{"Error", http.StatusInternalServerError, errors.New("failed")},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockDataManager.On("CancelReplay").Return(test.ExpectedError).Once()

			req, err := http.NewRequest(http.MethodGet, recordRoute, nil)
			require.NoError(t, err)

			testRecorder := httptest.NewRecorder()
			handler.ServeHTTP(testRecorder, req)

			require.Equal(t, test.ExpectedStatus, testRecorder.Code)
		})
	}
}

func TestHttpController_ExportRecordedData(t *testing.T) {
	noRecordedData := dtos.RecordedData{}
	recordedData := dtos.RecordedData{
		RecordedEvents: []coreDtos.Event{
			coreDtos.Event{
				DeviceName:  "test",
				ProfileName: "test",
				Readings: []coreDtos.BaseReading{
					coreDtos.BaseReading{
						SimpleReading: coreDtos.SimpleReading{
							Value: "1456.0",
						},
					},
				},
			},
			coreDtos.Event{
				DeviceName:  "test",
				ProfileName: "test",
				Readings: []coreDtos.BaseReading{
					coreDtos.BaseReading{
						SimpleReading: coreDtos.SimpleReading{
							Value: "1456.0",
						},
					},
				},
			},
		},
		Devices: []coreDtos.Device{
			coreDtos.Device{
				Name:        "test_device",
				ProfileName: "test",
			},
		},
		Profiles: []coreDtos.DeviceProfile{
			coreDtos.DeviceProfile{
				DeviceProfileBasicInfo: coreDtos.DeviceProfileBasicInfo{
					Name: "test",
				},
			},
		},
	}

	tests := []struct {
		Name             string
		ExpectedResponse *dtos.RecordedData
		ExpectedStatus   int
		ExpectedError    error
		QueryParam       string
	}{
		{
			Name:             "Valid with data",
			ExpectedResponse: &recordedData,
			ExpectedStatus:   http.StatusOK,
			ExpectedError:    nil,
		},
		{
			Name:             "Valid - no events",
			ExpectedResponse: &noRecordedData,
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedError:    fmt.Errorf("failed"),
		},
		{
			Name:             "Valid with data with GZIP query parameter",
			ExpectedResponse: &recordedData,
			ExpectedStatus:   http.StatusOK,
			ExpectedError:    nil,
			QueryParam:       gzipCompression,
		},
		{
			Name:             "Valid with data with ZLIB query parameter",
			ExpectedResponse: &recordedData,
			ExpectedStatus:   http.StatusOK,
			ExpectedError:    nil,
			QueryParam:       zlibCompression,
		},
		{
			Name:             "Valid with data with invalid GZ query parameter",
			ExpectedResponse: &recordedData,
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedError:    nil,
			QueryParam:       "GZ",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			target, mockDataManager, _ := createTargetAndMocks()
			mockDataManager.On("ExportRecordedData").Return(test.ExpectedResponse, test.ExpectedError)

			handler := http.HandlerFunc(WrapEchoHandler(t, target.exportRecordedData))

			req, err := http.NewRequest(http.MethodGet, dataRoute, nil)
			require.NoError(t, err)

			query := req.URL.Query()
			query.Add("compression", test.QueryParam)
			req.URL.RawQuery = query.Encode()

			testRecorder := httptest.NewRecorder()
			handler.ServeHTTP(testRecorder, req)

			require.Equal(t, test.ExpectedStatus, testRecorder.Code)
			if test.ExpectedStatus != http.StatusOK {
				return
			}

			require.NotNil(t, testRecorder.Body)
			actualResponse := &dtos.RecordedData{}
			if test.QueryParam == "" {
				err = json.Unmarshal(testRecorder.Body.Bytes(), actualResponse)
				require.NoError(t, err)
			} else {
				actualResponse = uncompressData(t, test.QueryParam, testRecorder.Body)
			}

			require.Equal(t, test.ExpectedResponse, actualResponse)

		})
	}
}

func TestHttpController_ImportRecordedData(t *testing.T) {
	emptyDataRequest := dtos.RecordedData{}
	recordedEventRequest := dtos.RecordedData{
		RecordedEvents: []coreDtos.Event{
			coreDtos.Event{
				DeviceName:  "test",
				ProfileName: "test",
				Readings: []coreDtos.BaseReading{
					coreDtos.BaseReading{
						SimpleReading: coreDtos.SimpleReading{
							Value: "1456.0",
						},
					},
				},
			},
			coreDtos.Event{
				DeviceName:  "test",
				ProfileName: "test",
				Readings: []coreDtos.BaseReading{
					coreDtos.BaseReading{
						SimpleReading: coreDtos.SimpleReading{
							Value: "1456.0",
						},
					},
				},
			},
		},
		Devices: []coreDtos.Device{
			coreDtos.Device{
				Name:        "test_device",
				ProfileName: "test",
			},
		},
		Profiles: []coreDtos.DeviceProfile{
			coreDtos.DeviceProfile{
				DeviceProfileBasicInfo: coreDtos.DeviceProfileBasicInfo{
					Name: "test",
				},
			},
		},
	}

	trueParam := "true"
	falseParam := "false"

	tests := []struct {
		Name             string
		ContentEncoding  string
		ContentType      string
		OverwriteParam   *string
		ExpectedResponse []byte
		ExpectedStatus   int
		ExpectedError    error
	}{
		{
			Name:             "valid - data with 2 events",
			ExpectedResponse: marshal(t, recordedEventRequest),
			ExpectedStatus:   http.StatusAccepted,
			ExpectedError:    nil,
			OverwriteParam:   &falseParam,
			ContentType:      common.ContentTypeJSON,
		},
		{
			Name:             "valid - data with 2 events using json file",
			ExpectedResponse: readJsonFile(t, "recordedDataJsonUncompressed.json"),
			ExpectedStatus:   http.StatusAccepted,
			ExpectedError:    nil,
			OverwriteParam:   &falseParam,
			ContentType:      common.ContentTypeJSON,
		},
		{
			Name:             "invalid - no data",
			ExpectedResponse: marshal(t, emptyDataRequest),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedError:    errors.New(noDataFound),
			OverwriteParam:   &falseParam,
			ContentType:      common.ContentTypeJSON,
		},
		{
			Name:             "valid - data with 2 events and compressed gzip",
			ContentEncoding:  contentEncodingGzip,
			ExpectedResponse: compressData(t, "GZIP", recordedEventRequest),
			ExpectedStatus:   http.StatusAccepted,
			ExpectedError:    nil,
			OverwriteParam:   &falseParam,
			ContentType:      common.ContentTypeJSON,
		},
		{
			Name:             "valid - data with 2 events and compressed zlib",
			ContentEncoding:  contentEncodingZlib,
			ExpectedResponse: compressData(t, "ZLIB", recordedEventRequest),
			ExpectedStatus:   http.StatusAccepted,
			ExpectedError:    nil,
			OverwriteParam:   &falseParam,
			ContentType:      common.ContentTypeJSON,
		},
		{
			Name:             "valid - data with 2 events and compressed gzip w/ overwrite set to true",
			ContentEncoding:  contentEncodingGzip,
			ExpectedResponse: compressData(t, "GZIP", recordedEventRequest),
			ExpectedStatus:   http.StatusAccepted,
			ExpectedError:    nil,
			OverwriteParam:   &trueParam,
			ContentType:      common.ContentTypeJSON,
		},
		{
			Name:             "valid - data with 2 events and compressed zlib  w/ overwrite set to nil",
			ContentEncoding:  contentEncodingZlib,
			ExpectedResponse: compressData(t, "ZLIB", recordedEventRequest),
			ExpectedStatus:   http.StatusAccepted,
			ExpectedError:    nil,
			OverwriteParam:   nil,
			ContentType:      common.ContentTypeJSON,
		},
		{
			Name:             "invalid - bad Content-type",
			ExpectedResponse: marshal(t, emptyDataRequest),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedError:    nil,
			OverwriteParam:   &falseParam,
			ContentType:      common.ContentTypeTOML,
		},
		{
			Name:             "invalid - no Content-type",
			ExpectedResponse: marshal(t, emptyDataRequest),
			ExpectedStatus:   http.StatusBadRequest,
			ExpectedError:    nil,
			OverwriteParam:   &falseParam,
			ContentType:      "",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			target, mockDataManager, _ := createTargetAndMocks()
			handler := http.HandlerFunc(WrapEchoHandler(t, target.importRecordedData))
			mockDataManager.On("ImportRecordedData", mock.Anything, mock.Anything).Return(test.ExpectedError)

			req, err := http.NewRequest(http.MethodPost, dataRoute, bytes.NewReader(test.ExpectedResponse))
			require.NoError(t, err)

			req.Header.Set(common.ContentType, test.ContentType)
			if len(test.ContentEncoding) > 0 {
				req.Header.Set("Content-Encoding", test.ContentEncoding)
			}

			if test.OverwriteParam != nil {
				query := req.URL.Query()
				query.Add("overwrite", *test.OverwriteParam)
				req.URL.RawQuery = query.Encode()
			}

			testRecorder := httptest.NewRecorder()
			handler.ServeHTTP(testRecorder, req)

			require.Equal(t, test.ExpectedStatus, testRecorder.Code)
		})
	}

}

func marshal(t *testing.T, v any) []byte {
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func readJsonFile(t *testing.T, file string) []byte {
	jsonData, err := os.ReadFile(file)
	require.NoError(t, err)
	return jsonData
}

func createTargetAndMocks() (*httpController, *mocks.DataManager, *appMocks.ApplicationService) {
	mockDataManager := &mocks.DataManager{}
	mockSdk := &appMocks.ApplicationService{}
	mockSdk.On("LoggingClient").Return(logger.NewMockClient())

	target := New(mockDataManager, mockSdk).(*httpController)
	return target, mockDataManager, mockSdk
}

func uncompressData(t *testing.T, compressionType string, r io.Reader) *dtos.RecordedData {
	data := dtos.RecordedData{}
	switch compressionType {
	case gzipCompression:
		reader, err := gzip.NewReader(r)
		require.NoError(t, err)
		defer reader.Close()
		err = json.NewDecoder(reader).Decode(&data)
		require.NoError(t, err)
	case zlibCompression:
		reader, err := zlib.NewReader(r)
		require.NoError(t, err)
		defer reader.Close()
		err = json.NewDecoder(reader).Decode(&data)
		require.NoError(t, err)
	}

	return &data
}

func compressData(t *testing.T, compressionType string, data dtos.RecordedData) []byte {
	buf := &bytes.Buffer{}
	var jsondata []byte
	var err error
	switch compressionType {
	case "GZIP":
		gZipWriter := gzip.NewWriter(buf)
		jsondata, err = json.Marshal(&data)
		require.NoError(t, err)
		_, err = gZipWriter.Write(jsondata)
		require.NoError(t, err)
		gZipWriter.Close()
	case "ZLIB":
		zlibWriter := zlib.NewWriter(buf)
		jsondata, err = json.Marshal(&data)
		require.NoError(t, err)
		_, err = zlibWriter.Write(jsondata)
		require.NoError(t, err)
		zlibWriter.Close()
	}

	return buf.Bytes()
}

// WrapHandler wraps `handler func(http.ResponseWriter, *http.Request)` into `echo.HandlerFunc`
func WrapEchoHandler(t *testing.T, handler echo.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	t.Helper()
	return func(writer http.ResponseWriter, req *http.Request) {
		ctx := echo.New().NewContext(req, writer)
		err := handler(ctx)
		require.NoError(t, err)
	}
}
