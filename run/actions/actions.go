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

package actions

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/actions/action/bench"
	"github.com/bukka/wst/run/actions/action/expect"
	"github.com/bukka/wst/run/actions/action/not"
	"github.com/bukka/wst/run/actions/action/parallel"
	"github.com/bukka/wst/run/actions/action/reload"
	"github.com/bukka/wst/run/actions/action/request"
	"github.com/bukka/wst/run/actions/action/restart"
	"github.com/bukka/wst/run/actions/action/start"
	"github.com/bukka/wst/run/actions/action/stop"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
)

type ActionMaker interface {
	MakeAction(config types.Action, sl services.ServiceLocator, defaultTimeout int) (action.Action, error)
}

type nativeActionMaker struct {
	fnd           app.Foundation
	benchMaker    bench.Maker
	expectMaker   expect.Maker
	notMaker      not.Maker
	parallelMaker parallel.Maker
	requestMaker  request.Maker
	reloadMaker   reload.Maker
	restartMaker  restart.Maker
	startMaker    start.Maker
	stopMaker     stop.Maker
}

func CreateActionMaker(
	fnd app.Foundation,
	expectationsMaker expectations.Maker,
	parametersMaker parameters.Maker,
) ActionMaker {
	return &nativeActionMaker{
		fnd:           fnd,
		benchMaker:    bench.CreateActionMaker(fnd),
		expectMaker:   expect.CreateExpectationActionMaker(fnd, expectationsMaker, parametersMaker),
		notMaker:      not.CreateActionMaker(fnd),
		parallelMaker: parallel.CreateActionMaker(fnd),
		requestMaker:  request.CreateActionMaker(fnd),
		reloadMaker:   reload.CreateActionMaker(fnd),
		restartMaker:  restart.CreateActionMaker(fnd),
		startMaker:    start.CreateActionMaker(fnd),
		stopMaker:     stop.CreateActionMaker(fnd),
	}
}

func (m *nativeActionMaker) MakeAction(
	config types.Action,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	switch action := config.(type) {
	case *types.BenchAction:
		return m.benchMaker.Make(action, sl, defaultTimeout)
	case *types.CustomExpectationAction:
		return m.expectMaker.MakeCustomAction(action, sl, defaultTimeout)
	case *types.MetricsExpectationAction:
		return m.expectMaker.MakeMetricsAction(action, sl, defaultTimeout)
	case *types.OutputExpectationAction:
		return m.expectMaker.MakeOutputAction(action, sl, defaultTimeout)
	case *types.ResponseExpectationAction:
		return m.expectMaker.MakeResponseAction(action, sl, defaultTimeout)
	case *types.NotAction:
		return m.notMaker.Make(action, sl, defaultTimeout, m)
	case *types.ParallelAction:
		return m.parallelMaker.Make(action, sl, defaultTimeout, m)
	case *types.RequestAction:
		return m.requestMaker.Make(action, sl, defaultTimeout)
	case *types.ReloadAction:
		return m.reloadMaker.Make(action, sl, defaultTimeout)
	case *types.RestartAction:
		return m.restartMaker.Make(action, sl, defaultTimeout)
	case *types.StartAction:
		return m.startMaker.Make(action, sl, defaultTimeout)
	case *types.StopAction:
		return m.stopMaker.Make(action, sl, defaultTimeout)
	default:
		return nil, fmt.Errorf("unsupported action type: %T", config)
	}
}
