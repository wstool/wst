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
	"context"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"sync"
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
	config *types.ParallelAction,
	sl services.ServiceLocator,
	defaultTimeout int,
	actionMaker action.ActionMaker,
) (action.Action, error) {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	var parallelActions []action.Action
	for _, configAction := range config.Actions {
		newAction, err := actionMaker.MakeAction(configAction, sl, config.Timeout)
		if err != nil {
			return nil, err
		}
		parallelActions = append(parallelActions, newAction)
	}
	return &Action{
		fnd:     m.fnd,
		actions: parallelActions,
		timeout: time.Duration(config.Timeout * 1e6),
	}, nil
}

type Action struct {
	fnd     app.Foundation
	actions []action.Action
	timeout time.Duration
}

func (a *Action) Timeout() time.Duration {
	return a.timeout
}

func (a *Action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing parallel action")
	// Use a WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup
	wg.Add(len(a.actions))

	// Use an error channel to collect potential errors from actions.
	errs := make(chan error, len(a.actions))

	for pos, act := range a.actions {
		go func(act action.Action, pos int) {
			defer wg.Done()

			// Execute the action, passing the context.
			a.fnd.Logger().Debugf("Executing parallel action %d", pos)
			success, err := act.Execute(ctx, runData)
			if err != nil || !success {
				errs <- fmt.Errorf("parallel action %d failed with error %v", pos, err)
			}
		}(act, pos)
	}

	// Wait for all actions to complete.
	wg.Wait()
	close(errs)

	// Check if there were any errors.
	failedActionsCount := 0
	for err := range errs {
		if err != nil {
			failedActionsCount++
			a.fnd.Logger().Errorf("Parallel execution error: %v", err)
		}
	}
	if failedActionsCount > 0 {
		multipleSuffix := ""
		if failedActionsCount > 1 {
			multipleSuffix = "s"
		}
		return false, fmt.Errorf("failed %d parallel action%s", failedActionsCount, multipleSuffix)
	}

	return true, nil
}
