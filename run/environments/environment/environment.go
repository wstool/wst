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

package environment

import (
	"context"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/environments/task"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/resources"
	"github.com/wstool/wst/run/resources/certificates"
	"github.com/wstool/wst/run/sandboxes/containers"
	"io"
	"os"
)

type Command struct {
	Name string
	Args []string
}

type Ports struct {
	Start int32
	Used  int32
	End   int32
}

type ContainerRegistryAuth struct {
	Username string
	Password string
}

type ContainerRegistry struct {
	Auth ContainerRegistryAuth
}

type ServiceSettings struct {
	Name                   string
	FullName               string
	UniqueName             string
	Port                   int32
	Public                 bool
	ContainerConfig        *containers.ContainerConfig
	ServerPort             int32
	ServerParameters       parameters.Parameters
	EnvironmentConfigPaths map[string]string
	EnvironmentScriptPaths map[string]string
	WorkspaceConfigPaths   map[string]string
	WorkspaceScriptPaths   map[string]string
	Certificates           map[string]*certificates.RenderedCertificate
}

type Environment interface {
	Init(ctx context.Context) error
	Destroy(ctx context.Context) error
	RootPath(workspace string) string
	Mkdir(serviceName string, path string, perm os.FileMode) error
	ServiceLocalAddress(serviceName string, servicePort, serverPort int32) string
	ServiceLocalPort(servicePort, serverPort int32) int32
	ServicePrivateAddress(serviceName string, servicePort, serverPort int32) string
	RunTask(ctx context.Context, ss *ServiceSettings, cmd *Command) (task.Task, error)
	ExecTaskCommand(ctx context.Context, ss *ServiceSettings, target task.Task, cmd *Command, oc output.Collector) error
	ExecTaskSignal(ctx context.Context, ss *ServiceSettings, target task.Task, signal os.Signal) error
	Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error)
	PortsStart() int32
	PortsEnd() int32
	ReservePort() int32
	Resources() *resources.Resources
	ContainerRegistry() *ContainerRegistry
	MarkUsed()
	IsUsed() bool
}

type Maker interface {
	MakeCommonEnvironment(config *types.CommonEnvironment) (*CommonEnvironment, error)
	MakeLocalEnvironment(config *types.LocalEnvironment) (*LocalEnvironment, error)
	MakeContainerEnvironment(config *types.ContainerEnvironment) (*ContainerEnvironment, error)
	MakeDockerEnvironment(config *types.DockerEnvironment) (*DockerEnvironment, error)
	MakeKubernetesEnvironment(config *types.KubernetesEnvironment) (*KubernetesEnvironment, error)
}

type CommonMaker struct {
	Fnd            app.Foundation
	ResourcesMaker resources.Maker
	OutputMaker    output.Maker
}

func CreateCommonMaker(fnd app.Foundation, resourcesMaker resources.Maker) *CommonMaker {
	return &CommonMaker{
		Fnd:            fnd,
		ResourcesMaker: resourcesMaker,
		OutputMaker:    output.CreateMaker(fnd),
	}
}

// Merge helper functions
func mergePorts(base, override types.EnvironmentPorts) types.EnvironmentPorts {
	result := base
	if override.Start != 0 {
		result.Start = override.Start
	}
	if override.End != 0 {
		result.End = override.End
	}
	return result
}

func mergeResources(base, override types.Resources) types.Resources {
	result := types.Resources{
		Certificates: make(map[string]types.Certificate),
		Scripts:      make(map[string]types.Script),
	}

	// Copy base certificates
	for name, cert := range base.Certificates {
		result.Certificates[name] = cert
	}

	// Copy base scripts
	for name, script := range base.Scripts {
		result.Scripts[name] = script
	}

	// Override with new certificates (overwrites existing ones with same name)
	for name, cert := range override.Certificates {
		result.Certificates[name] = cert
	}

	// Override with new scripts (overwrites existing ones with same name)
	for name, script := range override.Scripts {
		result.Scripts[name] = script
	}

	return result
}

func mergeContainerRegistry(base, override types.ContainerRegistry) types.ContainerRegistry {
	result := base
	if override.Auth.Username != "" {
		result.Auth.Username = override.Auth.Username
	}
	if override.Auth.Password != "" {
		result.Auth.Password = override.Auth.Password
	}
	return result
}

// Environment implementations
type CommonEnvironment struct {
	Fnd          app.Foundation
	OutputMaker  output.Maker
	Used         bool
	Ports        Ports
	EnvResources *resources.Resources
}

func (m *CommonMaker) MakeCommonEnvironment(config *types.CommonEnvironment) (*CommonEnvironment, error) {
	rscs, err := m.ResourcesMaker.Make(config.Resources)
	if err != nil {
		return nil, err
	}
	return &CommonEnvironment{
		Fnd:         m.Fnd,
		OutputMaker: m.OutputMaker,
		Used:        false,
		Ports: Ports{
			Start: config.Ports.Start,
			Used:  config.Ports.Start,
			End:   config.Ports.End,
		},
		EnvResources: rscs,
	}, nil
}

func (e *CommonEnvironment) MarkUsed() {
	e.Used = true
}

func (e *CommonEnvironment) IsUsed() bool {
	return e.Used
}

func (e *CommonEnvironment) PortsStart() int32 {
	return e.Ports.Start
}

