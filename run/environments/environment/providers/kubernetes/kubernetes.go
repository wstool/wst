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
	"github.com/bukka/wst/run/environments/task"
	"io"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type Maker struct {
	environment.Maker
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		Maker: *environment.CreateMaker(fnd),
	}
}

func (m *Maker) Make(config *types.KubernetesEnvironment) (environment.Environment, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	if err != nil {
		return nil, err
	}

	// Create a clientset for interacting with the Kubernetes API
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return &kubernetesEnvironment{
		ContainerEnvironment: *m.MakeContainerEnvironment(&types.ContainerEnvironment{
			Ports:    config.Ports,
			Registry: config.Registry,
		}),
		kubeconfigPath:   config.Kubeconfig,
		namespace:        config.Namespace,
		useFullName:      false,
		deploymentClient: clientset.AppsV1().Deployments(config.Namespace),
		configMapClient:  clientset.CoreV1().ConfigMaps(config.Namespace),
		podClient:        clientset.CoreV1().Pods(config.Namespace),
		serviceClient:    clientset.CoreV1().Services(config.Namespace),
	}, nil
}

type kubernetesEnvironment struct {
	environment.ContainerEnvironment
	fnd              app.Foundation
	kubeconfigPath   string
	namespace        string
	useFullName      bool
	deploymentClient clientappsv1.DeploymentInterface
	configMapClient  clientcorev1.ConfigMapInterface
	podClient        clientcorev1.PodInterface
	serviceClient    clientcorev1.ServiceInterface
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

func (e *kubernetesEnvironment) Destroy(ctx context.Context) error {
	var err error

	deleteOptions := metav1.DeleteOptions{DryRun: e.dryRunOption()}
	// Iterate over all tasks to delete services and deployments
	for _, kubeTask := range e.tasks {
		// Delete the service
		if kubeTask.service != nil {
			err = e.serviceClient.Delete(ctx, kubeTask.serviceName, deleteOptions)
			if err != nil {
				return fmt.Errorf("failed to delete service %s: %w", kubeTask.serviceName, err)
			}
		}

		// Delete the deployment
		if kubeTask.deployment != nil {
			err = e.deploymentClient.Delete(ctx, kubeTask.deployment.Name, deleteOptions)
			if err != nil {
				return fmt.Errorf("failed to delete deployment %s: %w", kubeTask.deployment.Name, err)
			}
		}
	}

	// Clear the tasks map for potential reuse of the environment
	e.tasks = make(map[string]*kubernetesTask)

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
func loadFileContent(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file at %s: %w", filePath, err)
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
) error {
	for name, hostPath := range workspacePaths {
		envPath, found := envPaths[name]
		if !found {
			return fmt.Errorf("environment path not found for %s", name)
		}

		// Load the content of the file at hostPath
		content, err := loadFileContent(hostPath)
		if err != nil {
			return err
		}

		// Create a ConfigMap for the file content
		configMapName := sanitizeName(fmt.Sprintf("%s-%s", serviceName, name))
		data := map[string]string{
			filepath.Base(hostPath): content, // Use the file name as the key
		}

		_, err = e.createConfigMap(ctx, configMapName, data)
		if err != nil {
			return fmt.Errorf("failed to create configMap %s: %w", configMapName, err)
		}

		// Prepare volume and volume mount for this ConfigMap
		volumeName := configMapName + "-volume"
		*volumes = append(*volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		})

		*volumeMounts = append(*volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: envPath,
		})
	}

	return nil
}

func (e *kubernetesEnvironment) createDeployment(
	ctx context.Context,
	serviceName string,
	ss *environment.ServiceSettings,
	cmd *environment.Command,
) (*kubernetesTask, error) {
	containerConfig, err := ss.Sandbox.ContainerConfig()
	if err != nil {
		return nil, err
	}

	var volumeMounts []corev1.VolumeMount
	var volumes []corev1.Volume

	err = e.processWorkspacePaths(
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
	err = e.processWorkspacePaths(
		ctx,
		serviceName,
		ss.WorkspaceScriptPaths,
		ss.EnvironmentScriptPaths,
		&volumeMounts,
		&volumes,
	)
	if err != nil {
		return nil, err
	}

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
									ContainerPort: ss.Port,
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
		return nil, err
	}
	kubeTask := &kubernetesTask{deployment: result}
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
					return nil, fmt.Errorf("expected Deployment object, but got something else")
				}
				if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
					kubeTask.deploymentReady = true
					return kubeTask, nil
				}
			case watch.Deleted:
				fallthrough
			case watch.Error:
				return nil, fmt.Errorf("watch error: %v", event.Object)
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled or timed out when waiting on deployment to be ready")
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
					Port:       ss.Port,
					TargetPort: intstr.FromInt32(ss.Port),
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
		return fmt.Errorf("failed to create service: %w", err)
	}

	kubeTask.service = kubeService

	if e.Fnd.DryRun() {
		if ss.Public {
			kubeTask.servicePublicUrl = "http://127.0.0.1"
		}
		kubeTask.servicePrivateUrl = fmt.Sprintf("http://%s:%s", serviceName, ss.Port)
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
					return fmt.Errorf("expected Service object, but got something else")
				}
				if len(svc.Status.LoadBalancer.Ingress) > 0 {
					ip := svc.Status.LoadBalancer.Ingress[0].IP
					if ss.Public {
						kubeTask.servicePublicUrl = fmt.Sprintf("http://%s", ip)
					}
					kubeTask.servicePrivateUrl = fmt.Sprintf("http://%s:%s", serviceName, ss.Port)
					return nil
				}
			case watch.Deleted:
				fallthrough
			case watch.Error:
				return fmt.Errorf("watch error: %v", event.Object)
			}
		case <-ctx.Done():
			return fmt.Errorf("context canceled or timed out when waiting on service IP")
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
	return fmt.Errorf("executing command is not currently supported in Kubernetes environment")
}

func (e *kubernetesEnvironment) ExecTaskSignal(ctx context.Context, ss *environment.ServiceSettings, target task.Task, signal os.Signal) error {
	return fmt.Errorf("executing signal is not currently supported in Kubernetes environment")
}

func (e *kubernetesEnvironment) logsStream(ctx context.Context, pod corev1.Pod) (io.ReadCloser, error) {
	if e.Fnd.DryRun() {
		return &app.DummyReaderCloser{}, nil
	}
	return e.podClient.GetLogs(pod.Name, &corev1.PodLogOptions{}).Stream(ctx)
}

func (e *kubernetesEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	if outputType != output.Any {
		return nil, fmt.Errorf("only any output type is supported by Kubernetes environment")
	}
	pods, err := e.podClient.List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", target.Name()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	readers := make([]io.Reader, 0, len(pods.Items))
	for _, pod := range pods.Items {
		podLogs, err := e.logsStream(ctx, pod)
		if err != nil {
			return nil, fmt.Errorf("error in opening stream: %w", err)
		}
		defer podLogs.Close()

		readers = append(readers, podLogs)
	}

	// Combine all readers into a single one
	combinedReader := io.MultiReader(readers...)

	return combinedReader, nil
}

func (e *kubernetesEnvironment) RootPath(workspace string) string {
	return ""
}

type kubernetesTask struct {
	deployment        *appsv1.Deployment
	service           *corev1.Service
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

func (k *kubernetesTask) Type() providers.Type {
	return providers.KubernetesType
}

func (t *kubernetesTask) PublicUrl() string {
	return t.servicePublicUrl
}

func (t *kubernetesTask) PrivateUrl() string {
	return t.servicePrivateUrl
}
