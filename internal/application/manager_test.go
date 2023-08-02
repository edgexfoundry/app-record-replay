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
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/app-record-replay/pkg/dtos"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	coreDtos "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	expectedProfileName = "testProfile"
	expectedDeviceName  = "testDevice"
	expectedSourceName  = "testSource"
)

var expectedEventData = []coreDtos.Event{
	coreDtos.NewEvent(expectedProfileName, expectedDeviceName, expectedSourceName),
	coreDtos.NewEvent(expectedProfileName, expectedDeviceName, expectedSourceName),
	coreDtos.NewEvent(expectedProfileName, expectedDeviceName, expectedSourceName),
}

func TestMain(m *testing.M) {
	for i := range expectedEventData {
		addValue := time.Duration(int64(i) * int64(time.Second))
		expectedEventData[i].Origin = time.Now().Add(addValue).UnixNano()
		_ = expectedEventData[i].AddSimpleReading(expectedSourceName, common.ValueTypeString, "test1")
	}

	os.Exit(m.Run())
}

func TestNewManager(t *testing.T) {
	target := NewManager(&mocks.ApplicationService{}, 0)
	require.NotNil(t, target)
	d := target.(*dataManager)
	require.NotNil(t, d)
}

func TestDataManager_StartRecording(t *testing.T) {
	countAndTimeNoFiltersRequest := dtos.RecordRequest{
		Duration:   10 * time.Second,
		EventLimit: 100,
	}

	tests := []struct {
		Name                    string
		StartRequest            dtos.RecordRequest
		RecordingAlreadyRunning bool
		ReplayRunning           bool
		ExpectedStartError      error
		SetPipelineError        error
	}{
		{
			Name:         "Happy Path - By Count & Time - no filters",
			StartRequest: countAndTimeNoFiltersRequest,
		},
		{
			Name: "Happy Path - By Count - 3 include filters",
			StartRequest: dtos.RecordRequest{
				EventLimit:            100,
				IncludeDeviceProfiles: []string{"test-profile1", "test-profile2"},
				IncludeDevices:        []string{"test-device1", "test-device2"},
				IncludeSources:        []string{"test-source1", "test-source2"},
			},
		},
		{
			Name: "Happy Path - By Duration - 3 exclude filters",
			StartRequest: dtos.RecordRequest{
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
			Name:               "Fail Path - replay is running",
			StartRequest:       countAndTimeNoFiltersRequest,
			ReplayRunning:      true,
			ExpectedStartError: replayInProgressError,
		},
		{
			Name:             "Fail Path - pipeline set error",
			StartRequest:     countAndTimeNoFiltersRequest,
			SetPipelineError: errors.New("failed"),
		},
		{
			Name:               "Fail Path - No count or duration set",
			StartRequest:       dtos.RecordRequest{},
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
			target := NewManager(mockSdk, 0).(*dataManager)

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

			if test.ReplayRunning {
				target.replayStartedAt = &now
			}

			// simulate previous recorded data is present
			target.recordedData = &recordedData{}
			target.recordedEventCount = 100

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
			assert.Zero(t, target.recordedEventCount)
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

func TestDataManager_RecordingStatus(t *testing.T) {
	tests := []struct {
		Name           string
		ExpectedStatus *dtos.RecordStatus
	}{
		{
			Name: "Happy Path - nothing recoded",
			ExpectedStatus: &dtos.RecordStatus{
				InProgress: false,
				EventCount: 0,
				Duration:   0,
			},
		},
		{
			Name: "Happy Path - Recording ending",
			ExpectedStatus: &dtos.RecordStatus{
				InProgress: false,
				EventCount: 10,
				Duration:   6000000000,
			},
		},
		{
			Name: "Happy Path - Recording in progress",
			ExpectedStatus: &dtos.RecordStatus{
				InProgress: true,
				EventCount: 10,
				Duration:   6000000000,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			target := NewManager(nil, 0).(*dataManager)

			if test.ExpectedStatus.InProgress {
				// Set up case when recording is in progress
				startTime := time.Now().Add(test.ExpectedStatus.Duration * -1)
				target.recordingStartedAt = &startTime
				target.recordedEventCount = test.ExpectedStatus.EventCount
			} else if test.ExpectedStatus.EventCount > 0 || test.ExpectedStatus.Duration > 0 {
				// Set up case when recording is finished and using recorded data
				target.recordedData = &recordedData{
					Duration: test.ExpectedStatus.Duration,
				}

				for i := 0; i < test.ExpectedStatus.EventCount; i++ {
					target.recordedData.Events = append(target.recordedData.Events, coreDtos.Event{})
				}
			}

			// default case (neither conditions above) is when recording is not in progress & no previous recorded data exists
			// Nothing to set up.

			actual := target.RecordingStatus()

			assert.Equal(t, test.ExpectedStatus.InProgress, actual.InProgress)
			assert.Equal(t, test.ExpectedStatus.EventCount, actual.EventCount)
			assert.Equal(t, test.ExpectedStatus.Duration, actual.Duration.Round(test.ExpectedStatus.Duration))
		})
	}
}

func TestDataManager_CancelRecording(t *testing.T) {
	tests := []struct {
		Name             string
		RecordingRunning bool
		ExpectedError    error
	}{
		{
			Name:             "Happy Path - Running recording canceled",
			RecordingRunning: true,
		},
		{
			Name:             "Error Path - No recording running to be canceled",
			RecordingRunning: false,
			ExpectedError:    noRecordingRunningToCancelError,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Debug", mock.Anything)
			mockSdk := &mocks.ApplicationService{}
			mockSdk.On("LoggingClient").Return(mockLogger)
			mockSdk.On("RemoveAllFunctionPipelines")

			target := NewManager(mockSdk, 0).(*dataManager)

			if test.RecordingRunning {
				now := time.Now()
				target.recordingStartedAt = &now
			}

			err := target.CancelRecording()
			if test.ExpectedError != nil {
				assert.Equal(t, err, test.ExpectedError)
				return
			}

			require.NoError(t, err)
			assert.Nil(t, target.recordingStartedAt)

			mockSdk.AssertExpectations(t)
		})
	}
}

func TestDataManager_StartReplay(t *testing.T) {
	expectedTopic := common.BuildTopic(strings.Replace(common.CoreDataEventSubscribeTopic, "/#", "", 1),
		expectedProfileName, expectedDeviceName, expectedSourceName)
	goodRequest := dtos.ReplayRequest{
		ReplayRate:  1,
		RepeatCount: 2,
	}

	tests := []struct {
		Name                 string
		StartRequest         dtos.ReplayRequest
		RecordingRunning     bool
		ReplayAlreadyRunning bool
		MaxReplayDelayLimit  time.Duration
		RecordedData         *recordedData
		ExpectedPublishError error
		ExpectedReplayError  error
		ExpectedStartError   error
	}{
		{
			Name:                "Happy Path",
			StartRequest:        goodRequest,
			MaxReplayDelayLimit: time.Minute,
			RecordedData: &recordedData{
				Events: expectedEventData,
			},
		},
		{
			Name:         "Error Path - failed to publish",
			StartRequest: goodRequest,
			RecordedData: &recordedData{
				Events: expectedEventData,
			},
			ExpectedPublishError: errors.New("publish failed"),
		},
		{
			Name: "Error Path - calculated delay too large",
			StartRequest: dtos.ReplayRequest{
				ReplayRate: 0.00001,
			},
			MaxReplayDelayLimit: 1 * time.Second,
			RecordedData: &recordedData{
				Events: expectedEventData,
			},
			ExpectedReplayError: errors.New("delay exceeds the maximum replay delay"),
		},
		{
			Name: "Error Path - Bad ReplayRate -1",
			StartRequest: dtos.ReplayRequest{
				ReplayRate:  -1,
				RepeatCount: 0,
			},
			RecordedData:       &recordedData{},
			ExpectedStartError: invalidReplayRate,
		},
		{
			Name: "Error Path - Bad ReplayRate 0",
			StartRequest: dtos.ReplayRequest{
				ReplayRate:  0,
				RepeatCount: 0,
			},
			RecordedData:       &recordedData{},
			ExpectedStartError: invalidReplayRate,
		},
		{
			Name: "Error Path - Bad RepeatCount -1",
			StartRequest: dtos.ReplayRequest{
				ReplayRate:  1,
				RepeatCount: -1,
			},
			RecordedData:       &recordedData{},
			ExpectedStartError: invalidReplayCount,
		},
		{
			Name:               "Error Path - Recording in progress",
			RecordingRunning:   true,
			ExpectedStartError: recordingInProgressError,
		},
		{
			Name:                 "Error Path - Replay in progress",
			ReplayAlreadyRunning: true,
			ExpectedStartError:   replayInProgressError,
		},
		{
			Name:               "Error Path - No recorded data",
			ExpectedStartError: noRecordedData,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			mockSdk := &mocks.ApplicationService{}
			mockSdk.On("LoggingClient").Return(mockLogger)
			mockSdk.On("AppContext").Return(context.Background())
			mockSdk.On("PublishWithTopic", expectedTopic, mock.Anything, common.ContentTypeJSON).Return(test.ExpectedPublishError)
			target := NewManager(mockSdk, test.MaxReplayDelayLimit).(*dataManager)

			target.recordingStartedAt = nil
			target.replayStartedAt = nil
			if test.RecordedData != nil {
				target.recordedData = test.RecordedData
			}

			now := time.Now()

			if test.RecordingRunning {
				target.recordingStartedAt = &now
			}

			if test.ReplayAlreadyRunning {
				target.replayStartedAt = &now
			}

			if test.RecordedData != nil {
				target.recordedData = test.RecordedData
			}

			err := target.StartReplay(test.StartRequest)

			if test.ExpectedStartError != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.ExpectedStartError.Error())
				return
			}

			require.NoError(t, err)
			target.recordingMutex.Lock()
			assert.NotNil(t, target.replayStartedAt)
			assert.Zero(t, target.replayedDuration)
			assert.Zero(t, target.replayedEventCount)
			target.recordingMutex.Unlock()

			// Wait for the replay to complete
			for {
				target.recordingMutex.Lock()
				replayStartedAt := target.replayStartedAt
				target.recordingMutex.Unlock()

				if replayStartedAt == nil {
					break
				}

				time.Sleep(500 * time.Millisecond)
			}

			target.recordingMutex.Lock()
			defer target.recordingMutex.Unlock()

			if test.ExpectedPublishError != nil {
				require.Error(t, target.replayError)
				assert.ErrorContains(t, target.replayError, test.ExpectedPublishError.Error())
				return
			}

			if test.ExpectedReplayError != nil {
				require.Error(t, target.replayError)
				assert.ErrorContains(t, target.replayError, test.ExpectedReplayError.Error())
				return
			}

			expectedEventCount := len(test.RecordedData.Events) * test.StartRequest.RepeatCount
			assert.Equal(t, expectedEventCount, target.replayedEventCount)
			assert.NotZero(t, target.replayedDuration)
		})
	}
}