func (e *CommonEnvironment) PortsEnd() int32 {
	return e.Ports.End
}

func (e *CommonEnvironment) ReservePort() int32 {
	used := e.Ports.Used
	e.Ports.Used++
	return used
}

func (e *CommonEnvironment) Resources() *resources.Resources {
	return e.EnvResources
}

func (e *CommonEnvironment) ContainerRegistry() *ContainerRegistry {
	return nil
}

type LocalEnvironment struct {
	CommonEnvironment
}

func (m *CommonMaker) MakeLocalEnvironment(config *types.LocalEnvironment) (*LocalEnvironment, error) {
	commonEnv, err := m.MakeCommonEnvironment(&types.CommonEnvironment{
		Ports:     config.Ports,
		Resources: config.Resources,
	})
	if err != nil {
		return nil, err
	}
	return &LocalEnvironment{
		CommonEnvironment: *commonEnv,
	}, nil
}

type ContainerEnvironment struct {
	CommonEnvironment
	Registry ContainerRegistry
}

func (m *CommonMaker) MakeContainerEnvironment(config *types.ContainerEnvironment) (*ContainerEnvironment, error) {
	commonEnv, err := m.MakeCommonEnvironment(&types.CommonEnvironment{
		Ports:     config.Ports,
		Resources: config.Resources,
	})
	if err != nil {
		return nil, err
	}
	return &ContainerEnvironment{
		CommonEnvironment: *commonEnv,
		Registry: ContainerRegistry{
			Auth: ContainerRegistryAuth{
				Username: config.Registry.Auth.Username,
				Password: config.Registry.Auth.Password,
			},
		},
	}, nil
}

func (e *ContainerEnvironment) ContainerRegistry() *ContainerRegistry {
	return &e.Registry
}

type DockerEnvironment struct {
	ContainerEnvironment
	NamePrefix string
}

func (m *CommonMaker) MakeDockerEnvironment(config *types.DockerEnvironment) (*DockerEnvironment, error) {
	containerEnv, err := m.MakeContainerEnvironment(&types.ContainerEnvironment{
		Ports:     config.Ports,
		Resources: config.Resources,
		Registry:  config.Registry,
	})
	if err != nil {
		return nil, err
	}
	return &DockerEnvironment{
		ContainerEnvironment: *containerEnv,
		NamePrefix:           config.NamePrefix,
	}, nil
}

type KubernetesEnvironment struct {
	ContainerEnvironment
	Namespace  string
	Kubeconfig string
}

func (m *CommonMaker) MakeKubernetesEnvironment(config *types.KubernetesEnvironment) (*KubernetesEnvironment, error) {
	containerEnv, err := m.MakeContainerEnvironment(&types.ContainerEnvironment{
		Ports:     config.Ports,
		Resources: config.Resources,
		Registry:  config.Registry,
	})
	if err != nil {
		return nil, err
	}
	return &KubernetesEnvironment{
		ContainerEnvironment: *containerEnv,
		Namespace:            config.Namespace,
		Kubeconfig:           config.Kubeconfig,
	}, nil
}

// Environment merging functions for inheritance
func (m *CommonMaker) MergeCommonEnvironments(base, override *types.CommonEnvironment) *types.CommonEnvironment {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	return &types.CommonEnvironment{
		Ports:     mergePorts(base.Ports, override.Ports),
		Resources: mergeResources(base.Resources, override.Resources),
	}
}

func (m *CommonMaker) MergeLocalEnvironments(base, override *types.LocalEnvironment) *types.LocalEnvironment {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	return &types.LocalEnvironment{
		Ports:     mergePorts(base.Ports, override.Ports),
		Resources: mergeResources(base.Resources, override.Resources),
	}
}

func (m *CommonMaker) MergeContainerEnvironments(base, override *types.ContainerEnvironment) *types.ContainerEnvironment {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	return &types.ContainerEnvironment{
		Ports:     mergePorts(base.Ports, override.Ports),
		Resources: mergeResources(base.Resources, override.Resources),
		Registry:  mergeContainerRegistry(base.Registry, override.Registry),
	}
}

func (m *CommonMaker) MergeDockerEnvironments(base, override *types.DockerEnvironment) *types.DockerEnvironment {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := &types.DockerEnvironment{
		Ports:     mergePorts(base.Ports, override.Ports),
		Resources: mergeResources(base.Resources, override.Resources),
		Registry:  mergeContainerRegistry(base.Registry, override.Registry),
	}

	if override.NamePrefix != "" {
		result.NamePrefix = override.NamePrefix
	} else {
		result.NamePrefix = base.NamePrefix
	}

	return result
}

func (m *CommonMaker) MergeKubernetesEnvironments(base, override *types.KubernetesEnvironment) *types.KubernetesEnvironment {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := &types.KubernetesEnvironment{
		Ports:     mergePorts(base.Ports, override.Ports),
		Resources: mergeResources(base.Resources, override.Resources),
		Registry:  mergeContainerRegistry(base.Registry, override.Registry),
	}

	if override.Namespace != "" {
		result.Namespace = override.Namespace
	} else {
		result.Namespace = base.Namespace
	}

	if override.Kubeconfig != "" {
		result.Kubeconfig = override.Kubeconfig
	} else {
		result.Kubeconfig = base.Kubeconfig
	}

	return result
}
