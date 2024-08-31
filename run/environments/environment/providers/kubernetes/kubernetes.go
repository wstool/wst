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

package kubernetes

import (
	"context"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/environments/environment/providers/kubernetes/clients"
	"github.com/bukka/wst/run/environments/task"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type Maker interface {
	Make(config *types.KubernetesEnvironment) (environment.Environment, error)
}

type kubernetesMaker struct {
	*environment.CommonMaker
	clientsMaker clients.Maker
}

func CreateMaker(fnd app.Foundation) Maker {
	return &kubernetesMaker{
		CommonMaker:  environment.CreateCommonMaker(fnd),
		clientsMaker: clients.CreateMaker(fnd),
	}
}

func (m *kubernetesMaker) Make(config *types.KubernetesEnvironment) (environment.Environment, error) {
	configMapClient, err := m.clientsMaker.MakeConfigMapClient(config)
	if err != nil {
		return nil, errors.Errorf("failed to create kubernetes client: %v", err)
	}
	deploymentClient, err := m.clientsMaker.MakeDeploymentClient(config)
	if err != nil {
		return nil, errors.Errorf("failed to create kubernetes client: %v", err)
	}
	podClient, err := m.clientsMaker.MakePodClient(config)
	if err != nil {
		return nil, errors.Errorf("failed to create kubernetes client: %v", err)
	}
	serviceClient, err := m.clientsMaker.MakeServiceClient(config)
	if err != nil {
		return nil, errors.Errorf("failed to create kubernetes client: %v", err)
	}

	return &kubernetesEnvironment{
		ContainerEnvironment: *m.MakeContainerEnvironment(&types.ContainerEnvironment{
			Ports:    config.Ports,
			Registry: config.Registry,
		}),
		kubeconfigPath:   config.Kubeconfig,
		namespace:        config.Namespace,
		useFullName:      false,
		configMapClient:  configMapClient,
		deploymentClient: deploymentClient,
		podClient:        podClient,
		serviceClient:    serviceClient,
		tasks:            make(map[string]*kubernetesTask),
	}, nil
}

type kubernetesEnvironment struct {
	environment.ContainerEnvironment
	kubeconfigPath   string
	namespace        string
	useFullName      bool
	deploymentClient clients.DeploymentClient
	configMapClient  clients.ConfigMapClient
	podClient        clients.PodClient
	serviceClient    clients.ServiceClient
	tasks            map[string]*kubernetesTask
}

func (e *kubernetesEnvironment) dryRunOption() []string {
	if e.Fnd.DryRun() {
		return []string{metav1.DryRunAll}
	}
	return nil
}

func (e *kubernetesEnvironment) Init(ctx context.Context) error {
	return nil
}

func (e *kubernetesEnvironment) destroyConfigMaps(
	ctx context.Context,
	configMaps []*corev1.ConfigMap,
	opts metav1.DeleteOptions,
) error {
	hasError := false
	for _, configMap := range configMaps {
		if err := e.configMapClient.Delete(ctx, configMap.Name, opts); err != nil {
			e.Fnd.Logger().Errorf("Failed to remove config map %s: %v", configMap.Name, err)
			hasError = true
		}
	}
	if hasError {
		return errors.Errorf("failed to delete config maps")
	}
	return nil
}

func (e *kubernetesEnvironment) destroyDeployment(
	ctx context.Context,
	deployment *appsv1.Deployment,
	opts metav1.DeleteOptions,
) error {
	if deployment != nil {
		err := e.deploymentClient.Delete(ctx, deployment.Name, opts)
		if err != nil {
			e.Fnd.Logger().Errorf("Failed to delete deployment %s: %v", deployment.Name, err)
			return errors.Errorf("failed to delete deployment")
		}
	}
	return nil
}

func (e *kubernetesEnvironment) destroyService(
	ctx context.Context,
	service *corev1.Service,
	opts metav1.DeleteOptions,
) error {
	if service != nil {
		err := e.serviceClient.Delete(ctx, service.Name, opts)
		if err != nil {
			e.Fnd.Logger().Errorf("Failed to delete service %s: %v", service.Name, err)
			return errors.Errorf("failed to delete service")
		}
	}
	return nil
}

func (e *kubernetesEnvironment) destroyTask(
	ctx context.Context,
	kubeTask *kubernetesTask,
	deleteOptions metav1.DeleteOptions,
) error {
	hasError := false
	var err error
	// Close output
	if kubeTask.outputReader != nil && kubeTask.outputReader.Close() != nil {
		hasError = true
	}

	// Delete the service
	if err = e.destroyService(ctx, kubeTask.service, deleteOptions); err != nil {
		hasError = true
	}

	// Delete the deployment
	if err = e.destroyDeployment(ctx, kubeTask.deployment, deleteOptions); err != nil {
		hasError = true
	}

	// Delete config maps
	if err = e.destroyConfigMaps(ctx, kubeTask.configMaps, deleteOptions); err != nil {
		hasError = true
	}

	if hasError {
		return errors.Errorf("failed to delete task")
	}

	return nil
}

func (e *kubernetesEnvironment) Destroy(ctx context.Context) error {
	var err error

	deleteOptions := metav1.DeleteOptions{DryRun: e.dryRunOption()}
	hasError := false
	// Iterate over all tasks to delete services and deployments
	for _, kubeTask := range e.tasks {
		// Delete the tasks
		if err = e.destroyTask(ctx, kubeTask, deleteOptions); err != nil {
			hasError = true
		}
	}

	// Clear the tasks map for potential reuse of the environment
	e.tasks = make(map[string]*kubernetesTask)

	if hasError {
		return errors.Errorf("failed to destroy kubernetes environment")
	}
	return nil
}

func int32Ptr(i int32) *int32 {
	return &i
}

func (e *kubernetesEnvironment) serviceName(ss *environment.ServiceSettings) string {
	if e.useFullName {
		return ss.FullName
	} else {
		return ss.Name
	}
}

// sanitizeName prepares a string to be used as a Kubernetes resource name
func sanitizeName(input string) string {
	// Replace invalid characters with '-'
	reg := regexp.MustCompile("[^a-zA-Z0-9-]+")
	sanitized := reg.ReplaceAllString(input, "-")

	// Trim to maximum length for ConfigMap names
	maxLength := 253
	if len(sanitized) > maxLength {
		sanitized = sanitized[:maxLength]
	}
	return strings.ToLower(sanitized)
}

// createConfigMap creates a single ConfigMap from the provided data
func (e *kubernetesEnvironment) createConfigMap(ctx context.Context, configMapName string, data map[string]string) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapName,
		},
		Data: data,
	}

	return e.configMapClient.Create(ctx, configMap, metav1.CreateOptions{DryRun: e.dryRunOption()})
}