func TestDataManager_StartReplay_Cancel(t *testing.T) {
	// These values should allow time to cancel.
	replayRequest := dtos.ReplayRequest{
		ReplayRate:  0.10,
		RepeatCount: 100,
	}

	tests := []struct {
		Name                string
		ReplayCancel        bool
		AppTerminated       bool
		ExpectedReplayError error
	}{
		{
			Name:                "Replay cancel",
			ReplayCancel:        true,
			ExpectedReplayError: errors.New("replay canceled"),
		},
		{
			Name:                "App cancel",
			AppTerminated:       true,
			ExpectedReplayError: errors.New("app terminated"),
		},
	}

	appCtx, appCancelFunc := context.WithCancel(context.Background())
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Info", replayExiting)
			mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			mockSdk := &mocks.ApplicationService{}
			mockSdk.On("LoggingClient").Return(mockLogger)
			mockSdk.On("AppContext").Return(appCtx)
			mockSdk.On("PublishWithTopic", mock.Anything, mock.Anything, mock.Anything).Return(nil)

			target := NewManager(mockSdk, time.Minute).(*dataManager)

			target.recordedData = &recordedData{
				Events: expectedEventData,
			}

			err := target.StartReplay(replayRequest)
			require.NoError(t, err)

			require.True(t, test.AppTerminated || test.ReplayCancel)

			if test.AppTerminated {
				appCancelFunc()
			} else if test.ReplayCancel {
				target.replayCancelFunc()
			}

			// Wait for the replay to cancel
			for {
				target.recordingMutex.Lock()
				replayStartedAt := target.replayStartedAt
				target.recordingMutex.Unlock()

				if replayStartedAt == nil {
					break
				}

				time.Sleep(500 * time.Millisecond)
			}

			if test.AppTerminated {
				mockLogger.AssertCalled(t, "Info", replayExiting)
				return
			}

			target.recordingMutex.Lock()
			defer target.recordingMutex.Unlock()

			require.Error(t, target.replayError)
			assert.Equal(t, test.ExpectedReplayError, target.replayError)

		})
	}

	// This is need to appease the linter.
	appCancelFunc()
}

