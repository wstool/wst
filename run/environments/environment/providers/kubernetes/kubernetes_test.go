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
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	outputMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/output"
	k8sClientMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/providers/kubernetes/clients"
	resourcesMocks "github.com/wstool/wst/mocks/generated/run/resources"
	certificatesMocks "github.com/wstool/wst/mocks/generated/run/resources/certificates"
	scriptsMocks "github.com/wstool/wst/mocks/generated/run/resources/scripts"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/environments/environment/providers/kubernetes/clients"
	"github.com/wstool/wst/run/environments/task"
	"github.com/wstool/wst/run/resources"
	"github.com/wstool/wst/run/resources/certificates"
	"github.com/wstool/wst/run/resources/scripts"
	"github.com/wstool/wst/run/sandboxes/containers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
)

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	resourcesMaker := resourcesMocks.NewMockMaker(t)
	tests := []struct {
		name           string
		fnd            app.Foundation
		resourcesMaker resources.Maker
	}{
		{
			name:           "create maker",
			fnd:            fndMock,
			resourcesMaker: resourcesMaker,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateMaker(tt.fnd, resourcesMaker)
			maker, ok := got.(*kubernetesMaker)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, maker.Fnd)
			assert.Equal(t, tt.resourcesMaker, maker.ResourcesMaker)
		})
	}
}

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name       string
		config     *types.KubernetesEnvironment
		setupMocks func(*testing.T, *k8sClientMocks.MockMaker, *resourcesMocks.MockMaker, *types.KubernetesEnvironment) (
			*k8sClientMocks.MockConfigMapClient,
			*k8sClientMocks.MockDeploymentClient,
			*k8sClientMocks.MockPodClient,
			*k8sClientMocks.MockServiceClient,
			*resources.Resources,
		)
		getExpectedEnv func(
			fndMock *appMocks.MockFoundation,
			configMapClient *k8sClientMocks.MockConfigMapClient,
			deploymentClient *k8sClientMocks.MockDeploymentClient,
			podClient *k8sClientMocks.MockPodClient,
			serviceClient *k8sClientMocks.MockServiceClient,
			rscs *resources.Resources,
		) *kubernetesEnvironment
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful kubernetes environment maker creation with resources",
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
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{
						"k8s_ssl": {
							Certificate: "-----BEGIN CERTIFICATE-----\nk8s cert\n-----END CERTIFICATE-----",
							PrivateKey:  "-----BEGIN PRIVATE KEY-----\nk8s key\n-----END PRIVATE KEY-----",
						},
					},
					Scripts: map[string]types.Script{
						"k8s_script": {
							Content: "#!/bin/bash\necho 'Kubernetes init'",
							Path:    "/k8s/init.sh",
							Mode:    "0755",
						},
					},
				},
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, r *resourcesMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
				*resources.Resources,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				dc := k8sClientMocks.NewMockDeploymentClient(t)
				m.On("MakeDeploymentClient", config).Return(dc, nil)
				pc := k8sClientMocks.NewMockPodClient(t)
				m.On("MakePodClient", config).Return(pc, nil)
				sc := k8sClientMocks.NewMockServiceClient(t)
				m.On("MakeServiceClient", config).Return(sc, nil)

				expectedResources := &resources.Resources{
					Certificates: map[string]certificates.Certificate{
						"k8s_ssl": certificatesMocks.NewMockCertificate(t),
					},
					Scripts: map[string]scripts.Script{
						"k8s_script": scriptsMocks.NewMockScript(t),
					},
				}

				r.On("Make", types.Resources{
					Certificates: map[string]types.Certificate{
						"k8s_ssl": {
							Certificate: "-----BEGIN CERTIFICATE-----\nk8s cert\n-----END CERTIFICATE-----",
							PrivateKey:  "-----BEGIN PRIVATE KEY-----\nk8s key\n-----END PRIVATE KEY-----",
						},
					},
					Scripts: map[string]types.Script{
						"k8s_script": {
							Content: "#!/bin/bash\necho 'Kubernetes init'",
							Path:    "/k8s/init.sh",
							Mode:    "0755",
						},
					},
				}).Return(expectedResources, nil)

				return cmc, dc, pc, sc, expectedResources
			},
			getExpectedEnv: func(
				fndMock *appMocks.MockFoundation,
				configMapClient *k8sClientMocks.MockConfigMapClient,
				deploymentClient *k8sClientMocks.MockDeploymentClient,
				podClient *k8sClientMocks.MockPodClient,
				serviceClient *k8sClientMocks.MockServiceClient,
				rscs *resources.Resources,
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
							EnvResources: rscs,
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
					useUniqueName:    true,
					namespace:        "test",
					kubeconfigPath:   "/home/kubeconfig/config.yaml",
					tasks:            make(map[string]*kubernetesTask),
				}
			},
		},
		{
			name: "successful kubernetes environment maker creation with empty resources",
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
				Resources:  types.Resources{},
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, r *resourcesMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
				*resources.Resources,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				dc := k8sClientMocks.NewMockDeploymentClient(t)
				m.On("MakeDeploymentClient", config).Return(dc, nil)
				pc := k8sClientMocks.NewMockPodClient(t)
				m.On("MakePodClient", config).Return(pc, nil)
				sc := k8sClientMocks.NewMockServiceClient(t)
				m.On("MakeServiceClient", config).Return(sc, nil)

				expectedResources := &resources.Resources{}
				r.On("Make", types.Resources{}).Return(expectedResources, nil)

				return cmc, dc, pc, sc, expectedResources
			},
			getExpectedEnv: func(
				fndMock *appMocks.MockFoundation,
				configMapClient *k8sClientMocks.MockConfigMapClient,
				deploymentClient *k8sClientMocks.MockDeploymentClient,
				podClient *k8sClientMocks.MockPodClient,
				serviceClient *k8sClientMocks.MockServiceClient,
				rscs *resources.Resources,
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
							EnvResources: rscs,
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
					useUniqueName:    true,
					namespace:        "test",
					kubeconfigPath:   "/home/kubeconfig/config.yaml",
					tasks:            make(map[string]*kubernetesTask),
				}
			},
		},
		{
			name: "failed kubernetes environment maker creation due to resource failure",
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
				Resources: types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Content: "echo test",
							Mode:    "invalid_mode",
						},
					},
				},
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, r *resourcesMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
				*resources.Resources,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				dc := k8sClientMocks.NewMockDeploymentClient(t)
				m.On("MakeDeploymentClient", config).Return(dc, nil)
				pc := k8sClientMocks.NewMockPodClient(t)
				m.On("MakePodClient", config).Return(pc, nil)
				sc := k8sClientMocks.NewMockServiceClient(t)
				m.On("MakeServiceClient", config).Return(sc, nil)

				r.On("Make", types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Content: "echo test",
							Path:    "",
							Mode:    "invalid_mode",
						},
					},
				}).Return(nil, errors.New("resource creation failed"))

				return cmc, dc, pc, sc, nil
			},
			expectError:      true,
			expectedErrorMsg: "resource creation failed",
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
				Resources:  types.Resources{},
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, r *resourcesMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
				*resources.Resources,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				dc := k8sClientMocks.NewMockDeploymentClient(t)
				m.On("MakeDeploymentClient", config).Return(dc, nil)
				pc := k8sClientMocks.NewMockPodClient(t)
				m.On("MakePodClient", config).Return(pc, nil)
				m.On("MakeServiceClient", config).Return(nil, errors.New("failed sc"))
				// No resource mocking needed since service client fails first
				return cmc, dc, pc, nil, nil
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
				Resources:  types.Resources{},
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, r *resourcesMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
				*resources.Resources,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				dc := k8sClientMocks.NewMockDeploymentClient(t)
				m.On("MakeDeploymentClient", config).Return(dc, nil)
				m.On("MakePodClient", config).Return(nil, errors.New("failed sc"))
				// No resource mocking needed since pod client fails
				return cmc, dc, nil, nil, nil
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
				Resources:  types.Resources{},
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, r *resourcesMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
				*resources.Resources,
			) {
				cmc := k8sClientMocks.NewMockConfigMapClient(t)
				m.On("MakeConfigMapClient", config).Return(cmc, nil)
				m.On("MakeDeploymentClient", config).Return(nil, errors.New("failed dc"))
				// No resource mocking needed since deployment client fails
				return cmc, nil, nil, nil, nil
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
				Resources:  types.Resources{},
			},
			setupMocks: func(t *testing.T, m *k8sClientMocks.MockMaker, r *resourcesMocks.MockMaker, config *types.KubernetesEnvironment) (
				*k8sClientMocks.MockConfigMapClient,
				*k8sClientMocks.MockDeploymentClient,
				*k8sClientMocks.MockPodClient,
				*k8sClientMocks.MockServiceClient,
				*resources.Resources,
			) {
				m.On("MakeConfigMapClient", config).Return(nil, errors.New("failed cmc"))
				// No resource mocking needed since config map client fails first
				return nil, nil, nil, nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to create kubernetes client: failed cmc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			resourcesMaker := resourcesMocks.NewMockMaker(t)
			clientsMakerMock := k8sClientMocks.NewMockMaker(t)
			m := &kubernetesMaker{
				CommonMaker: &environment.CommonMaker{
					Fnd:            fndMock,
					ResourcesMaker: resourcesMaker,
				},
				clientsMaker: clientsMakerMock,
			}

			cmc, dc, pc, sc, expectedResources := tt.setupMocks(t, clientsMakerMock, resourcesMaker, tt.config)

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
				expectedEnv := tt.getExpectedEnv(fndMock, cmc, dc, pc, sc, expectedResources)
				assert.Equal(t, expectedEnv, actualEnv)

				// Assert that resources are properly set
				assert.Equal(t, expectedResources, actualEnv.Resources())
			}

			// Only assert resource expectations if they were set up
			if !tt.expectError || tt.expectedErrorMsg == "resource creation failed" {
				resourcesMaker.AssertExpectations(t)
			}
			clientsMakerMock.AssertExpectations(t)
		})
	}
}

