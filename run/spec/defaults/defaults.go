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

package defaults

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/parameters"
)

type Defaults struct {
	Service    ServiceDefaults
	Timeouts   TimeoutsDefaults
	Parameters parameters.Parameters
}

type ServiceDefaults struct {
	Sandbox string
	Server  ServiceServerDefaults
}

type ServiceServerDefaults struct {
	Tag string
}

type TimeoutsDefaults struct {
	Actions int
	Action  int
}

type Maker interface {
	Make(config *types.SpecDefaults) (*Defaults, error)
}

type nativeMaker struct {
	fnd             app.Foundation
	parametersMaker parameters.Maker
}

func CreateMaker(fnd app.Foundation, parametersMaker parameters.Maker) Maker {
	return &nativeMaker{
		fnd:             fnd,
		parametersMaker: parametersMaker,
	}
}

func (m *nativeMaker) Make(config *types.SpecDefaults) (*Defaults, error) {
	params, err := m.parametersMaker.Make(config.Parameters)
	if err != nil {
		return nil, err
	}
	return &Defaults{
		Service: ServiceDefaults{
			Sandbox: config.Service.Sandbox,
			Server: ServiceServerDefaults{
				Tag: config.Service.Server.Tag,
			},
		},
		Timeouts: TimeoutsDefaults{
			Actions: config.Timeouts.Actions,
			Action:  config.Timeouts.Action,
		},
		Parameters: params,
	}, nil
}
