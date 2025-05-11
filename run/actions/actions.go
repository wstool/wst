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
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/actions/action/bench"
	"github.com/wstool/wst/run/actions/action/execute"
	"github.com/wstool/wst/run/actions/action/expect"
	"github.com/wstool/wst/run/actions/action/not"
	"github.com/wstool/wst/run/actions/action/parallel"
	"github.com/wstool/wst/run/actions/action/reload"
	"github.com/wstool/wst/run/actions/action/request"
	"github.com/wstool/wst/run/actions/action/restart"
	"github.com/wstool/wst/run/actions/action/sequential"
	"github.com/wstool/wst/run/actions/action/start"
	"github.com/wstool/wst/run/actions/action/stop"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/services"
)

type ActionMaker interface {
	MakeAction(config types.Action, sl services.ServiceLocator, defaultTimeout int) (action.Action, error)
}

type nativeActionMaker struct {
	fnd             app.Foundation
	runtimeMaker    runtime.Maker
	benchMaker      bench.Maker
	executeMaker    execute.Maker
	expectMaker     expect.Maker
	notMaker        not.Maker
	parallelMaker   parallel.Maker
	requestMaker    request.Maker
	reloadMaker     reload.Maker
	restartMaker    restart.Maker
	sequentialMaker sequential.Maker
	startMaker      start.Maker
	stopMaker       stop.Maker
}

func CreateActionMaker(
	fnd app.Foundation,
	expectationsMaker expectations.Maker,
	parametersMaker parameters.Maker,
	runtimeMaker runtime.Maker,
) ActionMaker {
	return &nativeActionMaker{
		fnd:             fnd,
		runtimeMaker:    runtimeMaker,
		benchMaker:      bench.CreateActionMaker(fnd),
		executeMaker:    execute.CreateActionMaker(fnd),
		expectMaker:     expect.CreateExpectationActionMaker(fnd, expectationsMaker, parametersMaker),
		notMaker:        not.CreateActionMaker(fnd, runtimeMaker),
		parallelMaker:   parallel.CreateActionMaker(fnd, runtimeMaker),
		requestMaker:    request.CreateActionMaker(fnd),
		reloadMaker:     reload.CreateActionMaker(fnd),
		restartMaker:    restart.CreateActionMaker(fnd),
		sequentialMaker: sequential.CreateActionMaker(fnd, runtimeMaker),
		startMaker:      start.CreateActionMaker(fnd),
		stopMaker:       stop.CreateActionMaker(fnd),
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
	case *types.ExecuteAction:
		return m.executeMaker.Make(action, sl, defaultTimeout)
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
	case *types.SequentialAction:
		return m.sequentialMaker.Make(action, sl, defaultTimeout, m)
	case *types.StartAction:
		return m.startMaker.Make(action, sl, defaultTimeout)
	case *types.StopAction:
		return m.stopMaker.Make(action, sl, defaultTimeout)
	default:
		return nil, errors.Errorf("unsupported action type: %T", config)
	}
}
