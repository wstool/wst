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

type EnvironmentType string

const (
	CommonEnvironmentType     EnvironmentType = "common"
	LocalEnvironmentType                      = "local"
	ContainerEnvironmentType                  = "container"
	DockerEnvironmentType                     = "docker"
	KubernetesEnvironmentType                 = "kubernetes"
)

type Environment interface {
}

type EnvironmentPorts struct {
	Start int32 `wst:"start"`
	End   int32 `wst:"end"`
}

type CommonEnvironment struct {
	Ports     EnvironmentPorts `wst:"ports"`
	Resources Resources        `wst:"resources"`
}

type LocalEnvironment struct {
	Ports     EnvironmentPorts `wst:"ports"`
	Resources Resources        `wst:"resources"`
}

type ContainerEnvironment struct {
	Ports     EnvironmentPorts  `wst:"ports"`
	Resources Resources         `wst:"resources"`
	Registry  ContainerRegistry `wst:"registry"`
}

type DockerEnvironment struct {
	Ports      EnvironmentPorts  `wst:"ports"`
	Resources  Resources         `wst:"resources"`
	Registry   ContainerRegistry `wst:"registry"`
	NamePrefix string            `wst:"name_prefix"`
}

type KubernetesEnvironment struct {
	Ports      EnvironmentPorts  `wst:"ports"`
	Resources  Resources         `wst:"resources"`
	Registry   ContainerRegistry `wst:"registry"`
	Namespace  string            `wst:"namespace"`
	Kubeconfig string            `wst:"kubeconfig,path=virtual"`
}
