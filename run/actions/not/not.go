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
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
)

type Action struct {
	Action  actions.Action
	Timeout int
}

type ActionMaker struct {
	env app.Env
}

func CreateActionMaker(env app.Env) *ActionMaker {
	return &ActionMaker{
		env: env,
	}
}

func (m *ActionMaker) Make(
	config *types.NotAction,
	svcs services.Services,
	defaultTimeout int,
	actionMaker *actions.ActionMaker,
) (*Action, error) {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	action, err := actionMaker.MakeAction(config.Action, svcs, config.Timeout)
	if err != nil {
		return nil, err
	}

	return &Action{
		Action:  action,
		Timeout: config.Timeout,
	}, nil
}

func (a Action) Execute(runData runtime.Data, dryRun bool) (bool, error) {
	success, err := a.Action.Execute(runData, dryRun)
	if err != nil {
		return false, err
	}
	return !success, nil
}
