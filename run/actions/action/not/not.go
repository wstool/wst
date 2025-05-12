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

package not

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
		config *types.NotAction,
		sl services.ServiceLocator,
		defaultTimeout int,
		actionMaker action.Maker,
	) (action.Action, error)
}

type ActionMaker struct {
	fnd          app.Foundation
	runtimeMaker runtime.Maker
}

func CreateActionMaker(fnd app.Foundation, runtimeMaker runtime.Maker) *ActionMaker {
	return &ActionMaker{
		fnd:          fnd,
		runtimeMaker: runtimeMaker,
	}
}

func (m *ActionMaker) Make(
	config *types.NotAction,
	sl services.ServiceLocator,
	defaultTimeout int,
	actionMaker action.Maker,
) (action.Action, error) {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	newAction, err := actionMaker.MakeAction(config.Action, sl, config.Timeout)
	if err != nil {
		return nil, err
	}

	return &Action{
		fnd:          m.fnd,
		runtimeMaker: m.runtimeMaker,
		action:       newAction,
		timeout:      time.Duration(config.Timeout * 1e6),
		when:         action.When(config.When),
		onFailure:    action.OnFailureType(config.OnFailure),
	}, nil
}

type Action struct {
	fnd          app.Foundation
	runtimeMaker runtime.Maker
	action       action.Action
	timeout      time.Duration
	when         action.When
	onFailure    action.OnFailureType
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
	actTimeout := a.action.Timeout()
	a.fnd.Logger().Infof("Executing not action with timeout %s", actTimeout)
	actCtx, cancel := a.runtimeMaker.MakeContextWithTimeout(ctx, actTimeout)
	defer cancel()
	success, err := a.action.Execute(actCtx, runData)
	if err != nil {
		return false, err
	}
	a.fnd.Logger().Infof("Executed action resulted to %t - inverting to %t", success, !success)
	if a.fnd.DryRun() {
		// always return success for dry run
		return true, nil
	}
	return !success, nil
}
