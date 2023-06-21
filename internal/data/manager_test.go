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
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDefaultDataManger(t *testing.T) {
	target := CreateDefaultDataManger(context.Background(), &mocks.ApplicationService{})
	require.NotNil(t, target)
	d := target.(*defaultDataManager)
	require.NotNil(t, d)
	assert.NotNil(t, d.dataChan)
	assert.NotNil(t, d.ctx)
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
	// TODO: Implement using TDD
	target := defaultDataManager{}
	target.countEvents(nil, nil)
}

func TestDefaultDataManager_ProcessBatchedData(t *testing.T) {
	// TODO: Implement using TDD
	target := defaultDataManager{}
	target.processBatchedData(nil, nil)
}
