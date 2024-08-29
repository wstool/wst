package kubernetes

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	k8sClientMocks "github.com/bukka/wst/mocks/generated/run/environments/environment/providers/kubernetes/clients"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/environments/task"
	"github.com/bukka/wst/run/sandboxes/containers"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"os"
	"strings"
	"testing"
)

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	tests := []struct {
		name string
		fnd  app.Foundation
	}{
		{
			name: "create maker",
			fnd:  fndMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateMaker(tt.fnd)
			maker, ok := got.(*kubernetesMaker)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, maker.Fnd)
		})
	}
}

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name       string
		config     *types.KubernetesEnvironment
		setupMocks func(*testing.T, *k8sClientMocks.MockMaker, *types.KubernetesEnvironment) (
			*k8sClientMocks.MockConfigMapClient,
			*k8sClientMocks.MockDeploymentClient,
			*k8sClientMocks.MockPodClient,
			*k8sClientMocks.MockServiceClient,
		)
		getExpectedEnv func(
			fndMock *appMocks.MockFoundation,
			configMapClient *k8sClientMocks.MockConfigMapClient,
			deploymentClient *k8sClientMocks.MockDeploymentClient,
			podClient *k8sClientMocks.MockPodClient,
			serviceClient *k8sClientMocks.MockServiceClient,
		) *kubernetesEnvironment
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful kubernetes environment maker creation",
			config: &types.KubernetesEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "u1",
						Password: "p1",
					},
				},
				Namespace:  "test",
				Kubeconfig: "/home/kubeconfig/config.yaml",
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				dc := k8sClientMocks.NewMockDeploymentClient(t)
				m.On("MakeDeploymentClient", config).Return(dc, nil)
				pc := k8sClientMocks.NewMockPodClient(t)
				m.On("MakePodClient", config).Return(pc, nil)
				sc := k8sClientMocks.NewMockServiceClient(t)
				m.On("MakeServiceClient", config).Return(sc, nil)
				return cmc, dc, pc, sc
			},
			getExpectedEnv: func(
				fndMock *appMocks.MockFoundation,
				configMapClient *k8sClientMocks.MockConfigMapClient,
				deploymentClient *k8sClientMocks.MockDeploymentClient,
				podClient *k8sClientMocks.MockPodClient,
				serviceClient *k8sClientMocks.MockServiceClient,
			) *kubernetesEnvironment {
				return &kubernetesEnvironment{
					ContainerEnvironment: environment.ContainerEnvironment{
						CommonEnvironment: environment.CommonEnvironment{
							Fnd:  fndMock,
							Used: false,
							Ports: environment.Ports{
								Start: 8000,
								Used:  8000,
								End:   8500,
							},
						},
						Registry: environment.ContainerRegistry{
							Auth: environment.ContainerRegistryAuth{
								Username: "u1",
								Password: "p1",
							},
						},
					},
					configMapClient:  configMapClient,
					deploymentClient: deploymentClient,
					podClient:        podClient,
					serviceClient:    serviceClient,
					namespace:        "test",
					kubeconfigPath:   "/home/kubeconfig/config.yaml",
					tasks:            make(map[string]*kubernetesTask),
				}
			},
		},
		{
			name: "failed kubernetes environment maker creation due to service client failure",
			config: &types.KubernetesEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "u1",
						Password: "p1",
					},
				},
				Namespace:  "test",
				Kubeconfig: "/home/kubeconfig/config.yaml",
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				dc := k8sClientMocks.NewMockDeploymentClient(t)
				m.On("MakeDeploymentClient", config).Return(dc, nil)
				pc := k8sClientMocks.NewMockPodClient(t)
				m.On("MakePodClient", config).Return(pc, nil)
				m.On("MakeServiceClient", config).Return(nil, errors.New("failed sc"))
				return cmc, dc, pc, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to create kubernetes client: failed sc",
		},
		{
			name: "failed kubernetes environment maker creation due to pod client failure",
			config: &types.KubernetesEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "u1",
						Password: "p1",
					},
				},
				Namespace:  "test",
				Kubeconfig: "/home/kubeconfig/config.yaml",
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				dc := k8sClientMocks.NewMockDeploymentClient(t)
				m.On("MakeDeploymentClient", config).Return(dc, nil)
				m.On("MakePodClient", config).Return(nil, errors.New("failed sc"))
				return cmc, dc, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to create kubernetes client: failed sc",
		},
		{
			name: "failed kubernetes environment maker creation due to deployment client failure",
			config: &types.KubernetesEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "u1",
						Password: "p1",
					},
				},
				Namespace:  "test",
				Kubeconfig: "/home/kubeconfig/config.yaml",
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				m.On("MakeDeploymentClient", config).Return(nil, errors.New("failed dc"))
				return cmc, nil, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to create kubernetes client: failed dc",
		},
		{
			name: "failed kubernetes environment maker creation due to config map client failure",
			config: &types.KubernetesEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "u1",
						Password: "p1",
					},
				},
				Namespace:  "test",
				Kubeconfig: "/home/kubeconfig/config.yaml",
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
			) {
				m.On("MakeConfigMapClient", config).Return(nil, errors.New("failed cmc"))
				return nil, nil, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to create kubernetes client: failed cmc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			clientsMakerMock := k8sClientMocks.NewMockMaker(t)
			m := &kubernetesMaker{
				CommonMaker: &environment.CommonMaker{
					Fnd: fndMock,
				},
				clientsMaker: clientsMakerMock,
			}

			cmc, dc, pc, sc := tt.setupMocks(t, clientsMakerMock, tt.config)

			got, err := m.Make(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				actualEnv, ok := got.(*kubernetesEnvironment)
				assert.True(t, ok)
				expectedEnv := tt.getExpectedEnv(fndMock, cmc, dc, pc, sc)
				assert.Equal(t, expectedEnv, actualEnv)
			}
		})
	}
}

