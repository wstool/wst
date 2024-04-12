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

func (a Sandboxes) Inherit(parentSandboxes Sandboxes) error {
	for sandboxName, parentSandbox := range parentSandboxes {
		sb, ok := a[sandboxName]
		if ok {
			err := sb.Inherit(parentSandbox)
			if err != nil {
				return err
			}
		} else {
			a[sandboxName] = parentSandbox
		}
	}

	return nil
}

type Maker struct {
	fnd             app.Foundation
	localMaker      *local.Maker
	dockerMaker     *docker.Maker
	kubernetesMaker *kubernetes.Maker
}

func CreateMaker(fnd app.Foundation) *Maker {
	hooksMaker := hooks.CreateMaker(fnd)
	commonMaker := common.CreateMaker(fnd, hooksMaker)
	containerMaker := container.CreateMaker(fnd, commonMaker)
	return &Maker{
		fnd:             fnd,
		localMaker:      local.CreateMaker(fnd, commonMaker),
		dockerMaker:     docker.CreateMaker(fnd, containerMaker),
		kubernetesMaker: kubernetes.CreateMaker(fnd, containerMaker),
	}
}

func (m *Maker) MakeSandboxes(
	rootSandboxes map[string]types.Sandbox,
	serverSandboxes map[string]types.Sandbox,
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
	mergedCommon, err := m.mergeCommonSandbox(&types.CommonSandbox{
		Available: localSandbox.Available,
		Dirs:      localSandbox.Dirs,
		Hooks:     localSandbox.Hooks,
	}, common)
	if err != nil {
		return nil, err
	}
	mergedCommonDeref := mergedCommon.(*types.CommonSandbox)
	localSandbox.Available = mergedCommonDeref.Available
	localSandbox.Dirs = mergedCommonDeref.Dirs
	localSandbox.Hooks = mergedCommonDeref.Hooks

	return localSandbox, nil
}

func (m *Maker) mergeContainerAndCommon(container, common types.Sandbox) (types.Sandbox, error) {
	containerSandbox, containerSandboxOk := container.(*types.ContainerSandbox)
	if !containerSandboxOk {
		return nil, errors.New("type assertion to *ContainerSandbox failed")
	}
	mergedCommon, err := m.mergeCommonSandbox(&types.CommonSandbox{
		Available: containerSandbox.Available,
		Dirs:      containerSandbox.Dirs,
		Hooks:     containerSandbox.Hooks,
	}, common)
	if err != nil {
		return nil, err
	}
	mergedCommonDeref := mergedCommon.(*types.CommonSandbox)
	containerSandbox.Available = mergedCommonDeref.Available
	containerSandbox.Dirs = mergedCommonDeref.Dirs
	containerSandbox.Hooks = mergedCommonDeref.Hooks

	return containerSandbox, nil
}

func (m *Maker) mergeDockerAndContainer(docker, container types.Sandbox) (types.Sandbox, error) {
	dockerSandbox, dockerSandboxOk := docker.(*types.DockerSandbox)
	if !dockerSandboxOk {
		return nil, errors.New("type assertion to *DockerSandbox failed")
	}
	mergedContainer, err := m.mergeContainerSandbox(&types.ContainerSandbox{
		Available: dockerSandbox.Available,
		Dirs:      dockerSandbox.Dirs,
		Hooks:     dockerSandbox.Hooks,
		Image:     dockerSandbox.Image,
		Registry:  dockerSandbox.Registry,
	}, container)
	if err != nil {
		return nil, err
	}
	mergedContainerDeref := mergedContainer.(*types.ContainerSandbox)
	dockerSandbox.Available = mergedContainerDeref.Available
	dockerSandbox.Dirs = mergedContainerDeref.Dirs
	dockerSandbox.Hooks = mergedContainerDeref.Hooks
	dockerSandbox.Image = mergedContainerDeref.Image
	dockerSandbox.Registry = mergedContainerDeref.Registry

	return dockerSandbox, nil
}

func (m *Maker) mergeKubernetesAndContainer(kubernetes, container types.Sandbox) (types.Sandbox, error) {
	kubernetesSandbox, kubernetesSandboxOk := kubernetes.(*types.KubernetesSandbox)
	if !kubernetesSandboxOk {
		return nil, errors.New("type assertion to *KubernetesSandbox failed")
	}
	mergedContainer, err := m.mergeContainerSandbox(&types.ContainerSandbox{
		Available: kubernetesSandbox.Available,
		Dirs:      kubernetesSandbox.Dirs,
		Hooks:     kubernetesSandbox.Hooks,
		Image:     kubernetesSandbox.Image,
		Registry:  kubernetesSandbox.Registry,
	}, container)
	if err != nil {
		return nil, err
	}
	mergedContainerDeref := mergedContainer.(*types.ContainerSandbox)
	kubernetesSandbox.Available = mergedContainerDeref.Available
	kubernetesSandbox.Dirs = mergedContainerDeref.Dirs
	kubernetesSandbox.Hooks = mergedContainerDeref.Hooks
	kubernetesSandbox.Image = mergedContainerDeref.Image
	kubernetesSandbox.Registry = mergedContainerDeref.Registry

	return kubernetesSandbox, nil
}

type mergeFunc func(root, server types.Sandbox) (types.Sandbox, error)

