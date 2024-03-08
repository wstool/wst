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

package configs

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/parameters"
)

type Config interface {
	FilePath() string
	Parameters() parameters.Parameters
}

type Configs map[string]Config

func (a Configs) Inherit(parentConfigs Configs) {
	for configName, config := range parentConfigs {
		_, ok := a[configName]
		if !ok {
			a[configName] = config
		}
	}
}

type Maker struct {
	fnd             app.Foundation
	parametersMaker *parameters.Maker
}

func CreateMaker(fnd app.Foundation, parametersMaker *parameters.Maker) *Maker {
	return &Maker{
		fnd:             fnd,
		parametersMaker: parametersMaker,
	}
}

func (m *Maker) Make(config map[string]types.ServerConfig) (Configs, error) {
	configs := make(Configs)
	for name, serverConfig := range config {
		params, err := m.parametersMaker.Make(serverConfig.Parameters)
		if err != nil {
			return nil, err
		}
		configs[name] = &nativeConfig{
			file:       serverConfig.File,
			parameters: params,
		}
	}
	return configs, nil
}

type nativeConfig struct {
	file       string
	parameters parameters.Parameters
}

func (c *nativeConfig) FilePath() string {
	return c.file
}

func (c *nativeConfig) Parameters() parameters.Parameters {
	return c.parameters
}