func Test_kubernetesEnvironment_Init(t *testing.T) {
	env := &kubernetesEnvironment{}
	ctx := context.Background()
	assert.Nil(t, env.Init(ctx))
}

// MockWatchResult is a watch result backed by a channel pre-filled with events.
type MockWatchResult struct {
	Events   chan watch.Event
	stopOnce sync.Once
	Stopped  bool
}

func newMockWatchResult(events ...watch.Event) *MockWatchResult {
	ch := make(chan watch.Event, len(events)+8)
	for _, event := range events {
		ch <- event
	}
	return &MockWatchResult{Events: ch}
}

func newClosedMockWatchResult() *MockWatchResult {
	ch := make(chan watch.Event)
	close(ch)
	return &MockWatchResult{Events: ch}
}

func (m *MockWatchResult) Stop() {
	m.stopOnce.Do(func() {
		m.Stopped = true
	})
}

func (m *MockWatchResult) ResultChan() <-chan watch.Event {
	return m.Events
}

type fakeCloser struct {
	err    error
	closed int
}

func (c *fakeCloser) Close() error {
	c.closed++
	return c.err
}

type pullReaderCloser struct {
	msg    string
	err    string
	closed int
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
	b.closed++
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

func (t *invalidTask) PublicUrl(scheme string) string {
	return scheme
}

func (t *invalidTask) PrivateUrl(scheme string) string {
	return scheme
}

func notFoundError() error {
	return apierrors.NewNotFound(schema.GroupResource{Resource: "tests"}, "name")
}

func deploymentEvent(eventType watch.EventType, name string, ready int32) watch.Event {
	return watch.Event{
		Type: eventType,
		Object: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(1)},
			Status:     appsv1.DeploymentStatus{ReadyReplicas: ready},
		},
	}
}

func podEvent(eventType watch.EventType, pod *corev1.Pod) watch.Event {
	return watch.Event{Type: eventType, Object: pod}
}

func runningPod(name string, ready bool) *corev1.Pod {
	readyStatus := corev1.ConditionFalse
	if ready {
		readyStatus = corev1.ConditionTrue
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodScheduled, Status: corev1.ConditionTrue},
				{Type: corev1.PodReady, Status: readyStatus},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "c", State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
			},
		},
	}
}

func pendingPod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "c", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ContainerCreating"}}},
			},
		},
	}
}

func terminatedPod(name string, exitCode int32) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "c",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{ExitCode: exitCode, Reason: "Error", Message: "boom"},
					},
				},
			},
		},
	}
}

func waitingPod(name string, reason string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "c",
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason, Message: "msg"}},
				},
			},
		},
	}
}

func restartedPod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "c",
					RestartCount: 1,
					State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
					LastTerminationState: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{ExitCode: 2, Reason: "Error", Message: "died"},
					},
				},
			},
		},
	}
}

func failedPod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     corev1.PodStatus{Phase: corev1.PodFailed, Reason: "Evicted", Message: "no memory"},
	}
}

func getTestingConfigMaps(serviceName string) []*corev1.ConfigMap {
	return []*corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceName + "-configs-1",
			},
			Data: map[string]string{
				"main.conf": "main: data",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceName + "-scripts-1",
			},
			Data: map[string]string{
				"test.php": "<?php echo 1; ?>",
			},
		},
	}
}