func TestDataManager_ReplayStatus(t *testing.T) {
	// These values should allow time get status .
	longReplayRequest := dtos.ReplayRequest{
		ReplayRate:  1,
		RepeatCount: 100000,
	}

	// These values should replay fast so status is after completed .
	shortReplayRequest := dtos.ReplayRequest{
		ReplayRate:  10,
		RepeatCount: 2,
	}

	tests := []struct {
		Name                string
		Request             dtos.ReplayRequest
		StatusWhileRunning  bool
		NoReplayRan         bool
		ExpectedStatus      dtos.ReplayStatus
		ExpectedReplayError error
	}{
		{
			Name:               "Replay Status while replaying",
			Request:            longReplayRequest,
			StatusWhileRunning: true,
			ExpectedStatus: dtos.ReplayStatus{
				Running: true,
			},
		},
		{
			Name:    "Replay Status after replayed",
			Request: shortReplayRequest,
			ExpectedStatus: dtos.ReplayStatus{
				Running:     false,
				EventCount:  6,
				RepeatCount: 2,
			},
		},
		{
			Name:        "Replay Status nothing replayed",
			NoReplayRan: true,
			Request:     shortReplayRequest,
			ExpectedStatus: dtos.ReplayStatus{
				ErrorMessage: noReplayExists,
			},
		},
		{
			Name:                "Replay error publishing",
			Request:             shortReplayRequest,
			ExpectedReplayError: errors.New("publish failed"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

			mockSdk := &mocks.ApplicationService{}
			mockSdk.On("LoggingClient").Return(mockLogger)
			mockSdk.On("AppContext").Return(context.Background())
			mockSdk.On("PublishWithTopic", mock.Anything, mock.Anything, mock.Anything).Return(test.ExpectedReplayError)

			target := NewManager(mockSdk, time.Minute).(*dataManager)

			target.recordedData = &recordedData{
				Events: expectedEventData,
			}

			target.replayStartedAt = nil
			target.replayError = nil
			target.replayedEventCount = 0
			target.replayedRepeatCount = 0
			target.replayedDuration = 0

			if !test.NoReplayRan {
				err := target.StartReplay(test.Request)
				require.NoError(t, err)
			}

			if test.StatusWhileRunning {

				time.Sleep(time.Second)

				actualStatus := target.ReplayStatus()

				target.recordingMutex.Lock()
				defer target.recordingMutex.Unlock()

				// Verify replay is still running
				require.NotNil(t, target.replayStartedAt)

				assert.Equal(t, test.ExpectedStatus.Running, actualStatus.Running)
				assert.NotZero(t, actualStatus.EventCount)
				assert.NotZero(t, actualStatus.Duration)
				assert.Empty(t, actualStatus.ErrorMessage)
				return
			}

			// Wait for the replay to complete
			for {
				target.recordingMutex.Lock()
				replayStartedAt := target.replayStartedAt
				target.recordingMutex.Unlock()

				if replayStartedAt == nil {
					break
				}

				time.Sleep(500 * time.Millisecond)
			}

			actualStatus := target.ReplayStatus()

			if test.NoReplayRan {
				// In this case we can compare the whole status struct
				assert.Equal(t, test.ExpectedStatus, actualStatus)
				return
			}

			if test.ExpectedReplayError != nil {
				assert.Equal(t, test.ExpectedStatus.Running, actualStatus.Running)
				require.NotEmpty(t, actualStatus.ErrorMessage)
				assert.Contains(t, actualStatus.ErrorMessage, test.ExpectedReplayError.Error())
				return
			}

			// Since Duration will vary, we can't compare the whole status struct,
			// so have to verify  the individual fields
			assert.Equal(t, test.ExpectedStatus.Running, actualStatus.Running)
			assert.Equal(t, test.ExpectedStatus.EventCount, actualStatus.EventCount)
			assert.Equal(t, test.ExpectedStatus.RepeatCount, actualStatus.RepeatCount)
			assert.NotZero(t, actualStatus.Duration)
			assert.Empty(t, actualStatus.ErrorMessage)
		})
	}
}