// loadFileContent reads the content of the file at the given path.
func (e *kubernetesEnvironment) loadFileContent(filePath string) (string, error) {
	content, err := afero.ReadFile(e.Fnd.Fs(), filePath)
	if err != nil {
		return "", errors.Errorf("failed to read file at %s: %v", filePath, err)
	}
	return string(content), nil
}

// processWorkspacePaths creates ConfigMaps from files specified in workspacePaths and updates volumes and volumeMounts slices.
// workspacePaths maps a logical name to a file path on the host.
// envPaths maps the same logical name to a mount path within the container.
// The function assumes volumeMounts and volumes are pre-initialized and passed by reference.
func (e *kubernetesEnvironment) processWorkspacePaths(
	ctx context.Context,
	serviceName string,
	workspacePaths,
	envPaths map[string]string,
	volumeMounts *[]corev1.VolumeMount,
	volumes *[]corev1.Volume,
) ([]*corev1.ConfigMap, error) {
	configMaps := make([]*corev1.ConfigMap, 0, len(workspacePaths))
	for name, hostPath := range workspacePaths {
		envPath, found := envPaths[name]
		if !found {
			return configMaps, errors.Errorf("environment path not found for %s", name)
		}

		// Load the content of the file at hostPath
		content, err := e.loadFileContent(hostPath)
		if err != nil {
			return configMaps, err
		}

		// Create a ConfigMap for the file content
		configMapName := sanitizeName(fmt.Sprintf("%s-%s", serviceName, name))
		baseHostPath := filepath.Base(hostPath)
		data := map[string]string{
			baseHostPath: content, // Use the file name as the key
		}

		configMap, err := e.createConfigMap(ctx, configMapName, data)
		if err != nil {
			return configMaps, errors.Errorf("failed to create configMap %s: %v", configMapName, err)
		}
		configMaps = append(configMaps, configMap)

		// Prepare volume and volume mount for this ConfigMap
		volumeName := configMapName + "-volume"
		volume := corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		}
		baseEnvPath := filepath.Base(envPath)
		if baseEnvPath != baseHostPath {
			volume.VolumeSource.ConfigMap.Items = []corev1.KeyToPath{
				{
					Key:  baseHostPath,
					Path: baseEnvPath,
				},
			}
		}

		*volumes = append(*volumes, volume)

		*volumeMounts = append(*volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: filepath.Dir(envPath),
		})
	}

	return configMaps, nil
}

