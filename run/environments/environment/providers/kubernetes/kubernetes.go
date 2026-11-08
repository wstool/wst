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
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/environments/environment/providers/kubernetes/clients"
	"github.com/wstool/wst/run/environments/task"
	"github.com/wstool/wst/run/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
)

type Maker interface {
	Make(config *types.KubernetesEnvironment) (environment.Environment, error)
}

type kubernetesMaker struct {
	*environment.CommonMaker
	clientsMaker clients.Maker
}

func CreateMaker(fnd app.Foundation, resourcesMaker resources.Maker) Maker {
	return &kubernetesMaker{
		CommonMaker:  environment.CreateCommonMaker(fnd, resourcesMaker),
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
	containerEnv, err := m.MakeContainerEnvironment(&types.ContainerEnvironment{
		Ports:     config.Ports,
		Resources: config.Resources,
		Registry:  config.Registry,
	})
	if err != nil {
		return nil, errors.Errorf("failed to create kubernetes client: %v", err)
	}

	return &kubernetesEnvironment{
		ContainerEnvironment: *containerEnv,
		kubeconfigPath:       config.Kubeconfig,
		namespace:            config.Namespace,
		useUniqueName:        true,
		configMapClient:      configMapClient,
		deploymentClient:     deploymentClient,
		podClient:            podClient,
		serviceClient:        serviceClient,
		tasks:                make(map[string]*kubernetesTask),
	}, nil
}

type kubernetesEnvironment struct {
	environment.ContainerEnvironment
	kubeconfigPath   string
	namespace        string
	useUniqueName    bool
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

// isNotFound reports whether the delete error can be ignored because the object is already gone.
func isNotFound(err error) bool {
	return apierrors.IsNotFound(err)
}

func (e *kubernetesEnvironment) destroyConfigMaps(
	ctx context.Context,
	configMaps []*corev1.ConfigMap,
	opts metav1.DeleteOptions,
) error {
	hasError := false
	for _, configMap := range configMaps {
		if err := e.configMapClient.Delete(ctx, configMap.Name, opts); err != nil && !isNotFound(err) {
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
		if err != nil && !isNotFound(err) {
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
		if err != nil && !isNotFound(err) {
			e.Fnd.Logger().Errorf("Failed to delete service %s: %v", service.Name, err)
			return errors.Errorf("failed to delete service")
		}
	}
	return nil
}

// closePortForward stops the public port forwarding of the task if it is active.
func (e *kubernetesEnvironment) closePortForward(kubeTask *kubernetesTask) error {
	if kubeTask.portForward == nil {
		return nil
	}
	err := kubeTask.portForward.Close()
	kubeTask.portForward = nil
	if err != nil {
		e.Fnd.Logger().Errorf("Failed to stop port forwarding for task %s: %v", kubeTask.serviceName, err)
		return errors.Errorf("failed to stop port forwarding")
	}
	return nil
}

// closeOutput stops the log collection of the task if it is active.
func (e *kubernetesEnvironment) closeOutput(kubeTask *kubernetesTask) error {
	if kubeTask.outputCollector == nil {
		return nil
	}
	err := kubeTask.outputCollector.Close()
	kubeTask.outputCollector = nil
	if err != nil {
		e.Fnd.Logger().Errorf("Failed to close output for task %s: %v", kubeTask.serviceName, err)
		return errors.Errorf("failed to close output")
	}
	return nil
}

func (e *kubernetesEnvironment) destroyTask(
	ctx context.Context,
	kubeTask *kubernetesTask,
	deleteOptions metav1.DeleteOptions,
) error {
	hasError := false

	// Stop the port forwarding
	if e.closePortForward(kubeTask) != nil {
		hasError = true
	}

	// Close output
	if e.closeOutput(kubeTask) != nil {
		hasError = true
	}

	// Delete the service
	if e.destroyService(ctx, kubeTask.service, deleteOptions) != nil {
		hasError = true
	}

	// Delete the deployment
	if e.destroyDeployment(ctx, kubeTask.deployment, deleteOptions) != nil {
		hasError = true
	}

	// Delete config maps
	if e.destroyConfigMaps(ctx, kubeTask.configMaps, deleteOptions) != nil {
		hasError = true
	}

	if hasError {
		return errors.Errorf("failed to delete task")
	}

	return nil
}

func (e *kubernetesEnvironment) Destroy(ctx context.Context) error {
	deleteOptions := metav1.DeleteOptions{DryRun: e.dryRunOption()}
	hasError := false
	// Iterate over all tasks to delete services and deployments
	for _, kubeTask := range e.tasks {
		// Delete the tasks
		if err := e.destroyTask(ctx, kubeTask, deleteOptions); err != nil {
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

func int64Ptr(i int64) *int64 {
	return &i
}

func (e *kubernetesEnvironment) serviceName(ss *environment.ServiceSettings) string {
	if e.useUniqueName {
		return ss.UniqueName
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
	configType string,
	serviceName string,
	workspacePaths,
	envPaths map[string]string,
	volumeMounts *[]corev1.VolumeMount,
	volumes *[]corev1.Volume,
) ([]*corev1.ConfigMap, error) {
	configMaps := make([]*corev1.ConfigMap, 0, len(workspacePaths))
	data := make(map[string]map[string]string)
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
		baseEnvPath := filepath.Base(envPath)
		dirEnvPath := filepath.Dir(envPath)

		if _, ok := data[dirEnvPath]; !ok {
			data[dirEnvPath] = make(map[string]string)
		}
		data[dirEnvPath][baseEnvPath] = content
	}

	configMapIndex := 1
	for dirEnvPath, baseData := range data {
		configMapName := sanitizeName(fmt.Sprintf("%s-%s-%d", serviceName, configType, configMapIndex))
		configMap, err := e.createConfigMap(ctx, configMapName, baseData)
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
		*volumes = append(*volumes, volume)

		*volumeMounts = append(*volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: dirEnvPath,
		})

		configMapIndex++
	}

	return configMaps, nil
}

// waitingFailureReasons lists the container waiting reasons that mean the container will not start.
var waitingFailureReasons = map[string]bool{
	"CrashLoopBackOff":           true,
	"ErrImagePull":               true,
	"ImagePullBackOff":           true,
	"ErrImageNeverPull":          true,
	"InvalidImageName":           true,
	"CreateContainerConfigError": true,
	"CreateContainerError":       true,
	"RunContainerError":          true,
}

// podFailure returns an error if the pod status shows that the pod cannot become ready.
func podFailure(pod *corev1.Pod) error {
	if pod.Status.Phase == corev1.PodFailed {
		return errors.Errorf("pod failed: %s %s", pod.Status.Reason, pod.Status.Message)
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && waitingFailureReasons[cs.State.Waiting.Reason] {
			return errors.Errorf("container %s is waiting: %s: %s",
				cs.Name, cs.State.Waiting.Reason, cs.State.Waiting.Message)
		}
		if cs.State.Terminated != nil {
			return errors.Errorf("container %s terminated with exit code %d: %s: %s",
				cs.Name, cs.State.Terminated.ExitCode, cs.State.Terminated.Reason, cs.State.Terminated.Message)
		}
		if cs.RestartCount > 0 && cs.LastTerminationState.Terminated != nil {
			return errors.Errorf("container %s restarted after termination with exit code %d: %s: %s",
				cs.Name, cs.LastTerminationState.Terminated.ExitCode, cs.LastTerminationState.Terminated.Reason,
				cs.LastTerminationState.Terminated.Message)
		}
	}
	return nil
}

// isPodReady reports whether the pod is running and ready.
func isPodReady(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// waitForDeployment watches the deployment and its pods until the deployment is ready and a ready pod is known
// or until a pod failure is detected.
func (e *kubernetesEnvironment) waitForDeployment(ctx context.Context, kubeTask *kubernetesTask) error {
	serviceName := kubeTask.serviceName
	deploymentWatcher, err := e.deploymentClient.Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", serviceName),
	})
	if err != nil {
		return errors.Errorf("failed to watch deployment %s: %v", serviceName, err)
	}
	defer deploymentWatcher.Stop()

	podWatcher, err := e.podClient.Watch(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", serviceName),
	})
	if err != nil {
		return errors.Errorf("failed to watch pods of deployment %s: %v", serviceName, err)
	}
	defer podWatcher.Stop()

	// Listen for events and handle context cancellation or timeout
	for {
		select {
		case event, ok := <-deploymentWatcher.ResultChan():
			if !ok {
				return errors.Errorf("watching deployment %s ended unexpectedly", serviceName)
			}
			switch event.Type {
			case watch.Added, watch.Modified:
				deployment, ok := event.Object.(*appsv1.Deployment)
				if !ok {
					return errors.Errorf("expected Deployment object, but got something else")
				}
				if deployment.Spec.Replicas != nil && deployment.Status.ReadyReplicas >= *deployment.Spec.Replicas {
					kubeTask.deploymentReady = true
				}
			case watch.Deleted, watch.Error:
				return errors.New("watching deployment did not result to addition and modification")
			}
		case event, ok := <-podWatcher.ResultChan():
			if !ok {
				return errors.Errorf("watching pods of deployment %s ended unexpectedly", serviceName)
			}
			switch event.Type {
			case watch.Added, watch.Modified:
				pod, ok := event.Object.(*corev1.Pod)
				if !ok {
					return errors.Errorf("expected Pod object, but got something else")
				}
				if err := podFailure(pod); err != nil {
					return errors.Errorf("pod %s of deployment %s failed: %v", pod.Name, serviceName, err)
				}
				if isPodReady(pod) {
					kubeTask.podName = pod.Name
				}
			case watch.Deleted:
				pod, ok := event.Object.(*corev1.Pod)
				if ok && pod.Name == kubeTask.podName {
					kubeTask.podName = ""
				}
			case watch.Error:
				return errors.Errorf("watching pods of deployment %s resulted in error", serviceName)
			}
		case <-ctx.Done():
			return errors.Errorf("context canceled or timed out when waiting on deployment to be ready")
		}
		if kubeTask.deploymentReady && kubeTask.podName != "" {
			return nil
		}
	}
}

func (e *kubernetesEnvironment) createDeployment(
	ctx context.Context,
	serviceName string,
	shortServiceName string,
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
		"configs",
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
		"scripts",
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
	var executable string
	if cmd != nil && cmd.Name != "" {
		executable = cmd.Name
		command = append(command, executable)
		args = cmd.Args
	} else {
		executable = shortServiceName
	}

	// Define the deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
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
							Name:    serviceName,
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
		executable:  executable,
	}
	// Register the task before waiting so that it is cleaned up by Destroy even if the waiting fails.
	e.tasks[serviceName] = kubeTask

	if e.Fnd.DryRun() {
		kubeTask.deploymentReady = true
		kubeTask.podName = serviceName
		return kubeTask, nil
	}

	if err = e.waitForDeployment(ctx, kubeTask); err != nil {
		return nil, err
	}

	return kubeTask, nil
}

func (e *kubernetesEnvironment) createService(
	ctx context.Context,
	kubeTask *kubernetesTask,
	serviceName string,
	ss *environment.ServiceSettings,
) error {
	kubeServiceSpec := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
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
	kubeTask.servicePrivateUrl = fmt.Sprintf("://%s:%d", serviceName, ss.ServerPort)

	return nil
}

// startOutput opens the pod log stream and collects it for the whole life of the task. The stream is not bound
// to the action context because the output is read by later actions.
func (e *kubernetesEnvironment) startOutput(ctx context.Context, kubeTask *kubernetesTask) error {
	if e.Fnd.DryRun() {
		return nil
	}
	logs, err := e.podClient.StreamLogs(
		context.WithoutCancel(ctx),
		kubeTask.podName,
		&corev1.PodLogOptions{Follow: true},
	)
	if err != nil {
		return errors.Errorf("failed to open log stream of pod %s: %v", kubeTask.podName, err)
	}
	collector := e.OutputMaker.MakeCollector(kubeTask.serviceName)
	if err = collector.Start(logs, &app.DummyReaderCloser{}); err != nil {
		_ = logs.Close()
		return errors.Errorf("failed to start output collection of pod %s: %v", kubeTask.podName, err)
	}
	kubeTask.outputCollector = collector
	return nil
}

// forwardPublicPort makes a public service reachable on the local host by forwarding the service port reserved
// in the environment to the pod server port.
func (e *kubernetesEnvironment) forwardPublicPort(
	ctx context.Context,
	kubeTask *kubernetesTask,
	ss *environment.ServiceSettings,
) error {
	if !ss.Public {
		return nil
	}
	kubeTask.servicePublicUrl = fmt.Sprintf("://127.0.0.1:%d", ss.Port)
	if e.Fnd.DryRun() {
		return nil
	}
	portForward, err := e.podClient.PortForward(ctx, kubeTask.podName, ss.Port, ss.ServerPort)
	if err != nil {
		return errors.Errorf("failed to forward port %d to pod %s port %d: %v",
			ss.Port, kubeTask.podName, ss.ServerPort, err)
	}
	kubeTask.portForward = portForward
	return nil
}

func (e *kubernetesEnvironment) RunTask(ctx context.Context, ss *environment.ServiceSettings, cmd *environment.Command) (task.Task, error) {
	serviceName := e.serviceName(ss)
	kubeTask, err := e.createDeployment(ctx, serviceName, ss.Name, ss, cmd)
	if err != nil {
		return nil, err
	}
	if err = e.createService(ctx, kubeTask, serviceName, ss); err != nil {
		return nil, err
	}
	if err = e.startOutput(ctx, kubeTask); err != nil {
		return nil, err
	}
	if err = e.forwardPublicPort(ctx, kubeTask, ss); err != nil {
		return nil, err
	}

	return kubeTask, nil
}

func convertTask(target task.Task) (*kubernetesTask, error) {
	kubeTask, ok := target.(*kubernetesTask)
	if !ok || kubeTask == nil {
		return nil, errors.Errorf("task in not a Kubernetes task")
	}
	return kubeTask, nil
}

func (e *kubernetesEnvironment) exec(
	ctx context.Context,
	kubeTask *kubernetesTask,
	command []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	return e.podClient.Exec(ctx, kubeTask.podName, &clients.ExecOptions{
		Container: kubeTask.serviceName,
		Command:   command,
		Stdout:    stdout,
		Stderr:    stderr,
	})
}

func (e *kubernetesEnvironment) ExecTaskCommand(
	ctx context.Context,
	ss *environment.ServiceSettings,
	target task.Task,
	cmd *environment.Command,
	oc output.Collector,
) error {
	kubeTask, err := convertTask(target)
	if err != nil {
		return err
	}
	if e.Fnd.DryRun() {
		return nil
	}

	var stdout, stderr io.Writer = io.Discard, io.Discard
	if oc != nil && !reflect.ValueOf(oc).IsNil() {
		stdout = oc.StdoutWriter()
		stderr = oc.StderrWriter()
	}

	command := append([]string{cmd.Name}, cmd.Args...)
	if err = e.exec(ctx, kubeTask, command, stdout, stderr); err != nil {
		return errors.Errorf("failed to execute command %s in pod %s: %v", cmd.Name, kubeTask.podName, err)
	}
	return nil
}

// terminateTask stops the task by deleting its deployment. The pods are then terminated by Kubernetes which sends
// SIGTERM to the containers. When gracePeriod is set, the pod is deleted directly with that grace period so that
// zero can be used to kill the container immediately.
func (e *kubernetesEnvironment) terminateTask(ctx context.Context, kubeTask *kubernetesTask, gracePeriod *int64) error {
	if kubeTask.deployment == nil {
		return errors.Errorf("task %s has already been terminated", kubeTask.serviceName)
	}
	// The pod is going away so the forwarding is stopped first.
	if err := e.closePortForward(kubeTask); err != nil {
		return err
	}
	if err := e.destroyDeployment(ctx, kubeTask.deployment, metav1.DeleteOptions{DryRun: e.dryRunOption()}); err != nil {
		return err
	}
	kubeTask.deployment = nil
	if gracePeriod != nil && kubeTask.podName != "" {
		err := e.podClient.Delete(ctx, kubeTask.podName, metav1.DeleteOptions{
			DryRun:             e.dryRunOption(),
			GracePeriodSeconds: gracePeriod,
		})
		if err != nil && !isNotFound(err) {
			return errors.Errorf("failed to delete pod %s: %v", kubeTask.podName, err)
		}
	}
	return nil
}

func (e *kubernetesEnvironment) ExecTaskSignal(ctx context.Context, ss *environment.ServiceSettings, target task.Task, signal os.Signal) error {
	kubeTask, err := convertTask(target)
	if err != nil {
		return err
	}

	switch signal {
	case syscall.SIGTERM:
		return e.terminateTask(ctx, kubeTask, nil)
	case syscall.SIGKILL:
		return e.terminateTask(ctx, kubeTask, int64Ptr(0))
	}

	signalName, ok := environment.SignalName(signal)
	if !ok {
		return errors.Errorf("signal %v is not supported in Kubernetes environment", signal)
	}
	if e.Fnd.DryRun() {
		return nil
	}
	command := []string{"kill", "-s", signalName, strconv.Itoa(kubeTask.Pid())}
	if err = e.exec(ctx, kubeTask, command, io.Discard, io.Discard); err != nil {
		return errors.Errorf("failed to send signal %s to pod %s: %v", signalName, kubeTask.podName, err)
	}
	return nil
}

func (e *kubernetesEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	kubeTask, err := convertTask(target)
	if err != nil {
		return nil, err
	}
	if e.Fnd.DryRun() {
		return &app.DummyReaderCloser{}, nil
	}
	if kubeTask.outputCollector == nil {
		return nil, errors.Errorf("output of task %s is not available", kubeTask.serviceName)
	}
	if outputType != output.Any {
		e.Fnd.Logger().Debugf(
			"Kubernetes pod logs do not separate stdout and stderr so the combined output is used for task %s",
			kubeTask.serviceName,
		)
	}
	return kubeTask.outputCollector.AnyReader(ctx), nil
}

func (e *kubernetesEnvironment) RootPath(workspace string) string {
	return ""
}

func (e *kubernetesEnvironment) Mkdir(serviceName string, path string, perm os.FileMode) error {
	// Currently it is a user responsibility to make sure that directory exists in the container
	return nil
}

func (e *kubernetesEnvironment) ServiceLocalAddress(serviceName string, servicePort, serverPort int32) string {
	// The server has to listen on all interfaces so that it is reachable through the Kubernetes service.
	return fmt.Sprintf("0.0.0.0:%d", serverPort)
}

func (e *kubernetesEnvironment) ServiceLocalPort(servicePort, serverPort int32) int32 {
	return serverPort
}

func (e *kubernetesEnvironment) ServicePrivateAddress(serviceName string, servicePort, serverPort int32) string {
	return serviceName
}

type kubernetesTask struct {
	deployment        *appsv1.Deployment
	service           *corev1.Service
	configMaps        []*corev1.ConfigMap
	outputCollector   output.Collector
	portForward       io.Closer
	podName           string
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

func (t *kubernetesTask) PublicUrl(scheme string) string {
	return scheme + t.servicePublicUrl
}

func (t *kubernetesTask) PrivateUrl(scheme string) string {
	return scheme + t.servicePrivateUrl
}
