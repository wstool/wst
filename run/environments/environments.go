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

type Maker struct {
	fnd             app.Foundation
	localMaker      *local.Maker
	dockerMaker     *docker.Maker
	kubernetesMaker *kubernetes.Maker
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd:             fnd,
		localMaker:      local.CreateMaker(fnd),
		dockerMaker:     docker.CreateMaker(fnd),
		kubernetesMaker: kubernetes.CreateMaker(fnd),
	}
}

func (m *Maker) Make(
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
	dockerSb, dockerFound := mergedEnvironments[types.ContainerEnvironmentType]
	kubernetesSb, kubernetesFound := mergedEnvironments[types.ContainerEnvironmentType]
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

	Environments := make(Environments)

	if localFound {
		Environments[providers.LocalType], err = m.localMaker.Make(localSb.(*types.LocalEnvironment), instanceWorkspace)
		if err != nil {
			return nil, err
		}
	}
	if dockerFound {
		Environments[providers.DockerType], err = m.dockerMaker.Make(dockerSb.(*types.DockerEnvironment))
		if err != nil {
			return nil, err
		}
	}
	if kubernetesFound {
		Environments[providers.KubernetesType], err = m.kubernetesMaker.Make(dockerSb.(*types.KubernetesEnvironment))
		if err != nil {
			return nil, err
		}
	}

	return Environments, nil
}

func (m *Maker) mergeLocalAndCommon(local, common types.Environment) (types.Environment, error) {
	localEnvironment, localEnvironmentOk := local.(*types.LocalEnvironment)
	if !localEnvironmentOk {
		return nil, errors.New("type assertion to *LocalEnvironment failed")
	}
	mergedCommon, err := m.mergeCommonEnvironment(&localEnvironment.CommonEnvironment, common)
	if err != nil {
		return nil, err
	}
	localEnvironment.CommonEnvironment = *mergedCommon.(*types.CommonEnvironment)

	return localEnvironment, nil
}

func (m *Maker) mergeContainerAndCommon(container, common types.Environment) (types.Environment, error) {
	containerEnvironment, containerEnvironmentOk := container.(*types.ContainerEnvironment)
	if !containerEnvironmentOk {
		return nil, errors.New("type assertion to *ContainerEnvironment failed")
	}
	mergedCommon, err := m.mergeCommonEnvironment(&containerEnvironment.CommonEnvironment, common)
	if err != nil {
		return nil, err
	}
	containerEnvironment.CommonEnvironment = *mergedCommon.(*types.CommonEnvironment)

	return containerEnvironment, nil
}

func (m *Maker) mergeDockerAndContainer(docker, container types.Environment) (types.Environment, error) {
	dockerEnvironment, dockerEnvironmentOk := docker.(*types.DockerEnvironment)
	if !dockerEnvironmentOk {
		return nil, errors.New("type assertion to *DockerEnvironment failed")
	}
	mergedContainer, err := m.mergeContainerEnvironment(&dockerEnvironment.ContainerEnvironment, container)
	if err != nil {
		return nil, err
	}
	dockerEnvironment.ContainerEnvironment = *mergedContainer.(*types.ContainerEnvironment)

	return dockerEnvironment, nil
}

func (m *Maker) mergeKubernetesAndContainer(kubernetes, container types.Environment) (types.Environment, error) {
	kubernetesEnvironment, kubernetesEnvironmentOk := kubernetes.(*types.KubernetesEnvironment)
	if !kubernetesEnvironmentOk {
		return nil, errors.New("type assertion to *KubernetesEnvironment failed")
	}
	mergedContainer, err := m.mergeContainerEnvironment(&kubernetesEnvironment.ContainerEnvironment, container)
	if err != nil {
		return nil, err
	}
	kubernetesEnvironment.ContainerEnvironment = *mergedContainer.(*types.ContainerEnvironment)

	return kubernetesEnvironment, nil
}

type mergeFunc func(spec, instance types.Environment) (types.Environment, error)

func (m *Maker) mergeConfigMaps(
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

	for sandboxType, merge := range mergeFuncs {
		sandboxTypeStr := string(sandboxType)
		specEnvironment, specExists := specEnvironments[sandboxTypeStr]
		instanceEnvironment, instanceExists := instanceEnvironments[sandboxTypeStr]

		if specExists && instanceExists {
			// Use the merge function, now handling errors.
			mergedEnvironment, err := merge(specEnvironment, instanceEnvironment)
			if err != nil {
				// Handle the error, e.g., by returning it or logging it.
				return nil, err // Return an error if merging fails.
			}
			mergedEnvironments[sandboxType] = mergedEnvironment
		} else if !specExists && instanceExists {
			mergedEnvironments[sandboxType] = instanceEnvironment
		} else {
			mergedEnvironments[sandboxType] = specEnvironment
		}
	}

	return mergedEnvironments, nil
}

