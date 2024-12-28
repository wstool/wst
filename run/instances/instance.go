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
	IsChild() bool
	IsAbstract() bool
	Extend(instsMap map[string]Instance) error
	Init() error
	Parameters() parameters.Parameters
	InstanceTimeout() time.Duration
	ActionTimeout() int
	ConfigActions() []types.Action
	ConfigServices() map[string]types.Service
	ConfigInstanceEnvs() map[string]types.Environment
	ConfigResources() types.Resources
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
	parametersMaker  parameters.Maker
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
		parametersMaker:  parametersMaker,
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
	name := instanceConfig.Name

	instanceTimeout := instanceConfig.Timeouts.Actions
	instanceTimeoutDefault := instanceTimeout == 0
	if instanceTimeoutDefault {
		instanceTimeout = dflts.Timeouts.Actions
	}

	actionTimeout := instanceConfig.Timeouts.Action
	actionTimeoutDefault := actionTimeout == 0
	if actionTimeoutDefault {
		actionTimeout = dflts.Timeouts.Action
	}

	var err error
	var extendParams parameters.Parameters
	extendName := instanceConfig.Extends.Name
	if extendName != "" {
		extendParams, err = m.parametersMaker.Make(instanceConfig.Extends.Parameters)
		if err != nil {
			return nil, err
		}
	}

	params, err := m.parametersMaker.Make(instanceConfig.Parameters)
	if err != nil {
		return nil, err
	}

	runData := m.runtimeMaker.MakeData()
	return &nativeInstance{
		fnd:                    m.fnd,
		runtimeMaker:           m.runtimeMaker,
		environmentMaker:       m.environmentMaker,
		actionMaker:            m.actionMaker,
		scriptsMaker:           m.scriptsMaker,
		servicesMaker:          m.servicesMaker,
		configActions:          instanceConfig.Actions,
		configServices:         instanceConfig.Services,
		configEnvs:             envsConfig,
		configInstanceEnvs:     instanceConfig.Environments,
		configResources:        instanceConfig.Resources,
		name:                   name,
		index:                  instanceIdx,
		specWorkspace:          specWorkspace,
		abstract:               instanceConfig.Abstract,
		extendName:             extendName,
		extendParams:           extendParams,
		params:                 params,
		instanceTimeout:        time.Duration(instanceTimeout) * time.Millisecond,
		instanceTimeoutDefault: instanceTimeoutDefault,
		actionTimeout:          actionTimeout,
		actionTimeoutDefault:   actionTimeoutDefault,
		runData:                runData,
		servers:                srvs,
		defaults:               dflts,
	}, nil
}

type nativeInstance struct {
	fnd                app.Foundation
	runtimeMaker       runtime.Maker
	scriptsMaker       scripts.Maker
	environmentMaker   environments.Maker
	servicesMaker      services.Maker
	actionMaker        actions.ActionMaker
	configActions      []types.Action
	configServices     map[string]types.Service
	configEnvs         map[string]types.Environment
	configInstanceEnvs map[string]types.Environment
	configResources    types.Resources
	// Make runtime fields
	name             string
	index            int
	specWorkspace    string
	initialized      bool
	abstract         bool
	extendingStarted bool
	extendName       string
	extendParams     parameters.Parameters
	params           parameters.Parameters
	defaults         *defaults.Defaults
	servers          servers.Servers
	// Init runtime fields
	actions                []action.Action
	services               services.Services
	envs                   environments.Environments
	runData                runtime.Data
	instanceTimeout        time.Duration
	instanceTimeoutDefault bool
	actionTimeout          int
	actionTimeoutDefault   bool
	workspace              string
}

func (i *nativeInstance) InstanceTimeout() time.Duration {
	return i.instanceTimeout
}

func (i *nativeInstance) ActionTimeout() int {
	return i.actionTimeout
}

func (i *nativeInstance) ConfigActions() []types.Action {
	return i.configActions
}

func (i *nativeInstance) ConfigServices() map[string]types.Service {
	return i.configServices
}

func (i *nativeInstance) ConfigInstanceEnvs() map[string]types.Environment {
	return i.configInstanceEnvs
}

