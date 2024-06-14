package clients

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
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

func (m *nativeMaker) getClientSet(config *types.KubernetesEnvironment) (*kubernetes.Clientset, error) {
	if m.clientSet == nil && m.clientSetError == nil {
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
		if err != nil {
			m.clientSetError = err
		} else {
			// Create a clientset for interacting with the Kubernetes API
			m.clientSet, m.clientSetError = kubernetes.NewForConfig(kubeConfig)
		}

	}
	return m.clientSet, m.clientSetError
}

func (m *nativeMaker) MakeConfigMapClient(config *types.KubernetesEnvironment) (ConfigMapClient, error) {
	clientSet, err := m.getClientSet(config)
	if err != nil {
		return nil, err
	}
	return &configMapClient{client: clientSet.CoreV1().ConfigMaps(config.Namespace)}, nil
}

func (m *nativeMaker) MakeDeploymentClient(config *types.KubernetesEnvironment) (DeploymentClient, error) {
	clientSet, err := m.getClientSet(config)
	if err != nil {
		return nil, err
	}
	return &deploymentClient{client: clientSet.AppsV1().Deployments(config.Namespace)}, nil
}

func (m *nativeMaker) MakePodClient(config *types.KubernetesEnvironment) (PodClient, error) {
	clientSet, err := m.getClientSet(config)
	if err != nil {
		return nil, err
	}
	return &podClient{client: clientSet.CoreV1().Pods(config.Namespace)}, nil
}

func (m *nativeMaker) MakeServiceClient(config *types.KubernetesEnvironment) (ServiceClient, error) {
	clientSet, err := m.getClientSet(config)
	if err != nil {
		return nil, err
	}
	return &serviceClient{client: clientSet.CoreV1().Services(config.Namespace)}, nil
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
	GetLogs(name string, opts *corev1.PodLogOptions) *restclient.Request
	List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error)
}

type ServiceClient interface {
	Create(ctx context.Context, service *corev1.Service, opts metav1.CreateOptions) (*corev1.Service, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Watch(ctx context.Context, opts metav1.ListOptions) (WatchResult, error)
}

type configMapClient struct {
	client clientcorev1.ConfigMapInterface
}

func (c *configMapClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete(ctx, name, opts)
}

func (c *configMapClient) Create(
	ctx context.Context,
	configMap *corev1.ConfigMap,
	opts metav1.CreateOptions,
) (*corev1.ConfigMap, error) {
	return c.client.Create(ctx, configMap, opts)
}

type deploymentClient struct {
	client clientappsv1.DeploymentInterface
}

func (d *deploymentClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return d.client.Delete(ctx, name, opts)
}

func (d *deploymentClient) Create(
	ctx context.Context,
	deployment *appsv1.Deployment,
	opts metav1.CreateOptions,
) (*appsv1.Deployment, error) {
	return d.client.Create(ctx, deployment, opts)
}

func (d *deploymentClient) Watch(ctx context.Context, opts metav1.ListOptions) (WatchResult, error) {
	return d.client.Watch(ctx, opts)
}

type podClient struct {
	client clientcorev1.PodInterface
}

func (p *podClient) GetLogs(name string, opts *corev1.PodLogOptions) *restclient.Request {
	return p.GetLogs(name, opts)
}

func (p *podClient) List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error) {
	return p.client.List(ctx, opts)
}

type serviceClient struct {
	client clientcorev1.ServiceInterface
}

func (s *serviceClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return s.client.Delete(ctx, name, opts)
}

func (s *serviceClient) Create(
	ctx context.Context,
	service *corev1.Service,
	opts metav1.CreateOptions,
) (*corev1.Service, error) {
	return s.client.Create(ctx, service, opts)
}

func (s *serviceClient) Watch(ctx context.Context, opts metav1.ListOptions) (WatchResult, error) {
	return s.client.Watch(ctx, opts)
}
