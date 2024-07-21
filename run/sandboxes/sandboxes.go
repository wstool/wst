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

type Maker interface {
	MakeSandboxes(
		rootSandboxes map[string]types.Sandbox,
		serverSandboxes map[string]types.Sandbox,
	) (Sandboxes, error)
}

type nativeMaker struct {
	fnd             app.Foundation
	localMaker      local.Maker
	dockerMaker     docker.Maker
	kubernetesMaker kubernetes.Maker
}

func CreateMaker(fnd app.Foundation) Maker {
	hooksMaker := hooks.CreateMaker(fnd)
	commonMaker := common.CreateMaker(fnd, hooksMaker)
	containerMaker := container.CreateMaker(fnd, commonMaker)
	return &nativeMaker{
		fnd:             fnd,
		localMaker:      local.CreateMaker(fnd, commonMaker),
		dockerMaker:     docker.CreateMaker(fnd, containerMaker),
		kubernetesMaker: kubernetes.CreateMaker(fnd, containerMaker),
	}
}

func (m *nativeMaker) MakeSandboxes(
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
	dockerSb, dockerFound := mergedSandboxes[types.DockerSandboxType]
	kubernetesSb, kubernetesFound := mergedSandboxes[types.KubernetesSandboxType]
	if commonFound {
		// Local merging
		localSb = m.mergeLocalAndCommon(localSb, commonSb)
		localFound = true
		// Container merging
		containerSb = m.mergeContainerAndCommon(containerSb, commonSb)
		containerFound = true
	}
	if containerFound {
		// Docker merging
		dockerSb = m.mergeDockerAndContainer(dockerSb, containerSb)
		dockerFound = true
		// Kubernetes merging
		kubernetesSb = m.mergeKubernetesAndContainer(kubernetesSb, containerSb)
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
		sandboxes[providers.KubernetesType], err = m.kubernetesMaker.MakeSandbox(kubernetesSb.(*types.KubernetesSandbox))
		if err != nil {
			return nil, err
		}
	}

	return sandboxes, nil
}

func (m *nativeMaker) mergeLocalAndCommon(local, common types.Sandbox) types.Sandbox {
	commonSandbox := common.(*types.CommonSandbox)
	if local == nil {
		return &types.LocalSandbox{
			Available: commonSandbox.Available,
			Dirs:      commonSandbox.Dirs,
			Hooks:     commonSandbox.Hooks,
		}
	}
	localSandbox := local.(*types.LocalSandbox)
	localSandbox.Dirs = m.mergeDirs(commonSandbox.Dirs, localSandbox.Dirs)
	localSandbox.Hooks = m.mergeHooks(commonSandbox.Hooks, localSandbox.Hooks)

	return localSandbox
}

func (m *nativeMaker) mergeContainerAndCommon(container, common types.Sandbox) types.Sandbox {
	commonSandbox := common.(*types.CommonSandbox)
	if container == nil {
		return &types.ContainerSandbox{
			Available: commonSandbox.Available,
			Dirs:      commonSandbox.Dirs,
			Hooks:     commonSandbox.Hooks,
		}
	}
	containerSandbox := container.(*types.ContainerSandbox)
	containerSandbox.Dirs = m.mergeDirs(commonSandbox.Dirs, containerSandbox.Dirs)
	containerSandbox.Hooks = m.mergeHooks(commonSandbox.Hooks, containerSandbox.Hooks)

	return containerSandbox
}

