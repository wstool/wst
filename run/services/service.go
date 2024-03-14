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
	"github.com/bukka/wst/run/environments/task"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/resources/scripts"
	"github.com/bukka/wst/run/sandboxes/hooks"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/servers/configs"
	"github.com/bukka/wst/run/services/template"
	"io"
	"os"
	"path/filepath"
)

type Service interface {
	BaseUrl() (string, error)
	Name() string
	User() string
	Group() string
	Dirs() map[string]string
	ConfigPaths() map[string]string
	Environment() environment.Environment
	Task() task.Task
	RenderTemplate(text string, params parameters.Parameters) (string, error)
	OutputScanner(ctx context.Context, outputType output.Type) (*bufio.Scanner, error)
	Sandbox() sandbox.Sandbox
	Server() servers.Server
	Reload(ctx context.Context) error
	Restart(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Workspace() string
	SetTemplate(template template.Template)
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
	fnd             app.Foundation
	parametersMaker *parameters.Maker
	templateMaker   *template.Maker
}

func CreateMaker(fnd app.Foundation, parametersMaker *parameters.Maker) *Maker {
	return &Maker{
		fnd:             fnd,
		parametersMaker: parametersMaker,
		templateMaker:   template.CreateMaker(fnd),
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

		server, ok := srvs.GetServer(serviceConfig.Server.Name)
		if !ok {
			return nil, fmt.Errorf("server %s not found for service %s", serviceConfig.Server, serviceName)
		}

		serverParameters, err := m.parametersMaker.Make(serviceConfig.Server.Parameters)
		if err != nil {
			return nil, err
		}

		sandboxName := serviceConfig.Server.Sandbox
		providerType := providers.Type(sandboxName)

		sb, ok := server.Sandbox(providerType)
		if !ok {
			return nil, fmt.Errorf("sandbox %s not found for service %s", sandboxName, serviceName)
		}

		env, ok := environments[providerType]
		if !ok {
			return nil, fmt.Errorf("environment %s not found for service %s", sandboxName, serviceName)
		}

		nativeConfigs := make(map[string]nativeServiceConfig)

		for configName, serviceServerConfig := range serviceConfig.Server.Configs {
			config, found := server.Config(configName)
			if !found {
				return nil, fmt.Errorf("server config %s not found for service %s", configName, serviceName)
			}

			serviceServerConfigParameters, err := m.parametersMaker.Make(serviceServerConfig.Parameters)
			if err != nil {
				return nil, err
			}

			nativeConfigs[configName] = nativeServiceConfig{
				parameters:          serviceServerConfigParameters.Inherit(serverParameters),
				overwriteParameters: serviceServerConfig.OverwriteParameters,
				config:              config,
			}
		}

		service := &nativeService{
			name:             serviceName,
			environment:      env,
			scripts:          includedScripts,
			server:           server,
			serverParameters: serverParameters,
			sandbox:          sb,
			configs:          nativeConfigs,
			workspace:        filepath.Join(instanceWorkspace, serviceName),
		}

		svcs[serviceName] = service
	}

	for _, svc := range svcs {
		svc.SetTemplate(m.templateMaker.Make(svc, svcs))
	}

	return svcs, nil
}

type nativeServiceConfig struct {
	parameters          parameters.Parameters
	overwriteParameters bool
	config              configs.Config
}

type nativeService struct {
	name             string
	scripts          scripts.Scripts
	server           servers.Server
	serverParameters parameters.Parameters
	sandbox          sandbox.Sandbox
	task             task.Task
	environment      environment.Environment
	configs          map[string]nativeServiceConfig
	configPaths      map[string]string
	workspace        string
	template         template.Template
}

func (s *nativeService) ConfigPaths() map[string]string {
	return s.configPaths
}

func (s *nativeService) User() string {
	return s.server.User()
}

func (s *nativeService) Group() string {
	return s.server.Group()
}

func (s *nativeService) Dirs() map[string]string {
	return s.Sandbox().Dirs()
}

func (s *nativeService) Server() servers.Server {
	return s.server
}

func (s *nativeService) SetTemplate(template template.Template) {
	s.template = template
}

func (s *nativeService) Workspace() string {
	return s.workspace
}

func (s *nativeService) OutputScanner(ctx context.Context, outputType output.Type) (*bufio.Scanner, error) {
	reader, err := s.environment.Output(ctx, s.task, outputType)
	if err != nil {
		return nil, err
	}
	return bufio.NewScanner(reader), nil
}

func (s *nativeService) configPath(config configs.Config) string {
	return filepath.Join(s.workspace, filepath.Base(s.configPath(config)))
}

func (s *nativeService) renderConfig(config configs.Config) (string, error) {
	file, err := os.Open(config.FilePath())
	if err != nil {
		return "", err
	}
	defer file.Close()

	configContent, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	configPath := s.configPath(config)

	err = s.template.RenderToFile(string(configContent), config.Parameters(), configPath)
	if err != nil {
		return "", err
	}

	return configPath, nil
}

func (s *nativeService) renderConfigs() error {
	configs := s.server.Configs()
	configPaths := make(map[string]string, len(configs))
	for configName, config := range configs {
		path, err := s.renderConfig(config)
		if err != nil {
			return err
		}
		configPaths[configName] = path
	}
	s.configPaths = configPaths
	return nil
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

	// Render configs.
	err = s.renderConfigs()
	if err != nil {
		return err
	}

	// Execute start hook
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

func (s *nativeService) RenderTemplate(text string, params parameters.Parameters) (string, error) {
	return s.template.RenderToString(text, params)
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
