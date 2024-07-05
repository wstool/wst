package environments

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	envMocks "github.com/bukka/wst/mocks/generated/run/environments/environment"
	dockerMocks "github.com/bukka/wst/mocks/generated/run/environments/environment/providers/docker"
	kubernetesMocks "github.com/bukka/wst/mocks/generated/run/environments/environment/providers/kubernetes"
	localMocks "github.com/bukka/wst/mocks/generated/run/environments/environment/providers/local"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
				"local": &types.LocalEnvironment{
					Ports: types.EnvironmentPorts{Start: 1000, End: 2000},
				},
				"docker": &types.DockerEnvironment{
					Ports: types.EnvironmentPorts{Start: 3000, End: 4000},
				},
				"kubernetes": &types.KubernetesEnvironment{
					Ports: types.EnvironmentPorts{Start: 5000, End: 6000},
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
					},
					"/workspace",
				).Return(localEnv, nil)

				// Setup Docker environment maker
				dockerEnv := envMocks.NewMockEnvironment(t)
				dockerMock.On(
					"Make",
					&types.DockerEnvironment{
						Ports: types.EnvironmentPorts{Start: 3000, End: 4000},
					},
				).Return(dockerEnv, nil)

				// Setup Kubernetes environment maker
				kubernetesEnv := envMocks.NewMockEnvironment(t)
				kubernetesMock.On(
					"Make",
					&types.KubernetesEnvironment{
						Ports: types.EnvironmentPorts{Start: 5000, End: 6000},
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
					Ports: types.EnvironmentPorts{Start: 2000, End: 3000},
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
					},
					"/workspace",
				).Return(localEnv, nil)

				// Setup Docker environment maker
				dockerEnv := envMocks.NewMockEnvironment(t)
				dockerMock.On(
					"Make",
					&types.DockerEnvironment{
						Ports: types.EnvironmentPorts{Start: 3000, End: 4000},
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
				}, "/workspace").Return(localEnv, nil)
				return Environments{providers.LocalType: localEnv}
			},
		},
		{
			name: "mixed success and errors in environment creation",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			localMock := localMocks.NewMockMaker(t)
			dockerMock := dockerMocks.NewMockMaker(t)
			kubernetesMock := kubernetesMocks.NewMockMaker(t)
			maker := CreateMaker(fndMock)
			nm := maker.(*nativeMaker)
			nm.localMaker = localMock
			nm.dockerMaker = dockerMock
			nm.kubernetesMaker = kubernetesMock

			expectEnvironments := tt.setupMocks(t, localMock, dockerMock, kubernetesMock)

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