func (e *kubernetesEnvironment) createDeployment(
	ctx context.Context,
	serviceName string,
	ss *environment.ServiceSettings,
	cmd *environment.Command,
) (*kubernetesTask, error) {
	containerConfig := ss.ContainerConfig
	if containerConfig == nil {
		return nil, errors.New("container config is not set")
	}

	var volumeMounts []corev1.VolumeMount
	var volumes []corev1.Volume
	deleteOptions := metav1.DeleteOptions{DryRun: e.dryRunOption()}

	configConfigMaps, err := e.processWorkspacePaths(
		ctx,
		serviceName,
		ss.WorkspaceConfigPaths,
		ss.EnvironmentConfigPaths,
		&volumeMounts,
		&volumes,
	)
	if err != nil {
		return nil, err
	}
	scriptConfigMaps, err := e.processWorkspacePaths(
		ctx,
		serviceName,
		ss.WorkspaceScriptPaths,
		ss.EnvironmentScriptPaths,
		&volumeMounts,
		&volumes,
	)
	if err != nil {
		_ = e.destroyConfigMaps(ctx, configConfigMaps, deleteOptions)
		return nil, err
	}
	configMaps := append(configConfigMaps, scriptConfigMaps...)

	var command []string
	var args []string
	if cmd.Name != "" {
		command = append(command, cmd.Name)
		args = cmd.Args
	}

	// Define the deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2), // Specify the number of replicas
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": serviceName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": serviceName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    ss.Name,
							Image:   containerConfig.Image(),
							Command: command,
							Args:    args,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: ss.ServerPort,
								},
							},
							VolumeMounts: volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	result, err := e.deploymentClient.Create(ctx, deployment, metav1.CreateOptions{DryRun: e.dryRunOption()})
	if err != nil {
		_ = e.destroyConfigMaps(ctx, configMaps, deleteOptions)
		return nil, err
	}
	kubeTask := &kubernetesTask{
		serviceName: serviceName,
		configMaps:  configMaps,
		deployment:  result,
		executable:  cmd.Name,
	}
	e.tasks[serviceName] = kubeTask

	if e.Fnd.DryRun() {
		kubeTask.deploymentReady = true
		return kubeTask, nil
	}

	watcher, err := e.deploymentClient.Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", serviceName),
	})
	if err != nil {
		return nil, err
	}
	defer watcher.Stop()

	// Listen for events and handle context cancellation or timeout
	for {
		select {
		case event := <-watcher.ResultChan():
			switch event.Type {
			case watch.Added, watch.Modified:
				deployment, ok := event.Object.(*appsv1.Deployment)
				if !ok {
					return nil, errors.Errorf("expected Deployment object, but got something else")
				}
				if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
					kubeTask.deploymentReady = true
					return kubeTask, nil
				}
			case watch.Deleted:
				fallthrough
			case watch.Error:
				return nil, errors.New("watching deployment did not result to addition and modification")
			}
		case <-ctx.Done():
			return nil, errors.Errorf("context canceled or timed out when waiting on deployment to be ready")
		}
	}
}

func (e *kubernetesEnvironment) createService(
	ctx context.Context,
	kubeTask *kubernetesTask,
	serviceName string,
	ss *environment.ServiceSettings,
) error {
	var kubeServiceType corev1.ServiceType
	if ss.Public {
		kubeServiceType = corev1.ServiceTypeLoadBalancer
	} else {
		kubeServiceType = corev1.ServiceTypeClusterIP
	}
	kubeServiceSpec := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Type: kubeServiceType,
			Ports: []corev1.ServicePort{
				{
					Port:       ss.ServerPort,
					TargetPort: intstr.FromInt32(ss.ServerPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": serviceName,
			},
		},
	}

	kubeService, err := e.serviceClient.Create(ctx, kubeServiceSpec, metav1.CreateOptions{DryRun: e.dryRunOption()})
	if err != nil {
		return errors.Errorf("failed to create service: %v", err)
	}

	kubeTask.service = kubeService

	if e.Fnd.DryRun() {
		if ss.Public {
			kubeTask.servicePublicUrl = "http://127.0.0.1"
		}
		kubeTask.servicePrivateUrl = fmt.Sprintf("http://%s:%d", serviceName, ss.ServerPort)
		return nil
	}

	watcher, err := e.serviceClient.Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", serviceName),
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	// Listen for events and handle context cancellation or timeout
	for {
		select {
		case event := <-watcher.ResultChan():
			switch event.Type {
			case watch.Added, watch.Modified:
				svc, ok := event.Object.(*corev1.Service)
				if !ok {
					return errors.Errorf("expected Service object, but got something else")
				}
				if ss.Public && len(svc.Status.LoadBalancer.Ingress) > 0 {
					ip := svc.Status.LoadBalancer.Ingress[0].IP
					kubeTask.servicePublicUrl = fmt.Sprintf("http://%s", ip)
				}
				kubeTask.servicePrivateUrl = fmt.Sprintf("http://%s:%d", serviceName, ss.ServerPort)
				return nil
			case watch.Deleted:
				fallthrough
			case watch.Error:
				return errors.Errorf("watching service did not result to addition and modification")
			}
		case <-ctx.Done():
			return errors.Errorf("context canceled or timed out when waiting on service IP")
		}
	}
}

