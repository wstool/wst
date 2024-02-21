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
	"github.com/bukka/wst/run/sandboxes"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/spec"
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

type Runner struct {
	env            app.Env
	configMaker    *conf.ConfigMaker
	sandboxesMaker *sandboxes.Maker
	serversMaker   *servers.Maker
	specMaker      *spec.Maker
}

var DefaultsFs = afero.NewOsFs()

func CreateRunner(env app.Env) *Runner {
	return &Runner{
		env:            env,
		configMaker:    conf.CreateConfigMaker(env),
		sandboxesMaker: sandboxes.CreateMaker(env),
		serversMaker:   servers.CreateMaker(env),
		specMaker:      spec.CreateMaker(env),
	}
}

func (r *Runner) Execute(options *Options) error {
	var configPaths []string
	if options.IncludeAll {
		extraPaths := r.getConfigPaths()
		configPaths = append(options.ConfigPaths, extraPaths...)
	} else {
		configPaths = options.ConfigPaths
	}
	configPaths = r.removeDuplicates(configPaths)

	config, err := r.configMaker.Make(configPaths, options.Overwrites)
	if err != nil {
		return err
	}

	serversMap, err := r.serversMaker.Make(config)
	if err != nil {
		return err
	}

	specification, err := r.specMaker.Make(&config.Spec, serversMap)
	if err != nil {
		return err
	}

	return specification.ExecuteInstances(options.Instances, options.DryRun)
}

func (r *Runner) getConfigPaths() []string {
	var paths []string
	home, _ := r.env.UserHomeDir()
	r.validateAndAppendPath("wst.yaml", &paths)
	r.validateAndAppendPath(filepath.Join(home, ".wst/wst.yaml"), &paths)
	r.validateAndAppendPath(filepath.Join(home, ".config/wst/wst.yaml"), &paths)

	return paths
}

func (r *Runner) validateAndAppendPath(path string, paths *[]string) {
	if _, err := r.env.Fs().Stat(path); !os.IsNotExist(err) {
		*paths = append(*paths, path)
	}
}

func (r *Runner) removeDuplicates(elements []string) []string {
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
