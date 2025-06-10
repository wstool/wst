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

package types

type SandboxType string

const (
	CommonSandboxType     SandboxType = "common"
	LocalSandboxType                  = "local"
	ContainerSandboxType              = "container"
	DockerSandboxType                 = "docker"
	KubernetesSandboxType             = "kubernetes"
)

type SandboxHookType string

const (
	ReloadSandboxHookType  SandboxHookType = "reload"
	RestartSandboxHookType                 = "restart"
	StartSandboxHookType                   = "start"
	StopSandboxHookType                    = "stop"
)

type SandboxHookNative struct {
	Enabled bool `wst:"enabled,default=true"`
	Force   bool `wst:"force,default=false"`
}

type SandboxHookShellCommand struct {
	Command string `wst:"command"`
	Shell   string `wst:"shell"`
}

type SandboxHookArgsCommand struct {
	Executable string   `wst:"executable"`
	Args       []string `wst:"args"`
}

type SandboxHookSignal struct {
	IsString    bool
	StringValue string
	IntValue    int
}

type SandboxHook interface {
}

type CommonSandbox struct {
	Available bool                   `wst:"available,default=true"`
	Dirs      map[string]string      `wst:"dirs,keys=cert|conf|run|script"`
	Hooks     map[string]SandboxHook `wst:"hooks,factory=createHooks"`
}

type LocalSandbox struct {
	Available bool                   `wst:"available,default=true"`
	Dirs      map[string]string      `wst:"dirs,keys=cert|conf|run|script"`
	Hooks     map[string]SandboxHook `wst:"hooks,factory=createHooks"`
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
	Available bool                   `wst:"available,default=true"`
	Dirs      map[string]string      `wst:"dirs,keys=cert|conf|run|script"`
	Hooks     map[string]SandboxHook `wst:"hooks,factory=createHooks"`
	Image     ContainerImage         `wst:"image,factory=createContainerImage"`
	Registry  ContainerRegistry      `wst:"registry"`
}

type DockerSandbox struct {
	Available bool                   `wst:"available,default=true"`
	Dirs      map[string]string      `wst:"dirs,keys=cert|conf|run|script"`
	Hooks     map[string]SandboxHook `wst:"hooks,factory=createHooks"`
	Image     ContainerImage         `wst:"image,factory=createContainerImage"`
	Registry  ContainerRegistry      `wst:"registry"`
}

type KubernetesSandbox struct {
	Available bool                   `wst:"available,default=true"`
	Dirs      map[string]string      `wst:"dirs,keys=cert|conf|run|script"`
	Hooks     map[string]SandboxHook `wst:"hooks,factory=createHooks"`
	Image     ContainerImage         `wst:"image,factory=createContainerImage"`
	Registry  ContainerRegistry      `wst:"registry"`
}

type Sandbox interface{}
