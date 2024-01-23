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

package conf

import "os"

type Config struct {
	Version     string             `wst:"version"`
	Name        string             `wst:"name"`
	Description string             `wst:"description"`
	Sandboxes   map[string]Sandbox `wst:"sandboxes,loadable,factory=createSandboxes"`
	Servers     []Server           `wst:"servers,loadable"`
	Spec        Spec               `wst:"spec"`
}

type Parameter interface {
	GetInt() int
	GetFloat() float64
	GetString() string
	GetParameters() Parameters
}

type Parameters map[string]Parameter

type SandboxHookNative struct {
	Type string `wst:type`
}

type SandboxHookShellCommand struct {
	Command string `wst:"command"`
	Shell   string `wst:"shell"`
}

type SandboxHookCommand struct {
	Executable string   `wst:"executable"`
	Args       []string `wst:"args"`
}

type SandboxHookSignal os.Signal

type SandboxHook interface {
	Execute(sandbox *Sandbox) error
}

const (
	CommonSandboxHook     string = "common"
	LocalSandboxHook             = "local"
	ContainerSandboxHook         = "container"
	DockerSandboxHook            = "docker"
	KubernetesSandboxHook        = "kubernetes"
)

const (
	SandboxHookStartType  string = "start"
	SandboxHookStopType          = "stop"
	SandboxHookReloadType        = "reload"
)

type Sandbox interface {
	ExecuteCommand(command *SandboxHookCommand) error
	ExecuteSignal(signal *SandboxHookSignal) error
}

type CommonSandbox struct {
	Dirs  map[string]string      `wst:"dirs,keys=conf|run|script"`
	Hooks map[string]SandboxHook `wst:"hooks,factory=createHooks"`
}

type LocalSandbox struct {
	CommonSandbox
}

type ContainerImage struct {
	Name string `wst:"name"`
	Tag  string `wst:"tag"`
}

type ContainerRegistryAuth struct {
	Username string `wst:"username"`
	Password string `wst:"password"`
}

type ContainerRegistry struct {
	Auth ContainerRegistryAuth `wst:"auth"`
}

type ContainerSandbox struct {
	CommonSandbox
	Image    ContainerImage    `wst:"image,factory=createContainerImage"`
	Registry ContainerRegistry `wst:"registry"`
}

type DockerSandbox struct {
	ContainerSandbox
}

type KubernetesAuth struct {
	Kubeconfig string `wst:"kubeconfig"`
}

type KubernetesSandbox struct {
	ContainerSandbox
	Auth KubernetesAuth `wst:"auth"`
}

type ServerConfig struct {
	File       string     `wst:"file"`
	Parameters Parameters `wst:"parameters,factory=createParameters"`
}

type Server struct {
	Name         string                  `wst:"name"`
	Extends      string                  `wst:"extends"`
	Configs      map[string]ServerConfig `wst:"configs"`
	Parameters   Parameters              `wst:"parameters,factory=createParameters"`
	Expectations map[string]Expectation  `wst:"expectations,factory=createExpectations"`
}

type OutputExpectation struct {
	Order    string   `wst:"order"`
	Match    string   `wst:"match"`
	Messages []string `wst:"messages"`
}

type OutputExpectationWrapper struct {
	Parameters Parameters        `wst:"parameters,factory=createParameters"`
	Output     OutputExpectation `wst:"output"`
}

type Headers map[string]string

type ResponseBody struct {
	Content        string `wst:"content"`
	Match          string `wst:"match"`
	RenderTemplate string `wst:"render_template"`
}

type ResponseExpectation struct {
	Request string       `wst:"request"`
	Headers Headers      `wst:"headers"`
	Body    ResponseBody `wst:"content,string=Content"`
}

type ResponseExpectationWrapper struct {
	Parameters Parameters          `wst:"parameters,factory=createParameters"`
	Response   ResponseExpectation `wst:"response"`
}

type Expectation interface {
	Verify(runtime *ActionRuntime) error
}

type Script struct {
	Content string `wst:"content"`
	Path    string `wst:"path"`
	Mode    string `wst:"mode"`
}

type ServiceConfig struct {
	Parameters          Parameters `wst:"parameters,factory=createParameters"`
	OverwriteParameters bool       `wst:"overwrite_parameters,factory=createParameters"`
}

type Service struct {
	Server  string                   `wst:"server"`
	Sandbox string                   `wst:"sandbox"`
	Scripts []string                 `wst:"scripts,bool=local|docker|kubernetes,enum=local|docker|kubernetes"`
	Configs map[string]ServiceConfig `wst:"configs"`
}

type Instance struct {
	Name     string             `wst:"name"`
	Scripts  map[string]Script  `wst:"scripts,string=Content"`
	Services map[string]Service `wst:"services"`
}

type Spec struct {
	Workspace string     `wst:"workspace"`
	Instances []Instance `wst:"instances,loadable"`
}

type ActionRuntime interface {
}
