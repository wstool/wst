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
	"context"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/expect"
	"github.com/bukka/wst/run/actions/not"
	"github.com/bukka/wst/run/actions/parallel"
	"github.com/bukka/wst/run/actions/request"
	"github.com/bukka/wst/run/actions/restart"
	"github.com/bukka/wst/run/actions/start"
	"github.com/bukka/wst/run/actions/stop"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"time"
)

type Action interface {
	Execute(ctx context.Context, runData runtime.Data, dryRun bool) (bool, error)
	Timeout() time.Duration
}

type ActionMaker struct {
	fnd                 app.Foundation
	expectOutputMaker   *expect.OutputExpectationActionMaker
	expectResponseMaker *expect.ResponseExpectationActionMaker
	notMaker            *not.ActionMaker
	parallelMaker       *parallel.ActionMaker
	requestMaker        *request.ActionMaker
	restartMaker        *restart.ActionMaker
	startMaker          *start.ActionMaker
	stopMaker           *stop.ActionMaker
}

func CreateActionMaker(fnd app.Foundation) *ActionMaker {
	return &ActionMaker{
		fnd:                 fnd,
		expectOutputMaker:   expect.CreateOutputExpectationActionMaker(fnd),
		expectResponseMaker: expect.CreateResponseExpectationActionMaker(fnd),
		notMaker:            not.CreateActionMaker(fnd),
		parallelMaker:       parallel.CreateActionMaker(fnd),
		requestMaker:        request.CreateActionMaker(fnd),
		restartMaker:        restart.CreateActionMaker(fnd),
		startMaker:          start.CreateActionMaker(fnd),
		stopMaker:           stop.CreateActionMaker(fnd),
	}
}

func (m *ActionMaker) MakeAction(config types.Action, svcs services.Services, defaultTimeout int) (Action, error) {
	switch action := config.(type) {
	case *types.OutputExpectationAction:
		return m.expectOutputMaker.MakeAction(action, svcs, defaultTimeout)
	case *types.ResponseExpectationAction:
		return m.expectResponseMaker.MakeAction(action, svcs, defaultTimeout)
	case *types.NotAction:
		return m.notMaker.Make(action, svcs, defaultTimeout, m)
	case *types.ParallelAction:
		return m.parallelMaker.Make(action, svcs, defaultTimeout, m)
	case *types.RequestAction:
		return m.requestMaker.Make(action, svcs, defaultTimeout)
	case *types.RestartAction:
		return m.restartMaker.Make(action, svcs, defaultTimeout)
	case *types.StartAction:
		return m.startMaker.Make(action, svcs, defaultTimeout)
	case *types.StopAction:
		return m.stopMaker.Make(action, svcs, defaultTimeout)
	default:
		return nil, fmt.Errorf("unsupported action type: %T", config)
	}
}