func Test_kubernetesEnvironment_Init(t *testing.T) {
	env := &kubernetesEnvironment{}
	ctx := context.Background()
	assert.Nil(t, env.Init(ctx))
}

func Test_kubernetesEnvironment_Destroy(t *testing.T) {
	tests := []struct {
		name       string
		tasks      map[string]*kubernetesTask
		setupMocks func(
			*testing.T,
			context.Context,
			*appMocks.MockFoundation,
			*k8sClientMocks.MockConfigMapClient,
			*k8sClientMocks.MockDeploymentClient,
			*k8sClientMocks.MockServiceClient,
		)
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful kubernetes env destroying",
			tasks: map[string]*kubernetesTask{
				"t1": {
					configMaps: []*corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c11",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c12",
							},
						},
					},
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "d1",
						},
					},
					service: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "s1",
						},
					},
					outputReader: &CombinedReader{readers: []io.ReadCloser{&app.DummyReaderCloser{}}},
				},
				"t2": {
					configMaps: []*corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c21",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c22",
							},
						},
					},
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "d2",
						},
					},
					service: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "s2",
						},
					},
				},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) {
				fnd.On("DryRun").Return(false)
				opts := metav1.DeleteOptions{}
				cmc.On("Delete", ctx, "c11", opts).Return(nil)
				cmc.On("Delete", ctx, "c12", opts).Return(nil)
				cmc.On("Delete", ctx, "c21", opts).Return(nil)
				cmc.On("Delete", ctx, "c22", opts).Return(nil)
				dc.On("Delete", ctx, "d1", opts).Return(nil)
				dc.On("Delete", ctx, "d2", opts).Return(nil)
				sc.On("Delete", ctx, "s1", opts).Return(nil)
				sc.On("Delete", ctx, "s2", opts).Return(nil)
			},
		},
		{
			name: "successful kubernetes env destroying",
			tasks: map[string]*kubernetesTask{
				"t1": {
					configMaps: []*corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c11",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c12",
							},
						},
					},
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "d1",
						},
					},
					service: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "s1",
						},
					},
				},
				"t2": {
					configMaps: []*corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c21",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c22",
							},
						},
					},
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "d2",
						},
					},
					service: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "s2",
						},
					},
				},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) {
				fnd.On("DryRun").Return(true)
				opts := metav1.DeleteOptions{DryRun: []string{metav1.DryRunAll}}
				cmc.On("Delete", ctx, "c11", opts).Return(nil)
				cmc.On("Delete", ctx, "c12", opts).Return(nil)
				cmc.On("Delete", ctx, "c21", opts).Return(nil)
				cmc.On("Delete", ctx, "c22", opts).Return(nil)
				dc.On("Delete", ctx, "d1", opts).Return(nil)
				dc.On("Delete", ctx, "d2", opts).Return(nil)
				sc.On("Delete", ctx, "s1", opts).Return(nil)
				sc.On("Delete", ctx, "s2", opts).Return(nil)
			},
		},
		{
			name: "failed kubernetes env destroying with single error",
			tasks: map[string]*kubernetesTask{
				"t1": {
					configMaps: []*corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c11",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c12",
							},
						},
					},
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "d1",
						},
					},
					service: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "s1",
						},
					},
				},
				"t2": {
					configMaps: []*corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c21",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c22",
							},
						},
					},
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "d2",
						},
					},
					service: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "s2",
						},
					},
				},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) {
				fnd.On("DryRun").Return(true)
				opts := metav1.DeleteOptions{DryRun: []string{metav1.DryRunAll}}
				cmc.On("Delete", ctx, "c11", opts).Return(errors.New("c11 fail"))
				cmc.On("Delete", ctx, "c12", opts).Return(nil)
				cmc.On("Delete", ctx, "c21", opts).Return(nil)
				cmc.On("Delete", ctx, "c22", opts).Return(nil)
				dc.On("Delete", ctx, "d1", opts).Return(nil)
				dc.On("Delete", ctx, "d2", opts).Return(nil)
				sc.On("Delete", ctx, "s1", opts).Return(nil)
				sc.On("Delete", ctx, "s2", opts).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "failed to destroy kubernetes environment",
		},
		{
			name: "failed kubernetes env destroying with multiple errors",
			tasks: map[string]*kubernetesTask{
				"t1": {
					configMaps: []*corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c11",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c12",
							},
						},
					},
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "d1",
						},
					},
					service: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "s1",
						},
					},
					outputReader: &CombinedReader{readers: []io.ReadCloser{&failingReader{}}},
				},
				"t2": {
					configMaps: []*corev1.ConfigMap{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c21",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "c22",
							},
						},
					},
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name: "d2",
						},
					},
					service: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name: "s2",
						},
					},
				},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) {
				fnd.On("DryRun").Return(true)
				opts := metav1.DeleteOptions{DryRun: []string{metav1.DryRunAll}}
				cmc.On("Delete", ctx, "c11", opts).Return(errors.New("c11 fail"))
				cmc.On("Delete", ctx, "c12", opts).Return(nil)
				cmc.On("Delete", ctx, "c21", opts).Return(errors.New("c21 fail"))
				cmc.On("Delete", ctx, "c22", opts).Return(nil)
				dc.On("Delete", ctx, "d1", opts).Return(errors.New("d1 fail"))
				dc.On("Delete", ctx, "d2", opts).Return(nil)
				sc.On("Delete", ctx, "s1", opts).Return(errors.New("s1 fail"))
				sc.On("Delete", ctx, "s2", opts).Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "failed to destroy kubernetes environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
			cmc := k8sClientMocks.NewMockConfigMapClient(t)
			dc := k8sClientMocks.NewMockDeploymentClient(t)
			sc := k8sClientMocks.NewMockServiceClient(t)
			ctx := context.Background()
			e := &kubernetesEnvironment{
				ContainerEnvironment: environment.ContainerEnvironment{
					CommonEnvironment: environment.CommonEnvironment{
						Fnd: fndMock,
					},
					Registry: environment.ContainerRegistry{},
				},
				configMapClient:  cmc,
				deploymentClient: dc,
				serviceClient:    sc,
				tasks:            tt.tasks,
			}

			tt.setupMocks(t, ctx, fndMock, cmc, dc, sc)

			err := e.Destroy(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type MockWatchResult struct {
	Events chan watch.Event
}

func (m *MockWatchResult) Stop() {
	close(m.Events)
}

func (m *MockWatchResult) ResultChan() <-chan watch.Event {
	return m.Events
}

func getTestingConfigMaps(serviceName string) []*corev1.ConfigMap {
	return []*corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceName + "-main-conf",
			},
			Data: map[string]string{
				"main.conf": "main: data",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceName + "-test-php",
			},
			Data: map[string]string{
				"my_test.php": "<?php echo 1; ?>",
			},
		},
	}
}