func (m *Maker) mergeConfigMaps(
	rootSandboxes map[string]types.Sandbox,
	serverSandboxes map[string]types.Sandbox,
) (map[types.SandboxType]types.Sandbox, error) {
	mergeFuncs := map[types.SandboxType]mergeFunc{
		types.CommonSandboxType:     m.mergeCommonSandbox,
		types.LocalSandboxType:      m.mergeLocalSandbox,
		types.ContainerSandboxType:  m.mergeContainerSandbox,
		types.DockerSandboxType:     m.mergeDockerSandbox,
		types.KubernetesSandboxType: m.mergeKubernetesSandbox,
	}
	mergedSandboxes := make(map[types.SandboxType]types.Sandbox)

	for sandboxType, merge := range mergeFuncs {
		sandboxTypeStr := string(sandboxType)
		rootSandbox, rootExists := rootSandboxes[sandboxTypeStr]
		serverSandbox, serverExists := serverSandboxes[sandboxTypeStr]

		if rootExists && serverExists {
			// Use the merge function, now handling errors.
			mergedSandbox, err := merge(rootSandbox, serverSandbox)
			if err != nil {
				// Handle the error, e.g., by returning it or logging it.
				return nil, err // Return an error if merging fails.
			}
			mergedSandboxes[sandboxType] = mergedSandbox
		} else if !rootExists && serverExists {
			mergedSandboxes[sandboxType] = serverSandbox
		} else {
			mergedSandboxes[sandboxType] = rootSandbox
		}
	}

	return mergedSandboxes, nil
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

	// Available is always set from the server
	mergedCommon.Available = serverCommon.Available

	// Return the new, merged CommonSandbox as a Sandbox interface.
	return mergedCommon, nil
}

func (m *Maker) mergeLocalSandbox(spec, server types.Sandbox) (types.Sandbox, error) {
	// Ensure both spec and server are of the correct type, using type assertion to *CommonSandbox.
	_, specOk := spec.(*types.LocalSandbox)
	_, serverOk := server.(*types.LocalSandbox)
	if !specOk || !serverOk {
		return nil, errors.New("type assertion to *LocalSandbox failed")
	}

	mergedCommon, err := m.mergeCommonSandbox(spec, server)
	if err != nil {
		return nil, err
	}

	mergedCommonDeref := mergedCommon.(*types.CommonSandbox)
	mergedLocal := &types.LocalSandbox{
		Available: mergedCommonDeref.Available,
		Dirs:      mergedCommonDeref.Dirs,
		Hooks:     mergedCommonDeref.Hooks,
	}

	return mergedLocal, nil
}

func (m *Maker) mergeContainerSandbox(spec, server types.Sandbox) (types.Sandbox, error) {
	specContainer, specOk := spec.(*types.ContainerSandbox)
	serverContainer, serverOk := server.(*types.ContainerSandbox)
	if !specOk || !serverOk {
		return nil, errors.New("type assertion to *ContainerSandbox failed")
	}

	mergedCommon, err := m.mergeCommonSandbox(spec, server)
	if err != nil {
		return nil, err
	}

	mergedCommonDeref := mergedCommon.(*types.CommonSandbox)
	mergedContainer := &types.ContainerSandbox{
		Available: mergedCommonDeref.Available,
		Dirs:      mergedCommonDeref.Dirs,
		Hooks:     mergedCommonDeref.Hooks,
		Image:     specContainer.Image,
		Registry:  specContainer.Registry,
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

func (m *Maker) mergeDockerSandbox(spec, server types.Sandbox) (types.Sandbox, error) {
	_, specOk := spec.(*types.DockerSandbox)
	_, serverOk := server.(*types.DockerSandbox)
	if !specOk || !serverOk {
		return nil, errors.New("type assertion to *DockerSandbox failed")
	}

	mergedContainer, err := m.mergeContainerSandbox(spec, server)
	if err != nil {
		return nil, err
	}

	mergedContainerDeref := mergedContainer.(*types.ContainerSandbox)
	mergedDocker := &types.DockerSandbox{
		Available: mergedContainerDeref.Available,
		Dirs:      mergedContainerDeref.Dirs,
		Hooks:     mergedContainerDeref.Hooks,
		Image:     mergedContainerDeref.Image,
		Registry:  mergedContainerDeref.Registry,
	}

	return mergedDocker, nil
}

func (m *Maker) mergeKubernetesSandbox(spec, server types.Sandbox) (types.Sandbox, error) {
	_, specOk := spec.(*types.KubernetesSandbox)
	_, serverOk := server.(*types.KubernetesSandbox)
	if !specOk || !serverOk {
		return nil, errors.New("type assertion to *KubernetesSandbox failed")
	}

	mergedContainer, err := m.mergeContainerSandbox(spec, server)
	if err != nil {
		return nil, err
	}

	mergedContainerDeref := mergedContainer.(*types.ContainerSandbox)
	mergedKubernetes := &types.KubernetesSandbox{
		Available: mergedContainerDeref.Available,
		Dirs:      mergedContainerDeref.Dirs,
		Hooks:     mergedContainerDeref.Hooks,
		Image:     mergedContainerDeref.Image,
		Registry:  mergedContainerDeref.Registry,
	}

	return mergedKubernetes, nil
}
