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
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/environments"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/resources/scripts"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/services"
	"github.com/pkg/errors"
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
		envsConfig map[string]types.Environment,
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
	return &nativeInstanceMaker{
		fnd:              fnd,
		actionMaker:      actions.CreateActionMaker(fnd, expectationsMaker, parametersMaker),
		servicesMaker:    services.CreateMaker(fnd, parametersMaker),
		scriptsMaker:     scripts.CreateMaker(fnd, parametersMaker),
		environmentMaker: environments.CreateMaker(fnd),
		runtimeMaker:     runtime.CreateMaker(fnd),
	}
}

func (m *nativeInstanceMaker) Make(
	instanceConfig types.Instance,
	envsConfig map[string]types.Environment,
	srvs servers.Servers,
	specWorkspace string,
) (Instance, error) {
	scriptResources, err := m.scriptsMaker.Make(instanceConfig.Resources.Scripts)
	if err != nil {
		return nil, err
	}

	name := instanceConfig.Name
	instanceWorkspace := filepath.Join(specWorkspace, name)

	envs, err := m.environmentMaker.Make(envsConfig, instanceConfig.Environments, instanceWorkspace)
	if err != nil {
		return nil, err
	}

	sl, err := m.servicesMaker.Make(instanceConfig.Services, scriptResources, srvs, envs, name, instanceWorkspace)
	if err != nil {
		return nil, err
	}

	instanceActions := make([]action.Action, len(instanceConfig.Actions))
	for i, actionConfig := range instanceConfig.Actions {
		act, err := m.actionMaker.MakeAction(actionConfig, sl, instanceConfig.Timeouts.Action)
		if err != nil {
			return nil, err
		}
		instanceActions[i] = act
	}
	runData := m.runtimeMaker.MakeData()
	return &nativeInstance{
		fnd:          m.fnd,
		runtimeMaker: m.runtimeMaker,
		name:         name,
		timeout:      time.Duration(instanceConfig.Timeouts.Actions * 1e6),
		actions:      instanceActions,
		envs:         envs,
		runData:      runData,
		workspace:    instanceWorkspace,
	}, nil
}

type nativeInstance struct {
	fnd          app.Foundation
	runtimeMaker runtime.Maker
	name         string
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
	for pos, act := range i.actions {
		i.fnd.Logger().Debugf("Executing action number %d", pos)
		if err = i.executeAction(ictx, act); err != nil {
			break
		}
	}

	destroyErr := i.destroyEnvironments(ctx, initializedEnvs)
	if err == nil {
		err = destroyErr
	}

	return err
}

func (i *nativeInstance) destroyEnvironments(ctx context.Context, initializedEnvs map[providers.Type]bool) error {
	var err error
	for envName := range initializedEnvs {
		env := i.envs[envName]
		i.fnd.Logger().Debugf("Destroying %s environment", envName)
		if destroyErr := env.Destroy(ctx); destroyErr != nil {
			i.fnd.Logger().Errorf("Failed to destroy %s environment: %v", envName, err)
			err = destroyErr
		}
	}
	return err
}

func (i *nativeInstance) executeAction(actionsCtx context.Context, action action.Action) error {
	ctx, cancel := i.runtimeMaker.MakeContextWithTimeout(actionsCtx, action.Timeout())
	defer cancel()
	success, err := action.Execute(ctx, i.runData)
	if err != nil {
		return err
	}
	if !success {
		return errors.Errorf("action execution failed")
	}
	return nil
}