func getTestingDeployment(serviceName string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
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
							Name:    "svc",
							Image:   "wst:test",
							Command: []string{"php"},
							Args:    []string{"test.php", "run"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 1234,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      serviceName + "-main-conf-volume",
									MountPath: "/etc",
								},
								{
									Name:      serviceName + "-test-php-volume",
									MountPath: "/www",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: serviceName + "-main-conf-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: serviceName + "-main-conf",
									},
								},
							},
						},
						{
							Name: serviceName + "-test-php-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: serviceName + "-test-php",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "my_test.php",
											Path: "test.php",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func getTestingService(serviceType corev1.ServiceType, serviceName string, port int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Type: serviceType,
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.FromInt32(port),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": serviceName,
			},
		},
	}
}

func Test_kubernetesEnvironment_RunTask(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		kubeconfigPath string
		useFullName    bool
		envStartPort   int32
		fs             map[string]string
		ss             *environment.ServiceSettings
		cmd            *environment.Command
		setupMocks     func(
			*testing.T,
			context.Context,
			context.CancelFunc,
			*appMocks.MockFoundation,
			*k8sClientMocks.MockConfigMapClient,
			*k8sClientMocks.MockDeploymentClient,
			*k8sClientMocks.MockServiceClient,
		) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service)
		contextSetup     func() (context.Context, context.CancelFunc)
		getExpectedTask  func([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) *kubernetesTask
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:           "successful kubernetes private run",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    true,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     false,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("mysvc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("mysvc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil)

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=mysvc",
				}).Return(mockDeploymentWatchResult, nil)
				go func() {
					mockDeploymentWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{Name: "mysv"},
							Status: appsv1.DeploymentStatus{
								ReadyReplicas: 3,
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: int32Ptr(3),
							},
						},
					}
				}()

				s := getTestingService(corev1.ServiceTypeClusterIP, "mysvc", 1234)
				sc.On("Create", ctx, s, metav1.CreateOptions{}).Return(s, nil)

				mockServiceWatchResult := &MockWatchResult{
					Events: make(chan watch.Event, 1),
				}
				sc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=mysvc",
				}).Return(mockServiceWatchResult, nil)
				go func() {
					mockServiceWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &corev1.Service{
							ObjectMeta: metav1.ObjectMeta{Name: "mysvc"},
							Status: corev1.ServiceStatus{
								LoadBalancer: corev1.LoadBalancerStatus{
									Ingress: []corev1.LoadBalancerIngress{
										{
											IP: "10.0.0.1",
										},
									},
								},
							},
						},
					}
				}()

				return cm, d, s
			},
			getExpectedTask: func(cm []*corev1.ConfigMap, d *appsv1.Deployment, s *corev1.Service) *kubernetesTask {
				return &kubernetesTask{
					configMaps:        cm,
					deployment:        d,
					service:           s,
					executable:        "php",
					serviceName:       "mysvc",
					servicePublicUrl:  "",
					servicePrivateUrl: "http://mysvc:1234",
					deploymentReady:   true,
				}
			},
		}, {
			name:           "successful kubernetes public run",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil)

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockDeploymentWatchResult, nil)
				go func() {
					mockDeploymentWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{Name: "svc"},
							Status: appsv1.DeploymentStatus{
								ReadyReplicas: 3,
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: int32Ptr(3),
							},
						},
					}
				}()

				s := getTestingService(corev1.ServiceTypeLoadBalancer, "svc", 1234)
				sc.On("Create", ctx, s, metav1.CreateOptions{}).Return(s, nil)

				mockServiceWatchResult := &MockWatchResult{
					Events: make(chan watch.Event, 1),
				}
				sc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockServiceWatchResult, nil)
				go func() {
					mockServiceWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &corev1.Service{
							ObjectMeta: metav1.ObjectMeta{Name: "svc"},
							Status: corev1.ServiceStatus{
								LoadBalancer: corev1.LoadBalancerStatus{
									Ingress: []corev1.LoadBalancerIngress{
										{
											IP: "10.0.0.1",
										},
									},
								},
							},
						},
					}
				}()

				return cm, d, s
			},
			getExpectedTask: func(cm []*corev1.ConfigMap, d *appsv1.Deployment, s *corev1.Service) *kubernetesTask {
				return &kubernetesTask{
					configMaps:        cm,
					deployment:        d,
					service:           s,
					executable:        "php",
					serviceName:       "svc",
					servicePublicUrl:  "http://10.0.0.1",
					servicePrivateUrl: "http://svc:1234",
					deploymentReady:   true,
				}
			},
		},
		{
			name:           "successful kubernetes public dry run",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(true)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}}).Return(d, nil)

				s := getTestingService(corev1.ServiceTypeLoadBalancer, "svc", 1234)
				sc.On("Create", ctx, s, metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}}).Return(s, nil)

				return cm, d, s
			},
			getExpectedTask: func(cm []*corev1.ConfigMap, d *appsv1.Deployment, s *corev1.Service) *kubernetesTask {
				return &kubernetesTask{
					configMaps:        cm,
					deployment:        d,
					service:           s,
					executable:        "php",
					serviceName:       "svc",
					servicePublicUrl:  "http://127.0.0.1",
					servicePrivateUrl: "http://svc:1234",
					deploymentReady:   true,
				}
			},
		},
		{
			name:           "failed run due to service watch result object not being Service",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil)

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockDeploymentWatchResult, nil)
				go func() {
					mockDeploymentWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{Name: "svc"},
							Status: appsv1.DeploymentStatus{
								ReadyReplicas: 3,
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: int32Ptr(3),
							},
						},
					}
				}()

				s := getTestingService(corev1.ServiceTypeLoadBalancer, "svc", 1234)
				sc.On("Create", ctx, s, metav1.CreateOptions{}).Return(s, nil)

				mockServiceWatchResult := &MockWatchResult{
					Events: make(chan watch.Event, 1),
				}
				sc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockServiceWatchResult, nil)
				go func() {
					mockServiceWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{Name: "svc"},
						},
					}
				}()

				return cm, d, s
			},
			expectError:      true,
			expectedErrorMsg: "expected Service object, but got something else",
		},
		{
			name:           "failed run due to service watch resulted in deleted event",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil)

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockDeploymentWatchResult, nil)
				go func() {
					mockDeploymentWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{Name: "svc"},
							Status: appsv1.DeploymentStatus{
								ReadyReplicas: 3,
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: int32Ptr(3),
							},
						},
					}
				}()

				s := getTestingService(corev1.ServiceTypeLoadBalancer, "svc", 1234)
				sc.On("Create", ctx, s, metav1.CreateOptions{}).Return(s, nil)

				mockServiceWatchResult := &MockWatchResult{
					Events: make(chan watch.Event, 1),
				}
				sc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockServiceWatchResult, nil)
				go func() {
					mockServiceWatchResult.Events <- watch.Event{
						Type: watch.Deleted,
					}
				}()

				return cm, d, s
			},
			expectError:      true,
			expectedErrorMsg: "watching service did not result to addition and modification",
		},
		{
			name:           "failed run due to context being cancelled before service watch",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil)

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockDeploymentWatchResult, nil)
				go func() {
					mockDeploymentWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{Name: "svc"},
							Status: appsv1.DeploymentStatus{
								ReadyReplicas: 3,
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: int32Ptr(3),
							},
						},
					}
				}()

				s := getTestingService(corev1.ServiceTypeLoadBalancer, "svc", 1234)
				sc.On("Create", ctx, s, metav1.CreateOptions{}).Return(s, nil).Run(func(args mock.Arguments) {
					cancel()
				})

				mockServiceWatchResult := &MockWatchResult{
					Events: make(chan watch.Event, 1),
				}
				sc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockServiceWatchResult, nil)

				return cm, d, s
			},
			expectError:      true,
			expectedErrorMsg: "context canceled or timed out when waiting on service IP",
		},
		{
			name:           "failed due to service create error",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil)

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockDeploymentWatchResult, nil)
				go func() {
					mockDeploymentWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{Name: "svc"},
							Status: appsv1.DeploymentStatus{
								ReadyReplicas: 3,
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: int32Ptr(3),
							},
						},
					}
				}()

				s := getTestingService(corev1.ServiceTypeLoadBalancer, "svc", 1234)
				sc.On("Create", ctx, s, metav1.CreateOptions{}).Return(nil, errors.New("service create fail"))

				return cm, d, nil
			},
			expectError:      true,
			expectedErrorMsg: "service create fail",
		},
		{
			name:           "failed due to service watch error",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil)

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockDeploymentWatchResult, nil)
				go func() {
					mockDeploymentWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &appsv1.Deployment{
							ObjectMeta: metav1.ObjectMeta{Name: "svc"},
							Status: appsv1.DeploymentStatus{
								ReadyReplicas: 3,
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: int32Ptr(3),
							},
						},
					}
				}()

				s := getTestingService(corev1.ServiceTypeLoadBalancer, "svc", 1234)
				sc.On("Create", ctx, s, metav1.CreateOptions{}).Return(s, nil).Run(func(args mock.Arguments) {
					cancel()
				})

				sc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(nil, errors.New("service watch fail"))

				return cm, d, s
			},
			expectError:      true,
			expectedErrorMsg: "service watch fail",
		},
		{
			name:           "failed due to deployment added object not being Deployment",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil)

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockDeploymentWatchResult, nil)
				go func() {
					mockDeploymentWatchResult.Events <- watch.Event{
						Type: watch.Added,
						Object: &corev1.Service{
							ObjectMeta: metav1.ObjectMeta{Name: "svc"},
						},
					}
				}()

				return cm, d, nil
			},
			expectError:      true,
			expectedErrorMsg: "expected Deployment object, but got something else",
		},
		{
			name:           "failed due to deployment watch deleted event",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil)

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockDeploymentWatchResult, nil)
				go func() {
					mockDeploymentWatchResult.Events <- watch.Event{
						Type: watch.Deleted,
					}
				}()

				return cm, d, nil
			},
			expectError:      true,
			expectedErrorMsg: "watching deployment did not result to addition and modification",
		},
		{
			name:           "failed due to context being done during deployment watching",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil).Run(func(args mock.Arguments) {
					cancel()
				})

				mockDeploymentWatchResult := &MockWatchResult{
					Events: make(chan watch.Event),
				}
				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(mockDeploymentWatchResult, nil)

				return cm, d, nil
			},
			expectError:      true,
			expectedErrorMsg: "context canceled or timed out when waiting on deployment to be ready",
		},
		{
			name:           "failed due to deployment watching failures",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(d, nil).Run(func(args mock.Arguments) {
					cancel()
				})

				dc.On("Watch", ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=svc",
				}).Return(nil, errors.New("deployment watch fail"))

				return cm, d, nil
			},
			expectError:      true,
			expectedErrorMsg: "deployment watch fail",
		},
		{
			name:           "failed due to deployment creation failures",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := getTestingDeployment("svc")
				dc.On("Create", ctx, d, metav1.CreateOptions{}).Return(nil, errors.New("dc fail"))
				cmc.On("Delete", ctx, "svc-main-conf", metav1.DeleteOptions{}).Return(nil)
				cmc.On("Delete", ctx, "svc-test-php", metav1.DeleteOptions{}).Return(nil)

				return cm, d, nil
			},
			expectError:      true,
			expectedErrorMsg: "dc fail",
		},
		{
			name:           "failed due to second config map creation failures",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := getTestingConfigMaps("svc")
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], errors.New("cm1 fail"))
				cmc.On("Delete", ctx, "svc-main-conf", metav1.DeleteOptions{}).Return(nil)

				return cm, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to create configMap svc-test-php: cm1 fail",
		},
		{
			name:           "failed due to first config map creation failures",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/main.conf":   "main: data",
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					strings.Repeat("test012345", 26): "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					strings.Repeat("test012345", 26): "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: "svc-test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test01234",
					},
					Data: map[string]string{
						"main.conf": "main: data",
					},
				}
				cmc.On("Create", ctx, cm, metav1.CreateOptions{}).Return(cm, errors.New("cm0 fail"))
				return nil, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to create configMap svc-test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test012345test01234: cm0 fail",
		},
		{
			name:           "failed due to not found first conf",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			fs: map[string]string{
				"/tmp/wst/my_test.php": "<?php echo 1; ?>",
			},
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				return nil, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to read file at /tmp/wst/main.conf: open /tmp/wst/main.conf: file does not exist",
		},
		{
			name:           "failed due to env and workspace maps mismatch",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    false,
			envStartPort:   8080,
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   "mysvc",
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				ContainerConfig: &containers.ContainerConfig{
					ImageName:        "wst",
					ImageTag:         "test",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
				EnvironmentConfigPaths: map[string]string{
					"my_main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				fnd.On("DryRun").Return(false)
				return nil, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "environment path not found for main_conf",
		},
		{
			name:           "failed due to missing container config",
			namespace:      "wt",
			kubeconfigPath: "/home/kubeconfig/config.yml",
			useFullName:    true,
			envStartPort:   8080,
			ss: &environment.ServiceSettings{
				Name:       "svc",
				FullName:   strings.Repeat("a", 260),
				Port:       8080,
				ServerPort: 1234,
				Public:     true,
				EnvironmentConfigPaths: map[string]string{
					"my_main_conf": "/etc/main.conf",
				},
				EnvironmentScriptPaths: map[string]string{
					"test_php": "/www/test.php",
				},
				WorkspaceConfigPaths: map[string]string{
					"main_conf": "/tmp/wst/main.conf",
				},
				WorkspaceScriptPaths: map[string]string{
					"test_php": "/tmp/wst/my_test.php",
				},
			},
			cmd: &environment.Command{
				Name: "php",
				Args: []string{"test.php", "run"},
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				cancel context.CancelFunc,
				fnd *appMocks.MockFoundation,
				cmc *k8sClientMocks.MockConfigMapClient,
				dc *k8sClientMocks.MockDeploymentClient,
				sc *k8sClientMocks.MockServiceClient,
			) ([]*corev1.ConfigMap, *appsv1.Deployment, *corev1.Service) {
				return nil, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "container config is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
			if tt.fs != nil {
				mockFs := afero.NewMemMapFs()
				for fn, fd := range tt.fs {
					err := afero.WriteFile(mockFs, fn, []byte(fd), 0644)
					assert.NoError(t, err)
				}
				fndMock.On("Fs").Return(mockFs)
			}
			cmc := k8sClientMocks.NewMockConfigMapClient(t)
			dc := k8sClientMocks.NewMockDeploymentClient(t)
			sc := k8sClientMocks.NewMockServiceClient(t)
			var ctx context.Context
			var cancel context.CancelFunc
			if tt.contextSetup == nil {
				ctx, cancel = context.WithCancel(context.Background())
			} else {
				ctx, cancel = tt.contextSetup()
			}
			e := &kubernetesEnvironment{
				ContainerEnvironment: environment.ContainerEnvironment{
					CommonEnvironment: environment.CommonEnvironment{
						Fnd: fndMock,
						Ports: environment.Ports{
							Start: tt.envStartPort,
							Used:  tt.envStartPort,
							End:   tt.envStartPort + 100,
						},
					},
					Registry: environment.ContainerRegistry{},
				},
				kubeconfigPath:   tt.kubeconfigPath,
				namespace:        tt.namespace,
				useFullName:      tt.useFullName,
				deploymentClient: dc,
				configMapClient:  cmc,
				serviceClient:    sc,
				tasks:            make(map[string]*kubernetesTask),
			}

			cm, d, s := tt.setupMocks(t, ctx, cancel, fndMock, cmc, dc, sc)

			got, err := e.RunTask(ctx, tt.ss, tt.cmd)

			if tt.expectError {
				require.Nil(t, got)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				actualTask, ok := got.(*kubernetesTask)
				require.True(t, ok)
				assert.Equal(t, tt.getExpectedTask(cm, d, s), actualTask)
			}
		})
	}
}

