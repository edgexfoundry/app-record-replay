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
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/stretchr/testify/assert"
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
	// TODO: Implement using TDD
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
		{"Valid", dtos.Event{}, 5, nil},
		{"Nil data", nil, 1, countsNoDataError},
		{"Not Event", dtos.Metric{}, 1, countsDataNotEventError},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			target := NewManager(nil).(*dataManager)
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
	expectedBatchedEvents := []dtos.Event{
		dtos.NewEvent("test-profile1", "test-device1", "test-source1"),
		dtos.NewEvent("test-profile2", "test-device2", "test-source2"),
		dtos.NewEvent("test-profile3", "test-device3", "test-source3"),
		dtos.NewEvent("test-profile4", "test-device4", "test-source4"),
	}

	tests := []struct {
		Name          string
		Data          any
		ExpectedError error
	}{
		{"Valid", expectedBatchedEvents, nil},
		{"Nil data", nil, batchNoDataError},
		{"Not Collection of Events", dtos.Event{}, batchDataNotEventCollectionError},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockSdk := &mocks.ApplicationService{}
			mockSdk.On("RemoveAllFunctionPipelines")

			target := NewManager(mockSdk).(*dataManager)
			target.recordingStartedAt = time.Now()

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
