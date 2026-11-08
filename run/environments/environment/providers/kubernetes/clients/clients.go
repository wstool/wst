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
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
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
	mu             sync.Mutex
	clientSet      *kubernetes.Clientset
	restConfig     *rest.Config
	clientSetError error
}

// getClientSet lazily creates the Kubernetes client set. When kubeconfigPath is empty, the standard loading rules
// are used (the KUBECONFIG environment variable or ~/.kube/config) with a fallback to the in-cluster configuration.
func (m *nativeMaker) getClientSet(kubeconfigPath string) (*kubernetes.Clientset, *rest.Config, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.clientSet == nil && m.clientSetError == nil {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.ExplicitPath = kubeconfigPath
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules,
			&clientcmd.ConfigOverrides{},
		)
		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			m.clientSetError = err
		} else {
			m.restConfig = restConfig
			// Create a clientset for interacting with the Kubernetes API
			m.clientSet, m.clientSetError = kubernetes.NewForConfig(restConfig)
		}
	}
	return m.clientSet, m.restConfig, m.clientSetError
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

// ExecOptions describe a command executed in a running container of a pod.
type ExecOptions struct {
	Container string
	Command   []string
	Stdout    io.Writer
	Stderr    io.Writer
}

type PodClient interface {
	StreamLogs(ctx context.Context, name string, opts *corev1.PodLogOptions) (io.ReadCloser, error)
	List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (WatchResult, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	// Exec runs the command in the pod container and waits for it to finish. A non-zero exit code is reported
	// as an error.
	Exec(ctx context.Context, name string, opts *ExecOptions) error
	// PortForward forwards the local port on 127.0.0.1 to the pod port. It returns once the forwarding is ready
	// and the returned closer stops the forwarding.
	PortForward(ctx context.Context, name string, localPort, podPort int32) (io.Closer, error)
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

func (c *configClient) getConfigSet() (*kubernetes.Clientset, *rest.Config, error) {
	return c.maker.getClientSet(c.kubeconfigPath)
}

type configMapClient struct {
	*configClient
	client clientcorev1.ConfigMapInterface
}

func (c *configMapClient) getClient() (clientcorev1.ConfigMapInterface, error) {
	if c.client == nil {
		clientSet, _, err := c.getConfigSet()
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
		clientSet, _, err := d.getConfigSet()
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
	client     clientcorev1.PodInterface
	restConfig *rest.Config
}

func (p *podClient) getClient() (clientcorev1.PodInterface, error) {
	if p.client == nil {
		clientSet, restConfig, err := p.getConfigSet()
		if err != nil {
			return nil, err
		}
		p.client = clientSet.CoreV1().Pods(p.namespace)
		p.restConfig = restConfig
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

func (p *podClient) Watch(ctx context.Context, opts metav1.ListOptions) (WatchResult, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	return client.Watch(ctx, opts)
}

func (p *podClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	client, err := p.getClient()
	if err != nil {
		return err
	}
	return client.Delete(ctx, name, opts)
}

func (p *podClient) Exec(ctx context.Context, name string, opts *ExecOptions) error {
	if _, err := p.getClient(); err != nil {
		return err
	}
	clientSet, _, err := p.getConfigSet()
	if err != nil {
		return err
	}
	request := clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(p.namespace).
		Name(name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: opts.Container,
			Command:   opts.Command,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(p.restConfig, http.MethodPost, request.URL())
	if err != nil {
		return errors.Errorf("failed to create executor: %v", err)
	}
	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
	})
}

// loggerWriter forwards the port forwarder messages to the debug log.
type loggerWriter struct {
	fnd    app.Foundation
	prefix string
}

func (w *loggerWriter) Write(p []byte) (int, error) {
	w.fnd.Logger().Debugf("%s: %s", w.prefix, string(p))
	return len(p), nil
}

type portForward struct {
	forwarder *portforward.PortForwarder
	stopCh    chan struct{}
	errCh     chan error
	stopOnce  sync.Once
}

func (f *portForward) Close() error {
	var err error
	f.stopOnce.Do(func() {
		close(f.stopCh)
		err = <-f.errCh
	})
	return err
}

func (p *podClient) PortForward(ctx context.Context, name string, localPort, podPort int32) (io.Closer, error) {
	if _, err := p.getClient(); err != nil {
		return nil, err
	}
	clientSet, _, err := p.getConfigSet()
	if err != nil {
		return nil, err
	}
	request := clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(p.namespace).
		Name(name).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(p.restConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create port forward transport: %v", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, request.URL())

	stopCh := make(chan struct{})
	readyCh := make(chan struct{})
	forwarder, err := portforward.NewOnAddresses(
		dialer,
		[]string{"127.0.0.1"},
		[]string{fmt.Sprintf("%d:%d", localPort, podPort)},
		stopCh,
		readyCh,
		&loggerWriter{fnd: p.maker.fnd, prefix: "port forward"},
		&loggerWriter{fnd: p.maker.fnd, prefix: "port forward error"},
	)
	if err != nil {
		return nil, errors.Errorf("failed to create port forwarder: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- forwarder.ForwardPorts()
	}()

	select {
	case <-readyCh:
		return &portForward{forwarder: forwarder, stopCh: stopCh, errCh: errCh}, nil
	case err := <-errCh:
		if err == nil {
			err = errors.New("port forwarding stopped before it was ready")
		}
		return nil, errors.Errorf("failed to forward port %d to pod %s port %d: %v", localPort, name, podPort, err)
	case <-ctx.Done():
		close(stopCh)
		<-errCh
		return nil, errors.Errorf("context canceled or timed out when waiting for port forwarding to pod %s", name)
	}
}

type serviceClient struct {
	*configClient
	client clientcorev1.ServiceInterface
}

func (s *serviceClient) getClient() (clientcorev1.ServiceInterface, error) {
	if s.client == nil {
		clientSet, _, err := s.getConfigSet()
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
