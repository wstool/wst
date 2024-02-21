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

package services

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/resources/scripts"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/servers/configs"
)

type Service interface {
	BaseUrl() string
	Name() string
	RenderTemplate(text string) (string, error)
	Sandbox() sandbox.Sandbox
	Restart(reload bool) error
	Start() error
	Stop() error
}

type Services map[string]Service

func (s Services) FindService(name string) (Service, error) {
	svc, ok := s[name]
	if !ok {
		return svc, fmt.Errorf("service %s not found", name)
	}
	return svc, nil
}

func (s Services) AddService(service Service) error {
	s[service.Name()] = service
	return nil
}

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(
	config map[string]types.Service,
	scriptResources scripts.Scripts,
	srvs servers.Servers,
) (Services, error) {
	svcs := make(Services)
	for serviceName, serviceConfig := range config {
		var includedScripts scripts.Scripts

		if serviceConfig.Resources.Scripts.IncludeAll {
			includedScripts = scriptResources
		} else {
			includedScripts = make(scripts.Scripts)
			for _, scriptName := range serviceConfig.Resources.Scripts.IncludeList {
				script, ok := scriptResources[scriptName]
				if !ok {
					return nil, fmt.Errorf("script %s not found for service %s", scriptName, serviceName)
				}
				includedScripts[scriptName] = script
			}
		}

		server, ok := srvs.GetServer(serviceConfig.Server)
		if !ok {
			return nil, fmt.Errorf("server %s not found for service %s", serviceConfig.Server, serviceName)
		}

		sandbox, ok := server.Sandbox(sandbox.Type(serviceConfig.Sandbox))
		if !ok {
			return nil, fmt.Errorf("sandbox %s not found for service %s", serviceConfig.Sandbox, serviceName)
		}

		nativeConfigs := make(map[string]nativeServiceConfig)

		for configName, serviceConfig := range serviceConfig.Configs {
			config, found := server.Config(configName)
			if !found {
				return nil, fmt.Errorf("server config %s not found for service %s", configName, serviceName)
			}
			nativeConfigs[configName] = nativeServiceConfig{
				parameters:          serviceConfig.Parameters,
				overwriteParameters: serviceConfig.OverwriteParameters,
				config:              config,
			}
		}

		service := &nativeService{
			name:    serviceName,
			scripts: includedScripts,
			server:  server,
			sandbox: sandbox,
			configs: nativeConfigs,
		}

		svcs[serviceName] = service
	}
	return svcs, nil
}

type nativeServiceConfig struct {
	parameters          types.Parameters
	overwriteParameters bool
	config              configs.Config
}

type nativeService struct {
	name    string
	scripts scripts.Scripts
	server  servers.Server
	sandbox sandbox.Sandbox
	configs map[string]nativeServiceConfig
}

func (s *nativeService) Restart(reload bool) error {
	//TODO implement me
	panic("implement me")
}

func (s *nativeService) Start() error {
	//TODO implement me
	panic("implement me")
}

func (s *nativeService) Stop() error {
	//TODO implement me
	panic("implement me")
}

func (s *nativeService) Name() string {
	return s.name
}

func (s *nativeService) BaseUrl() string {
	//TODO implement me
	panic("implement me")
}

func (s *nativeService) RenderTemplate(text string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *nativeService) Sandbox() sandbox.Sandbox {
	return s.sandbox
}