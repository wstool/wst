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

package sandboxes

import (
	"errors"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/sandboxes/hooks"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/sandboxes/sandbox/common"
	"github.com/bukka/wst/run/sandboxes/sandbox/container"
	"github.com/bukka/wst/run/sandboxes/sandbox/docker"
	"github.com/bukka/wst/run/sandboxes/sandbox/kubernetes"
	"github.com/bukka/wst/run/sandboxes/sandbox/local"
)

type Sandboxes map[providers.Type]sandbox.Sandbox

type Maker struct {
	env             app.Env
	localMaker      *local.Maker
	dockerMaker     *docker.Maker
	kubernetesMaker *kubernetes.Maker
}

func CreateMaker(env app.Env) *Maker {
	hooksMaker := hooks.CreateMaker(env)
	commonMaker := common.CreateMaker(env, hooksMaker)
	containerMaker := container.CreateMaker(env, commonMaker)
	return &Maker{
		env:             env,
		localMaker:      local.CreateMaker(env, commonMaker),
		dockerMaker:     docker.CreateMaker(env, containerMaker),
		kubernetesMaker: kubernetes.CreateMaker(env, containerMaker),
	}
}

func (m *Maker) MakeSandboxes(
	rootSandboxes map[types.SandboxType]types.Sandbox,
	serverSandboxes map[types.SandboxType]types.Sandbox,
) (Sandboxes, error) {
	var err error
	mergedSandboxes, err := m.mergeConfigMaps(rootSandboxes, serverSandboxes)
	if err != nil {
		return nil, err
	}

	commonSb, commonFound := mergedSandboxes[types.CommonSandboxType]
	localSb, localFound := mergedSandboxes[types.LocalSandboxType]
	containerSb, containerFound := mergedSandboxes[types.ContainerSandboxType]
	dockerSb, dockerFound := mergedSandboxes[types.ContainerSandboxType]
	kubernetesSb, kubernetesFound := mergedSandboxes[types.ContainerSandboxType]
	if commonFound {
		// Local merging
		localSb, err = m.mergeLocalAndCommon(localSb, commonSb)
		if err != nil {
			return nil, err
		}
		localFound = true
		// Container merging
		containerSb, err = m.mergeLocalAndCommon(containerSb, commonSb)
		if err != nil {
			return nil, err
		}
		containerFound = true
	}
	if containerFound {
		// Docker merging
		dockerSb, err = m.mergeDockerAndContainer(dockerSb, containerSb)
		if err != nil {
			return nil, err
		}
		dockerFound = true
		// Kubernetes merging
		kubernetesSb, err = m.mergeKubernetesAndContainer(kubernetesSb, containerSb)
		if err != nil {
			return nil, err
		}
		kubernetesFound = true
	}

	sandboxes := make(Sandboxes)

	if localFound {
		sandboxes[providers.LocalType], err = m.localMaker.MakeSandbox(localSb.(*types.LocalSandbox))
		if err != nil {
			return nil, err
		}
	}
	if dockerFound {
		sandboxes[providers.DockerType], err = m.dockerMaker.MakeSandbox(dockerSb.(*types.DockerSandbox))
		if err != nil {
			return nil, err
		}
	}
	if kubernetesFound {
		sandboxes[providers.KubernetesType], err = m.kubernetesMaker.MakeSandbox(dockerSb.(*types.KubernetesSandbox))
		if err != nil {
			return nil, err
		}
	}

	return sandboxes, nil
}

func (m *Maker) mergeLocalAndCommon(local, common types.Sandbox) (types.Sandbox, error) {
	localSandbox, localSandboxOk := local.(*types.LocalSandbox)
	if !localSandboxOk {
		return nil, errors.New("type assertion to *LocalSandbox failed")
	}
	mergedCommon, err := m.mergeCommonSandbox(&localSandbox.CommonSandbox, common)
	if err != nil {
		return nil, err
	}
	localSandbox.CommonSandbox = *mergedCommon.(*types.CommonSandbox)

	return localSandbox, nil
}

