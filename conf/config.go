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
	"github.com/bukka/wst/conf/overwrites"
	"github.com/bukka/wst/conf/parser"
	"github.com/bukka/wst/conf/types"
)

type Maker interface {
	Make(configPaths []string, overwrites map[string]string) (*types.Config, error)
}

type ConfigMaker struct {
	fnd        app.Foundation
	loader     loader.Loader
	parser     parser.Parser
	merger     merger.Merger
	overwriter overwrites.Overwriter
}

func CreateConfigMaker(fnd app.Foundation) Maker {
	ld := loader.CreateLoader(fnd)
	pr := parser.CreateParser(fnd, ld)
	return &ConfigMaker{
		fnd:        fnd,
		loader:     ld,
		parser:     pr,
		merger:     merger.CreateMerger(fnd),
		overwriter: overwrites.CreateOverwriter(fnd, pr),
	}
}

func (m *ConfigMaker) Make(configPaths []string, overwrites map[string]string) (*types.Config, error) {
	loadedConfigs, err := m.loader.LoadConfigs(configPaths)
	if err != nil {
		return nil, err
	}

	var configs []*types.Config
	for _, loadedConfig := range loadedConfigs {
		config := &types.Config{}
		err = m.parser.ParseConfig(loadedConfig.Data(), config, loadedConfig.Path())
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	config, err := m.merger.MergeConfigs(configs)
	if err != nil {
		return nil, err
	}

	if len(overwrites) > 0 {
		err = m.overwriter.Overwrite(config, overwrites)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
