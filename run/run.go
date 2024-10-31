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
	"encoding/json"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf"
	"github.com/wstool/wst/run/spec"
	"os"
	"path/filepath"
)

type Options struct {
	ConfigPaths []string
	IncludeAll  bool
	Overwrites  map[string]string
	NoEnvs      bool
	Instances   []string
}

type Runner struct {
	fnd         app.Foundation
	configMaker conf.Maker
	specMaker   spec.Maker
}

func CreateRunner(fnd app.Foundation) *Runner {
	return &Runner{
		fnd:         fnd,
		configMaker: conf.CreateConfigMaker(fnd),
		specMaker:   spec.CreateMaker(fnd),
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

	r.fnd.Logger().Info("Executing configuration")

	configPathsJSON, _ := json.Marshal(configPaths)
	overwritesJSON, _ := json.Marshal(options.Overwrites)
	r.fnd.Logger().Debugf("Creating config for paths %v and overwrites %v",
		string(configPathsJSON), string(overwritesJSON))
	config, err := r.configMaker.Make(configPaths, options.Overwrites)
	if err != nil {
		return err
	}

	r.fnd.Logger().Debug("Creating specification")
	specification, err := r.specMaker.Make(&config.Spec)
	if err != nil {
		return err
	}

	r.fnd.Logger().Debug("Running instances")
	return specification.Run(options.Instances)
}

func (r *Runner) getConfigPaths() []string {
	var paths []string
	home, _ := r.fnd.UserHomeDir()
	r.validateAndAppendPath("wst.yaml", &paths)
	r.validateAndAppendPath(filepath.Join(home, ".wst/wst.yaml"), &paths)
	r.validateAndAppendPath(filepath.Join(home, ".config/wst/wst.yaml"), &paths)

	return paths
}

func (r *Runner) validateAndAppendPath(path string, paths *[]string) {
	if _, err := r.fnd.Fs().Stat(path); !os.IsNotExist(err) {
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
