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
	expect2 "github.com/bukka/wst/run/actions/expect"
	"github.com/bukka/wst/run/actions/not"
	"github.com/bukka/wst/run/actions/parallel"
	"github.com/bukka/wst/run/actions/request"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
)

type Action interface {
	Execute(runData runtime.Data, dryRun bool) (bool, error)
}

type ActionMaker struct {
	env                 app.Env
	expectOutputMaker   *expect2.OutputExpectationActionMaker
	expectResponseMaker *expect2.ResponseExpectationActionMaker
	notMaker            *not.ActionMaker
	parallelMaker       *parallel.ActionMaker
	requestMaker        *request.ActionMaker
}

func CreateActionMaker(env app.Env) *ActionMaker {
	return &ActionMaker{
		env:                 env,
		expectOutputMaker:   expect2.CreateOutputExpectationActionMaker(env),
		expectResponseMaker: expect2.CreateResponseExpectationActionMaker(env),
		notMaker:            not.CreateActionMaker(env),
		parallelMaker:       parallel.CreateActionMaker(env),
		requestMaker:        request.CreateActionMaker(env),
	}
}

func (m *ActionMaker) MakeAction(config types.Action, svcs services.Services) (Action, error) {
	switch action := config.(type) {
	case *types.OutputExpectationAction:
		return m.expectOutputMaker.MakeAction(action, svcs)
	case *types.ResponseExpectationAction:
		return m.expectResponseMaker.MakeAction(action, svcs)
	case *types.NotAction:
		return m.notMaker.Make(action, svcs, m)
	case *types.ParallelAction:
		return m.parallelMaker.Make(action, svcs, m)
	case *types.RequestAction:
		return m.requestMaker.Make(action, svcs)
	default:
		return nil, fmt.Errorf("unsupported action type: %T", config)
	}
}