func (m *nativeMaker) mergeDockerAndContainer(docker, container types.Sandbox) types.Sandbox {
	containerSandbox := container.(*types.ContainerSandbox)
	if docker == nil {
		return &types.DockerSandbox{
			Available: containerSandbox.Available,
			Dirs:      containerSandbox.Dirs,
			Hooks:     containerSandbox.Hooks,
			Image:     containerSandbox.Image,
			Registry:  containerSandbox.Registry,
		}
	}

	dockerSandbox := docker.(*types.DockerSandbox)
	dockerSandbox.Dirs = m.mergeDirs(containerSandbox.Dirs, dockerSandbox.Dirs)
	dockerSandbox.Hooks = m.mergeHooks(containerSandbox.Hooks, dockerSandbox.Hooks)
	dockerSandbox.Image = m.mergeContainerImage(containerSandbox.Image, dockerSandbox.Image)
	dockerSandbox.Registry = m.mergeContainerRegistry(containerSandbox.Registry, dockerSandbox.Registry)

	return dockerSandbox
}

func (m *nativeMaker) mergeKubernetesAndContainer(kubernetes, container types.Sandbox) types.Sandbox {
	containerSandbox := container.(*types.ContainerSandbox)
	if kubernetes == nil {
		return &types.KubernetesSandbox{
			Available: containerSandbox.Available,
			Dirs:      containerSandbox.Dirs,
			Hooks:     containerSandbox.Hooks,
			Image:     containerSandbox.Image,
			Registry:  containerSandbox.Registry,
		}
	}

	kubernetesSandbox := kubernetes.(*types.KubernetesSandbox)
	kubernetesSandbox.Dirs = m.mergeDirs(containerSandbox.Dirs, kubernetesSandbox.Dirs)
	kubernetesSandbox.Hooks = m.mergeHooks(containerSandbox.Hooks, kubernetesSandbox.Hooks)
	kubernetesSandbox.Image = m.mergeContainerImage(containerSandbox.Image, kubernetesSandbox.Image)
	kubernetesSandbox.Registry = m.mergeContainerRegistry(containerSandbox.Registry, kubernetesSandbox.Registry)

	return kubernetesSandbox
}

type mergeFunc func(root, server types.Sandbox) (types.Sandbox, error)

func (m *nativeMaker) mergeConfigMaps(
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
		} else if rootExists {
			mergedSandboxes[sandboxType] = rootSandbox
		}
	}

	return mergedSandboxes, nil
}

func (m *nativeMaker) mergeDirs(firstDirs, secondDirs map[string]string) map[string]string {
	mergedDirs := make(map[string]string)
	for k, v := range firstDirs {
		mergedDirs[k] = v
	}
	for k, v := range secondDirs {
		mergedDirs[k] = v
	}
	return mergedDirs
}

func (m *nativeMaker) mergeHooks(firstHooks, secondHooks map[string]types.SandboxHook) map[string]types.SandboxHook {
	mergedHooks := make(map[string]types.SandboxHook)
	for k, v := range firstHooks {
		mergedHooks[k] = v
	}
	for k, v := range secondHooks {
		mergedHooks[k] = v
	}
	return mergedHooks
}

func (m *nativeMaker) mergeCommonSandbox(spec, server types.Sandbox) (types.Sandbox, error) {
	// Ensure both root and server are of the correct type, using type assertion to *CommonSandbox.
	specCommon, rootOk := spec.(*types.CommonSandbox)
	serverCommon, serverOk := server.(*types.CommonSandbox)
	if !rootOk || !serverOk {
		return nil, errors.New("common sandbox is not set in common field")
	}

	// Create a new instance of CommonSandbox for the merged result.
	mergedCommon := &types.CommonSandbox{
		Available: serverCommon.Available, // Available is always set from the server
		Dirs:      m.mergeDirs(specCommon.Dirs, serverCommon.Dirs),
		Hooks:     m.mergeHooks(specCommon.Hooks, serverCommon.Hooks),
	}

	// Return the new, merged CommonSandbox as a Sandbox interface.
	return mergedCommon, nil
}

