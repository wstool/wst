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

package environments

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/environments/environment/providers/docker"
	"github.com/bukka/wst/run/environments/environment/providers/kubernetes"
	"github.com/bukka/wst/run/environments/environment/providers/local"
)

type Environments map[providers.Type]environment.Environment

type Maker struct {
	env             app.Env
	localMaker      *local.Maker
	dockerMaker     *docker.Maker
	kubernetesMaker *kubernetes.Maker
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env:             env,
		localMaker:      local.CreateMaker(env),
		dockerMaker:     docker.CreateMaker(env),
		kubernetesMaker: kubernetes.CreateMaker(env),
	}
}

func (m *Maker) Make(specConfig, instanceConfig map[string]types.Environment) (Environments, error) {
	envs := Environments{}

	for _, providerType := range providers.Types() {
		specEnv, specExists := specConfig[string(providerType)]
		instanceEnv, instanceExists := instanceConfig[string(providerType)]

		var mergedEnv types.Environment
		if instanceExists {
			// If instanceConfig exists for this providerType, use it (possibly merging with specEnv).
			// TODO: deep merging
			mergedEnv = instanceEnv
		} else if specExists {
			// If only specConfig exists, use it.
			mergedEnv = specEnv
		} else {
			// If neither exists, fail as there should be always defaults in config
			return nil, fmt.Errorf("unsupported environment type: %s", providerType)
		}

		var env environment.Environment
		var err error

		switch providerType {
		case providers.LocalType:
			if localEnv, ok := mergedEnv.(*types.LocalEnvironment); ok {
				env, err = m.localMaker.Make(localEnv)
			} else {
				err = fmt.Errorf("local environment has unexpected data type %t", mergedEnv)
			}
		case providers.DockerType:
			if localEnv, ok := mergedEnv.(*types.DockerEnvironment); ok {
				env, err = m.dockerMaker.Make(localEnv)
			} else {
				err = fmt.Errorf("docker environment has unexpected data type %t", mergedEnv)
			}
		case providers.KubernetesType:
			if localEnv, ok := mergedEnv.(*types.KubernetesEnvironment); ok {
				env, err = m.kubernetesMaker.Make(localEnv)
			} else {
				err = fmt.Errorf("kubernetes environment has unexpected data type %t", mergedEnv)
			}
		default:
			err = fmt.Errorf("unsupported environment type: %s", providerType)
		}

		if err != nil {
			return nil, err
		}

		envs[providerType] = env
	}

	return envs, nil
}