func Test_kubernetesEnvironment_ExecTaskCommand(t *testing.T) {
	env := &kubernetesEnvironment{}
	ctx := context.Background()
	ss := &environment.ServiceSettings{
		Name:       "svc",
		FullName:   "mysvc",
		Port:       8080,
		ServerPort: 1234,
	}
	cmd := &environment.Command{Name: "test"}
	tsk := &kubernetesTask{}
	err := env.ExecTaskCommand(ctx, ss, tsk, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executing command is not currently supported in Kubernetes environment")
}

func Test_kubernetesEnvironment_ExecTaskSignal(t *testing.T) {
	env := &kubernetesEnvironment{}
	ctx := context.Background()
	ss := &environment.ServiceSettings{
		Name:       "svc",
		FullName:   "mysvc",
		Port:       8080,
		ServerPort: 1234,
	}
	tsk := &kubernetesTask{}
	err := env.ExecTaskSignal(ctx, ss, tsk, os.Interrupt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executing signal is not currently supported in Kubernetes environment")
}

type pullReaderCloser struct {
	msg string
	err string
}

func (b *pullReaderCloser) Read(p []byte) (n int, err error) {
	if len(b.err) > 0 {
		return 0, errors.New(b.err)
	}
	if len(b.msg) > 0 {
		n = copy(p, b.msg)
		b.msg = b.msg[n:]
		return n, nil
	}
	return 0, io.EOF
}

func (b *pullReaderCloser) Close() error {
	return nil
}

type invalidTask struct{}

func (t *invalidTask) Pid() int {
	return 1
}

func (t *invalidTask) Id() string {
	return ""
}

func (t *invalidTask) Executable() string {
	return ""
}

func (t *invalidTask) Name() string {
	return ""
}

func (t *invalidTask) Type() providers.Type {
	return providers.KubernetesType
}

func (t *invalidTask) PublicUrl() string {
	return ""
}

func (t *invalidTask) PrivateUrl() string {
	return ""
}

func Test_kubernetesEnvironment_Output(t *testing.T) {
	tests := []struct {
		name       string
		outputType output.Type
		target     task.Task
		setupMocks func(
			*testing.T,
			context.Context,
			*appMocks.MockFoundation,
			*k8sClientMocks.MockPodClient,
		)
		expectedLogData  string
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:       "successful output for any type",
			outputType: output.Any,
			target: &kubernetesTask{
				serviceName: "sn1",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				pc *k8sClientMocks.MockPodClient,
			) {
				fnd.On("DryRun").Return(false)
				p := corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "p1",
					},
				}
				pl := &corev1.PodList{Items: []corev1.Pod{p}}
				reader := &pullReaderCloser{
					msg: "data",
				}
				pc.On("List", ctx, metav1.ListOptions{
					LabelSelector: "app=sn1",
				}).Return(pl, nil)
				pc.On("StreamLogs", ctx, "p1", &corev1.PodLogOptions{}).Return(reader, nil)
			},
			expectedLogData: "data",
		},
		{
			name:       "successful output for dry run",
			outputType: output.Any,
			target: &kubernetesTask{
				serviceName: "sn1",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				pc *k8sClientMocks.MockPodClient,
			) {
				fnd.On("DryRun").Return(true)
			},
			expectedLogData: "",
		},
		{
			name:       "successful output when combined reader already set",
			outputType: output.Any,
			target: &kubernetesTask{
				serviceName:  "sn1",
				outputReader: &CombinedReader{readers: []io.ReadCloser{&app.DummyReaderCloser{}}},
			},
			expectedLogData: "",
		},
		{
			name:       "failed output due to failed log streaming",
			outputType: output.Any,
			target: &kubernetesTask{
				serviceName: "sn1",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				pc *k8sClientMocks.MockPodClient,
			) {
				fnd.On("DryRun").Return(false)
				p := corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "p1",
					},
				}
				pl := &corev1.PodList{Items: []corev1.Pod{p}}
				pc.On("List", ctx, metav1.ListOptions{
					LabelSelector: "app=sn1",
				}).Return(pl, nil)
				pc.On("StreamLogs", ctx, "p1", &corev1.PodLogOptions{}).Return(nil, errors.New("stream fail"))
			},
			expectError:      true,
			expectedErrorMsg: "error in opening stream: stream fail",
		},
		{
			name:       "failed output due to failed listing of pods",
			outputType: output.Any,
			target: &kubernetesTask{
				serviceName: "sn1",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fnd *appMocks.MockFoundation,
				pc *k8sClientMocks.MockPodClient,
			) {
				fnd.On("DryRun").Return(false)
				pc.On("List", ctx, metav1.ListOptions{
					LabelSelector: "app=sn1",
				}).Return(nil, errors.New("pod listing fail"))
			},
			expectError:      true,
			expectedErrorMsg: "failed to list pods: pod listing fail",
		},
		{
			name:             "failed output due to invalid task type",
			outputType:       output.Any,
			target:           &invalidTask{},
			expectError:      true,
			expectedErrorMsg: "task in not a Kubernetes task",
		},
		{
			name:             "failed output due to unsupported output type",
			outputType:       output.Stderr,
			target:           &invalidTask{},
			expectError:      true,
			expectedErrorMsg: "only any output type is supported by Kubernetes environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
			podClientMock := k8sClientMocks.NewMockPodClient(t)
			ctx := context.Background()
			e := &kubernetesEnvironment{
				ContainerEnvironment: environment.ContainerEnvironment{
					CommonEnvironment: environment.CommonEnvironment{
						Fnd: fndMock,
					},
					Registry: environment.ContainerRegistry{},
				},
				podClient: podClientMock,
			}

			if tt.setupMocks != nil {
				tt.setupMocks(t, ctx, fndMock, podClientMock)
			}
			actualReader, err := e.Output(ctx, tt.target, tt.outputType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				buf := new(strings.Builder)
				_, err := io.Copy(buf, actualReader)
				require.Nil(t, err)
				assert.Equal(t, tt.expectedLogData, buf.String())
			}
		})
	}
}

