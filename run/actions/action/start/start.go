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

package start

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"time"
)

type Maker interface {
	Make(
		config *types.StartAction,
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
	config *types.StartAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	var startServices services.Services

	if config.Service != "" {
		config.Services = append(config.Services, config.Service)
	}

	if len(config.Services) > 0 {
		startServices = make(services.Services, len(config.Services))
		for _, configService := range config.Services {
			svc, err := sl.Find(configService)
			if err != nil {
				return nil, err
			}
			startServices.AddService(svc)
		}
	} else {
		startServices = sl.Services()
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return &Action{
		fnd:      m.fnd,
		services: startServices,
		timeout:  time.Duration(config.Timeout) * time.Millisecond,
	}, nil
}

type Action struct {
	fnd      app.Foundation
	services services.Services
	timeout  time.Duration
}

func (a *Action) Timeout() time.Duration {
	return a.timeout
}

func (a *Action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing start action")
	for _, svc := range a.services {
		a.fnd.Logger().Debugf("Starting service %s", svc.Name())
		err := svc.Start(ctx)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
