package environments

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	envMocks "github.com/wstool/wst/mocks/generated/run/environments/environment"
	dockerMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/providers/docker"
	kubernetesMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/providers/kubernetes"
	localMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/providers/local"
	resourcesMocks "github.com/wstool/wst/mocks/generated/run/resources"
	"github.com/wstool/wst/run/environments/environment/providers"
	"testing"
)

func TestNativeMaker_Make(t *testing.T) {
	tests := []struct {
		name              string
		specConfig        map[string]types.Environment
		instanceConfig    map[string]types.Environment
		instanceWorkspace string
		setupMocks        func(
			*testing.T,
			*localMocks.MockMaker,
			*dockerMocks.MockMaker,
			*kubernetesMocks.MockMaker,
		) Environments
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful merge and environment creation",
			specConfig: map[string]types.Environment{
				"common": &types.CommonEnvironment{
					Resources: types.Resources{
						Certificates: map[string]types.Certificate{
							"local": {
								Certificate: "cert",
								PrivateKey:  "key",
							},
						},
						Scripts: map[string]types.Script{
							"index": {
								Content: "<?php echo 'hello';",
							},
						},
					},
				},
				"local": &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
				},
				"docker": &types.DockerEnvironment{
					Ports: types.EnvironmentPorts{Start: 3000, End: 4000},
				},
				"kubernetes": &types.KubernetesEnvironment{
					Ports: types.EnvironmentPorts{Start: 5000, End: 6000},
					Resources: types.Resources{
						Certificates: map[string]types.Certificate{
							"local": {
								Certificate: "kube cert",
								PrivateKey:  "kube key",
							},
						},
						Scripts: map[string]types.Script{
							"index": {
								Content: "<?php echo 'kube';",
							},
						},
					},
				},
			},
			instanceConfig: map[string]types.Environment{
				"local": &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1500, End: 2500},
				},
			},
			instanceWorkspace: "/workspace",
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Environments {
				// Setup local environment maker
				localEnv := envMocks.NewMockEnvironment(t)
				localMock.On(
					"Make",
					&types.LocalEnvironment{
						Ports: types.EnvironmentPorts{Start: 1500, End: 2500},
						Resources: types.Resources{
							Certificates: map[string]types.Certificate{
								"local": {
									Certificate: "cert",
									PrivateKey:  "key",
								},
							},
							Scripts: map[string]types.Script{
								"index": {
									Content: "<?php echo 'hello';",
								},
							},
						},
					},
					"/workspace",
				).Return(localEnv, nil)

				// Setup Docker environment maker
				dockerEnv := envMocks.NewMockEnvironment(t)
				dockerMock.On(
					"Make",
					&types.DockerEnvironment{
						Ports: types.EnvironmentPorts{Start: 3000, End: 4000},
						Resources: types.Resources{
							Certificates: map[string]types.Certificate{
								"local": {
									Certificate: "cert",
									PrivateKey:  "key",
								},
							},
							Scripts: map[string]types.Script{
								"index": {
									Content: "<?php echo 'hello';",
								},
							},
						},
					},
				).Return(dockerEnv, nil)

				// Setup Kubernetes environment maker
				kubernetesEnv := envMocks.NewMockEnvironment(t)
				kubernetesMock.On(
					"Make",
					&types.KubernetesEnvironment{
						Ports: types.EnvironmentPorts{Start: 5000, End: 6000},
						Resources: types.Resources{
							Certificates: map[string]types.Certificate{
								"local": {
									Certificate: "kube cert",
									PrivateKey:  "kube key",
								},
							},
							Scripts: map[string]types.Script{
								"index": {
									Content: "<?php echo 'kube';",
								},
							},
						},
					},
				).Return(kubernetesEnv, nil)

				return Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubernetesEnv,
				}
			},
		},
		{
			name: "successful merge of more environments",
			specConfig: map[string]types.Environment{
				"common": &types.CommonEnvironment{
					Ports: types.EnvironmentPorts{Start: 2000, End: 3000},
				},
				"container": &types.ContainerEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "test",
							Password: "pwd",
						},
					},
				},
				"docker": &types.DockerEnvironment{
					Ports: types.EnvironmentPorts{Start: 3000, End: 4000},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "user",
							Password: "1234",
						},
					},
				},
				"kubernetes": &types.KubernetesEnvironment{},
			},
			instanceConfig: map[string]types.Environment{
				"local": &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1500, End: 2500},
				},
			},
			instanceWorkspace: "/workspace",
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Environments {
				// Setup local environment maker
				localEnv := envMocks.NewMockEnvironment(t)
				localMock.On(
					"Make",
					&types.LocalEnvironment{
						Ports: types.EnvironmentPorts{Start: 1500, End: 2500},
						Resources: types.Resources{
							Scripts:      map[string]types.Script{},
							Certificates: map[string]types.Certificate{},
						},
					},
					"/workspace",
				).Return(localEnv, nil)

				// Setup Docker environment maker
				dockerEnv := envMocks.NewMockEnvironment(t)
				dockerMock.On(
					"Make",
					&types.DockerEnvironment{
						Ports: types.EnvironmentPorts{Start: 3000, End: 4000},
						Resources: types.Resources{
							Scripts:      map[string]types.Script{},
							Certificates: map[string]types.Certificate{},
						},
						Registry: types.ContainerRegistry{
							Auth: types.ContainerRegistryAuth{
								Username: "user",
								Password: "1234",
							},
						},
					},
				).Return(dockerEnv, nil)

				// Setup Kubernetes environment maker
				kubernetesEnv := envMocks.NewMockEnvironment(t)
				kubernetesMock.On(
					"Make",
					&types.KubernetesEnvironment{
						Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
						Resources: types.Resources{
							Scripts:      map[string]types.Script{},
							Certificates: map[string]types.Certificate{},
						},
						Registry: types.ContainerRegistry{
							Auth: types.ContainerRegistryAuth{
								Username: "test",
								Password: "pwd",
							},
						},
					},
				).Return(kubernetesEnv, nil)

				return Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubernetesEnv,
				}
			},
		},
		{
			name: "successful merge of all environments",
			specConfig: map[string]types.Environment{
				"common": &types.CommonEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
					Resources: types.Resources{
						Certificates: map[string]types.Certificate{
							"local": {
								Certificate: "spec cert",
								PrivateKey:  "spec key",
							},
						},
						Scripts: map[string]types.Script{
							"index": {
								Content: "<?php echo 'hello spec';",
							},
						},
					},
				},
				"local": &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 500, End: 1000},
				},
				"container": &types.ContainerEnvironment{
					Ports: types.EnvironmentPorts{Start: 10000, End: 20000},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "testx",
							Password: "pwdx",
						},
					},
				},
				"docker": &types.DockerEnvironment{
					Ports: types.EnvironmentPorts{Start: 2000, End: 3000},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "userx",
							Password: "1234x",
						},
					},
					NamePrefix: "t",
				},
				"kubernetes": &types.KubernetesEnvironment{
					Namespace:  "kube",
					Kubeconfig: "tmp/k.conf",
				},
			},
			instanceConfig: map[string]types.Environment{
				"common": &types.CommonEnvironment{
					Ports: types.EnvironmentPorts{Start: 2000, End: 3000},
					Resources: types.Resources{
						Certificates: map[string]types.Certificate{
							"local": {
								Certificate: "inst cert",
								PrivateKey:  "inst key",
							},
						},
						Scripts: map[string]types.Script{
							"index": {
								Content: "<?php echo 'hello inst';",
							},
						},
					},
				},
				"local": &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
				},
				"container": &types.ContainerEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "test",
							Password: "pwd",
						},
					},
				},
				"docker": &types.DockerEnvironment{
					Ports: types.EnvironmentPorts{Start: 3000, End: 4000},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "user",
							Password: "1234",
						},
					},
					NamePrefix: "tnp",
				},
				"kubernetes": &types.KubernetesEnvironment{
					Namespace:  "kubetest",
					Kubeconfig: "tmp/kube.conf",
				},
			},
			instanceWorkspace: "/workspace",
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Environments {
				// Setup local environment maker
				localEnv := envMocks.NewMockEnvironment(t)
				localMock.On(
					"Make",
					&types.LocalEnvironment{
						Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
						Resources: types.Resources{
							Certificates: map[string]types.Certificate{
								"local": {
									Certificate: "inst cert",
									PrivateKey:  "inst key",
								},
							},
							Scripts: map[string]types.Script{
								"index": {
									Content: "<?php echo 'hello inst';",
								},
							},
						},
					},
					"/workspace",
				).Return(localEnv, nil)

				// Setup Docker environment maker
				dockerEnv := envMocks.NewMockEnvironment(t)
				dockerMock.On(
					"Make",
					&types.DockerEnvironment{
						Ports: types.EnvironmentPorts{Start: 3000, End: 4000},
						Resources: types.Resources{
							Certificates: map[string]types.Certificate{
								"local": {
									Certificate: "inst cert",
									PrivateKey:  "inst key",
								},
							},
							Scripts: map[string]types.Script{
								"index": {
									Content: "<?php echo 'hello inst';",
								},
							},
						},
						Registry: types.ContainerRegistry{
							Auth: types.ContainerRegistryAuth{
								Username: "user",
								Password: "1234",
							},
						},
						NamePrefix: "tnp",
					},
				).Return(dockerEnv, nil)

				// Setup Kubernetes environment maker
				kubernetesEnv := envMocks.NewMockEnvironment(t)
				kubernetesMock.On(
					"Make",
					&types.KubernetesEnvironment{
						Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
						Resources: types.Resources{
							Certificates: map[string]types.Certificate{
								"local": {
									Certificate: "inst cert",
									PrivateKey:  "inst key",
								},
							},
							Scripts: map[string]types.Script{
								"index": {
									Content: "<?php echo 'hello inst';",
								},
							},
						},
						Registry: types.ContainerRegistry{
							Auth: types.ContainerRegistryAuth{
								Username: "test",
								Password: "pwd",
							},
						},
						Namespace:  "kubetest",
						Kubeconfig: "tmp/kube.conf",
					},
				).Return(kubernetesEnv, nil)

				return Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubernetesEnv,
				}
			},
		},
		{
			name: "successful merge of common only environment",
			specConfig: map[string]types.Environment{
				"common": &types.CommonEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
				},
			},
			instanceConfig:    map[string]types.Environment{},
			instanceWorkspace: "/workspace",
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Environments {
				// Setup local environment maker
				localEnv := envMocks.NewMockEnvironment(t)
				localMock.On(
					"Make",
					&types.LocalEnvironment{
						Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
					},
					"/workspace",
				).Return(localEnv, nil)

				// Setup Docker environment maker
				dockerEnv := envMocks.NewMockEnvironment(t)
				dockerMock.On(
					"Make",
					&types.DockerEnvironment{
						Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
					},
				).Return(dockerEnv, nil)

				// Setup Kubernetes environment maker
				kubernetesEnv := envMocks.NewMockEnvironment(t)
				kubernetesMock.On(
					"Make",
					&types.KubernetesEnvironment{
						Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
					},
				).Return(kubernetesEnv, nil)

				return Environments{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubernetesEnv,
				}
			},
		},
		{
			name: "partial configuration with only local defined",
			specConfig: map[string]types.Environment{
				"local": &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
				},
			},
			instanceWorkspace: "/workspace",
			setupMocks: func(t *testing.T, localMock *localMocks.MockMaker, dockerMock *dockerMocks.MockMaker, kubernetesMock *kubernetesMocks.MockMaker) Environments {
				localEnv := envMocks.NewMockEnvironment(t)
				localMock.On("Make", &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
				}, "/workspace").Return(localEnv, nil)
				return Environments{providers.LocalType: localEnv}
			},
		},
		{
			name: "conflicting ports resolved by instance config",
			specConfig: map[string]types.Environment{
				"local": &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
				},
			},
			instanceConfig: map[string]types.Environment{
				"local": &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1500, End: 2500},
				},
			},
			instanceWorkspace: "/workspace",
			setupMocks: func(t *testing.T, localMock *localMocks.MockMaker, dockerMock *dockerMocks.MockMaker, kubernetesMock *kubernetesMocks.MockMaker) Environments {
				localEnv := envMocks.NewMockEnvironment(t)
				localMock.On("Make", &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1500, End: 2500},
					Resources: types.Resources{
						Certificates: map[string]types.Certificate{},
						Scripts:      map[string]types.Script{},
					},
				}, "/workspace").Return(localEnv, nil)
				return Environments{providers.LocalType: localEnv}
			},
		},
		{
			name: "mixed success and errors in docker environment creation",
			specConfig: map[string]types.Environment{
				"local":  &types.LocalEnvironment{},
				"docker": &types.DockerEnvironment{},
			},
			instanceWorkspace: "/workspace",
			setupMocks: func(t *testing.T, localMock *localMocks.MockMaker, dockerMock *dockerMocks.MockMaker, kubernetesMock *kubernetesMocks.MockMaker) Environments {
				localEnv := envMocks.NewMockEnvironment(t)
				localMock.On("Make", mock.AnythingOfType("*types.LocalEnvironment"), "/workspace").Return(localEnv, nil)
				dockerMock.On("Make", mock.AnythingOfType("*types.DockerEnvironment")).Return(nil, errors.New("docker creation failed"))
				return Environments{providers.LocalType: localEnv}
			},
			expectError:      true,
			expectedErrorMsg: "docker creation failed",
		},
		{
			name: "mixed success and errors in kubernetes environment creation",
			specConfig: map[string]types.Environment{
				"local":      &types.LocalEnvironment{},
				"kubernetes": &types.KubernetesEnvironment{},
			},
			instanceWorkspace: "/workspace",
			setupMocks: func(t *testing.T, localMock *localMocks.MockMaker, dockerMock *dockerMocks.MockMaker, kubernetesMock *kubernetesMocks.MockMaker) Environments {
				localEnv := envMocks.NewMockEnvironment(t)
				localMock.On("Make", mock.AnythingOfType("*types.LocalEnvironment"), "/workspace").Return(localEnv, nil)
				kubernetesMock.On("Make", mock.AnythingOfType("*types.KubernetesEnvironment")).Return(nil, errors.New("k8s creation failed"))
				return Environments{providers.LocalType: localEnv}
			},
			expectError:      true,
			expectedErrorMsg: "k8s creation failed",
		},
		{
			name: "error during environment creation",
			specConfig: map[string]types.Environment{
				"local": &types.LocalEnvironment{},
			},
			instanceConfig: map[string]types.Environment{
				"local": &types.LocalEnvironment{},
			},
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Environments {
				localMock.On(
					"Make",
					mock.AnythingOfType("*types.LocalEnvironment"),
					mock.Anything,
				).Return(nil, errors.New("local err"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "local err",
		},
		{
			name: "error during common env merging",
			specConfig: map[string]types.Environment{
				"common": &types.LocalEnvironment{},
			},
			instanceConfig: map[string]types.Environment{
				"common": &types.LocalEnvironment{},
			},
			expectError:      true,
			expectedErrorMsg: "common environment is not set in common field",
		},
		{
			name: "error during local env merging",
			specConfig: map[string]types.Environment{
				"local": &types.CommonEnvironment{},
			},
			instanceConfig: map[string]types.Environment{
				"local": &types.CommonEnvironment{},
			},
			expectError:      true,
			expectedErrorMsg: "local environment is not set in local field",
		},
		{
			name: "error during container env merging",
			specConfig: map[string]types.Environment{
				"container": &types.LocalEnvironment{},
			},
			instanceConfig: map[string]types.Environment{
				"container": &types.LocalEnvironment{},
			},
			expectError:      true,
			expectedErrorMsg: "container environment is not set in container field",
		},
		{
			name: "error during docker env merging",
			specConfig: map[string]types.Environment{
				"docker": &types.LocalEnvironment{},
			},
			instanceConfig: map[string]types.Environment{
				"docker": &types.LocalEnvironment{},
			},
			expectError:      true,
			expectedErrorMsg: "docker environment is not set in docker field",
		},
		{
			name: "error during kubernetes env merging",
			specConfig: map[string]types.Environment{
				"kubernetes": &types.LocalEnvironment{},
			},
			instanceConfig: map[string]types.Environment{
				"kubernetes": &types.LocalEnvironment{},
			},
			expectError:      true,
			expectedErrorMsg: "kubernetes environment is not set in kubernetes field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			resourcesMaker := resourcesMocks.NewMockMaker(t)
			localMock := localMocks.NewMockMaker(t)
			dockerMock := dockerMocks.NewMockMaker(t)
			kubernetesMock := kubernetesMocks.NewMockMaker(t)
			maker := CreateMaker(fndMock, resourcesMaker)
			nm := maker.(*nativeMaker)
			nm.localMaker = localMock
			nm.dockerMaker = dockerMock
			nm.kubernetesMaker = kubernetesMock

			var expectEnvironments types.Environment
			if tt.setupMocks != nil {
				expectEnvironments = tt.setupMocks(t, localMock, dockerMock, kubernetesMock)
			}

			environments, err := nm.Make(tt.specConfig, tt.instanceConfig, tt.instanceWorkspace)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectEnvironments, environments)
			}

			localMock.AssertExpectations(t)
			dockerMock.AssertExpectations(t)
			kubernetesMock.AssertExpectations(t)
		})
	}
}
