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

package conf

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/loader"
	"github.com/bukka/wst/conf/merger"
	"github.com/bukka/wst/conf/parser"
	"github.com/bukka/wst/conf/processor"
	"github.com/bukka/wst/conf/types"
)

func MakeConfig(configPaths []string, overwrites map[string]string, env app.Env) (*types.Config, error) {
	loader := loader.CreateLoader(env)
	loadedConfigs, err := loader.LoadConfigs(configPaths)
	if err != nil {
		return nil, err
	}

	parser := parser.CreateParser(env, loader)
	var configs []*types.Config
	for _, loadedConfig := range loadedConfigs {
		config := &types.Config{}
		err = parser.ParseConfig(loadedConfig.Data(), config)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	config, err := merger.CreateMerger(env).MergeConfigs(configs, overwrites)
	if err != nil {
		return nil, err
	}

	if err = processor.CreateProcessor(env).ProcessConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}
