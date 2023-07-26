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

package application

import (
	"errors"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/app-record-replay/pkg/dtos"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	coreDtos "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	target := NewManager(&mocks.ApplicationService{})
	require.NotNil(t, target)
	d := target.(*dataManager)
	require.NotNil(t, d)
	assert.NotNil(t, d.dataChan)
}

func TestDefaultDataManager_StartRecording(t *testing.T) {
	countAndTimeNoFiltersRequest := &dtos.RecordRequest{
		Duration:   10 * time.Second,
		EventLimit: 100,
	}

	tests := []struct {
		Name                    string
		StartRequest            *dtos.RecordRequest
		RecordingAlreadyRunning bool
		ExpectedStartError      error
		SetPipelineError        error
	}{
		{
			Name:         "Happy Path - By Count & Time - no filters",
			StartRequest: countAndTimeNoFiltersRequest,
		},
		{
			Name: "Happy Path - By Count - 3 include filters",
			StartRequest: &dtos.RecordRequest{
				EventLimit:            100,
				IncludeDeviceProfiles: []string{"test-profile1", "test-profile2"},
				IncludeDevices:        []string{"test-device1", "test-device2"},
				IncludeSources:        []string{"test-source1", "test-source2"},
			},
		},
		{
			Name: "Happy Path - By Duration - 3 exclude filters",
			StartRequest: &dtos.RecordRequest{
				Duration:              10 * time.Second,
				ExcludeDeviceProfiles: []string{"test-profile3", "test-profile4"},
				ExcludeDevices:        []string{"test-device3", "test-device4"},
				ExcludeSources:        []string{"test-source3", "test-source4"},
			},
		},
		{
			Name:                    "Fail Path - recording already running",
			StartRequest:            countAndTimeNoFiltersRequest,
			RecordingAlreadyRunning: true,
			ExpectedStartError:      recordingInProgressError,
		},
		{
			Name:             "Fail Path - pipeline set error",
			StartRequest:     countAndTimeNoFiltersRequest,
			SetPipelineError: errors.New("failed"),
		},
		{
			Name:               "Fail Path - No count or duration set",
			StartRequest:       &dtos.RecordRequest{},
			ExpectedStartError: batchParametersNotSetError,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Debug", mock.Anything)
			mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
			mockSdk := &mocks.ApplicationService{}
			mockSdk.On("LoggingClient").Return(mockLogger)
			target := NewManager(mockSdk).(*dataManager)

			// Due to limitation of mocks with respect to function pointers, the best we can do is pass the expected number
			// of mock.Anything parameters to match the number of expected pipeline functions pointers in the actual call.
			var mockArgs []any

			// Add mock parameter for exclude profiles filter function
			if len(test.StartRequest.ExcludeDeviceProfiles) > 0 {
				mockArgs = append(mockArgs, mock.Anything)
			}

			// Add mock parameter for include profiles filter function
			if len(test.StartRequest.IncludeDeviceProfiles) > 0 {
				mockArgs = append(mockArgs, mock.Anything)
			}

			// Add mock parameter for exclude devices filter function
			if len(test.StartRequest.ExcludeDevices) > 0 {
				mockArgs = append(mockArgs, mock.Anything)
			}

			// Add mock parameter for include devices filter function
			if len(test.StartRequest.IncludeDevices) > 0 {
				mockArgs = append(mockArgs, mock.Anything)
			}

			// Add mock parameter for exclude sources filter function
			if len(test.StartRequest.ExcludeSources) > 0 {
				mockArgs = append(mockArgs, mock.Anything)
			}

			// Add mock parameter for include sources filter function
			if len(test.StartRequest.IncludeSources) > 0 {
				mockArgs = append(mockArgs, mock.Anything)
			}

			// Add three more for the expected countEvents, batch.Batch, processBatchedData pipeline functions
			mockArgs = append(mockArgs, mock.Anything, mock.Anything, mock.Anything)

			mockSdk.On("SetDefaultFunctionsPipeline", mockArgs...).Run(func(args mock.Arguments) {
				// Since the mock On will not complain if we have more mockArgs that actual pipeline functions passed,
				// we must verify we are getting the exact number of expected pipeline functions passed
				require.Equal(t, len(mockArgs), len(args), "Expect more pipeline functions to be passed to SetDefaultFunctionsPipeline")
			}).Return(test.SetPipelineError)

			now := time.Now()
			if test.RecordingAlreadyRunning {
				target.recordingStartedAt = &now
			}

			// simulate previous recorded data is present
			target.recordedData = &recordedData{}
			target.eventCount = 100

			startErr := target.StartRecording(test.StartRequest)

			if test.SetPipelineError != nil {
				require.Error(t, startErr)
				assert.ErrorContains(t, startErr, setPipelineFailedMessage)
				return
			}

			if test.ExpectedStartError != nil {
				require.Error(t, startErr)
				assert.ErrorContains(t, startErr, test.ExpectedStartError.Error())
				return
			}

			require.NoError(t, startErr)
			assert.Nil(t, target.recordedData)
			assert.Zero(t, target.eventCount)
			assert.NotNil(t, target.recordingStartedAt)

			mockSdk.AssertExpectations(t)

			if len(test.StartRequest.IncludeDeviceProfiles) > 0 {
				mockLogger.AssertCalled(t, "Debugf", debugFilterMessage, "for profile", test.StartRequest.IncludeDeviceProfiles)
			}

			if len(test.StartRequest.ExcludeDeviceProfiles) > 0 {
				mockLogger.AssertCalled(t, "Debugf", debugFilterMessage, "out profile", test.StartRequest.ExcludeDeviceProfiles)
			}

			if len(test.StartRequest.IncludeDevices) > 0 {
				mockLogger.AssertCalled(t, "Debugf", debugFilterMessage, "for device", test.StartRequest.IncludeDevices)
			}

			if len(test.StartRequest.ExcludeDevices) > 0 {
				mockLogger.AssertCalled(t, "Debugf", debugFilterMessage, "out device", test.StartRequest.ExcludeDevices)
			}

			if len(test.StartRequest.IncludeSources) > 0 {
				mockLogger.AssertCalled(t, "Debugf", debugFilterMessage, "for source", test.StartRequest.IncludeSources)
			}

			if len(test.StartRequest.ExcludeSources) > 0 {
				mockLogger.AssertCalled(t, "Debugf", debugFilterMessage, "out source", test.StartRequest.ExcludeSources)
			}

			mockLogger.AssertCalled(t, "Debug", debugPipelineFunctionsAddedMessage)
		})
	}
}

