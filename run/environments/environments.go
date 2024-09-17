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

package environments

import (
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/environments/environment/providers/docker"
	"github.com/wstool/wst/run/environments/environment/providers/kubernetes"
	"github.com/wstool/wst/run/environments/environment/providers/local"
)

type Environments map[providers.Type]environment.Environment

type Maker interface {
	Make(
		specConfig,
		instanceConfig map[string]types.Environment,
		instanceWorkspace string,
	) (Environments, error)
}

type nativeMaker struct {
	fnd             app.Foundation
	localMaker      local.Maker
	dockerMaker     docker.Maker
	kubernetesMaker kubernetes.Maker
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{
		fnd:             fnd,
		localMaker:      local.CreateMaker(fnd),
		dockerMaker:     docker.CreateMaker(fnd),
		kubernetesMaker: kubernetes.CreateMaker(fnd),
	}
}

func (m *nativeMaker) Make(
	specConfig,
	instanceConfig map[string]types.Environment,
	instanceWorkspace string,
) (Environments, error) {
	var err error
	mergedEnvironments, err := m.mergeConfigMaps(specConfig, instanceConfig)
	if err != nil {
		return nil, err
	}

	commonEnv, commonFound := mergedEnvironments[types.CommonEnvironmentType]
	localEnv, localFound := mergedEnvironments[types.LocalEnvironmentType]
	containerEnv, containerFound := mergedEnvironments[types.ContainerEnvironmentType]
	dockerEnv, dockerFound := mergedEnvironments[types.DockerEnvironmentType]
	kubernetesEnv, kubernetesFound := mergedEnvironments[types.KubernetesEnvironmentType]
	if commonFound {
		// Local merging
		localEnv = m.mergeLocalAndCommon(localEnv, commonEnv)
		localFound = true
		// Container merging
		containerEnv = m.mergeContainerAndCommon(containerEnv, commonEnv)
		containerFound = true
	}
	if containerFound {
		// Docker merging
		dockerEnv = m.mergeDockerAndContainer(dockerEnv, containerEnv)
		dockerFound = true
		// Kubernetes merging
		kubernetesEnv = m.mergeKubernetesAndContainer(kubernetesEnv, containerEnv)
		kubernetesFound = true
	}

	envs := make(Environments)

	if localFound {
		envs[providers.LocalType], err = m.localMaker.Make(localEnv.(*types.LocalEnvironment), instanceWorkspace)
		if err != nil {
			return nil, err
		}
	}
	if dockerFound {
		envs[providers.DockerType], err = m.dockerMaker.Make(dockerEnv.(*types.DockerEnvironment))
		if err != nil {
			return nil, err
		}
	}
	if kubernetesFound {
		envs[providers.KubernetesType], err = m.kubernetesMaker.Make(kubernetesEnv.(*types.KubernetesEnvironment))
		if err != nil {
			return nil, err
		}
	}

	return envs, nil
}

func (m *nativeMaker) mergeLocalAndCommon(local, common types.Environment) types.Environment {
	commonEnvironment := common.(*types.CommonEnvironment)
	if local == nil {
		return &types.LocalEnvironment{
			Ports: commonEnvironment.Ports,
		}
	}
	localEnvironment := local.(*types.LocalEnvironment)
	localEnvironment.Ports = m.mergePorts(&commonEnvironment.Ports, &localEnvironment.Ports)

	return localEnvironment
}

func (m *nativeMaker) mergeContainerAndCommon(container, common types.Environment) types.Environment {
	commonEnvironment := common.(*types.CommonEnvironment)
	if container == nil {
		return &types.ContainerEnvironment{
			Ports: commonEnvironment.Ports,
		}
	}
	containerEnvironment := container.(*types.ContainerEnvironment)
	containerEnvironment.Ports = m.mergePorts(&commonEnvironment.Ports, &containerEnvironment.Ports)

	return containerEnvironment
}

func (m *nativeMaker) mergeDockerAndContainer(docker, container types.Environment) types.Environment {
	containerEnvironment := container.(*types.ContainerEnvironment)
	if docker == nil {
		return &types.DockerEnvironment{
			Ports:    containerEnvironment.Ports,
			Registry: containerEnvironment.Registry,
		}
	}

	dockerEnvironment := docker.(*types.DockerEnvironment)
	dockerEnvironment.Ports = m.mergePorts(&containerEnvironment.Ports, &dockerEnvironment.Ports)
	dockerEnvironment.Registry = m.mergeContainerRegistry(&containerEnvironment.Registry, &dockerEnvironment.Registry)

	return dockerEnvironment
}

func (m *nativeMaker) mergeKubernetesAndContainer(kubernetes, container types.Environment) types.Environment {
	containerEnvironment := container.(*types.ContainerEnvironment)
	if kubernetes == nil {
		return &types.KubernetesEnvironment{
			Ports:    containerEnvironment.Ports,
			Registry: containerEnvironment.Registry,
		}
	}

	kubernetesEnvironment := kubernetes.(*types.KubernetesEnvironment)
	kubernetesEnvironment.Ports = m.mergePorts(&containerEnvironment.Ports, &kubernetesEnvironment.Ports)
	kubernetesEnvironment.Registry = m.mergeContainerRegistry(
		&containerEnvironment.Registry,
		&kubernetesEnvironment.Registry,
	)

	return kubernetesEnvironment
}

type mergeFunc func(spec, instance types.Environment) (types.Environment, error)