func (m *Maker) mergeContainerAndCommon(container, common types.Sandbox) (types.Sandbox, error) {
	containerSandbox, containerSandboxOk := container.(*types.ContainerSandbox)
	if !containerSandboxOk {
		return nil, errors.New("type assertion to *ContainerSandbox failed")
	}
	mergedCommon, err := m.mergeCommonSandbox(&containerSandbox.CommonSandbox, common)
	if err != nil {
		return nil, err
	}
	containerSandbox.CommonSandbox = *mergedCommon.(*types.CommonSandbox)

	return containerSandbox, nil
}

func (m *Maker) mergeDockerAndContainer(docker, container types.Sandbox) (types.Sandbox, error) {
	dockerSandbox, dockerSandboxOk := docker.(*types.DockerSandbox)
	if !dockerSandboxOk {
		return nil, errors.New("type assertion to *DockerSandbox failed")
	}
	mergedContainer, err := m.mergeCommonSandbox(&dockerSandbox.ContainerSandbox, docker)
	if err != nil {
		return nil, err
	}
	dockerSandbox.ContainerSandbox = *mergedContainer.(*types.ContainerSandbox)

	return dockerSandbox, nil
}

func (m *Maker) mergeKubernetesAndContainer(kubernetes, container types.Sandbox) (types.Sandbox, error) {
	kubernetesSandbox, kubernetesSandboxOk := kubernetes.(*types.KubernetesSandbox)
	if !kubernetesSandboxOk {
		return nil, errors.New("type assertion to *KubernetesSandbox failed")
	}
	mergedContainer, err := m.mergeCommonSandbox(&kubernetesSandbox.ContainerSandbox, kubernetes)
	if err != nil {
		return nil, err
	}
	kubernetesSandbox.ContainerSandbox = *mergedContainer.(*types.ContainerSandbox)

	return kubernetesSandbox, nil
}

type mergeFunc func(root, server types.Sandbox) (types.Sandbox, error)

func (m *Maker) mergeConfigMaps(
	rootSandboxes map[types.SandboxType]types.Sandbox,
	serverSandboxes map[types.SandboxType]types.Sandbox,
) (map[types.SandboxType]types.Sandbox, error) {
	mergeFuncs := map[types.SandboxType]mergeFunc{
		types.CommonSandboxType:     m.mergeCommonSandbox,
		types.LocalSandboxType:      m.mergeLocalSandbox,
		types.ContainerSandboxType:  m.mergeContainerSandbox,
		types.DockerSandboxType:     m.mergeDockerSandbox,
		types.KubernetesSandboxType: m.mergeKubernetesSandbox,
	}

	for sandboxType, merge := range mergeFuncs {
		rootSandbox, rootExists := rootSandboxes[sandboxType]
		serverSandbox, serverExists := serverSandboxes[sandboxType]

		if rootExists && serverExists {
			// Use the merge function, now handling errors.
			mergedSandbox, err := merge(rootSandbox, serverSandbox)
			if err != nil {
				// Handle the error, e.g., by returning it or logging it.
				return nil, err // Return an error if merging fails.
			}
			rootSandboxes[sandboxType] = mergedSandbox
		} else if !rootExists && serverExists {
			rootSandboxes[sandboxType] = serverSandbox
		}
		// No action needed if only root exists or neither exists.
	}

	return rootSandboxes, nil // No error encountered, return the merged map.
}

func (m *Maker) mergeCommonSandbox(root, server types.Sandbox) (types.Sandbox, error) {
	// Ensure both root and server are of the correct type, using type assertion to *CommonSandbox.
	rootCommon, rootOk := root.(*types.CommonSandbox)
	serverCommon, serverOk := server.(*types.CommonSandbox)
	if !rootOk || !serverOk {
		return nil, errors.New("type assertion to *CommonSandbox failed")
	}

	// Create a new instance of CommonSandbox for the merged result.
	mergedCommon := &types.CommonSandbox{
		Dirs:  make(map[string]string),
		Hooks: make(map[string]types.SandboxHook),
	}

	// Copy from root to mergedCommon.
	for k, v := range rootCommon.Dirs {
		mergedCommon.Dirs[k] = v
	}
	for k, v := range rootCommon.Hooks {
		mergedCommon.Hooks[k] = v
	}

	// Merge from server into mergedCommon, overwriting or adding new entries.
	for k, v := range serverCommon.Dirs {
		mergedCommon.Dirs[k] = v
	}
	for k, v := range serverCommon.Hooks {
		mergedCommon.Hooks[k] = v
	}

	// Return the new, merged CommonSandbox as a Sandbox interface.
	return mergedCommon, nil
}

