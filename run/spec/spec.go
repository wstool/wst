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

package spec

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/instances"
	"github.com/bukka/wst/run/servers"
	"strings"
)

type Spec interface {
	ExecuteInstances(filteredInstances []string, dryRun bool) error
}

type Maker struct {
	env           app.Env
	instanceMaker *instances.InstanceMaker
	serversMaker  *servers.Maker
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env:           env,
		instanceMaker: instances.CreateInstanceMaker(env),
		serversMaker:  servers.CreateMaker(env),
	}
}

func (m *Maker) Make(config *types.Spec) (Spec, error) {
	serversMap, err := m.serversMaker.Make(config)
	if err != nil {
		return nil, err
	}

	var instances []instances.Instance
	for _, instance := range config.Instances {
		inst, err := m.instanceMaker.Make(instance, config.Environments, serversMap)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}

	return &nativeSpec{
		workspace: config.Workspace,
		instances: instances,
	}, nil
}

type nativeSpec struct {
	workspace string
	instances []instances.Instance
}

func (n nativeSpec) ExecuteInstances(filteredInstances []string, dryRun bool) error {
	// Loop through the instances.
	for _, instance := range n.instances {
		// Determine the instance identifier or name.
		instanceName := instance.Name()

		// Execute if filteredInstances is empty or nil, meaning execute all instances.
		if len(filteredInstances) == 0 {
			if err := instance.ExecuteActions(dryRun); err != nil {
				return err // Return immediately if any execution fails.
			}
		} else {
			// If filteredInstances is not empty, check if the instance name starts with any of the filteredInstances strings.
			for _, filter := range filteredInstances {
				if strings.HasPrefix(instanceName, filter) {
					// Execute the instance if it matches the filter.
					if err := instance.ExecuteActions(dryRun); err != nil {
						return err // Return immediately if any execution fails.
					}
					break // Move to the next instance after successful execution.
				}
			}
		}
	}

	return nil // Return nil if all selected instances were executed successfully.
}
