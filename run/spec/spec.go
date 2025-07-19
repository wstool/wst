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
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/instances"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/servers"
	"github.com/wstool/wst/run/spec/defaults"
	"strings"
)

type Spec interface {
	Run(filteredInstances []string) error
}

type Maker interface {
	Make(config *types.Spec, filteredInstances []string) (Spec, error)
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

func (m *nativeMaker) Make(config *types.Spec, filteredInstances []string) (Spec, error) {
	serversMap, err := m.serversMaker.Make(config)
	if err != nil {
		return nil, err
	}

	dflts, err := m.defaultsMaker.Make(&config.Defaults)
	if err != nil {
		return nil, err
	}

	instsMap := make(map[string]instances.Instance)
	var childInstsList, runnableInstsList []instances.Instance
	var inst instances.Instance
	for i, configInst := range config.Instances {
		idx := i + 1
		if configInst.Name == "" {
			return nil, errors.Errorf("instance %d name is empty", idx)
		}
		if filteredInstances != nil && !isFiltered(configInst.Name, filteredInstances) {
			continue
		}
		inst, err = m.instanceMaker.Make(configInst, idx, config.Environments, dflts, serversMap, config.Workspace)
		if err != nil {
			return nil, err
		}
		instsMap[inst.Name()] = inst
		if inst.IsChild() {
			childInstsList = append(childInstsList, inst)
		}
		if !inst.IsAbstract() {
			runnableInstsList = append(runnableInstsList, inst)
		}
	}

	// Extend all child instances
	for _, inst = range childInstsList {
		if err = inst.Extend(instsMap); err != nil {
			return nil, err
		}
	}

	// Init instance
	for _, inst = range runnableInstsList {
		if err = inst.Init(); err != nil {
			return nil, err
		}
	}

	return &nativeSpec{
		fnd:       m.fnd,
		workspace: config.Workspace,
		instances: runnableInstsList,
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
