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
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/environments"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/environments/task"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/resources"
	"github.com/wstool/wst/run/resources/certificates"
	"github.com/wstool/wst/run/resources/scripts"
	"github.com/wstool/wst/run/sandboxes/dir"
	"github.com/wstool/wst/run/sandboxes/hooks"
	"github.com/wstool/wst/run/sandboxes/sandbox"
	"github.com/wstool/wst/run/servers"
	"github.com/wstool/wst/run/servers/configs"
	"github.com/wstool/wst/run/services/template"
	"github.com/wstool/wst/run/spec/defaults"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
)

type Service interface {
	LocalAddress() string
	LocalPort() int32
	PrivateAddress() string
	PrivateUrl() (string, error)
	PublicUrl(path string) (string, error)
	UdsPath(...string) (string, error)
	Pid() (int, error)
	Name() string
	FullName() string
	Executable() (string, error)
	User() string
	Group() string
	Dirs() map[dir.DirType]string
	ConfDir() (string, error)
	RunDir() (string, error)
	ScriptDir() (string, error)
	Port() int32
	EnvironmentConfigPaths() map[string]string
	WorkspaceConfigPaths() map[string]string
	EnvironmentScriptPaths() map[string]string
	WorkspaceScriptPaths() map[string]string
	Environment() environment.Environment
	FindCertificate(name string) (*certificates.RenderedCertificate, error)
	Task() task.Task
	RenderTemplate(text string, params parameters.Parameters) (string, error)
	OutputReader(ctx context.Context, outputType output.Type) (io.Reader, error)
	Sandbox() sandbox.Sandbox
	Server() servers.Server
	ServerParameters() parameters.Parameters
	ExecCommand(ctx context.Context, cmd *environment.Command, oc output.Collector) error
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
		return svc, errors.Errorf("service %s not found", name)
	}
	return svc, nil
}

