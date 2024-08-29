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
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/environments/task"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/sandboxes/containers"
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
	Port                   int32
	Public                 bool
	ContainerConfig        *containers.ContainerConfig
	ServerPort             int32
	ServerParameters       parameters.Parameters
	EnvironmentConfigPaths map[string]string
	EnvironmentScriptPaths map[string]string
	WorkspaceConfigPaths   map[string]string
	WorkspaceScriptPaths   map[string]string
}

type Environment interface {
	Init(ctx context.Context) error
	Destroy(ctx context.Context) error
	RootPath(workspace string) string
	ServiceAddress(serviceName string, port int32) string
	RunTask(ctx context.Context, ss *ServiceSettings, cmd *Command) (task.Task, error)
	ExecTaskCommand(ctx context.Context, ss *ServiceSettings, target task.Task, cmd *Command) error
	ExecTaskSignal(ctx context.Context, ss *ServiceSettings, target task.Task, signal os.Signal) error
	Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error)
	PortsStart() int32
	PortsEnd() int32
	ReservePort() int32
	ContainerRegistry() *ContainerRegistry
	MarkUsed()
	IsUsed() bool
}

type Maker interface {
	MakeCommonEnvironment(config *types.CommonEnvironment) *CommonEnvironment
	MakeContainerEnvironment(config *types.ContainerEnvironment) *ContainerEnvironment
}

type CommonMaker struct {
	Fnd         app.Foundation
	OutputMaker output.Maker
}

func CreateCommonMaker(fnd app.Foundation) *CommonMaker {
	return &CommonMaker{
		Fnd:         fnd,
		OutputMaker: output.CreateMaker(fnd),
	}
}

type CommonEnvironment struct {
	Fnd         app.Foundation
	OutputMaker output.Maker
	Used        bool
	Ports       Ports
}

func (m *CommonMaker) MakeCommonEnvironment(config *types.CommonEnvironment) *CommonEnvironment {
	return &CommonEnvironment{
		Fnd:         m.Fnd,
		OutputMaker: m.OutputMaker,
		Used:        false,
		Ports: Ports{
			Start: config.Ports.Start,
			Used:  config.Ports.Start,
			End:   config.Ports.End,
		},
	}
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

func (e *CommonEnvironment) ContainerRegistry() *ContainerRegistry {
	return nil
}

type ContainerEnvironment struct {
	CommonEnvironment
	Registry ContainerRegistry
}

func (m *CommonMaker) MakeContainerEnvironment(config *types.ContainerEnvironment) *ContainerEnvironment {
	return &ContainerEnvironment{
		CommonEnvironment: *m.MakeCommonEnvironment(&types.CommonEnvironment{
			Ports: config.Ports,
		}),
		Registry: ContainerRegistry{
			Auth: ContainerRegistryAuth{
				Username: config.Registry.Auth.Username,
				Password: config.Registry.Auth.Password,
			},
		},
	}
}

func (e *ContainerEnvironment) ContainerRegistry() *ContainerRegistry {
	return &e.Registry
}
