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
//

package app

import (
	"fmt"
	"testing"

	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
)

// This is an example of how to test the code that would typically be in the main() function use mocks
// Not to helpful for a simple main() , but can be if the main() has more complexity that should be unit tested

func TestCreateAndRunService_Success(t *testing.T) {
	app := New()

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.Mock.On("ApplicationSettings").Return(map[string]string{MaxReplayDelayAppSetting: "1s"})
		mockAppService.On("DeviceClient").Return(&clientMocks.DeviceClient{})
		mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mockAppService.On("Run").Return(nil)
		return mockAppService, true
	}

	expected := 0
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_NewService_Failed(t *testing.T) {
	app := New()

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		return nil, false
	}
	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_MaxReplayDelayAppSetting_Failed(t *testing.T) {
	app := New()

	mockLogger := &loggerMocks.LoggingClient{}
	mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything)

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(mockLogger)
		mockAppService.Mock.On("ApplicationSettings").Return(map[string]string{MaxReplayDelayAppSetting: "junk"})
		mockAppService.On("DeviceClient").Return(&clientMocks.DeviceClient{})
		return mockAppService, true
	}

	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
	mockLogger.AssertExpectations(t)
}

func TestCreateAndRunService_DeviceClient_Failed(t *testing.T) {
	app := New()

	// ensure failure is from DeviceClient()
	DeviceClientCalled := false

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("DeviceClient").Return(nil).Run(func(args mock.Arguments) {
			DeviceClientCalled = true
		})

		return mockAppService, true
	}

	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
	require.True(t, DeviceClientCalled, "DeviceClient never called")
}

func TestCreateAndRunService_Run_Failed(t *testing.T) {
	app := New()

	// ensure failure is from Run
	RunCalled := false

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.Mock.On("ApplicationSettings").Return(map[string]string{MaxReplayDelayAppSetting: "1s"})
		mockAppService.On("DeviceClient").Return(&clientMocks.DeviceClient{})
		mockAppService.On("AddRoute", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mockAppService.On("Run").Return(fmt.Errorf("failed")).Run(func(args mock.Arguments) {
			RunCalled = true
		})

		return mockAppService, true
	}

	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	require.True(t, RunCalled, "Run never called")
	assert.Equal(t, expected, actual)
}