func getTestingDeployment(serviceName string, command []string, args []string) *appsv1.Deployment {
	return &appsv1.Deployment{
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
							Image:   "wst:test",
							Command: command,
							Args:    args,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 1234,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      serviceName + "-configs-1-volume",
									MountPath: "/etc",
								},
								{
									Name:      serviceName + "-scripts-1-volume",
									MountPath: "/www",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: serviceName + "-configs-1-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: serviceName + "-configs-1",
									},
								},
							},
						},
						{
							Name: serviceName + "-scripts-1-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: serviceName + "-scripts-1",
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

func getTestingService(serviceName string, port int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
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

func getTestingServiceSettings(public bool) *environment.ServiceSettings {
	return &environment.ServiceSettings{
		Name:       "svc",
		FullName:   "mysvc",
		UniqueName: "i1-svc",
		Port:       8080,
		ServerPort: 1234,
		Public:     public,
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
	}
}

var testingFs = map[string]string{
	"/tmp/wst/main.conf":   "main: data",
	"/tmp/wst/my_test.php": "<?php echo 1; ?>",
}

// testEnv groups the environment under test with all its mocks.
type testEnv struct {
	env *kubernetesEnvironment
	fnd *appMocks.MockFoundation
	cmc *k8sClientMocks.MockConfigMapClient
	dc  *k8sClientMocks.MockDeploymentClient
	pc  *k8sClientMocks.MockPodClient
	sc  *k8sClientMocks.MockServiceClient
	om  *outputMocks.MockMaker
}

func newTestEnv(t *testing.T, fs map[string]string, tasks map[string]*kubernetesTask) *testEnv {
	fndMock := appMocks.NewMockFoundation(t)
	mockLogger := external.NewMockLogger()
	fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)
	if fs != nil {
		mockFs := afero.NewMemMapFs()
		for fn, fd := range fs {
			err := afero.WriteFile(mockFs, fn, []byte(fd), 0644)
			assert.NoError(t, err)
		}
		fndMock.On("Fs").Maybe().Return(mockFs)
	}
	te := &testEnv{
		fnd: fndMock,
		cmc: k8sClientMocks.NewMockConfigMapClient(t),
		dc:  k8sClientMocks.NewMockDeploymentClient(t),
		pc:  k8sClientMocks.NewMockPodClient(t),
		sc:  k8sClientMocks.NewMockServiceClient(t),
		om:  outputMocks.NewMockMaker(t),
	}
	if tasks == nil {
		tasks = make(map[string]*kubernetesTask)
	}
	te.env = &kubernetesEnvironment{
		ContainerEnvironment: environment.ContainerEnvironment{
			CommonEnvironment: environment.CommonEnvironment{
				Fnd:         fndMock,
				OutputMaker: te.om,
				Ports: environment.Ports{
					Start: 8080,
					Used:  8080,
					End:   8180,
				},
			},
			Registry: environment.ContainerRegistry{},
		},
		kubeconfigPath:   "/home/kubeconfig/config.yml",
		namespace:        "wt",
		useUniqueName:    true,
		deploymentClient: te.dc,
		configMapClient:  te.cmc,
		podClient:        te.pc,
		serviceClient:    te.sc,
		tasks:            tasks,
	}
	return te
}

func Test_kubernetesEnvironment_Destroy(t *testing.T) {
	deleteOpts := metav1.DeleteOptions{}
	dryRunDeleteOpts := metav1.DeleteOptions{DryRun: []string{metav1.DryRunAll}}
	makeTask := func(id string, collector output.Collector, portForward io.Closer) *kubernetesTask {
		return &kubernetesTask{
			serviceName: "t" + id,
			configMaps: []*corev1.ConfigMap{
				{ObjectMeta: metav1.ObjectMeta{Name: "c" + id + "1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "c" + id + "2"}},
			},
			deployment:      &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "d" + id}},
			service:         &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "s" + id}},
			outputCollector: collector,
			portForward:     portForward,
		}
	}
	tests := []struct {
		name             string
		dryRun           bool
		setupMocks       func(t *testing.T, ctx context.Context, te *testEnv) map[string]*kubernetesTask
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful kubernetes env destroying",
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*kubernetesTask {
				collector := outputMocks.NewMockCollector(t)
				collector.On("Close").Return(nil)
				closer := &fakeCloser{}
				te.cmc.On("Delete", ctx, "c11", deleteOpts).Return(nil)
				te.cmc.On("Delete", ctx, "c12", deleteOpts).Return(nil)
				te.cmc.On("Delete", ctx, "c21", deleteOpts).Return(nil)
				te.cmc.On("Delete", ctx, "c22", deleteOpts).Return(nil)
				te.dc.On("Delete", ctx, "d1", deleteOpts).Return(nil)
				te.sc.On("Delete", ctx, "s1", deleteOpts).Return(nil)
				t2 := makeTask("2", nil, nil)
				t2.deployment = nil
				t2.service = nil
				return map[string]*kubernetesTask{
					"t1": makeTask("1", collector, closer),
					"t2": t2,
				}
			},
		},
		{
			name:   "successful kubernetes env destroying in dry run",
			dryRun: true,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*kubernetesTask {
				te.cmc.On("Delete", ctx, "c11", dryRunDeleteOpts).Return(nil)
				te.cmc.On("Delete", ctx, "c12", dryRunDeleteOpts).Return(nil)
				te.dc.On("Delete", ctx, "d1", dryRunDeleteOpts).Return(nil)
				te.sc.On("Delete", ctx, "s1", dryRunDeleteOpts).Return(nil)
				return map[string]*kubernetesTask{"t1": makeTask("1", nil, nil)}
			},
		},
		{
			name: "successful kubernetes env destroying when objects are already gone",
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*kubernetesTask {
				te.cmc.On("Delete", ctx, "c11", deleteOpts).Return(notFoundError())
				te.cmc.On("Delete", ctx, "c12", deleteOpts).Return(nil)
				te.dc.On("Delete", ctx, "d1", deleteOpts).Return(notFoundError())
				te.sc.On("Delete", ctx, "s1", deleteOpts).Return(notFoundError())
				return map[string]*kubernetesTask{"t1": makeTask("1", nil, nil)}
			},
		},
		{
			name: "failed kubernetes env destroying due to all delete errors",
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*kubernetesTask {
				collector := outputMocks.NewMockCollector(t)
				collector.On("Close").Return(errors.New("close err"))
				closer := &fakeCloser{err: errors.New("pf err")}
				te.cmc.On("Delete", ctx, "c11", deleteOpts).Return(errors.New("cm err"))
				te.cmc.On("Delete", ctx, "c12", deleteOpts).Return(nil)
				te.dc.On("Delete", ctx, "d1", deleteOpts).Return(errors.New("dep err"))
				te.sc.On("Delete", ctx, "s1", deleteOpts).Return(errors.New("svc err"))
				return map[string]*kubernetesTask{"t1": makeTask("1", collector, closer)}
			},
			expectError:      true,
			expectedErrorMsg: "failed to destroy kubernetes environment",
		},
		{
			name: "failed kubernetes env destroying due to config map delete error",
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) map[string]*kubernetesTask {
				te.cmc.On("Delete", ctx, "c11", deleteOpts).Return(nil)
				te.cmc.On("Delete", ctx, "c12", deleteOpts).Return(errors.New("cm err"))
				te.dc.On("Delete", ctx, "d1", deleteOpts).Return(nil)
				te.sc.On("Delete", ctx, "s1", deleteOpts).Return(nil)
				return map[string]*kubernetesTask{"t1": makeTask("1", nil, nil)}
			},
			expectError:      true,
			expectedErrorMsg: "failed to destroy kubernetes environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			te := newTestEnv(t, nil, nil)
			te.fnd.On("DryRun").Return(tt.dryRun)
			tasks := tt.setupMocks(t, ctx, te)
			te.env.tasks = tasks

			err := te.env.Destroy(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Empty(t, te.env.tasks)
			for _, tsk := range tasks {
				assert.Nil(t, tsk.outputCollector)
				assert.Nil(t, tsk.portForward)
			}
		})
	}
}

