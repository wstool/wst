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
	"github.com/bukka/wst/run/services"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type Maker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd: fnd,
	}
}

func (m *Maker) Make(config *types.KubernetesEnvironment) (environment.Environment, error) {
	return &kubernetesEnvironment{
		fnd:            m.fnd,
		kubeconfigPath: config.Kubeconfig,
		namespace:      config.Namespace,
		useFullName:    false,
	}, nil
}

type kubernetesEnvironment struct {
	fnd              app.Foundation
	kubeconfigPath   string
	namespace        string
	useFullName      bool
	deploymentClient clientappsv1.DeploymentInterface
	podClient        clientcorev1.PodInterface
	serviceClient    clientcorev1.ServiceInterface
	tasks            map[string]*kubernetesTask
}

func (l *kubernetesEnvironment) Init(ctx context.Context) error {
	config, err := clientcmd.BuildConfigFromFlags("", l.kubeconfigPath)
	if err != nil {
		return err
	}

	// Create a clientset for interacting with the Kubernetes API
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	l.deploymentClient = clientset.AppsV1().Deployments(l.namespace)
	l.podClient = clientset.CoreV1().Pods(l.namespace)
	l.serviceClient = clientset.CoreV1().Services(l.namespace)

	return nil
}

func (l *kubernetesEnvironment) Destroy(ctx context.Context) error {
	var err error

	// Iterate over all tasks to delete services and deployments
	for _, task := range l.tasks {
		// Delete the service
		if task.service != nil {
			err = l.serviceClient.Delete(ctx, task.serviceName, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete service %s: %w", task.serviceName, err)
			}
		}

		// Delete the deployment
		if task.deployment != nil {
			err = l.deploymentClient.Delete(ctx, task.deployment.Name, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete deployment %s: %w", task.deployment.Name, err)
			}
		}
	}

	// Clear the tasks map for potential reuse of the environment
	l.tasks = make(map[string]*kubernetesTask)

	return nil
}

func int32Ptr(i int32) *int32 {
	return &i
}

func (l *kubernetesEnvironment) serviceName(service services.Service) string {
	if l.useFullName {
		return service.FullName()
	} else {
		return service.Name()
	}
}

func (l *kubernetesEnvironment) createDeployment(
	ctx context.Context,
	serviceName string,
	service services.Service,
	cmd *environment.Command,
) (*kubernetesTask, error) {
	containerConfig, err := service.Sandbox().ContainerConfig()
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
							Name:    service.Name(),
							Image:   containerConfig.Image(),
							Command: command,
							Args:    args,
							Ports: []corev1.ContainerPort{ // TODO: make it configurable
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := l.deploymentClient.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	kubeTask := &kubernetesTask{deployment: result}
	l.tasks[serviceName] = kubeTask

	watcher, err := l.deploymentClient.Watch(ctx, metav1.ListOptions{
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

func (l *kubernetesEnvironment) createService(
	ctx context.Context,
	kubeTask *kubernetesTask,
	serviceName string,
	service services.Service,
) error {
	kubeServiceSpec := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": serviceName,
			},
		},
	}

	kubeService, err := l.serviceClient.Create(ctx, kubeServiceSpec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	kubeTask.service = kubeService

	watcher, err := l.serviceClient.Watch(ctx, metav1.ListOptions{
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
					kubeTask.serviceUrl = fmt.Sprintf("http://%s", ip)
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

func (l *kubernetesEnvironment) RunTask(ctx context.Context, service services.Service, cmd *environment.Command) (task.Task, error) {
	serviceName := l.serviceName(service)
	kubeTask, err := l.createDeployment(ctx, serviceName, service, cmd)
	if err != nil {
		return nil, err
	}
	err = l.createService(ctx, kubeTask, serviceName, service)
	if err != nil {
		return nil, err
	}

	return kubeTask, nil
}

func (l *kubernetesEnvironment) ExecTaskCommand(ctx context.Context, service services.Service, target task.Task, cmd *environment.Command) error {
	return fmt.Errorf("executing command is not currently supported by Kubernetes environment")
}

func (l *kubernetesEnvironment) ExecTaskSignal(ctx context.Context, service services.Service, target task.Task, signal os.Signal) error {
	return fmt.Errorf("executing signal is not currently supported by Kubernetes environment")
}

func (l *kubernetesEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	if outputType != output.Any {
		return nil, fmt.Errorf("only any output type is supported by Kubernetes environment")
	}
	pods, err := l.podClient.List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", target.Name()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	readers := make([]io.Reader, 0, len(pods.Items))
	for _, pod := range pods.Items {
		req := l.podClient.GetLogs(pod.Name, &corev1.PodLogOptions{})
		podLogs, err := req.Stream(ctx)
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

func (l *kubernetesEnvironment) RootPath(service services.Service) string {
	return ""
}

type kubernetesTask struct {
	deployment      *appsv1.Deployment
	service         *corev1.Service
	serviceName     string
	serviceUrl      string
	deploymentReady bool
}

func (t *kubernetesTask) Name() string {
	return t.serviceName
}

func (k *kubernetesTask) Type() providers.Type {
	return providers.KubernetesType
}

func (t *kubernetesTask) BaseUrl() string {
	return t.serviceUrl
}