func TestDefaultDataManager_RecordingStatus(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDefaultDataManager_CancelRecording(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDefaultDataManager_StartReplay(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDefaultDataManager_ReplayStatus(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDefaultDataManager_CancelReplay(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDefaultDataManager_ExportRecordedData(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDefaultDataManager_ImportRecordedData(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDefaultDataManager_CountEvents(t *testing.T) {
	tests := []struct {
		Name          string
		Data          any
		ExpectedCount int
		ExpectedError error
	}{
		{"Valid", coreDtos.Event{}, 5, nil},
		{"Nil data", nil, 1, countsNoDataError},
		{"Not Event", coreDtos.Metric{}, 1, countsDataNotEventError},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Debugf", mock.Anything, mock.Anything)
			mockSdk := &mocks.ApplicationService{}
			mockSdk.On("LoggingClient").Return(mockLogger)

			target := NewManager(mockSdk).(*dataManager)
			for i := 0; i < test.ExpectedCount; i++ {
				continueExecution, actual := target.countEvents(nil, test.Data)
				if test.ExpectedError != nil {
					require.Equal(t, test.ExpectedError, actual)
					return
				}

				require.True(t, continueExecution)
				require.Equal(t, test.Data, actual)
			}

			assert.Equal(t, test.ExpectedCount, target.eventCount)
		})
	}
}

func TestDefaultDataManager_ProcessBatchedData(t *testing.T) {
	expectedBatchedEvents := []coreDtos.Event{
		coreDtos.NewEvent("test-profile1", "test-device1", "test-source1"),
		coreDtos.NewEvent("test-profile2", "test-device2", "test-source2"),
		coreDtos.NewEvent("test-profile3", "test-device3", "test-source3"),
		coreDtos.NewEvent("test-profile4", "test-device4", "test-source4"),
	}

	tests := []struct {
		Name          string
		Data          any
		ExpectedError error
	}{
		{"Valid", expectedBatchedEvents, nil},
		{"Nil data", nil, batchNoDataError},
		{"Not Collection of Events", coreDtos.Event{}, batchDataNotEventCollectionError},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Debug", mock.Anything)
			mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
			mockSdk := &mocks.ApplicationService{}
			mockSdk.On("RemoveAllFunctionPipelines")
			mockSdk.On("LoggingClient").Return(mockLogger)

			target := NewManager(mockSdk).(*dataManager)
			now := time.Now()
			target.recordingStartedAt = &now

			continueExecution, actual := target.processBatchedData(nil, test.Data)
			if test.ExpectedError != nil {
				require.Equal(t, test.ExpectedError, actual)
				return
			}

			require.False(t, continueExecution)
			require.NotNil(t, target.recordedData)
			assert.Equal(t, expectedBatchedEvents, target.recordedData.Events)
			assert.NotZero(t, target.recordedData.Duration)

			mockSdk.AssertExpectations(t)
		})
	}
}