func Test_kubernetesEnvironment_RunTask(t *testing.T) {
	createOpts := metav1.CreateOptions{}
	deleteOpts := metav1.DeleteOptions{}
	dryRunCreateOpts := metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}}
	phpCmd := &environment.Command{Name: "php", Args: []string{"test.php", "run"}}
	logOpts := &corev1.PodLogOptions{Follow: true}
	deploymentListOpts := metav1.ListOptions{FieldSelector: "metadata.name=i1-svc"}
	podListOpts := metav1.ListOptions{LabelSelector: "app=i1-svc"}

	// expectConfigMaps sets up the successful creation of the testing config maps.
	expectConfigMaps := func(ctx context.Context, te *testEnv, opts metav1.CreateOptions) []*corev1.ConfigMap {
		cm := getTestingConfigMaps("i1-svc")
		te.cmc.On("Create", ctx, cm[0], opts).Return(cm[0], nil)
		te.cmc.On("Create", ctx, cm[1], opts).Return(cm[1], nil)
		return cm
	}
	// expectDeployment sets up the successful creation of the testing deployment.
	expectDeployment := func(ctx context.Context, te *testEnv, opts metav1.CreateOptions, cmd *environment.Command) *appsv1.Deployment {
		var command, args []string
		if cmd != nil {
			command = []string{cmd.Name}
			args = cmd.Args
		}
		d := getTestingDeployment("i1-svc", command, args)
		te.dc.On("Create", ctx, d, opts).Return(d, nil)
		return d
	}
	// expectWatches sets up the deployment and pod watches with the given events.
	expectWatches := func(ctx context.Context, te *testEnv, dw, pw *MockWatchResult) {
		te.dc.On("Watch", ctx, deploymentListOpts).Return(dw, nil)
		te.pc.On("Watch", ctx, podListOpts).Return(pw, nil)
	}
	expectReadyWatches := func(ctx context.Context, te *testEnv) {
		expectWatches(ctx, te,
			newMockWatchResult(deploymentEvent(watch.Added, "i1-svc", 1)),
			newMockWatchResult(podEvent(watch.Added, runningPod("p1", true))),
		)
	}
	expectService := func(ctx context.Context, te *testEnv, opts metav1.CreateOptions) *corev1.Service {
		s := getTestingService("i1-svc", 1234)
		te.sc.On("Create", ctx, s, opts).Return(s, nil)
		return s
	}
	expectOutput := func(t *testing.T, te *testEnv) (*outputMocks.MockCollector, *pullReaderCloser) {
		logs := &pullReaderCloser{msg: "data"}
		te.pc.On("StreamLogs", mock.Anything, "p1", logOpts).Return(logs, nil)
		collector := outputMocks.NewMockCollector(t)
		collector.On("Start", logs, &app.DummyReaderCloser{}).Return(nil)
		te.om.On("MakeCollector", "i1-svc").Return(collector)
		return collector, logs
	}

	tests := []struct {
		name            string
		dryRun          bool
		fs              map[string]string
		ss              *environment.ServiceSettings
		cmd             *environment.Command
		setupMocks      func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask
		expectError     bool
		expectedErrMsg  string
		expectTaskSaved bool
	}{
		{
			name: "successful private run",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				cm := expectConfigMaps(ctx, te, createOpts)
				d := expectDeployment(ctx, te, createOpts, phpCmd)
				expectReadyWatches(ctx, te)
				s := expectService(ctx, te, createOpts)
				collector, _ := expectOutput(t, te)
				return &kubernetesTask{
					configMaps:        cm,
					deployment:        d,
					service:           s,
					outputCollector:   collector,
					podName:           "p1",
					executable:        "php",
					serviceName:       "i1-svc",
					servicePrivateUrl: "://i1-svc:1234",
					deploymentReady:   true,
				}
			},
			expectTaskSaved: true,
		},
		{
			name: "successful public run with port forwarding",
			fs:   testingFs,
			ss:   getTestingServiceSettings(true),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				cm := expectConfigMaps(ctx, te, createOpts)
				d := expectDeployment(ctx, te, createOpts, phpCmd)
				expectReadyWatches(ctx, te)
				s := expectService(ctx, te, createOpts)
				collector, _ := expectOutput(t, te)
				closer := &fakeCloser{}
				te.pc.On("PortForward", ctx, "p1", int32(8080), int32(1234)).Return(closer, nil)
				return &kubernetesTask{
					configMaps:        cm,
					deployment:        d,
					service:           s,
					outputCollector:   collector,
					portForward:       closer,
					podName:           "p1",
					executable:        "php",
					serviceName:       "i1-svc",
					servicePublicUrl:  "://127.0.0.1:8080",
					servicePrivateUrl: "://i1-svc:1234",
					deploymentReady:   true,
				}
			},
			expectTaskSaved: true,
		},
		{
			name:   "successful public dry run",
			dryRun: true,
			fs:     testingFs,
			ss:     getTestingServiceSettings(true),
			cmd:    phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				cm := expectConfigMaps(ctx, te, dryRunCreateOpts)
				d := expectDeployment(ctx, te, dryRunCreateOpts, phpCmd)
				s := expectService(ctx, te, dryRunCreateOpts)
				return &kubernetesTask{
					configMaps:        cm,
					deployment:        d,
					service:           s,
					podName:           "i1-svc",
					executable:        "php",
					serviceName:       "i1-svc",
					servicePublicUrl:  "://127.0.0.1:8080",
					servicePrivateUrl: "://i1-svc:1234",
					deploymentReady:   true,
				}
			},
			expectTaskSaved: true,
		},
		{
			name: "successful run without command uses service name as executable",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  nil,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				cm := expectConfigMaps(ctx, te, createOpts)
				d := expectDeployment(ctx, te, createOpts, nil)
				expectReadyWatches(ctx, te)
				s := expectService(ctx, te, createOpts)
				collector, _ := expectOutput(t, te)
				return &kubernetesTask{
					configMaps:        cm,
					deployment:        d,
					service:           s,
					outputCollector:   collector,
					podName:           "p1",
					executable:        "svc",
					serviceName:       "i1-svc",
					servicePrivateUrl: "://i1-svc:1234",
					deploymentReady:   true,
				}
			},
			expectTaskSaved: true,
		},
		{
			name: "successful run with gradual readiness and ignored events",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				cm := expectConfigMaps(ctx, te, createOpts)
				d := expectDeployment(ctx, te, createOpts, phpCmd)
				expectWatches(ctx, te,
					newMockWatchResult(
						deploymentEvent(watch.Added, "i1-svc", 0),
						watch.Event{Type: watch.Bookmark, Object: &appsv1.Deployment{}},
						deploymentEvent(watch.Modified, "i1-svc", 1),
					),
					newMockWatchResult(
						podEvent(watch.Added, pendingPod("p1")),
						podEvent(watch.Modified, runningPod("p1", false)),
						watch.Event{Type: watch.Bookmark, Object: &corev1.Pod{}},
						watch.Event{Type: watch.Deleted, Object: &runtime.Unknown{}},
						podEvent(watch.Modified, runningPod("p1", true)),
					),
				)
				s := expectService(ctx, te, createOpts)
				collector, _ := expectOutput(t, te)
				return &kubernetesTask{
					configMaps:        cm,
					deployment:        d,
					service:           s,
					outputCollector:   collector,
					podName:           "p1",
					executable:        "php",
					serviceName:       "i1-svc",
					servicePrivateUrl: "://i1-svc:1234",
					deploymentReady:   true,
				}
			},
			expectTaskSaved: true,
		},
		{
			name: "successful run when ready pod is replaced",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				cm := expectConfigMaps(ctx, te, createOpts)
				d := expectDeployment(ctx, te, createOpts, phpCmd)
				pw := newMockWatchResult(
					podEvent(watch.Added, runningPod("p1", true)),
					podEvent(watch.Deleted, runningPod("p1", true)),
					podEvent(watch.Deleted, runningPod("other", true)),
					podEvent(watch.Added, runningPod("p2", true)),
				)
				dw := newMockWatchResult()
				// The deployment becomes ready only after all pod events are processed.
				go func() {
					for len(pw.Events) > 0 {
						time.Sleep(time.Millisecond)
					}
					dw.Events <- deploymentEvent(watch.Added, "i1-svc", 1)
				}()
				expectWatches(ctx, te, dw, pw)
				s := expectService(ctx, te, createOpts)
				logs := &pullReaderCloser{msg: "data"}
				te.pc.On("StreamLogs", mock.Anything, "p2", logOpts).Return(logs, nil)
				collector := outputMocks.NewMockCollector(t)
				collector.On("Start", logs, &app.DummyReaderCloser{}).Return(nil)
				te.om.On("MakeCollector", "i1-svc").Return(collector)
				return &kubernetesTask{
					configMaps:        cm,
					deployment:        d,
					service:           s,
					outputCollector:   collector,
					podName:           "p2",
					executable:        "php",
					serviceName:       "i1-svc",
					servicePrivateUrl: "://i1-svc:1234",
					deploymentReady:   true,
				}
			},
			expectTaskSaved: true,
		},
		{
			name: "failed run due to missing container config",
			fs:   testingFs,
			ss: func() *environment.ServiceSettings {
				ss := getTestingServiceSettings(false)
				ss.ContainerConfig = nil
				return ss
			}(),
			cmd: phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				return nil
			},
			expectError:    true,
			expectedErrMsg: "container config is not set",
		},
		{
			name: "failed run due to missing config environment path",
			fs:   testingFs,
			ss: func() *environment.ServiceSettings {
				ss := getTestingServiceSettings(false)
				ss.EnvironmentConfigPaths = map[string]string{}
				return ss
			}(),
			cmd: phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				return nil
			},
			expectError:    true,
			expectedErrMsg: "environment path not found for main_conf",
		},
		{
			name: "failed run due to unreadable config file",
			fs:   map[string]string{},
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to read file at /tmp/wst/main.conf",
		},
		{
			name: "failed run due to missing script environment path removes config maps",
			fs:   testingFs,
			ss: func() *environment.ServiceSettings {
				ss := getTestingServiceSettings(false)
				ss.EnvironmentScriptPaths = map[string]string{}
				return ss
			}(),
			cmd: phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				cm := getTestingConfigMaps("i1-svc")
				te.cmc.On("Create", ctx, cm[0], createOpts).Return(cm[0], nil)
				te.cmc.On("Delete", ctx, "i1-svc-configs-1", deleteOpts).Return(nil)
				return nil
			},
			expectError:    true,
			expectedErrMsg: "environment path not found for test_php",
		},
		{
			name: "failed run due to config map creation error",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				cm := getTestingConfigMaps("i1-svc")
				te.cmc.On("Create", ctx, cm[0], createOpts).Return(nil, errors.New("cm err"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "failed to create configMap i1-svc-configs-1: cm err",
		},
		{
			name: "failed run due to deployment creation error removes config maps",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				d := getTestingDeployment("i1-svc", []string{"php"}, []string{"test.php", "run"})
				te.dc.On("Create", ctx, d, createOpts).Return(nil, errors.New("dep err"))
				te.cmc.On("Delete", ctx, "i1-svc-configs-1", deleteOpts).Return(nil)
				te.cmc.On("Delete", ctx, "i1-svc-scripts-1", deleteOpts).Return(nil)
				return nil
			},
			expectError:    true,
			expectedErrMsg: "dep err",
		},
		{
			name: "failed run due to deployment watch error",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				te.dc.On("Watch", ctx, deploymentListOpts).Return(nil, errors.New("watch err"))
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed to watch deployment i1-svc: watch err",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to pod watch error",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				dw := newMockWatchResult()
				te.dc.On("Watch", ctx, deploymentListOpts).Return(dw, nil)
				te.pc.On("Watch", ctx, podListOpts).Return(nil, errors.New("pod watch err"))
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed to watch pods of deployment i1-svc: pod watch err",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to closed deployment watch",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectWatches(ctx, te, newClosedMockWatchResult(), newMockWatchResult())
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "watching deployment i1-svc ended unexpectedly",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to closed pod watch",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectWatches(ctx, te, newMockWatchResult(), newClosedMockWatchResult())
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "watching pods of deployment i1-svc ended unexpectedly",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to invalid deployment event object",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectWatches(ctx, te,
					newMockWatchResult(watch.Event{Type: watch.Added, Object: &corev1.Pod{}}),
					newMockWatchResult(),
				)
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "expected Deployment object, but got something else",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to deleted deployment",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectWatches(ctx, te,
					newMockWatchResult(deploymentEvent(watch.Deleted, "i1-svc", 0)),
					newMockWatchResult(),
				)
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "watching deployment did not result to addition and modification",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to invalid pod event object",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectWatches(ctx, te,
					newMockWatchResult(),
					newMockWatchResult(watch.Event{Type: watch.Added, Object: &appsv1.Deployment{}}),
				)
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "expected Pod object, but got something else",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to terminated container",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectWatches(ctx, te,
					newMockWatchResult(),
					newMockWatchResult(podEvent(watch.Modified, terminatedPod("p1", 1))),
				)
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "pod p1 of deployment i1-svc failed: container c terminated with exit code 1: Error: boom",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to pod watch error event",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectWatches(ctx, te,
					newMockWatchResult(),
					newMockWatchResult(watch.Event{Type: watch.Error, Object: &metav1.Status{}}),
				)
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "watching pods of deployment i1-svc resulted in error",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to context cancellation",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectWatches(ctx, te, newMockWatchResult(), newMockWatchResult())
				cancel()
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "context canceled or timed out when waiting on deployment to be ready",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to service creation error",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectReadyWatches(ctx, te)
				s := getTestingService("i1-svc", 1234)
				te.sc.On("Create", ctx, s, createOpts).Return(nil, errors.New("svc err"))
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed to create service: svc err",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to log stream error",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectReadyWatches(ctx, te)
				expectService(ctx, te, createOpts)
				te.pc.On("StreamLogs", mock.Anything, "p1", logOpts).Return(nil, errors.New("logs err"))
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed to open log stream of pod p1: logs err",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to output collector start error",
			fs:   testingFs,
			ss:   getTestingServiceSettings(false),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectReadyWatches(ctx, te)
				expectService(ctx, te, createOpts)
				logs := &pullReaderCloser{msg: "data"}
				te.pc.On("StreamLogs", mock.Anything, "p1", logOpts).Return(logs, nil)
				collector := outputMocks.NewMockCollector(t)
				collector.On("Start", logs, &app.DummyReaderCloser{}).Return(errors.New("start err")).Run(
					func(args mock.Arguments) {
						t.Cleanup(func() {
							assert.Equal(t, 1, logs.closed)
						})
					})
				te.om.On("MakeCollector", "i1-svc").Return(collector)
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed to start output collection of pod p1: start err",
			expectTaskSaved: true,
		},
		{
			name: "failed run due to port forwarding error",
			fs:   testingFs,
			ss:   getTestingServiceSettings(true),
			cmd:  phpCmd,
			setupMocks: func(t *testing.T, ctx context.Context, cancel context.CancelFunc, te *testEnv) *kubernetesTask {
				expectConfigMaps(ctx, te, createOpts)
				expectDeployment(ctx, te, createOpts, phpCmd)
				expectReadyWatches(ctx, te)
				expectService(ctx, te, createOpts)
				expectOutput(t, te)
				te.pc.On("PortForward", ctx, "p1", int32(8080), int32(1234)).Return(nil, errors.New("pf err"))
				return nil
			},
			expectError:     true,
			expectedErrMsg:  "failed to forward port 8080 to pod p1 port 1234: pf err",
			expectTaskSaved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTestEnv(t, tt.fs, nil)
			te.fnd.On("DryRun").Maybe().Return(tt.dryRun)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			expectedTask := tt.setupMocks(t, ctx, cancel, te)

			got, err := te.env.RunTask(ctx, tt.ss, tt.cmd)

			if tt.expectError {
				require.Nil(t, got)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				actualTask, ok := got.(*kubernetesTask)
				require.True(t, ok)
				assert.Equal(t, expectedTask, actualTask)
			}
			if tt.expectTaskSaved {
				assert.Contains(t, te.env.tasks, "i1-svc")
			} else {
				assert.Empty(t, te.env.tasks)
			}
		})
	}
}

