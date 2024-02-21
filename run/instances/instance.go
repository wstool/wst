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
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/resources/scripts"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/services"
)

type Instance interface {
	ExecuteActions(dryRun bool) error
	Name() string
}

type InstanceMaker struct {
	env           app.Env
	actionMaker   *actions.ActionMaker
	servicesMaker *services.Maker
	scriptsMaker  *scripts.Maker
}

func CreateInstanceMaker(env app.Env) *InstanceMaker {
	return &InstanceMaker{
		env:           env,
		actionMaker:   actions.CreateActionMaker(env),
		servicesMaker: services.CreateMaker(env),
		scriptsMaker:  scripts.CreateMaker(env),
	}
}

func (m *InstanceMaker) Make(config types.Instance, srvs servers.Servers) (Instance, error) {
	scriptResources, err := m.scriptsMaker.Make(config.Resources.Scripts)
	if err != nil {
		return nil, err
	}

	svcs, err := m.servicesMaker.Make(config.Services, scriptResources, srvs)
	if err != nil {
		return nil, err
	}

	actions := make([]actions.Action, len(config.Actions))
	for i, actionConfig := range config.Actions {
		action, err := m.actionMaker.MakeAction(actionConfig, svcs)
		if err != nil {
			return nil, err
		}
		actions[i] = action
	}
	runData := runtime.CreateData()
	return &nativeInstance{
		name:    config.Name,
		actions: actions,
		runData: runData,
	}, nil
}

type nativeInstance struct {
	name    string
	actions []actions.Action
	runData runtime.Data
}

func (i *nativeInstance) Name() string {
	return i.name
}

func (i *nativeInstance) ExecuteActions(dryRun bool) error {
	for _, action := range i.actions {
		success, err := action.Execute(i.runData, dryRun)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("action execution failed")
		}
	}
	return nil
}
