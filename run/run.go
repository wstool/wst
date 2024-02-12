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

package run

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf"
	"github.com/bukka/wst/sandboxes"
	"github.com/bukka/wst/servers"
	"github.com/bukka/wst/spec"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

type Options struct {
	ConfigPaths []string
	IncludeAll  bool
	Overwrites  map[string]string
	NoEnvs      bool
	DryRun      bool
	Instances   []string
}

var DefaultsFs = afero.NewOsFs()

func Execute(options *Options, env app.Env) error {
	var configPaths []string
	if options.IncludeAll {
		extraPaths := GetConfigPaths(env)
		configPaths = append(options.ConfigPaths, extraPaths...)
	} else {
		configPaths = options.ConfigPaths
	}
	configPaths = removeDuplicates(configPaths)

	config, err := conf.MakeConfig(configPaths, options.Overwrites, env)
	if err != nil {
		return err
	}

	sbs, err := sandboxes.MakeSandboxes(config)
	if err != nil {
		return err
	}

	srvs, err := servers.MakeServers(config, sbs)
	if err != nil {
		return err
	}

	specification, err := spec.MakeSpec(config, srvs)
	if err != nil {
		return err
	}

	return specification.ExecuteInstances(options.Instances, options.DryRun)
}

func GetConfigPaths(env app.Env) []string {
	var paths []string
	home, _ := env.GetUserHomeDir()
	validateAndAppendPath("wst.yaml", &paths, env)
	validateAndAppendPath(filepath.Join(home, ".wst/wst.yaml"), &paths, env)
	validateAndAppendPath(filepath.Join(home, ".config/wst/wst.yaml"), &paths, env)

	return paths
}

func validateAndAppendPath(path string, paths *[]string, env app.Env) {
	if _, err := env.Fs().Stat(path); !os.IsNotExist(err) {
		*paths = append(*paths, path)
	}
}

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	var result []string

	for v := range elements {
		if !encountered[elements[v]] {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}
	return result
}
