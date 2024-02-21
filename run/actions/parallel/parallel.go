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

package parallel

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"sync"
)

type Action struct {
	Actions []actions.Action
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
	config *types.ParallelAction,
	svcs services.Services,
	defaultTimeout int,
	actionMaker *actions.ActionMaker,
) (*Action, error) {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	var actions []actions.Action
	for _, configAction := range config.Actions {
		action, err := actionMaker.MakeAction(configAction, svcs, config.Timeout)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return &Action{
		Actions: actions,
		Timeout: config.Timeout,
	}, nil
}

func (a Action) Execute(runData runtime.Data, dryRun bool) (bool, error) {
	// Use a WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup
	wg.Add(len(a.Actions))

	// Use an error channel to collect potential errors from actions.
	errs := make(chan error, len(a.Actions))

	for _, action := range a.Actions {
		go func(act actions.Action) {
			defer wg.Done()

			// Execute the action, passing the context.
			success, err := act.Execute(runData, dryRun)
			if err != nil || !success {
				errs <- err
			}
		}(action)
	}

	// Wait for all actions to complete.
	wg.Wait()
	close(errs)

	// Check if there were any errors.
	for err := range errs {
		if err != nil {
			return false, err
		}
	}

	return true, nil
}