func (i *nativeInstance) ConfigResources() types.Resources {
	return i.configResources
}

func (i *nativeInstance) Parameters() parameters.Parameters {
	return i.params
}

func (i *nativeInstance) IsChild() bool {
	return i.extendName != ""
}

func (i *nativeInstance) IsAbstract() bool {
	return i.abstract
}

func (i *nativeInstance) Extend(instsMap map[string]Instance) error {
	// Do nothing for non child instance
	if i.extendName == "" {
		return nil
	}
	// Skip if all defined
	if len(i.configActions) == 0 && len(i.configInstanceEnvs) == 0 && len(i.configResources.Scripts) == 0 &&
		len(i.configServices) == 0 && !i.instanceTimeoutDefault && !i.actionTimeoutDefault {
		return nil
	}
	// Make sure there is no circular extending
	if i.extendingStarted {
		return errors.Errorf("instance %s already extending", i.name)
	}
	i.extendingStarted = true
	extendInst, ok := instsMap[i.extendName]
	if !ok {
		return errors.Errorf("failed to extend instance %s: instance %s not found", i.name, i.extendName)
	}
	// Make sure parent is also extended
	if err := extendInst.Extend(instsMap); err != nil {
		return err
	}
	i.extendingStarted = false
	// Extend actions if not already defined
	if len(i.configActions) == 0 {
		i.configActions = extendInst.ConfigActions()
	}
	// Extend instance envs if not already defined
	if len(i.configInstanceEnvs) == 0 {
		i.configInstanceEnvs = extendInst.ConfigInstanceEnvs()
	}
	// Extend services if not already defined
	if len(i.configServices) == 0 {
		i.configServices = extendInst.ConfigServices()
	}
	// Extend resource script if not already defined
	if len(i.configResources.Scripts) == 0 {
		i.configResources.Scripts = extendInst.ConfigResources().Scripts
	}
	// Extend instance timeout if it was not explicitly defined (default used)
	if i.instanceTimeoutDefault {
		i.instanceTimeout = extendInst.InstanceTimeout()
		i.instanceTimeoutDefault = false
	}
	// Extend action timeout if it was not explicitly defined (default used)
	if i.actionTimeoutDefault {
		i.actionTimeout = extendInst.ActionTimeout()
		i.actionTimeoutDefault = false
	}
	// inherit parameters
	i.params.Inherit(i.extendParams).Inherit(extendInst.Parameters())
	return nil
}

func (i *nativeInstance) Init() error {
	scrpts, err := i.scriptsMaker.Make(i.configResources.Scripts)
	if err != nil {
		return err
	}

	i.workspace = filepath.Join(i.specWorkspace, i.name)

	envs, err := i.environmentMaker.Make(i.configEnvs, i.configInstanceEnvs, i.workspace)
	if err != nil {
		return err
	}
	i.envs = envs

	sl, err := i.servicesMaker.Make(
		i.configServices, i.defaults, scrpts, i.servers, envs, i.name, i.index, i.workspace, i.params)
	if err != nil {
		return err
	}
	i.services = sl.Services()

	instanceActions := make([]action.Action, len(i.configActions))
	var act action.Action
	for idx, actionConfig := range i.configActions {
		act, err = i.actionMaker.MakeAction(actionConfig, sl, i.actionTimeout)
		if err != nil {
			return err
		}
		instanceActions[idx] = act
	}
	i.actions = instanceActions

	i.initialized = true

	return nil
}

func (i *nativeInstance) Workspace() string {
	return i.workspace
}

func (i *nativeInstance) Name() string {
	return i.name
}

func (i *nativeInstance) Run() error {
	if i.abstract {
		return errors.Errorf("instance %s is abstract and cannot be run", i.name)
	}

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

	ictx, cancel := i.runtimeMaker.MakeContextWithTimeout(ctx, i.instanceTimeout)
	defer cancel()
	var actionErr error = nil
	for pos, act := range i.actions {
		i.fnd.Logger().Debugf("Executing action number %d with timeout %d", pos, i.instanceTimeout)
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