func Test_kubernetesEnvironment_ExecTaskCommand(t *testing.T) {
	ss := getTestingServiceSettings(false)
	cmd := &environment.Command{Name: "php", Args: []string{"-v"}}
	tests := []struct {
		name           string
		dryRun         bool
		target         task.Task
		withCollector  bool
		setupMocks     func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:          "successful command execution with collector",
			target:        &kubernetesTask{serviceName: "i1-svc", podName: "p1"},
			withCollector: true,
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) {
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}
				oc.On("StdoutWriter").Return(stdout)
				oc.On("StderrWriter").Return(stderr)
				te.pc.On("Exec", ctx, "p1", &clients.ExecOptions{
					Container: "i1-svc",
					Command:   []string{"php", "-v"},
					Stdout:    stdout,
					Stderr:    stderr,
				}).Return(nil)
			},
		},
		{
			name:   "successful command execution without collector",
			target: &kubernetesTask{serviceName: "i1-svc", podName: "p1"},
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) {
				te.pc.On("Exec", ctx, "p1", &clients.ExecOptions{
					Container: "i1-svc",
					Command:   []string{"php", "-v"},
					Stdout:    io.Discard,
					Stderr:    io.Discard,
				}).Return(nil)
			},
		},
		{
			name:       "successful command execution in dry run",
			dryRun:     true,
			target:     &kubernetesTask{serviceName: "i1-svc", podName: "p1"},
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) {},
		},
		{
			name:   "failed command execution due to exec error",
			target: &kubernetesTask{serviceName: "i1-svc", podName: "p1"},
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) {
				te.pc.On("Exec", ctx, "p1", mock.Anything).Return(errors.New("exec err"))
			},
			expectError:    true,
			expectedErrMsg: "failed to execute command php in pod p1: exec err",
		},
		{
			name:           "failed command execution due to invalid task",
			target:         &invalidTask{},
			setupMocks:     func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) {},
			expectError:    true,
			expectedErrMsg: "task in not a Kubernetes task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTestEnv(t, nil, nil)
			te.fnd.On("DryRun").Maybe().Return(tt.dryRun)
			ctx := context.Background()
			var oc output.Collector
			var ocMock *outputMocks.MockCollector
			if tt.withCollector {
				ocMock = outputMocks.NewMockCollector(t)
				oc = ocMock
			}
			tt.setupMocks(t, ctx, te, ocMock)

			err := te.env.ExecTaskCommand(ctx, ss, tt.target, cmd, oc)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_kubernetesEnvironment_ExecTaskSignal(t *testing.T) {
	ss := getTestingServiceSettings(false)
	deleteOpts := metav1.DeleteOptions{}
	killOpts := metav1.DeleteOptions{GracePeriodSeconds: int64Ptr(0)}
	makeTask := func(portForward io.Closer) *kubernetesTask {
		return &kubernetesTask{
			serviceName: "i1-svc",
			podName:     "p1",
			deployment:  &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "i1-svc"}},
			portForward: portForward,
		}
	}
	tests := []struct {
		name               string
		dryRun             bool
		signal             os.Signal
		target             task.Task
		setupMocks         func(t *testing.T, ctx context.Context, te *testEnv)
		expectError        bool
		expectedErrMsg     string
		expectDeployment   bool
		expectPortForward  bool
		expectClosedClosed int
	}{
		{
			name:   "successful SIGTERM deletes deployment and stops forwarding",
			signal: syscall.SIGTERM,
			target: makeTask(&fakeCloser{}),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.dc.On("Delete", ctx, "i1-svc", deleteOpts).Return(nil)
			},
		},
		{
			name:   "successful SIGTERM without port forwarding",
			signal: syscall.SIGTERM,
			target: makeTask(nil),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.dc.On("Delete", ctx, "i1-svc", deleteOpts).Return(nil)
			},
		},
		{
			name:   "successful SIGKILL deletes deployment and pod immediately",
			signal: syscall.SIGKILL,
			target: makeTask(nil),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.dc.On("Delete", ctx, "i1-svc", deleteOpts).Return(nil)
				te.pc.On("Delete", ctx, "p1", killOpts).Return(nil)
			},
		},
		{
			name:   "successful SIGKILL when pod is already gone",
			signal: syscall.SIGKILL,
			target: makeTask(nil),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.dc.On("Delete", ctx, "i1-svc", deleteOpts).Return(nil)
				te.pc.On("Delete", ctx, "p1", killOpts).Return(notFoundError())
			},
		},
		{
			name:   "successful SIGKILL in dry run",
			dryRun: true,
			signal: syscall.SIGKILL,
			target: makeTask(nil),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				dryRun := []string{metav1.DryRunAll}
				te.dc.On("Delete", ctx, "i1-svc", metav1.DeleteOptions{DryRun: dryRun}).Return(nil)
				te.pc.On("Delete", ctx, "p1", metav1.DeleteOptions{DryRun: dryRun, GracePeriodSeconds: int64Ptr(0)}).Return(nil)
			},
		},
		{
			name:   "successful SIGUSR2 executes kill in the container",
			signal: syscall.SIGUSR2,
			target: makeTask(&fakeCloser{}),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.pc.On("Exec", ctx, "p1", &clients.ExecOptions{
					Container: "i1-svc",
					Command:   []string{"kill", "-s", "USR2", "1"},
					Stdout:    io.Discard,
					Stderr:    io.Discard,
				}).Return(nil)
			},
			expectDeployment:  true,
			expectPortForward: true,
		},
		{
			name:              "successful SIGHUP in dry run",
			dryRun:            true,
			signal:            syscall.SIGHUP,
			target:            makeTask(nil),
			setupMocks:        func(t *testing.T, ctx context.Context, te *testEnv) {},
			expectDeployment:  true,
			expectPortForward: false,
		},
		{
			name:   "failed SIGINT due to exec error",
			signal: syscall.SIGINT,
			target: makeTask(nil),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.pc.On("Exec", ctx, "p1", mock.Anything).Return(errors.New("exec err"))
			},
			expectError:      true,
			expectedErrMsg:   "failed to send signal INT to pod p1: exec err",
			expectDeployment: true,
		},
		{
			name:             "failed signal due to unsupported signal",
			signal:           syscall.SIGALRM,
			target:           makeTask(nil),
			setupMocks:       func(t *testing.T, ctx context.Context, te *testEnv) {},
			expectError:      true,
			expectedErrMsg:   "signal alarm clock is not supported in Kubernetes environment",
			expectDeployment: true,
		},
		{
			name:   "failed SIGTERM due to port forward close error",
			signal: syscall.SIGTERM,
			target: makeTask(&fakeCloser{err: errors.New("pf err")}),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
			},
			expectError:      true,
			expectedErrMsg:   "failed to stop port forwarding",
			expectDeployment: true,
		},
		{
			name:   "failed SIGTERM due to deployment delete error",
			signal: syscall.SIGTERM,
			target: makeTask(nil),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.dc.On("Delete", ctx, "i1-svc", deleteOpts).Return(errors.New("dep err"))
			},
			expectError:      true,
			expectedErrMsg:   "failed to delete deployment",
			expectDeployment: true,
		},
		{
			name:   "failed SIGKILL due to pod delete error",
			signal: syscall.SIGKILL,
			target: makeTask(nil),
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv) {
				te.dc.On("Delete", ctx, "i1-svc", deleteOpts).Return(nil)
				te.pc.On("Delete", ctx, "p1", killOpts).Return(errors.New("pod err"))
			},
			expectError:    true,
			expectedErrMsg: "failed to delete pod p1: pod err",
		},
		{
			name:           "failed SIGTERM due to already terminated task",
			signal:         syscall.SIGTERM,
			target:         &kubernetesTask{serviceName: "i1-svc", podName: "p1"},
			setupMocks:     func(t *testing.T, ctx context.Context, te *testEnv) {},
			expectError:    true,
			expectedErrMsg: "task i1-svc has already been terminated",
		},
		{
			name:           "failed signal due to invalid task",
			signal:         syscall.SIGTERM,
			target:         &invalidTask{},
			setupMocks:     func(t *testing.T, ctx context.Context, te *testEnv) {},
			expectError:    true,
			expectedErrMsg: "task in not a Kubernetes task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTestEnv(t, nil, nil)
			te.fnd.On("DryRun").Maybe().Return(tt.dryRun)
			ctx := context.Background()
			tt.setupMocks(t, ctx, te)

			err := te.env.ExecTaskSignal(ctx, ss, tt.target, tt.signal)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
			if kubeTask, ok := tt.target.(*kubernetesTask); ok {
				if tt.expectDeployment {
					assert.NotNil(t, kubeTask.deployment)
				} else {
					assert.Nil(t, kubeTask.deployment)
				}
				if tt.expectPortForward {
					assert.NotNil(t, kubeTask.portForward)
				} else {
					assert.Nil(t, kubeTask.portForward)
				}
			}
		})
	}
}

