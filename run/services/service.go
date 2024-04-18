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
	PublicUrl(path string) (string, error)
	PrivateUrl() (string, error)
	Pid() (int, error)
	Name() string
	FullName() string
	User() string
	Group() string
	Dirs() map[sandbox.DirType]string
	Port() int32
	EnvironmentConfigPaths() map[string]string
	WorkspaceConfigPaths() map[string]string
	EnvironmentScriptPaths() map[string]string
	WorkspaceScriptPaths() map[string]string
	Environment() environment.Environment
	Task() task.Task
	RenderTemplate(text string, params parameters.Parameters) (string, error)
	OutputScanner(ctx context.Context, outputType output.Type) (*bufio.Scanner, error)
	Sandbox() sandbox.Sandbox
	Server() servers.Server
	ServerParameters() parameters.Parameters
	Reload(ctx context.Context) error
	Restart(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsPublic() bool
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

type ServiceLocator interface {
	Find(name string) (Service, error)
	Services() Services
}

type nativeServiceLocator struct {
	services Services
}

func (sl *nativeServiceLocator) Services() Services {
	return sl.services
}

func (sl *nativeServiceLocator) Find(name string) (Service, error) {
	return sl.services.FindService(name)
}

func NewServiceLocator(services Services) ServiceLocator {
	return &nativeServiceLocator{
		services: services,
	}
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
	instanceName string,
	instanceWorkspace string,
) (ServiceLocator, error) {
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
		if !sb.Available() {
			return nil, fmt.Errorf("sandbox %s is not available for service %s", sandboxName, serviceName)
		}

		env, ok := environments[providerType]
		if !ok {
			return nil, fmt.Errorf("environment %s not found for service %s", sandboxName, serviceName)
		}
		env.MarkUsed()

		nativeConfigs := make(map[string]nativeServiceConfig)

		for configName, serviceServerConfig := range serviceConfig.Server.Configs {
			if serviceServerConfig.Include {
				config, found := server.Config(configName)
				if !found {
					return nil, fmt.Errorf("server config %s not found for service %s", configName, serviceName)
				}

				serviceServerConfigParameters, err := m.parametersMaker.Make(serviceServerConfig.Parameters)
				if err != nil {
					return nil, err
				}

				nativeConfigs[configName] = nativeServiceConfig{
					parameters: serviceServerConfigParameters.Inherit(serverParameters),
					config:     config,
				}
			}
		}

		service := &nativeService{
			name:             serviceName,
			fullName:         instanceName + "-" + serviceName,
			public:           serviceConfig.Public,
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

	sl := NewServiceLocator(svcs)

	for _, svc := range svcs {
		svc.SetTemplate(m.templateMaker.Make(svc, sl))
	}

	return sl, nil
}

type nativeServiceConfig struct {
	parameters parameters.Parameters
	config     configs.Config
}

type nativeService struct {
	name                   string
	fullName               string
	public                 bool
	scripts                scripts.Scripts
	server                 servers.Server
	serverParameters       parameters.Parameters
	sandbox                sandbox.Sandbox
	task                   task.Task
	environment            environment.Environment
	configs                map[string]nativeServiceConfig
	environmentConfigPaths map[string]string
	workspaceConfigPaths   map[string]string
	environmentScriptPaths map[string]string
	workspaceScriptPaths   map[string]string
	workspace              string
	template               template.Template
}

func (s *nativeService) Port() int32 {
	return s.server.Port()
}

func (s *nativeService) IsPublic() bool {
	return s.public
}

func (s *nativeService) EnvironmentConfigPaths() map[string]string {
	return s.environmentConfigPaths
}

func (s *nativeService) WorkspaceConfigPaths() map[string]string {
	return s.workspaceConfigPaths
}

func (s *nativeService) EnvironmentScriptPaths() map[string]string {
	return s.environmentScriptPaths
}

func (s *nativeService) WorkspaceScriptPaths() map[string]string {
	return s.workspaceScriptPaths
}

func (s *nativeService) User() string {
	return s.server.User()
}

func (s *nativeService) Group() string {
	return s.server.Group()
}

func (s *nativeService) Dirs() map[sandbox.DirType]string {
	return s.Sandbox().Dirs()
}

func (s *nativeService) Server() servers.Server {
	return s.server
}

func (s *nativeService) ServerParameters() parameters.Parameters {
	return s.serverParameters
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

func (s *nativeService) configPaths(configPath string) (string, string, error) {
	environmentRootPath := s.environment.RootPath(s.workspace)
	sandboxConfDir, err := s.sandbox.Dir(sandbox.ConfDirType)
	if err != nil {
		return "", "", err
	}
	confPath := filepath.Join(sandboxConfDir, filepath.Base(configPath))
	return filepath.Join(s.workspace, confPath), filepath.Join(environmentRootPath, confPath), nil
}

func (s *nativeService) renderConfig(config configs.Config) (string, string, error) {
	file, err := os.Open(config.FilePath())
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	configContent, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}

	workspaceConfigPath, environmentConfigPath, err := s.configPaths(config.FilePath())
	if err != nil {
		return "", "", err
	}

	err = s.template.RenderToFile(string(configContent), config.Parameters(), workspaceConfigPath, 0644)
	if err != nil {
		return "", "", err
	}

	return workspaceConfigPath, environmentConfigPath, nil
}

func (s *nativeService) renderConfigs() error {
	cfgs := s.server.Configs()
	envConfigPaths := make(map[string]string, len(cfgs))
	wsConfigPaths := make(map[string]string, len(cfgs))
	for cfgName, cfg := range cfgs {
		wsPath, envPath, err := s.renderConfig(cfg)
		if err != nil {
			return err
		}
		envConfigPaths[cfgName] = envPath
		wsConfigPaths[cfgName] = wsPath
	}
	s.environmentConfigPaths = envConfigPaths
	s.workspaceConfigPaths = envConfigPaths
	return nil
}

func (s *nativeService) renderScript(script scripts.Script) (string, string, error) {
	sandboxScriptDir, err := s.sandbox.Dir(sandbox.ScriptDirType)
	if err != nil {
		return "", "", err
	}
	environmentScriptPath := filepath.Join(sandboxScriptDir, script.Path())
	workspaceScriptPath := filepath.Join(s.workspace, environmentScriptPath)

	err = s.template.RenderToFile(script.Content(), script.Parameters(), workspaceScriptPath, script.Mode())
	if err != nil {
		return "", "", err
	}

	return workspaceScriptPath, environmentScriptPath, nil
}

func (s *nativeService) renderScripts() error {
	includedScripts := s.scripts
	envScriptPaths := make(map[string]string, len(includedScripts))
	wsScriptPaths := make(map[string]string, len(includedScripts))
	for scriptName, script := range includedScripts {
		wsPath, envPath, err := s.renderScript(script)
		if err != nil {
			return err
		}
		envScriptPaths[scriptName] = envPath
		wsScriptPaths[scriptName] = wsPath
	}
	s.environmentScriptPaths = envScriptPaths
	s.workspaceScriptPaths = wsScriptPaths
	return nil
}

func (s *nativeService) makeEnvServiceSettings() *environment.ServiceSettings {
	return &environment.ServiceSettings{
		Name:                   s.name,
		FullName:               s.fullName,
		Port:                   s.Port(),
		Public:                 s.public,
		Sandbox:                s.sandbox,
		ServerParameters:       s.serverParameters,
		EnvironmentConfigPaths: s.environmentConfigPaths,
		EnvironmentScriptPaths: s.environmentScriptPaths,
		WorkspaceConfigPaths:   s.workspaceConfigPaths,
		WorkspaceScriptPaths:   s.workspaceScriptPaths,
	}
}

func (s *nativeService) Reload(ctx context.Context) error {
	hook, err := s.sandbox.Hook(hooks.ReloadHookType)
	if err != nil {
		return err
	}

	ss := s.makeEnvServiceSettings()
	_, err = hook.Execute(ctx, ss, s.template, s.environment, s.task)

	return err
}

func (s *nativeService) Restart(ctx context.Context) error {
	hook, err := s.sandbox.Hook(hooks.RestartHookType)
	if err != nil {
		return err
	}

	ss := s.makeEnvServiceSettings()
	_, err = hook.Execute(ctx, ss, s.template, s.environment, s.task)

	return err
}

func (s *nativeService) Start(ctx context.Context) error {
	hook, err := s.sandbox.Hook(hooks.StartHookType)
	if err != nil {
		return err
	}

	// Render configs
	err = s.renderConfigs()
	if err != nil {
		return err
	}

	// Render scripts
	err = s.renderScripts()
	if err != nil {
		return err
	}

	// Execute start hook
	ss := s.makeEnvServiceSettings()
	t, err := hook.Execute(ctx, ss, s.template, s.environment, s.task)
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

	ss := s.makeEnvServiceSettings()
	_, err = hook.Execute(ctx, ss, s.template, s.environment, s.task)
	if err != nil {
		s.task = nil
	}

	return err
}

func (s *nativeService) Name() string {
	return s.name
}

func (s *nativeService) FullName() string {
	return s.fullName
}

func (s *nativeService) PublicUrl(path string) (string, error) {
	if s.task == nil {
		return "", fmt.Errorf("service has not started yet")
	}
	if !s.IsPublic() {
		return "", fmt.Errorf("only public service has public URL")
	}

	return filepath.Join(s.task.PublicUrl(), path), nil
}

func (s *nativeService) PrivateUrl() (string, error) {
	if s.task == nil {
		return "", fmt.Errorf("service has not started yet")
	}

	return s.task.PrivateUrl(), nil
}

func (s *nativeService) Pid() (int, error) {
	if s.task == nil {
		return 0, fmt.Errorf("service has not started yet")
	}

	return s.task.Pid(), nil
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
