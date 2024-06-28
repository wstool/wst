package kubernetes

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	k8sClientMocks "github.com/bukka/wst/mocks/generated/run/environments/environment/providers/kubernetes/clients"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/sandboxes/containers"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"os"
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
				Name:     "svc",
				FullName: "mysvc",
				Port:     1234,
				Public:   true,
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
				cm := []*corev1.ConfigMap{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "svc-main-conf",
						},
						Data: map[string]string{
							"main.conf": "main: data",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "svc-test-php",
						},
						Data: map[string]string{
							"my_test.php": "<?php echo 1; ?>",
						},
					},
				}
				cmc.On("Create", ctx, cm[0], metav1.CreateOptions{}).Return(cm[0], nil)
				cmc.On("Create", ctx, cm[1], metav1.CreateOptions{}).Return(cm[1], nil)
				d := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "svc",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(2), // Specify the number of replicas
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "svc",
							},
						},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app": "svc",
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
												Name:      "svc-main-conf-volume",
												MountPath: "/etc",
											},
											{
												Name:      "svc-test-php-volume",
												MountPath: "/www",
											},
										},
									},
								},
								Volumes: []corev1.Volume{
									{
										Name: "svc-main-conf-volume",
										VolumeSource: corev1.VolumeSource{
											ConfigMap: &corev1.ConfigMapVolumeSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: "svc-main-conf",
												},
											},
										},
									},
									{
										Name: "svc-test-php-volume",
										VolumeSource: corev1.VolumeSource{
											ConfigMap: &corev1.ConfigMapVolumeSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: "svc-test-php",
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

				s := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name: "svc",
					},
					Spec: corev1.ServiceSpec{
						Type: corev1.ServiceTypeLoadBalancer,
						Ports: []corev1.ServicePort{
							{
								Port:       1234,
								TargetPort: intstr.FromInt32(1234),
								Protocol:   corev1.ProtocolTCP,
							},
						},
						Selector: map[string]string{
							"app": "svc",
						},
					},
				}
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
					serviceName:       "svc",
					servicePublicUrl:  "http://10.0.0.1",
					servicePrivateUrl: "http://svc:1234",
					deploymentReady:   true,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
			mockFs := afero.NewMemMapFs()
			for fn, fd := range tt.fs {
				err := afero.WriteFile(mockFs, fn, []byte(fd), 0644)
				assert.NoError(t, err)
			}
			fndMock.On("Fs").Return(mockFs)
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
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				actualTask, ok := got.(*kubernetesTask)
				assert.True(t, ok)
				assert.Equal(t, tt.getExpectedTask(cm, d, s), actualTask)
			}
		})
	}
}

func Test_kubernetesEnvironment_ExecTaskCommand(t *testing.T) {
	env := &kubernetesEnvironment{}
	ctx := context.Background()
	ss := &environment.ServiceSettings{
		Name:     "svc",
		FullName: "mysvc",
		Port:     1234,
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
		Name:     "svc",
		FullName: "mysvc",
		Port:     1234,
	}
	tsk := &kubernetesTask{}
	err := env.ExecTaskSignal(ctx, ss, tsk, os.Interrupt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executing signal is not currently supported in Kubernetes environment")
}

func Test_kubernetesEnvironment_Output(t *testing.T) {
	// TODO: implement
}

func Test_kubernetesEnvironment_RootPath(t *testing.T) {
	env := &kubernetesEnvironment{}
	assert.Equal(t, "", env.RootPath("/www/ws"))
}

func getTestTask() *kubernetesTask {
	return &kubernetesTask{
		deployment:        nil,
		service:           nil,
		configMaps:        nil,
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
