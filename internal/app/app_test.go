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

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
)

// This is an example of how to test the code that would typically be in the main() function use mocks
// Not to helpful for a simple main() , but can be if the main() has more complexity that should be unit tested

func TestCreateAndRunService_Success(t *testing.T) {
	app := recordReplayApp{}

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("Run").Return(nil)
		return mockAppService, true
	}

	expected := 0
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_NewService_Failed(t *testing.T) {
	app := recordReplayApp{}

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		return nil, false
	}
	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_Run_Failed(t *testing.T) {
	app := recordReplayApp{}

	// ensure failure is from Run
	RunCalled := false

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("Run").Return(fmt.Errorf("Failed")).Run(func(args mock.Arguments) {
			RunCalled = true
		})

		return mockAppService, true
	}

	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	require.True(t, RunCalled, "Run never called")
	assert.Equal(t, expected, actual)
}