func (m *nativeMaker) mergeConfigMaps(
	specEnvironments map[string]types.Environment,
	instanceEnvironments map[string]types.Environment,
) (map[types.EnvironmentType]types.Environment, error) {
	mergeFuncs := map[types.EnvironmentType]mergeFunc{
		types.CommonEnvironmentType:     m.mergeCommonEnvironment,
		types.LocalEnvironmentType:      m.mergeLocalEnvironment,
		types.ContainerEnvironmentType:  m.mergeContainerEnvironment,
		types.DockerEnvironmentType:     m.mergeDockerEnvironment,
		types.KubernetesEnvironmentType: m.mergeKubernetesEnvironment,
	}
	mergedEnvironments := make(map[types.EnvironmentType]types.Environment)

	for envType, merge := range mergeFuncs {
		envTypeStr := string(envType)
		specEnvironment, specExists := specEnvironments[envTypeStr]
		instanceEnvironment, instanceExists := instanceEnvironments[envTypeStr]

		if specExists && instanceExists {
			// Both environment defs exist, use the merge function
			mergedEnvironment, err := merge(specEnvironment, instanceEnvironment)
			if err != nil {
				return nil, err
			}
			mergedEnvironments[envType] = mergedEnvironment
		} else if !specExists && instanceExists {
			mergedEnvironments[envType] = instanceEnvironment
		} else if specExists {
			mergedEnvironments[envType] = specEnvironment
		}
	}

	return mergedEnvironments, nil
}

func (m *nativeMaker) mergePorts(specPorts, instancePorts *types.EnvironmentPorts) types.EnvironmentPorts {
	mergedPorts := *specPorts
	if instancePorts.Start > 0 {
		mergedPorts.Start = instancePorts.Start
	}
	if instancePorts.End > 0 {
		mergedPorts.End = instancePorts.End
	}
	return mergedPorts
}

func (m *nativeMaker) mergeCommonEnvironment(spec, instance types.Environment) (types.Environment, error) {
	// Ensure both spec and instance are of the correct type, using type assertion to *CommonEnvironment.
	specCommon, specOk := spec.(*types.CommonEnvironment)
	instanceCommon, instanceOk := instance.(*types.CommonEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("common environment is not set in common field")
	}

	// Create a new instance of CommonEnvironment for the merged result.
	mergedCommon := &types.CommonEnvironment{
		Ports: m.mergePorts(&specCommon.Ports, &instanceCommon.Ports),
	}

	// Return the new, merged CommonEnvironment as an Environment interface.
	return mergedCommon, nil
}

func (m *nativeMaker) mergeLocalEnvironment(spec, instance types.Environment) (types.Environment, error) {
	// Ensure both spec and instance are of the correct type, using type assertion to *CommonEnvironment.
	specLocal, specOk := spec.(*types.LocalEnvironment)
	instanceLocal, instanceOk := instance.(*types.LocalEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("local environment is not set in local field")
	}

	mergedLocal := &types.LocalEnvironment{
		Ports: m.mergePorts(&specLocal.Ports, &instanceLocal.Ports),
	}

	return mergedLocal, nil
}

func (m *nativeMaker) mergeContainerRegistry(specRegistry, instanceRegistry *types.ContainerRegistry) types.ContainerRegistry {
	mergedRegistry := *specRegistry
	if instanceRegistry.Auth.Username != "" {
		mergedRegistry.Auth.Username = instanceRegistry.Auth.Username
	}
	if instanceRegistry.Auth.Password != "" {
		mergedRegistry.Auth.Password = instanceRegistry.Auth.Password
	}
	return mergedRegistry
}

func (m *nativeMaker) mergeContainerEnvironment(spec, instance types.Environment) (types.Environment, error) {
	specContainer, specOk := spec.(*types.ContainerEnvironment)
	instanceContainer, instanceOk := instance.(*types.ContainerEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("container environment is not set in container field")
	}

	mergedContainer := &types.ContainerEnvironment{
		Ports:    m.mergePorts(&specContainer.Ports, &instanceContainer.Ports),
		Registry: m.mergeContainerRegistry(&specContainer.Registry, &instanceContainer.Registry),
	}

	return mergedContainer, nil
}

func (m *nativeMaker) mergeDockerEnvironment(spec, instance types.Environment) (types.Environment, error) {
	specDocker, specOk := spec.(*types.DockerEnvironment)
	instanceDocker, instanceOk := instance.(*types.DockerEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("docker environment is not set in docker field")
	}

	mergedDocker := &types.DockerEnvironment{
		Ports:      m.mergePorts(&specDocker.Ports, &instanceDocker.Ports),
		Registry:   m.mergeContainerRegistry(&specDocker.Registry, &instanceDocker.Registry),
		NamePrefix: specDocker.NamePrefix,
	}

	if instanceDocker.NamePrefix != "" {
		mergedDocker.NamePrefix = instanceDocker.NamePrefix
	}

	return mergedDocker, nil
}

func (m *nativeMaker) mergeKubernetesEnvironment(spec, instance types.Environment) (types.Environment, error) {
	specKubernetes, specOk := spec.(*types.KubernetesEnvironment)
	instanceKubernetes, instanceOk := instance.(*types.KubernetesEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("kubernetes environment is not set in kubernetes field")
	}

	mergedKubernetes := &types.KubernetesEnvironment{
		Ports:      m.mergePorts(&specKubernetes.Ports, &instanceKubernetes.Ports),
		Registry:   m.mergeContainerRegistry(&specKubernetes.Registry, &instanceKubernetes.Registry),
		Namespace:  specKubernetes.Namespace,
		Kubeconfig: specKubernetes.Kubeconfig,
	}

	if instanceKubernetes.Namespace != "" {
		mergedKubernetes.Namespace = instanceKubernetes.Namespace
	}
	if instanceKubernetes.Kubeconfig != "" {
		mergedKubernetes.Kubeconfig = instanceKubernetes.Kubeconfig
	}

	return mergedKubernetes, nil
}