func Test_kubernetesEnvironment_RootPath(t *testing.T) {
	env := &kubernetesEnvironment{}
	assert.Equal(t, "", env.RootPath("/www/ws"))
}

func Test_kubernetesEnvironment_ServiceAddress(t *testing.T) {
	env := &kubernetesEnvironment{}
	assert.Equal(t, "svc", env.ServiceAddress("svc", 1234))
}

func getTestTask() *kubernetesTask {
	return &kubernetesTask{
		deployment:        nil,
		service:           nil,
		configMaps:        nil,
		executable:        "epk",
		serviceName:       "kubes",
		servicePublicUrl:  "http://localhost:1234",
		servicePrivateUrl: "http://kubes:8080",
		deploymentReady:   true,
	}
}

func Test_kubernetesTask_Id(t *testing.T) {
	assert.Equal(t, "kubes", getTestTask().Id())
}

func Test_kubernetesTask_Name(t *testing.T) {
	assert.Equal(t, "kubes", getTestTask().Name())
}

func Test_kubernetesTask_Executable(t *testing.T) {
	assert.Equal(t, "epk", getTestTask().Executable())
}

func Test_kubernetesTask_Pid(t *testing.T) {
	assert.Equal(t, 1, getTestTask().Pid())
}

func Test_kubernetesTask_PrivateUrl(t *testing.T) {
	assert.Equal(t, "http://kubes:8080", getTestTask().PrivateUrl())
}

func Test_kubernetesTask_PublicUrl(t *testing.T) {
	assert.Equal(t, "http://localhost:1234", getTestTask().PublicUrl())
}

func Test_kubernetesTask_Type(t *testing.T) {
	assert.Equal(t, providers.KubernetesType, getTestTask().Type())
}