func (m *Maker) mergeCommonEnvironment(spec, instance types.Environment) (types.Environment, error) {
	// Ensure both spec and instance are of the correct type, using type assertion to *CommonEnvironment.
	specCommon, specOk := spec.(*types.CommonEnvironment)
	instanceCommon, instanceOk := instance.(*types.CommonEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("type assertion to *CommonEnvironment failed")
	}

	// Create a new instance of CommonEnvironment for the merged result.
	mergedCommon := &types.CommonEnvironment{
		Ports: specCommon.Ports,
	}

	if instanceCommon.Ports.Start > 0 {
		mergedCommon.Ports.Start = instanceCommon.Ports.Start
	}
	if instanceCommon.Ports.End > 0 {
		mergedCommon.Ports.End = instanceCommon.Ports.End
	}

	// Return the new, merged CommonEnvironment as an Environment interface.
	return mergedCommon, nil
}

func (m *Maker) mergeLocalEnvironment(spec, instance types.Environment) (types.Environment, error) {
	// Ensure both spec and instance are of the correct type, using type assertion to *CommonEnvironment.
	_, specOk := spec.(*types.LocalEnvironment)
	_, instanceOk := instance.(*types.LocalEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("type assertion to *LocalEnvironment failed")
	}

	mergedCommon, err := m.mergeCommonEnvironment(spec, instance)
	if err != nil {
		return nil, err
	}

	mergedLocal := &types.LocalEnvironment{
		CommonEnvironment: *mergedCommon.(*types.CommonEnvironment),
	}

	return mergedLocal, nil
}

func (m *Maker) mergeContainerEnvironment(spec, instance types.Environment) (types.Environment, error) {
	specContainer, specOk := spec.(*types.ContainerEnvironment)
	instanceContainer, instanceOk := instance.(*types.ContainerEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("type assertion to *ContainerEnvironment failed")
	}

	mergedCommon, err := m.mergeCommonEnvironment(spec, instance)
	if err != nil {
		return nil, err
	}

	mergedContainer := &types.ContainerEnvironment{
		CommonEnvironment: *mergedCommon.(*types.CommonEnvironment),
		Registry:          specContainer.Registry,
	}

	if instanceContainer.Registry.Auth.Username != "" {
		mergedContainer.Registry.Auth.Username = instanceContainer.Registry.Auth.Username
	}
	if instanceContainer.Registry.Auth.Password != "" {
		mergedContainer.Registry.Auth.Password = instanceContainer.Registry.Auth.Password
	}

	return mergedContainer, nil
}

func (m *Maker) mergeDockerEnvironment(spec, instance types.Environment) (types.Environment, error) {
	specDocker, specOk := spec.(*types.DockerEnvironment)
	instanceDocker, instanceOk := instance.(*types.DockerEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("type assertion to *DockerEnvironment failed")
	}

	mergedContainer, err := m.mergeContainerEnvironment(spec, instance)
	if err != nil {
		return nil, err
	}

	mergedDocker := &types.DockerEnvironment{
		ContainerEnvironment: *mergedContainer.(*types.ContainerEnvironment),
		NamePrefix:           specDocker.NamePrefix,
	}

	if instanceDocker.NamePrefix != "" {
		mergedDocker.NamePrefix = instanceDocker.NamePrefix
	}

	return mergedDocker, nil
}

func (m *Maker) mergeKubernetesEnvironment(spec, instance types.Environment) (types.Environment, error) {
	specKubernetes, specOk := spec.(*types.KubernetesEnvironment)
	instanceKubernetes, instanceOk := instance.(*types.KubernetesEnvironment)
	if !specOk || !instanceOk {
		return nil, errors.New("type assertion to *KubernetesEnvironment failed")
	}

	mergedContainer, err := m.mergeContainerEnvironment(spec, instance)
	if err != nil {
		return nil, err
	}

	mergedKubernetes := &types.KubernetesEnvironment{
		ContainerEnvironment: *mergedContainer.(*types.ContainerEnvironment),
		Namespace:            specKubernetes.Namespace,
		Kubeconfig:           specKubernetes.Kubeconfig,
	}

	if instanceKubernetes.Namespace != "" {
		mergedKubernetes.Namespace = instanceKubernetes.Namespace
	}
	if instanceKubernetes.Kubeconfig != "" {
		mergedKubernetes.Kubeconfig = instanceKubernetes.Kubeconfig
	}

	return mergedKubernetes, nil
}