func Test_kubernetesEnvironment_Output(t *testing.T) {
	tests := []struct {
		name           string
		dryRun         bool
		outputType     output.Type
		target         task.Task
		setupMocks     func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) io.Reader
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:       "successful output for any type",
			outputType: output.Any,
			target:     &kubernetesTask{serviceName: "i1-svc"},
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) io.Reader {
				reader := strings.NewReader("data")
				oc.On("AnyReader", ctx).Return(reader)
				return reader
			},
		},
		{
			name:       "successful output for stdout type uses combined logs",
			outputType: output.Stdout,
			target:     &kubernetesTask{serviceName: "i1-svc"},
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) io.Reader {
				reader := strings.NewReader("data")
				oc.On("AnyReader", ctx).Return(reader)
				return reader
			},
		},
		{
			name:       "successful output for dry run",
			dryRun:     true,
			outputType: output.Any,
			target:     &kubernetesTask{serviceName: "i1-svc"},
			setupMocks: func(t *testing.T, ctx context.Context, te *testEnv, oc *outputMocks.MockCollector) io.Reader {
				return &app.DummyReaderCloser{}
			},
		},
		{
			name:           "failed output due to missing collector",
			outputType:     output.Any,
			target:         &kubernetesTask{serviceName: "i1-svc"},
			expectError:    true,
			expectedErrMsg: "output of task i1-svc is not available",
		},
		{
			name:           "failed output due to invalid task",
			outputType:     output.Any,
			target:         &invalidTask{},
			expectError:    true,
			expectedErrMsg: "task in not a Kubernetes task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTestEnv(t, nil, nil)
			te.fnd.On("DryRun").Maybe().Return(tt.dryRun)
			ctx := context.Background()
			var expectedReader io.Reader
			if tt.setupMocks != nil {
				oc := outputMocks.NewMockCollector(t)
				if kubeTask, ok := tt.target.(*kubernetesTask); ok && !tt.dryRun {
					kubeTask.outputCollector = oc
				}
				expectedReader = tt.setupMocks(t, ctx, te, oc)
			}

			reader, err := te.env.Output(ctx, tt.target, tt.outputType)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, reader)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, expectedReader, reader)
			}
		})
	}
}

