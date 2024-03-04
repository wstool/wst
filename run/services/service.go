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
	"bufio"
	"context"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/resources/scripts"
	"github.com/bukka/wst/run/sandboxes/hooks"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/servers/configs"
	"github.com/bukka/wst/run/task"
	"path/filepath"
)

type Service interface {
	BaseUrl() (string, error)
	Name() string
	Environment() environment.Environment
	Task() task.Task
	RenderTemplate(text string) (string, error)
	OutputScanner(ctx context.Context, outputType output.Type) (*bufio.Scanner, error)
	Sandbox() sandbox.Sandbox
	Reload(ctx context.Context) error
	Restart(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Workspace() string
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
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd: fnd,
	}
}

func (m *Maker) Make(
	config map[string]types.Service,
	scriptResources scripts.Scripts,
	srvs servers.Servers,
	environments environments.Environments,
	instanceWorkspace string,
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

		providerType := providers.Type(serviceConfig.Sandbox)

		sb, ok := server.Sandbox(providerType)
		if !ok {
			return nil, fmt.Errorf("sandbox %s not found for service %s", serviceConfig.Sandbox, serviceName)
		}

		env, ok := environments[providerType]
		if !ok {
			return nil, fmt.Errorf("environment %s not found for service %s", serviceConfig.Sandbox, serviceName)
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
			name:        serviceName,
			environment: env,
			scripts:     includedScripts,
			server:      server,
			sandbox:     sb,
			configs:     nativeConfigs,
			workspace:   filepath.Join(instanceWorkspace, serviceName),
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
	name        string
	scripts     scripts.Scripts
	server      servers.Server
	sandbox     sandbox.Sandbox
	task        task.Task
	environment environment.Environment
	configs     map[string]nativeServiceConfig
	workspace   string
}

func (s *nativeService) Workspace() string {
	return s.workspace
}

func (s *nativeService) OutputScanner(ctx context.Context, outputType output.Type) (*bufio.Scanner, error) {
	reader, err := s.environment.Output(ctx, outputType)
	if err != nil {
		return nil, err
	}
	return bufio.NewScanner(reader), nil
}

func (s *nativeService) Reload(ctx context.Context) error {
	hook, err := s.sandbox.Hook(hooks.ReloadHookType)
	if err != nil {
		return err
	}

	_, err = hook.Execute(ctx, s)

	return err
}

func (s *nativeService) Restart(ctx context.Context) error {
	hook, err := s.sandbox.Hook(hooks.RestartHookType)
	if err != nil {
		return err
	}

	_, err = hook.Execute(ctx, s)

	return err
}

func (s *nativeService) Start(ctx context.Context) error {
	hook, err := s.sandbox.Hook(hooks.StartHookType)
	if err != nil {
		return err
	}

	t, err := hook.Execute(ctx, s)
	if err != nil {
		return err
	}

	s.task = t
	return nil
}

func (s *nativeService) Stop(ctx context.Context) error {
	hook, err := s.sandbox.Hook(hooks.StopHookType)
	if err != nil {
		return err
	}

	_, err = hook.Execute(ctx, s)
	if err != nil {
		s.task = nil
	}

	return err
}

func (s *nativeService) Name() string {
	return s.name
}

func (s *nativeService) BaseUrl() (string, error) {
	if s.task == nil {
		return "", fmt.Errorf("service has not started yet")
	}

	return s.task.BaseUrl(), nil
}

func (s *nativeService) RenderTemplate(text string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *nativeService) Sandbox() sandbox.Sandbox {
	return s.sandbox
}

func (s *nativeService) Environment() environment.Environment {
	return s.environment
}

func (s *nativeService) Task() task.Task {
	return s.task
}
