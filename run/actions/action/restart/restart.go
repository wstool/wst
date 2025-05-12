// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package restart

import (
	"context"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/services"
	"time"
)

type Maker interface {
	Make(
		config *types.RestartAction,
		sl services.ServiceLocator,
		defaultTimeout int,
	) (action.Action, error)
}

type ActionMaker struct {
	fnd app.Foundation
}

func CreateActionMaker(fnd app.Foundation) *ActionMaker {
	return &ActionMaker{
		fnd: fnd,
	}
}

func (m *ActionMaker) Make(
	config *types.RestartAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	var restartServices services.Services

	if config.Service != "" {
		config.Services = append(config.Services, config.Service)
	}

	if len(config.Services) > 0 {
		restartServices = make(services.Services, len(config.Services))
		for _, configService := range config.Services {
			svc, err := sl.Find(configService)
			if err != nil {
				return nil, err
			}
			restartServices.AddService(svc)
		}
	} else {
		restartServices = sl.Services()
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return &Action{
		fnd:       m.fnd,
		services:  restartServices,
		timeout:   time.Duration(config.Timeout * 1e6),
		when:      action.When(config.When),
		onFailure: action.OnFailureType(config.OnFailure),
	}, nil
}

type Action struct {
	fnd       app.Foundation
	services  services.Services
	timeout   time.Duration
	when      action.When
	onFailure action.OnFailureType
}

func (a *Action) When() action.When {
	return a.when
}

func (a *Action) OnFailure() action.OnFailureType {
	return a.onFailure
}

func (a *Action) Timeout() time.Duration {
	return a.timeout
}

func (a *Action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing restart action")
	for _, svc := range a.services {
		a.fnd.Logger().Debugf("Restarting service %s", svc.Name())
		err := svc.Restart(ctx)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
