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
	"testing"

	appMocks "github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/edgexfoundry/app-record-replay/internal/interfaces/mocks"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	"github.com/stretchr/testify/assert"
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

func TestHttpController_AddRoutes(t *testing.T) {
	// TODO: Implement using TDD
}

func TestHttpController_StartRecording(t *testing.T) {
	// TODO: Implement using TDD
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
