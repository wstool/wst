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

package reload

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"time"
)

type ActionMaker struct {
	fnd app.Foundation
}

func CreateActionMaker(fnd app.Foundation) *ActionMaker {
	return &ActionMaker{
		fnd: fnd,
	}
}

func (m *ActionMaker) Make(
	config *types.ReloadAction,
	svcs services.Services,
	defaultTimeout int,
) (actions.Action, error) {
	var reloadServices services.Services

	if config.Service != "" {
		config.Services = append(config.Services, config.Service)
	}

	if len(config.Services) > 0 {
		for _, configService := range config.Services {
			svc, err := svcs.FindService(configService)
			if err != nil {
				return nil, err
			}
			err = reloadServices.AddService(svc)
			if err != nil {
				return nil, err
			}
		}
	} else {
		reloadServices = svcs
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return &action{
		fnd:      m.fnd,
		services: reloadServices,
		timeout:  time.Duration(config.Timeout * 1e6),
	}, nil
}

type action struct {
	fnd      app.Foundation
	services services.Services
	timeout  time.Duration
}

func (a *action) Timeout() time.Duration {
	return a.timeout
}

func (a *action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing reload action")
	for _, svc := range a.services {
		a.fnd.Logger().Debugf("Reloading service %s", svc.Name())
		err := svc.Reload(ctx)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
