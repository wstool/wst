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

package sequential

import (
	"context"
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/services"
	"time"
)

type Maker interface {
	Make(
		config *types.SequentialAction,
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
	config *types.SequentialAction,
	sl services.ServiceLocator,
	defaultTimeout int,
	actionMaker action.Maker,
) (action.Action, error) {
	var actions []types.Action
	if config.Name == "" {
		actions = config.Actions
	} else {
		srv, err := sl.Find(config.Service)
		if err != nil {
			return nil, errors.Errorf("sequential action service not found: %v", err)
		}
		seqAct, ok := srv.Server().SequentialAction(config.Name)
		if !ok {
			return nil, errors.Errorf("sequential action %s not found", config.Name)
		}
		actions = seqAct.Actions()
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	var sequentialActions []action.Action
	for _, configAction := range actions {
		newAction, err := actionMaker.MakeAction(configAction, sl, config.Timeout)
		if err != nil {
			return nil, err
		}
		sequentialActions = append(sequentialActions, newAction)
	}
	return &Action{
		fnd:          m.fnd,
		runtimeMaker: m.runtimeMaker,
		actions:      sequentialActions,
		timeout:      time.Duration(config.Timeout * 1e6),
		when:         action.When(config.When),
		onFailure:    action.OnFailureType(config.OnFailure),
	}, nil
}

type Action struct {
	fnd          app.Foundation
	runtimeMaker runtime.Maker
	actions      []action.Action
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
	logger := a.fnd.Logger()
	logger.Infof("Executing sequential action")

	failedActionsCount := 0
	var lastErr error = nil
	for pos, act := range a.actions {
		// Execute action with context
		when := act.When()
		if when == action.Always ||
			(failedActionsCount == 0 && when == action.OnSuccess) ||
			(failedActionsCount > 0 && when == action.OnFailure) {
			actTimeout := act.Timeout()
			logger.Debugf("Executing sequential action %d with timeout %s", pos, actTimeout)
			// Create context for action
			actCtx, cancel := a.runtimeMaker.MakeContextWithTimeout(ctx, actTimeout)
			success, err := act.Execute(actCtx, runData)
			cancel() // Cancel the context immediately after action completion

			if err != nil {
				lastErr = err
				logger.Errorf("Sequential action %d failed with error: %v", pos, err)
			}

			if !success {
				failedActionsCount++
				logger.Debugf("Sequential action %d failed", pos)
			}
		}
	}

	if a.fnd.DryRun() {
		return true, nil
	}

	result := true
	if failedActionsCount > 0 {
		logger.Debugf("Sequential action failed on %d actions", failedActionsCount)
		result = false
	}

	if lastErr != nil {
		return result, errors.Errorf("Sequential action failed with error: %v", lastErr)
	}

	return result, lastErr
}
