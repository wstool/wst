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

type Maker interface {
	Make(
		config *types.ParallelAction,
		sl services.ServiceLocator,
		defaultTimeout int,
		actionMaker action.Maker,
	) (action.Action, error)
}

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
	actionMaker action.Maker,
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
		when:    action.When(config.When),
	}, nil
}

type Action struct {
	fnd     app.Foundation
	actions []action.Action
	timeout time.Duration
	when    action.When
}

func (a *Action) When() action.When {
	return a.when
}

func (a *Action) Timeout() time.Duration {
	return a.timeout
}

func (a *Action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	logger := a.fnd.Logger()
	logger.Infof("Executing parallel action")
	// Use a WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup
	wg.Add(len(a.actions))

	// Use an error channel to collect potential errors from actions.
	al := len(a.actions)
	errs := make(chan error, al)
	fails := make(chan int, al)

	for pos, act := range a.actions {
		go func(act action.Action, pos int) {
			defer wg.Done()

			// Execute the action, passing the context.
			logger.Debugf("Executing parallel action %d", pos)
			success, err := act.Execute(ctx, runData)
			if err != nil {
				errs <- fmt.Errorf("parallel action %d failed with error %v", pos, err)
			} else if !success {
				fails <- pos
			}
		}(act, pos)
	}

	// Wait for all actions to complete.
	wg.Wait()
	close(errs)
	close(fails)

	// Check if there were any errors.
	errActionsCount := 0
	for err := range errs {
		errActionsCount++
		logger.Errorf("Parallel execution error: %v", err)
	}
	if errActionsCount > 0 {
		multipleSuffix := ""
		if errActionsCount > 1 {
			multipleSuffix = "s"
		}
		return false, fmt.Errorf("failed %d parallel action%s", errActionsCount, multipleSuffix)
	}
	failedActionsCount := 0
	for range fails {
		failedActionsCount++
	}
	result := true
	if failedActionsCount > 0 {
		logger.Debugf("Executing parallel action failed on %d actions", failedActionsCount)
		if !a.fnd.DryRun() {
			result = false
		}
	}

	return result, nil
}
