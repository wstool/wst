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
	"fmt"
	"github.com/bukka/wst/app"
)

type Options struct {
	Configs    []string
	Overwrites map[string]string
	Instances  []string
	DryRun     bool
}

func ExecuteConfigs(options Options, env app.Env) error {
	loader := CreateLoader(env)
	loadedConfigs, err := loader.LoadConfigs(options.Configs)
	if err != nil {
		return err
	}
	// TODO: support other options

	parser := CreateParser(env, loader)
	for _, loadedConfig := range loadedConfigs {
		config := &Config{}
		err = parser.ParseConfig(loadedConfig.Data, config)
		if err != nil {
			return err
		}
		fmt.Println(config)
	}
	fmt.Println(options)
	return nil
}