func (s Services) AddService(service Service) {
	s[service.Name()] = service
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

type Maker interface {
	Make(
		config map[string]types.Service,
		dflts *defaults.Defaults,
		rscrs resources.Resources,
		srvs servers.Servers,
		environments environments.Environments,
		instanceName string,
		instanceIdx int,
		instanceWorkspace string,
		instanceParameters parameters.Parameters,
	) (ServiceLocator, error)
}

type nativeMaker struct {
	fnd             app.Foundation
	parametersMaker parameters.Maker
	templateMaker   template.Maker
}

func CreateMaker(fnd app.Foundation, parametersMaker parameters.Maker) Maker {
	return &nativeMaker{
		fnd:             fnd,
		parametersMaker: parametersMaker,
		templateMaker:   template.CreateMaker(fnd),
	}
}

func (m *nativeMaker) Make(
	config map[string]types.Service,
	dflts *defaults.Defaults,
	rscrs resources.Resources,
	srvs servers.Servers,
	environments environments.Environments,
	instanceName string,
	instanceIdx int,
	instanceWorkspace string,
	instanceParameters parameters.Parameters,
) (ServiceLocator, error) {
	svcs := make(Services)
	tmplSvcs := make(template.Services)
	for serviceName, serviceConfig := range config {
		tag := serviceConfig.Server.Tag
		if tag == "" {
			tag = dflts.Service.Server.Tag
		}
		server, ok := srvs.GetServer(serviceConfig.Server.Name, tag)
		if !ok {
			return nil, errors.Errorf("server %s not found for service %s", serviceConfig.Server.Name, serviceName)
		}

		serverParameters, err := m.parametersMaker.Make(serviceConfig.Server.Parameters)
		if err != nil {
			return nil, err
		}
		serverParameters.Inherit(server.Parameters()).Inherit(instanceParameters).Inherit(dflts.Parameters)

		sandboxName := serviceConfig.Server.Sandbox
		if sandboxName == "" {
			sandboxName = dflts.Service.Sandbox
		}
		providerType := providers.Type(sandboxName)

		sb, ok := server.Sandbox(providerType)
		if !ok {
			return nil, errors.Errorf("sandbox %s not found for service %s", sandboxName, serviceName)
		}
		if !sb.Available() {
			return nil, errors.Errorf("sandbox %s is not available for service %s", sandboxName, serviceName)
		}

		env, ok := environments[providerType]
		if !ok {
			return nil, errors.Errorf("environment %s not found for service %s", sandboxName, serviceName)
		}
		env.MarkUsed()
		envResources := env.Resources()

		var includedCertificates certificates.Certificates
		allCerts := make(certificates.Certificates).Inherit(rscrs.Certificates).Inherit(envResources.Certificates)
		if serviceConfig.Resources.Certificates.IncludeAll {
			includedCertificates = allCerts
		} else {
			includedCertificates = make(certificates.Certificates)
			for _, certName := range serviceConfig.Resources.Certificates.IncludeList {
				cert, ok := allCerts[certName]
				if !ok {
					return nil, errors.Errorf("certificates %s not found for service %s", certName, serviceName)
				}
				includedCertificates[certName] = cert
			}
		}

		var includedScripts scripts.Scripts
		allScripts := make(scripts.Scripts).Inherit(rscrs.Scripts).Inherit(envResources.Scripts)
		if serviceConfig.Resources.Scripts.IncludeAll {
			includedScripts = allScripts
		} else {
			includedScripts = make(scripts.Scripts)
			for _, scriptName := range serviceConfig.Resources.Scripts.IncludeList {
				script, ok := allScripts[scriptName]
				if !ok {
					return nil, errors.Errorf("script %s not found for service %s", scriptName, serviceName)
				}
				includedScripts[scriptName] = script
			}
		}

		nativeConfigs := make(map[string]nativeServiceConfig)

		for configName, serviceServerConfig := range serviceConfig.Server.Configs {
			if serviceServerConfig.Include {
				cfg, found := server.Config(configName)
				if !found {
					return nil, errors.Errorf("server config %s not found for service %s", configName, serviceName)
				}

				serviceServerConfigParameters, err := m.parametersMaker.Make(serviceServerConfig.Parameters)
				if err != nil {
					return nil, err
				}

				nativeConfigs[configName] = nativeServiceConfig{
					parameters: serviceServerConfigParameters.Inherit(cfg.Parameters()).Inherit(serverParameters),
					config:     cfg,
				}
			}
		}

		service := &nativeService{
			fnd:              m.fnd,
			name:             serviceName,
			fullName:         instanceName + "-" + serviceName,
			uniqueName:       fmt.Sprintf("i%d-%s", instanceIdx, serviceName),
			public:           serviceConfig.Public,
			port:             env.ReservePort(),
			environment:      env,
			certificates:     includedCertificates,
			scripts:          includedScripts,
			server:           server,
			serverParameters: serverParameters,
			sandbox:          sb,
			configs:          nativeConfigs,
			workspace:        filepath.Join(instanceWorkspace, serviceName),
		}

		svcs[serviceName] = service
		tmplSvcs[serviceName] = service
	}

	for _, svc := range svcs {
		svc.SetTemplate(m.templateMaker.Make(svc, tmplSvcs, svc.Server().Templates()))
	}

	return NewServiceLocator(svcs), nil
}

type nativeServiceConfig struct {
	parameters parameters.Parameters
	config     configs.Config
}

type nativeService struct {
	fnd                    app.Foundation
	name                   string
	fullName               string
	uniqueName             string
	public                 bool
	port                   int32
	certificates           certificates.Certificates
	scripts                scripts.Scripts
	server                 servers.Server
	serverParameters       parameters.Parameters
	sandbox                sandbox.Sandbox
	task                   task.Task
	environment            environment.Environment
	configs                map[string]nativeServiceConfig
	renderedCertificates   map[string]*certificates.RenderedCertificate
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

func (s *nativeService) Dirs() map[dir.DirType]string {
	return s.Sandbox().Dirs()
}

func (s *nativeService) rootPathDir(dirType dir.DirType) (string, error) {
	sandboxDir := s.Sandbox().Dirs()[dirType]
	rootPath := s.environment.RootPath(s.workspace)
	fullDir := filepath.Join(rootPath, sandboxDir, s.name)
	err := s.environment.Mkdir(s.name, fullDir, 0755)
	return fullDir, err
}

func (s *nativeService) ConfDir() (string, error) {
	return s.rootPathDir(dir.ConfDirType)
}

func (s *nativeService) RunDir() (string, error) {
	return s.rootPathDir(dir.RunDirType)
}

func (s *nativeService) ScriptDir() (string, error) {
	return s.rootPathDir(dir.ScriptDirType)
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

func (s *nativeService) OutputReader(ctx context.Context, outputType output.Type) (io.Reader, error) {
	if s.task == nil || reflect.ValueOf(s.task).IsNil() {
		return nil, errors.Errorf("service has not started yet")
	}

	reader, err := s.environment.Output(ctx, s.task, outputType)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (s *nativeService) renderingPaths(path string, dirType dir.DirType) (string, string, error) {
	environmentRootPath := s.environment.RootPath(s.workspace)
	sandboxDir, err := s.sandbox.Dir(dirType)
	if err != nil {
		return "", "", err
	}
	basePath := filepath.Base(path)
	wsPath := filepath.Join(s.workspace, sandboxDir, basePath)
	envPath := filepath.Join(environmentRootPath, sandboxDir, s.name, basePath)
	return wsPath, envPath, nil
}

func (s *nativeService) renderConfig(configParameters parameters.Parameters, configPath, wsPath string) error {
	file, err := s.fnd.Fs().Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	configContent, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = s.template.RenderToFile(string(configContent), configParameters, wsPath, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (s *nativeService) renderConfigs() error {
	envConfigPaths := make(map[string]string, len(s.configs))
	wsConfigPaths := make(map[string]string, len(s.configs))
	// fist get all config paths
	for cfgName, sc := range s.configs {
		wsPath, envPath, err := s.renderingPaths(sc.config.FilePath(), dir.ConfDirType)
		if err != nil {
			return err
		}
		envConfigPaths[cfgName] = envPath
		wsConfigPaths[cfgName] = wsPath
	}
	s.environmentConfigPaths = envConfigPaths
	s.workspaceConfigPaths = wsConfigPaths
	// and then render
	for cfgName, sc := range s.configs {
		err := s.renderConfig(sc.parameters, sc.config.FilePath(), wsConfigPaths[cfgName])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *nativeService) renderCertificatePart(name string, data string, mode os.FileMode) (string, error) {
	workspaceScriptPath, environmentScriptPath, err := s.renderingPaths(name, dir.CertDirType)
	if err != nil {
		return "", err
	}

	err = s.template.RenderToFile(data, parameters.Parameters{}, workspaceScriptPath, mode)
	if err != nil {
		return "", err
	}

	return environmentScriptPath, nil
}

func (s *nativeService) renderCertificates() error {
	renderedCerts := make(map[string]*certificates.RenderedCertificate, len(s.certificates))
	for name, cert := range s.certificates {
		envCertPath, err := s.renderCertificatePart(cert.CertificateName(), cert.CertificateData(), 0644)
		if err != nil {
			return err
		}
		envPrivKeyPath, err := s.renderCertificatePart(cert.PrivateKeyName(), cert.PrivateKeyData(), 0600)
		if err != nil {
			return err
		}
		renderedCerts[name] = &certificates.RenderedCertificate{
			Certificate:         cert,
			CertificateFilePath: envCertPath,
			PrivateKeyFilePath:  envPrivKeyPath,
		}
	}
	s.renderedCertificates = renderedCerts
	return nil
}

func (s *nativeService) renderScript(script scripts.Script, scriptName string) (string, string, error) {
	scriptPath := script.Path()
	if scriptPath == "" {
		scriptPath = scriptName
	}
	workspaceScriptPath, environmentScriptPath, err := s.renderingPaths(scriptPath, dir.ScriptDirType)
	if err != nil {
		return "", "", err
	}

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
		wsPath, envPath, err := s.renderScript(script, scriptName)
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
		UniqueName:             s.uniqueName,
		Port:                   s.port,
		ServerPort:             s.server.Port(),
		Public:                 s.public,
		ContainerConfig:        s.sandbox.ContainerConfig(),
		ServerParameters:       s.serverParameters,
		EnvironmentConfigPaths: s.environmentConfigPaths,
		EnvironmentScriptPaths: s.environmentScriptPaths,
		WorkspaceConfigPaths:   s.workspaceConfigPaths,
		WorkspaceScriptPaths:   s.workspaceScriptPaths,
	}
}

func (s *nativeService) ExecCommand(ctx context.Context, cmd *environment.Command, oc output.Collector) error {
	if s.task == nil || reflect.ValueOf(s.task).IsNil() {
		return errors.Errorf("service has not started yet")
	}

	return s.environment.ExecTaskCommand(ctx, s.makeEnvServiceSettings(), s.task, cmd, oc)
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

	return err
}

func (s *nativeService) Name() string {
	return s.name
}

func (s *nativeService) FullName() string {
	return s.fullName
}

func (s *nativeService) PrivateAddress() string {
	return s.environment.ServicePrivateAddress(s.name, s.port, s.server.Port())
}

func (s *nativeService) LocalAddress() string {
	return s.environment.ServiceLocalAddress(s.name, s.port, s.server.Port())
}

func (s *nativeService) LocalPort() int32 {
	return s.environment.ServiceLocalPort(s.port, s.server.Port())
}

func (s *nativeService) Executable() (string, error) {
	if s.task == nil || reflect.ValueOf(s.task).IsNil() {
		return "", errors.Errorf("service has not started yet")
	}

	return s.task.Executable(), nil
}

func (s *nativeService) PublicUrl(path string) (string, error) {
	if s.task == nil || reflect.ValueOf(s.task).IsNil() {
		return "", errors.Errorf("service has not started yet")
	}
	if !s.IsPublic() {
		return "", errors.Errorf("only public service has public URL")
	}

	return url.JoinPath(s.task.PublicUrl(), path)
}

func (s *nativeService) PrivateUrl() (string, error) {
	if s.task == nil || reflect.ValueOf(s.task).IsNil() {
		return "", errors.Errorf("service has not started yet")
	}

	return s.task.PrivateUrl(), nil
}

func (s *nativeService) UdsPath(args ...string) (string, error) {
	rd, err := s.RunDir()
	if err != nil {
		return "", err
	}
	var sockName string
	if len(args) > 0 && args[0] != "" {
		sockName = args[0]
	} else {
		sockName = s.name
	}
	return filepath.Join(rd, fmt.Sprintf("%s.sock", sockName)), nil
}

func (s *nativeService) Pid() (int, error) {
	if s.task == nil || reflect.ValueOf(s.task).IsNil() {
		return 0, errors.Errorf("service has not started yet")
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

func (s *nativeService) FindCertificate(name string) (*certificates.RenderedCertificate, error) {
	cert, ok := s.renderedCertificates[name]
	if !ok {
		return nil, errors.Errorf("certificate %s not found", name)
	}
	return cert, nil
}
