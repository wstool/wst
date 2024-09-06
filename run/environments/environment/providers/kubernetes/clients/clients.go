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

package clients

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type Maker interface {
	MakeConfigMapClient(config *types.KubernetesEnvironment) (ConfigMapClient, error)
	MakeDeploymentClient(config *types.KubernetesEnvironment) (DeploymentClient, error)
	MakePodClient(config *types.KubernetesEnvironment) (PodClient, error)
	MakeServiceClient(config *types.KubernetesEnvironment) (ServiceClient, error)
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{fnd: fnd}
}

type nativeMaker struct {
	fnd            app.Foundation
	clientSet      *kubernetes.Clientset
	clientSetError error
}

func (m *nativeMaker) getClientSet(kubeconfigPath string) (*kubernetes.Clientset, error) {
	if m.clientSet == nil && m.clientSetError == nil {
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			m.clientSetError = err
		} else {
			// Create a clientset for interacting with the Kubernetes API
			m.clientSet, m.clientSetError = kubernetes.NewForConfig(kubeConfig)
		}

	}
	return m.clientSet, m.clientSetError
}

func (m *nativeMaker) makeConfigClient(config *types.KubernetesEnvironment) *configClient {
	return &configClient{
		kubeconfigPath: config.Kubeconfig,
		maker:          m,
		namespace:      config.Namespace,
	}
}

func (m *nativeMaker) MakeConfigMapClient(config *types.KubernetesEnvironment) (ConfigMapClient, error) {
	return &configMapClient{configClient: m.makeConfigClient(config)}, nil
}

func (m *nativeMaker) MakeDeploymentClient(config *types.KubernetesEnvironment) (DeploymentClient, error) {
	return &deploymentClient{configClient: m.makeConfigClient(config)}, nil
}

func (m *nativeMaker) MakePodClient(config *types.KubernetesEnvironment) (PodClient, error) {
	return &podClient{configClient: m.makeConfigClient(config)}, nil
}

func (m *nativeMaker) MakeServiceClient(config *types.KubernetesEnvironment) (ServiceClient, error) {
	return &serviceClient{configClient: m.makeConfigClient(config)}, nil
}

type WatchResult interface {
	Stop()
	ResultChan() <-chan watch.Event
}

type ConfigMapClient interface {
	Create(ctx context.Context, configMap *corev1.ConfigMap, opts metav1.CreateOptions) (*corev1.ConfigMap, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

type DeploymentClient interface {
	Create(ctx context.Context, deployment *appsv1.Deployment, opts metav1.CreateOptions) (*appsv1.Deployment, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Watch(ctx context.Context, opts metav1.ListOptions) (WatchResult, error)
}

type PodClient interface {
	StreamLogs(ctx context.Context, name string, opts *corev1.PodLogOptions) (io.ReadCloser, error)
	List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error)
}

type ServiceClient interface {
	Create(ctx context.Context, service *corev1.Service, opts metav1.CreateOptions) (*corev1.Service, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Watch(ctx context.Context, opts metav1.ListOptions) (WatchResult, error)
}

type configClient struct {
	kubeconfigPath string
	maker          *nativeMaker
	namespace      string
}

func (c *configClient) getConfigSet() (*kubernetes.Clientset, error) {
	return c.maker.getClientSet(c.kubeconfigPath)
}

type configMapClient struct {
	*configClient
	client clientcorev1.ConfigMapInterface
}

func (c *configMapClient) getClient() (clientcorev1.ConfigMapInterface, error) {
	if c.client == nil {
		clientSet, err := c.getConfigSet()
		if err != nil {
			return nil, err
		}
		c.client = clientSet.CoreV1().ConfigMaps(c.namespace)
	}
	return c.client, nil
}

func (c *configMapClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}
	return client.Delete(ctx, name, opts)
}

func (c *configMapClient) Create(
	ctx context.Context,
	configMap *corev1.ConfigMap,
	opts metav1.CreateOptions,
) (*corev1.ConfigMap, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}
	return client.Create(ctx, configMap, opts)
}

type deploymentClient struct {
	*configClient
	client clientappsv1.DeploymentInterface
}

func (d *deploymentClient) getClient() (clientappsv1.DeploymentInterface, error) {
	if d.client == nil {
		clientSet, err := d.getConfigSet()
		if err != nil {
			return nil, err
		}
		d.client = clientSet.AppsV1().Deployments(d.namespace)
	}
	return d.client, nil
}

func (d *deploymentClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	return client.Delete(ctx, name, opts)
}

func (d *deploymentClient) Create(
	ctx context.Context,
	deployment *appsv1.Deployment,
	opts metav1.CreateOptions,
) (*appsv1.Deployment, error) {
	client, err := d.getClient()
	if err != nil {
		return nil, err
	}
	return client.Create(ctx, deployment, opts)
}

func (d *deploymentClient) Watch(ctx context.Context, opts metav1.ListOptions) (WatchResult, error) {
	client, err := d.getClient()
	if err != nil {
		return nil, err
	}
	return client.Watch(ctx, opts)
}

type podClient struct {
	*configClient
	client clientcorev1.PodInterface
}

func (p *podClient) getClient() (clientcorev1.PodInterface, error) {
	if p.client == nil {
		clientSet, err := p.getConfigSet()
		if err != nil {
			return nil, err
		}
		p.client = clientSet.CoreV1().Pods(p.namespace)
	}
	return p.client, nil
}

func (p *podClient) StreamLogs(ctx context.Context, name string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	return client.GetLogs(name, opts).Stream(ctx)
}

func (p *podClient) List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	return client.List(ctx, opts)
}

type serviceClient struct {
	*configClient
	client clientcorev1.ServiceInterface
}

func (s *serviceClient) getClient() (clientcorev1.ServiceInterface, error) {
	if s.client == nil {
		clientSet, err := s.getConfigSet()
		if err != nil {
			return nil, err
		}
		s.client = clientSet.CoreV1().Services(s.namespace)
	}
	return s.client, nil
}

func (s *serviceClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	return client.Delete(ctx, name, opts)
}

func (s *serviceClient) Create(
	ctx context.Context,
	service *corev1.Service,
	opts metav1.CreateOptions,
) (*corev1.Service, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	return client.Create(ctx, service, opts)
}

func (s *serviceClient) Watch(ctx context.Context, opts metav1.ListOptions) (WatchResult, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	return client.Watch(ctx, opts)
}
