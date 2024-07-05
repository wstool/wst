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
	"errors"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/environments/environment/providers/docker"
	"github.com/bukka/wst/run/environments/environment/providers/kubernetes"
	"github.com/bukka/wst/run/environments/environment/providers/local"
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

	commonSb, commonFound := mergedEnvironments[types.CommonEnvironmentType]
	localSb, localFound := mergedEnvironments[types.LocalEnvironmentType]
	containerSb, containerFound := mergedEnvironments[types.ContainerEnvironmentType]
	dockerSb, dockerFound := mergedEnvironments[types.DockerEnvironmentType]
	kubernetesSb, kubernetesFound := mergedEnvironments[types.KubernetesEnvironmentType]
	if commonFound {
		// Local merging
		localSb, err = m.mergeLocalAndCommon(localSb, commonSb)
		if err != nil {
			return nil, err
		}
		localFound = true
		// Container merging
		containerSb, err = m.mergeContainerAndCommon(containerSb, commonSb)
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

	envs := make(Environments)

	if localFound {
		envs[providers.LocalType], err = m.localMaker.Make(localSb.(*types.LocalEnvironment), instanceWorkspace)
		if err != nil {
			return nil, err
		}
	}
	if dockerFound {
		envs[providers.DockerType], err = m.dockerMaker.Make(dockerSb.(*types.DockerEnvironment))
		if err != nil {
			return nil, err
		}
	}
	if kubernetesFound {
		envs[providers.KubernetesType], err = m.kubernetesMaker.Make(dockerSb.(*types.KubernetesEnvironment))
		if err != nil {
			return nil, err
		}
	}

	return envs, nil
}

func (m *nativeMaker) mergeLocalAndCommon(local, common types.Environment) (types.Environment, error) {
	if local == nil {
		return &types.LocalEnvironment{
			Ports: common.(*types.CommonEnvironment).Ports,
		}, nil
	}
	localEnvironment, localEnvironmentOk := local.(*types.LocalEnvironment)
	if !localEnvironmentOk {
		return nil, errors.New("type assertion to *LocalEnvironment failed")
	}
	mergedCommon, err := m.mergeCommonEnvironment(&types.CommonEnvironment{Ports: localEnvironment.Ports}, common)
	if err != nil {
		return nil, err
	}
	localEnvironment.Ports = mergedCommon.(*types.CommonEnvironment).Ports

	return localEnvironment, nil
}

func (m *nativeMaker) mergeContainerAndCommon(container, common types.Environment) (types.Environment, error) {
	if container == nil {
		return &types.CommonEnvironment{
			Ports: common.(*types.CommonEnvironment).Ports,
		}, nil
	}
	containerEnvironment, containerEnvironmentOk := container.(*types.ContainerEnvironment)
	if !containerEnvironmentOk {
		return nil, errors.New("type assertion to *ContainerEnvironment failed")
	}
	mergedCommon, err := m.mergeCommonEnvironment(&types.CommonEnvironment{Ports: containerEnvironment.Ports}, common)
	if err != nil {
		return nil, err
	}
	containerEnvironment.Ports = mergedCommon.(*types.CommonEnvironment).Ports

	return containerEnvironment, nil
}

func (m *nativeMaker) mergeDockerAndContainer(docker, container types.Environment) (types.Environment, error) {
	if docker == nil {
		containerEnv := container.(*types.ContainerEnvironment)
		return &types.DockerEnvironment{
			Ports:    containerEnv.Ports,
			Registry: containerEnv.Registry,
		}, nil
	}

	dockerEnvironment, dockerEnvironmentOk := docker.(*types.DockerEnvironment)
	if !dockerEnvironmentOk {
		return nil, errors.New("type assertion to *DockerEnvironment failed")
	}
	mergedContainer, err := m.mergeContainerEnvironment(&types.ContainerEnvironment{
		Ports:    dockerEnvironment.Ports,
		Registry: dockerEnvironment.Registry,
	}, container)
	if err != nil {
		return nil, err
	}
	mergedContainerRef := mergedContainer.(*types.ContainerEnvironment)
	dockerEnvironment.Ports = mergedContainerRef.Ports
	dockerEnvironment.Registry = mergedContainerRef.Registry

	return dockerEnvironment, nil
}

func (m *nativeMaker) mergeKubernetesAndContainer(kubernetes, container types.Environment) (types.Environment, error) {
	if kubernetes == nil {
		containerEnv := container.(*types.ContainerEnvironment)
		return &types.KubernetesEnvironment{
			Ports:    containerEnv.Ports,
			Registry: containerEnv.Registry,
		}, nil
	}

	kubernetesEnvironment, kubernetesEnvironmentOk := kubernetes.(*types.KubernetesEnvironment)
	if !kubernetesEnvironmentOk {
		return nil, errors.New("type assertion to *KubernetesEnvironment failed")
	}
	mergedContainer, err := m.mergeContainerEnvironment(&types.ContainerEnvironment{
		Ports:    kubernetesEnvironment.Ports,
		Registry: kubernetesEnvironment.Registry,
	}, container)
	if err != nil {
		return nil, err
	}
	mergedContainerRef := mergedContainer.(*types.ContainerEnvironment)
	kubernetesEnvironment.Ports = mergedContainerRef.Ports
	kubernetesEnvironment.Registry = mergedContainerRef.Registry

	return kubernetesEnvironment, nil
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
			// Use the merge function, now handling errors.
			mergedEnvironment, err := merge(specEnvironment, instanceEnvironment)
			if err != nil {
				// Handle the error, e.g., by returning it or logging it.
				return nil, err // Return an error if merging fails.
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
		return nil, errors.New("type assertion to *CommonEnvironment failed")
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
		return nil, errors.New("type assertion to *LocalEnvironment failed")
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
		return nil, errors.New("type assertion to *ContainerEnvironment failed")
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
		return nil, errors.New("type assertion to *DockerEnvironment failed")
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
		return nil, errors.New("type assertion to *KubernetesEnvironment failed")
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