func (m *Maker) mergeLocalSandbox(root, server types.Sandbox) (types.Sandbox, error) {
	// Ensure both root and server are of the correct type, using type assertion to *CommonSandbox.
	_, rootOk := root.(*types.LocalSandbox)
	_, serverOk := server.(*types.LocalSandbox)
	if !rootOk || !serverOk {
		return nil, errors.New("type assertion to *LocalSandbox failed")
	}

	mergedCommon, err := m.mergeCommonSandbox(root, server)
	if err != nil {
		return nil, err
	}

	mergedLocal := &types.LocalSandbox{
		CommonSandbox: *mergedCommon.(*types.CommonSandbox),
	}

	return mergedLocal, nil
}

func (m *Maker) mergeContainerSandbox(root, server types.Sandbox) (types.Sandbox, error) {
	rootContainer, rootOk := root.(*types.ContainerSandbox)
	serverContainer, serverOk := server.(*types.ContainerSandbox)
	if !rootOk || !serverOk {
		return nil, errors.New("type assertion to *ContainerSandbox failed")
	}

	mergedCommon, err := m.mergeCommonSandbox(root, server)
	if err != nil {
		return nil, err
	}

	mergedContainer := &types.ContainerSandbox{
		CommonSandbox: *mergedCommon.(*types.CommonSandbox),
		Image:         rootContainer.Image,
		Registry:      rootContainer.Registry,
	}

	if serverContainer.Image.Name != "" {
		mergedContainer.Image.Name = serverContainer.Image.Name
	}
	if serverContainer.Image.Tag != "" {
		mergedContainer.Image.Tag = serverContainer.Image.Tag
	}
	if serverContainer.Registry.Auth.Username != "" {
		mergedContainer.Registry.Auth.Username = serverContainer.Registry.Auth.Username
	}
	if serverContainer.Registry.Auth.Password != "" {
		mergedContainer.Registry.Auth.Password = serverContainer.Registry.Auth.Password
	}

	return mergedContainer, nil
}

func (m *Maker) mergeDockerSandbox(root, server types.Sandbox) (types.Sandbox, error) {
	_, rootOk := root.(*types.DockerSandbox)
	_, serverOk := server.(*types.DockerSandbox)
	if !rootOk || !serverOk {
		return nil, errors.New("type assertion to *DockerSandbox failed")
	}

	mergedContainer, err := m.mergeContainerSandbox(root, server)
	if err != nil {
		return nil, err
	}

	mergedDocker := &types.DockerSandbox{
		ContainerSandbox: *mergedContainer.(*types.ContainerSandbox),
	}

	return mergedDocker, nil
}

func (m *Maker) mergeKubernetesSandbox(root, server types.Sandbox) (types.Sandbox, error) {
	rootKubernetes, rootOk := root.(*types.KubernetesSandbox)
	serverKubernetes, serverOk := server.(*types.KubernetesSandbox)
	if !rootOk || !serverOk {
		return nil, errors.New("type assertion to *KubernetesSandbox failed")
	}

	mergedContainer, err := m.mergeContainerSandbox(root, server)
	if err != nil {
		return nil, err
	}

	mergedKubernetes := &types.KubernetesSandbox{
		ContainerSandbox: *mergedContainer.(*types.ContainerSandbox),
		Auth:             rootKubernetes.Auth,
	}

	if serverKubernetes.Auth.Kubeconfig != "" {
		mergedKubernetes.Auth.Kubeconfig = serverKubernetes.Auth.Kubeconfig
	}

	return mergedKubernetes, nil
}
