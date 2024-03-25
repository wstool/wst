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

package bench

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
	config *types.BenchAction,
	svcs services.Services,
	defaultTimeout int,
) (actions.Action, error) {
	svc, err := svcs.FindService(config.Service)
	if err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		if defaultTimeout > config.Duration {
			config.Timeout = defaultTimeout
		} else {
			config.Timeout = config.Duration + 5000
		}
	}

	return &action{
		fnd:      m.fnd,
		service:  svc,
		timeout:  time.Duration(config.Timeout),
		duration: time.Duration(config.Duration),
		rate:     config.Rate,
		id:       config.Id,
		path:     config.Path,
		method:   config.Method,
		headers:  config.Headers,
	}, nil
}

type action struct {
	fnd      app.Foundation
	service  services.Service
	timeout  time.Duration
	duration time.Duration
	rate     int
	id       string
	path     string
	method   string
	headers  types.Headers
}

func (a *action) Timeout() time.Duration {
	return a.timeout
}

func (a *action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing bench action")
	// TODO: implementation
	return true, nil
}
