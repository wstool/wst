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

package instances

import (
	"context"
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/environments"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/resources/scripts"
	"github.com/wstool/wst/run/servers"
	"github.com/wstool/wst/run/services"
	"github.com/wstool/wst/run/spec/defaults"
	"path/filepath"
	"time"
)

type Instance interface {
	Run() error
	Name() string
	Workspace() string
}

type InstanceMaker interface {
	Make(
		instanceConfig types.Instance,
		instanceId int,
		envsConfig map[string]types.Environment,
		dflts *defaults.Defaults,
		srvs servers.Servers,
		specWorkspace string,
	) (Instance, error)
}

type nativeInstanceMaker struct {
	fnd              app.Foundation
	actionMaker      actions.ActionMaker
	servicesMaker    services.Maker
	scriptsMaker     scripts.Maker
	environmentMaker environments.Maker
	runtimeMaker     runtime.Maker
}

func CreateInstanceMaker(
	fnd app.Foundation,
	expectationsMaker expectations.Maker,
	parametersMaker parameters.Maker,
) InstanceMaker {
	runtimeMaker := runtime.CreateMaker(fnd)
	return &nativeInstanceMaker{
		fnd:              fnd,
		actionMaker:      actions.CreateActionMaker(fnd, expectationsMaker, parametersMaker, runtimeMaker),
		servicesMaker:    services.CreateMaker(fnd, parametersMaker),
		scriptsMaker:     scripts.CreateMaker(fnd, parametersMaker),
		environmentMaker: environments.CreateMaker(fnd),
		runtimeMaker:     runtimeMaker,
	}
}

func (m *nativeInstanceMaker) Make(
	instanceConfig types.Instance,
	instanceIdx int,
	envsConfig map[string]types.Environment,
	dflts *defaults.Defaults,
	srvs servers.Servers,
	specWorkspace string,
) (Instance, error) {
	scrpts, err := m.scriptsMaker.Make(instanceConfig.Resources.Scripts)
	if err != nil {
		return nil, err
	}

	name := instanceConfig.Name
	instanceWs := filepath.Join(specWorkspace, name)

	envs, err := m.environmentMaker.Make(envsConfig, instanceConfig.Environments, instanceWs)
	if err != nil {
		return nil, err
	}

	sl, err := m.servicesMaker.Make(instanceConfig.Services, dflts, scrpts, srvs, envs, name, instanceIdx, instanceWs)
	if err != nil {
		return nil, err
	}

	actTimeout := instanceConfig.Timeouts.Action
	if actTimeout == 0 {
		actTimeout = dflts.Timeouts.Action
	}

	instanceActions := make([]action.Action, len(instanceConfig.Actions))
	var act action.Action
	for i, actionConfig := range instanceConfig.Actions {
		act, err = m.actionMaker.MakeAction(actionConfig, sl, actTimeout)
		if err != nil {
			return nil, err
		}
		instanceActions[i] = act
	}

	instanceTimeout := instanceConfig.Timeouts.Actions
	if instanceTimeout == 0 {
		instanceTimeout = dflts.Timeouts.Actions
	}

	runData := m.runtimeMaker.MakeData()
	return &nativeInstance{
		fnd:          m.fnd,
		runtimeMaker: m.runtimeMaker,
		name:         name,
		index:        instanceIdx,
		timeout:      time.Duration(instanceTimeout) * time.Millisecond,
		actions:      instanceActions,
		envs:         envs,
		runData:      runData,
		workspace:    instanceWs,
	}, nil
}

type nativeInstance struct {
	fnd          app.Foundation
	runtimeMaker runtime.Maker
	name         string
	index        int
	actions      []action.Action
	envs         environments.Environments
	runData      runtime.Data
	timeout      time.Duration
	workspace    string
}

func (i *nativeInstance) Workspace() string {
	return i.workspace
}

func (i *nativeInstance) Name() string {
	return i.name
}

func (i *nativeInstance) Run() error {
	var err error

	fs := i.fnd.Fs()
	if err = fs.RemoveAll(i.workspace); err != nil {
		return errors.Errorf("failed to remove previous workspace for instance %s: %v", i.name, err)
	}

	ctx := i.runtimeMaker.MakeBackgroundContext()
	initializedEnvs := make(map[providers.Type]bool)
	for envName, env := range i.envs {
		if env.IsUsed() {
			i.fnd.Logger().Debugf("Initializing %s environment", envName)
			if err = env.Init(ctx); err != nil {
				i.fnd.Logger().Debugf("Failed to initialize %s environment", envName)
				_ = i.destroyEnvironments(ctx, initializedEnvs)
				return err
			}
			initializedEnvs[envName] = true
		}
	}

	ictx, cancel := i.runtimeMaker.MakeContextWithTimeout(ctx, i.timeout)
	defer cancel()
	var actionErr error = nil
	for pos, act := range i.actions {
		i.fnd.Logger().Debugf("Executing action number %d with timeout %s", pos, i.timeout)
		actionErr = i.executeAction(ictx, act, actionErr)
	}

	destroyErr := i.destroyEnvironments(ctx, initializedEnvs)
	if actionErr == nil {
		return destroyErr
	}

	return actionErr
}

func (i *nativeInstance) destroyEnvironments(ctx context.Context, initializedEnvs map[providers.Type]bool) error {
	var err error
	for envName := range initializedEnvs {
		env := i.envs[envName]
		i.fnd.Logger().Debugf("Destroying %s environment", envName)
		if destroyErr := env.Destroy(ctx); destroyErr != nil {
			err = destroyErr
		}
	}
	return err
}

func (i *nativeInstance) executeAction(actionsCtx context.Context, act action.Action, actErr error) error {
	if actErr != nil && act.When() == action.OnSuccess {
		return actErr
	}
	if actErr == nil && act.When() == action.OnFailure {
		return nil
	}
	ctx, cancel := i.runtimeMaker.MakeContextWithTimeout(actionsCtx, act.Timeout())
	defer cancel()
	success, err := act.Execute(ctx, i.runData)
	if err != nil {
		i.fnd.Logger().Errorf("Failed to to run action: %v", err)
		if actErr != nil {
			return actErr
		}
		return err
	}
	if !success {
		return errors.Errorf("action execution failed")
	}
	return actErr
}
