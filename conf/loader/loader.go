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

package loader

import (
	"encoding/json"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"path/filepath"
)

type LoadedConfig interface {
	Path() string
	Data() map[string]interface{}
}

type Loader interface {
	LoadConfig(path string) (LoadedConfig, error)
	LoadConfigs(paths []string) ([]LoadedConfig, error)
	GlobConfigs(pattern string) ([]LoadedConfig, error)
}

type LoadedConfigData struct {
	path string
	data map[string]interface{}
}

func (d LoadedConfigData) Path() string {
	return d.path
}

func (d LoadedConfigData) Data() map[string]interface{} {
	return d.data
}

type ConfigLoader struct {
	env app.Env
}

func (l ConfigLoader) LoadConfig(path string) (LoadedConfig, error) {
	fs := l.env.Fs()

	rawData, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}

	// Check file extension and choose appropriate unmarshal function
	extension := filepath.Ext(path)
	switch extension {
	case ".json":
		err = json.Unmarshal(rawData, &data)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(rawData, &data)
	case ".toml":
		err = toml.Unmarshal(rawData, &data)
	default:
		return nil, fmt.Errorf("unsupported extension: %s", extension)
	}

	if err != nil {
		return nil, err
	}

	loadedConfig := LoadedConfigData{
		path: path,
		data: data,
	}

	return loadedConfig, nil
}

func (l ConfigLoader) LoadConfigs(paths []string) ([]LoadedConfig, error) {
	configs := make([]LoadedConfig, 0)
	for _, path := range paths {
		config, err := l.LoadConfig(path)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (l ConfigLoader) GlobConfigs(pattern string) ([]LoadedConfig, error) {
	fs := l.env.Fs()
	paths, err := afero.Glob(fs, pattern)
	if err != nil {
		return nil, err
	}
	return l.LoadConfigs(paths)
}

func CreateLoader(env app.Env) Loader {
	return &ConfigLoader{
		env: env,
	}
}