func (m *nativeMaker) mergeLocalSandbox(spec, server types.Sandbox) (types.Sandbox, error) {
	// Ensure both spec and server are of the correct type, using type assertion to *CommonSandbox.
	specLocal, specOk := spec.(*types.LocalSandbox)
	serverLocal, serverOk := server.(*types.LocalSandbox)
	if !specOk || !serverOk {
		return nil, errors.New("local sandbox is not set in local field")
	}

	mergedLocal := &types.LocalSandbox{
		Available: serverLocal.Available, // Available is always set from the server
		Dirs:      m.mergeDirs(specLocal.Dirs, serverLocal.Dirs),
		Hooks:     m.mergeHooks(specLocal.Hooks, serverLocal.Hooks),
	}

	return mergedLocal, nil
}

func (m *nativeMaker) mergeContainerImage(firstImage, secondImage types.ContainerImage) types.ContainerImage {
	if secondImage.Name != "" {
		firstImage.Name = secondImage.Name
	}
	if secondImage.Tag != "" {
		firstImage.Tag = secondImage.Tag
	}
	return firstImage
}

func (m *nativeMaker) mergeContainerRegistry(
	firstRegistry,
	secondRegistry types.ContainerRegistry,
) types.ContainerRegistry {
	if secondRegistry.Auth.Username != "" {
		firstRegistry.Auth.Username = secondRegistry.Auth.Username
	}
	if secondRegistry.Auth.Password != "" {
		firstRegistry.Auth.Password = secondRegistry.Auth.Password
	}
	return firstRegistry
}

func (m *nativeMaker) mergeContainerSandbox(spec, server types.Sandbox) (types.Sandbox, error) {
	specContainer, specOk := spec.(*types.ContainerSandbox)
	serverContainer, serverOk := server.(*types.ContainerSandbox)
	if !specOk || !serverOk {
		return nil, errors.New("container sandbox is not set in container field")
	}

	mergedContainer := &types.ContainerSandbox{
		Available: serverContainer.Available,
		Dirs:      m.mergeDirs(specContainer.Dirs, serverContainer.Dirs),
		Hooks:     m.mergeHooks(specContainer.Hooks, serverContainer.Hooks),
		Image:     m.mergeContainerImage(specContainer.Image, serverContainer.Image),
		Registry:  m.mergeContainerRegistry(specContainer.Registry, serverContainer.Registry),
	}

	return mergedContainer, nil
}

func (m *nativeMaker) mergeDockerSandbox(spec, server types.Sandbox) (types.Sandbox, error) {
	specDocker, specOk := spec.(*types.DockerSandbox)
	serverDocker, serverOk := server.(*types.DockerSandbox)
	if !specOk || !serverOk {
		return nil, errors.New("docker sandbox is not set in docker field")
	}

	mergedDocker := &types.DockerSandbox{
		Available: serverDocker.Available,
		Dirs:      m.mergeDirs(specDocker.Dirs, serverDocker.Dirs),
		Hooks:     m.mergeHooks(specDocker.Hooks, serverDocker.Hooks),
		Image:     m.mergeContainerImage(specDocker.Image, serverDocker.Image),
		Registry:  m.mergeContainerRegistry(specDocker.Registry, serverDocker.Registry),
	}

	return mergedDocker, nil
}

func (m *nativeMaker) mergeKubernetesSandbox(spec, server types.Sandbox) (types.Sandbox, error) {
	specKubernetes, specOk := spec.(*types.KubernetesSandbox)
	serverKubernetes, serverOk := server.(*types.KubernetesSandbox)
	if !specOk || !serverOk {
		return nil, errors.New("kubernetes sandbox is not set in kubernetes field")
	}

	mergedKubernetes := &types.KubernetesSandbox{
		Available: serverKubernetes.Available,
		Dirs:      m.mergeDirs(specKubernetes.Dirs, serverKubernetes.Dirs),
		Hooks:     m.mergeHooks(specKubernetes.Hooks, serverKubernetes.Hooks),
		Image:     m.mergeContainerImage(specKubernetes.Image, serverKubernetes.Image),
		Registry:  m.mergeContainerRegistry(specKubernetes.Registry, serverKubernetes.Registry),
	}

	return mergedKubernetes, nil
}
