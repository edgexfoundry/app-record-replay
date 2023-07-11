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
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-record-replay/internal/controller"
	"github.com/edgexfoundry/app-record-replay/internal/data"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

type recordReplayApp struct {
	service interfaces.ApplicationService
	lc      logger.LoggingClient
}

func New() *recordReplayApp {
	return &recordReplayApp{}
}

func (app *recordReplayApp) CreateAndRunAppService(serviceKey string, newServiceFactory func(string) (interfaces.ApplicationService, bool)) int {
	var ok bool
	app.service, ok = newServiceFactory(serviceKey)
	if !ok {
		return -1
	}

	app.lc = app.service.LoggingClient()

	if err := controller.New(data.NewManager(app.service), app.service).AddRoutes(); err != nil {
		app.lc.Errorf("adding routes failed: %s", err.Error())
		return -1
	}

	if err := app.service.Run(); err != nil {
		app.lc.Errorf("running app service failed: %s", err.Error())
		return -1
	}

	return 0
}