func (e *kubernetesEnvironment) RunTask(ctx context.Context, ss *environment.ServiceSettings, cmd *environment.Command) (task.Task, error) {
	serviceName := e.serviceName(ss)
	kubeTask, err := e.createDeployment(ctx, serviceName, ss, cmd)
	if err != nil {
		return nil, err
	}
	err = e.createService(ctx, kubeTask, serviceName, ss)
	if err != nil {
		return nil, err
	}

	return kubeTask, nil
}

func (e *kubernetesEnvironment) ExecTaskCommand(ctx context.Context, ss *environment.ServiceSettings, target task.Task, cmd *environment.Command) error {
	return errors.Errorf("executing command is not currently supported in Kubernetes environment")
}

func (e *kubernetesEnvironment) ExecTaskSignal(ctx context.Context, ss *environment.ServiceSettings, target task.Task, signal os.Signal) error {
	return errors.Errorf("executing signal is not currently supported in Kubernetes environment")
}

func (e *kubernetesEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	if outputType != output.Any {
		return nil, errors.Errorf("only any output type is supported by Kubernetes environment")
	}
	kubeTask, ok := target.(*kubernetesTask)
	if !ok {
		return nil, errors.Errorf("task in not a Kubernetes task")
	}
	if kubeTask.outputReader != nil {
		return kubeTask.outputReader, nil
	}

	if e.Fnd.DryRun() {
		kubeTask.outputReader = &CombinedReader{readers: []io.ReadCloser{&app.DummyReaderCloser{}}}
		return kubeTask.outputReader, nil
	}

	pods, err := e.podClient.List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", kubeTask.Name()),
	})
	if err != nil {
		return nil, errors.Errorf("failed to list pods: %v", err)
	}

	readers := make([]io.ReadCloser, 0, len(pods.Items))
	combinedReader := &CombinedReader{readers: readers}
	for _, pod := range pods.Items {
		podLogs, err := e.podClient.StreamLogs(ctx, pod.Name, &corev1.PodLogOptions{})
		if err != nil {
			combinedReader.Close()
			return nil, errors.Errorf("error in opening stream: %v", err)
		}
		combinedReader.readers = append(combinedReader.readers, podLogs)
	}

	kubeTask.outputReader = combinedReader
	return combinedReader, nil
}

func (e *kubernetesEnvironment) RootPath(workspace string) string {
	return ""
}

func (e *kubernetesEnvironment) Mkdir(serviceName string, path string, perm os.FileMode) error {
	// Currently it is a user responsibility to make sure that directory exists in the container
	return nil
}

func (e *kubernetesEnvironment) ServiceAddress(serviceName string, port int32) string {
	return serviceName
}

type kubernetesTask struct {
	deployment        *appsv1.Deployment
	service           *corev1.Service
	configMaps        []*corev1.ConfigMap
	outputReader      *CombinedReader
	executable        string
	serviceName       string
	servicePublicUrl  string
	servicePrivateUrl string
	deploymentReady   bool
}

func (t *kubernetesTask) Pid() int {
	return 1
}

func (t *kubernetesTask) Id() string {
	return t.serviceName
}

func (t *kubernetesTask) Name() string {
	return t.serviceName
}

func (t *kubernetesTask) Executable() string {
	return t.executable
}

func (k *kubernetesTask) Type() providers.Type {
	return providers.KubernetesType
}

func (t *kubernetesTask) PublicUrl() string {
	return t.servicePublicUrl
}

func (t *kubernetesTask) PrivateUrl() string {
	return t.servicePrivateUrl
}