func Test_podFailure(t *testing.T) {
	tests := []struct {
		name           string
		pod            *corev1.Pod
		expectedErrMsg string
	}{
		{name: "running pod", pod: runningPod("p", true)},
		{name: "pending pod with non failure waiting reason", pod: pendingPod("p")},
		{name: "failed pod phase", pod: failedPod("p"), expectedErrMsg: "pod failed: Evicted no memory"},
		{
			name:           "waiting container with crash loop back off",
			pod:            waitingPod("p", "CrashLoopBackOff"),
			expectedErrMsg: "container c is waiting: CrashLoopBackOff: msg",
		},
		{
			name:           "waiting container with image pull error",
			pod:            waitingPod("p", "ErrImagePull"),
			expectedErrMsg: "container c is waiting: ErrImagePull: msg",
		},
		{
			name:           "terminated container",
			pod:            terminatedPod("p", 3),
			expectedErrMsg: "container c terminated with exit code 3: Error: boom",
		},
		{
			name:           "restarted container",
			pod:            restartedPod("p"),
			expectedErrMsg: "container c restarted after termination with exit code 2: Error: died",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := podFailure(tt.pod)
			if tt.expectedErrMsg == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			}
		})
	}
}

func Test_isPodReady(t *testing.T) {
	assert.True(t, isPodReady(runningPod("p", true)))
	assert.False(t, isPodReady(runningPod("p", false)))
	assert.False(t, isPodReady(pendingPod("p")))
	assert.False(t, isPodReady(&corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodRunning}}))
}

