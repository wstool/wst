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
)

type Config interface {
}

type Configs map[string]Config

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(config map[string]types.ServerConfig) (Configs, error) {
	configs := make(Configs)
	for name, serverConfig := range config {
		configs[name] = &nativeConfig{
			file:       serverConfig.File,
			parameters: serverConfig.Parameters,
		}
	}
	return configs, nil
}

type nativeConfig struct {
	file       string
	parameters types.Parameters
}
