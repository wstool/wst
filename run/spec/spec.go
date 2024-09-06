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
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/instances"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/spec/defaults"
	"github.com/pkg/errors"
	"strings"
)

type Spec interface {
	Run(filteredInstances []string) error
}

type Maker interface {
	Make(config *types.Spec) (Spec, error)
}

type nativeMaker struct {
	fnd           app.Foundation
	defaultsMaker defaults.Maker
	instanceMaker instances.InstanceMaker
	serversMaker  servers.Maker
}

func CreateMaker(fnd app.Foundation) Maker {
	parametersMaker := parameters.CreateMaker(fnd)
	expectationsMaker := expectations.CreateMaker(fnd, parametersMaker)
	return &nativeMaker{
		fnd:           fnd,
		defaultsMaker: defaults.CreateMaker(fnd, parametersMaker),
		instanceMaker: instances.CreateInstanceMaker(fnd, expectationsMaker, parametersMaker),
		serversMaker:  servers.CreateMaker(fnd, expectationsMaker, parametersMaker),
	}
}

func (m *nativeMaker) Make(config *types.Spec) (Spec, error) {
	serversMap, err := m.serversMaker.Make(config)
	if err != nil {
		return nil, err
	}

	dflts, err := m.defaultsMaker.Make(&config.Defaults)
	if err != nil {
		return nil, err
	}

	var insts []instances.Instance
	var inst instances.Instance
	for i, configInst := range config.Instances {
		if configInst.Name == "" {
			return nil, errors.Errorf("instance %d name is empty", i)
		}
		inst, err = m.instanceMaker.Make(configInst, config.Environments, dflts, serversMap, config.Workspace)
		if err != nil {
			return nil, err
		}
		insts = append(insts, inst)
	}

	return &nativeSpec{
		fnd:       m.fnd,
		workspace: config.Workspace,
		instances: insts,
	}, nil
}

type nativeSpec struct {
	fnd       app.Foundation
	workspace string
	instances []instances.Instance
}

func isFiltered(instanceName string, filteredInstances []string) bool {
	if len(filteredInstances) == 0 {
		return true
	}

	for _, filter := range filteredInstances {
		if strings.HasPrefix(instanceName, filter) {
			return true
		}
	}

	return false
}

func (s *nativeSpec) Run(filteredInstances []string) error {
	// Loop through the instances.
	for _, instance := range s.instances {
		instanceName := instance.Name()

		if isFiltered(instanceName, filteredInstances) {
			s.fnd.Logger().Infof("Running instance %s", instanceName)
			if err := instance.Run(); err != nil {
				return err
			}
		} else {
			s.fnd.Logger().Debugf("Skipping instance %s as it is not in filtered instances", instanceName)
		}
	}

	return nil
}