func Test_sanitizeName(t *testing.T) {
	assert.Equal(t, "my-svc-configs-1", sanitizeName("My_svc.configs/1"))
	long := strings.Repeat("a", 300)
	assert.Equal(t, strings.Repeat("a", 253), sanitizeName(long))
}

func Test_kubernetesEnvironment_RootPath(t *testing.T) {
	env := &kubernetesEnvironment{}
	assert.Equal(t, "", env.RootPath("/www/ws"))
}

func Test_kubernetesEnvironment_Mkdir(t *testing.T) {
	env := &kubernetesEnvironment{}
	assert.Nil(t, env.Mkdir("svc", "/www/ws", 0755))
}

func Test_kubernetesEnvironment_ServiceLocalAddress(t *testing.T) {
	env := &kubernetesEnvironment{}
	assert.Equal(t, "0.0.0.0:80", env.ServiceLocalAddress("svc", 1234, 80))
}

func Test_kubernetesEnvironment_ServiceLocalPort(t *testing.T) {
	env := &kubernetesEnvironment{}
	assert.Equal(t, int32(80), env.ServiceLocalPort(1234, 80))
}

func Test_kubernetesEnvironment_ServicePrivateAddress(t *testing.T) {
	env := &kubernetesEnvironment{}
	assert.Equal(t, "svc", env.ServicePrivateAddress("svc", 1234, 80))
}

func getTestTask() *kubernetesTask {
	return &kubernetesTask{
		deployment:        nil,
		service:           nil,
		configMaps:        nil,
		executable:        "epk",
		serviceName:       "kubes",
		servicePublicUrl:  "://localhost:1234",
		servicePrivateUrl: "://kubes:8080",
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
	assert.Equal(t, "http://kubes:8080", getTestTask().PrivateUrl("http"))
}

func Test_kubernetesTask_PublicUrl(t *testing.T) {
	assert.Equal(t, "https://localhost:1234", getTestTask().PublicUrl("https"))
}

func Test_kubernetesTask_Type(t *testing.T) {
	assert.Equal(t, providers.KubernetesType, getTestTask().Type())
}

func Test_kubernetesEnvironment_serviceName(t *testing.T) {
	ss := getTestingServiceSettings(false)
	assert.Equal(t, "i1-svc", (&kubernetesEnvironment{useUniqueName: true}).serviceName(ss))
	assert.Equal(t, "svc", (&kubernetesEnvironment{useUniqueName: false}).serviceName(ss))
}