func TestDataManager_CancelReplay(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDataManager_ExportRecordedData(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDataManager_ImportRecordedData(t *testing.T) {
	// TODO: Implement using TDD
}

func TestDataManager_CountEvents(t *testing.T) {
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

			target := NewManager(mockSdk, 0).(*dataManager)
			for i := 0; i < test.ExpectedCount; i++ {
				continueExecution, actual := target.countEvents(nil, test.Data)
				if test.ExpectedError != nil {
					require.Equal(t, test.ExpectedError, actual)
					return
				}

				require.True(t, continueExecution)
				require.Equal(t, test.Data, actual)
			}

			assert.Equal(t, test.ExpectedCount, target.recordedEventCount)
		})
	}
}

func TestDataManager_ProcessBatchedData(t *testing.T) {
	expectedBatchedEvents := []coreDtos.Event{
		coreDtos.NewEvent("test-profile1", "test-device1", "test-source1"),
		coreDtos.NewEvent("test-profile2", "test-device2", "test-source2"),
		coreDtos.NewEvent("test-profile3", "test-device3", "test-source3"),
		coreDtos.NewEvent("test-profile4", "test-device4", "test-source4"),
	}

	tests := []struct {
		Name                        string
		Data                        any
		RecordingPreviouslyCanceled bool
		ExpectedError               error
	}{
		{"Valid", expectedBatchedEvents, false, nil},
		{"Valid - Recording previously canceled", nil, true, nil},
		{"Nil data", nil, false, batchNoDataError},
		{"Not Collection of Events", coreDtos.Event{}, false, batchDataNotEventCollectionError},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLogger := &loggerMocks.LoggingClient{}
			mockLogger.On("Debug", mock.Anything)
			mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
			mockSdk := &mocks.ApplicationService{}
			mockSdk.On("RemoveAllFunctionPipelines")
			mockSdk.On("LoggingClient").Return(mockLogger)

			target := NewManager(mockSdk, 0).(*dataManager)

			if !test.RecordingPreviouslyCanceled {
				now := time.Now()
				target.recordingStartedAt = &now
			}

			continueExecution, actual := target.processBatchedData(nil, test.Data)
			if test.ExpectedError != nil {
				require.Equal(t, test.ExpectedError, actual)
				return
			}

			if test.RecordingPreviouslyCanceled {
				require.False(t, continueExecution)
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
