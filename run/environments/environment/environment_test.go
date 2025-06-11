package environment

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	outputMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/output"
	resourcesMocks "github.com/wstool/wst/mocks/generated/run/resources"
	"github.com/wstool/wst/run/resources"
	"testing"
)

func TestCommonMaker_MakeCommonEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		inputConfig    *types.CommonEnvironment
		setupMocks     func(resourceMaker *resourcesMocks.MockMaker) *resources.Resources
		expectedResult *CommonEnvironment
		expectError    bool
		errorMessage   string
	}{
		{
			name: "Basic configuration with resources",
			inputConfig: &types.CommonEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 3000,
					End:   4000,
				},
				Resources: types.Resources{
					Scripts: map[string]types.Script{
						"init": {
							Content: "echo 'init'",
							Path:    "/init.sh",
						},
					},
				},
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) *resources.Resources {
				mockResources := &resources.Resources{}
				resourceMaker.On("Make", types.Resources{
					Scripts: map[string]types.Script{
						"init": {
							Content: "echo 'init'",
							Path:    "/init.sh",
						},
					},
				}).Return(mockResources, nil)
				return mockResources
			},
			expectedResult: &CommonEnvironment{
				Ports: Ports{
					Start: 3000,
					Used:  3000,
					End:   4000,
				},
				Used: false,
			},
		},
		{
			name: "Resource creation error",
			inputConfig: &types.CommonEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 3000,
					End:   4000,
				},
				Resources: types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Content: "echo 'test'",
							Mode:    "invalid_mode",
						},
					},
				},
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) *resources.Resources {
				resourceMaker.On("Make", types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Content: "echo 'test'",
							Mode:    "invalid_mode",
						},
					},
				}).Return(nil, errors.New("resource creation failed"))
				return nil
			},
			expectError:  true,
			errorMessage: "resource creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			resourceMakerMock := resourcesMocks.NewMockMaker(t)

			expectedResources := tt.setupMocks(resourceMakerMock)

			maker := CreateCommonMaker(fndMock, resourceMakerMock)
			result, err := maker.MakeCommonEnvironment(tt.inputConfig)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				assert.Equal(t, fndMock, result.Fnd)
				assert.NotNil(t, result.OutputMaker)
				assert.Equal(t, tt.expectedResult.Ports.Start, result.Ports.Start)
				assert.Equal(t, tt.expectedResult.Ports.Used, result.Ports.Used)
				assert.Equal(t, tt.expectedResult.Ports.End, result.Ports.End)
				assert.Equal(t, tt.expectedResult.Used, result.Used)
				assert.Equal(t, expectedResources, result.EnvResources)
				assert.Equal(t, expectedResources, result.Resources())
			}

			resourceMakerMock.AssertExpectations(t)
		})
	}
}

func TestCommonMaker_MakeLocalEnvironment(t *testing.T) {
	tests := []struct {
		name         string
		config       *types.LocalEnvironment
		setupMocks   func(resourceMaker *resourcesMocks.MockMaker) *resources.Resources
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful local environment creation",
			config: &types.LocalEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 5000,
					End:   6000,
				},
				Resources: types.Resources{
					Scripts: map[string]types.Script{
						"local_script": {
							Content: "echo 'local'",
							Path:    "/local.sh",
						},
					},
				},
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) *resources.Resources {
				mockResources := &resources.Resources{}
				resourceMaker.On("Make", types.Resources{
					Scripts: map[string]types.Script{
						"local_script": {
							Content: "echo 'local'",
							Path:    "/local.sh",
						},
					},
				}).Return(mockResources, nil)
				return mockResources
			},
		},
		{
			name: "local environment with resource error",
			config: &types.LocalEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 5000,
					End:   6000,
				},
				Resources: types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Mode: "invalid",
						},
					},
				},
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) *resources.Resources {
				resourceMaker.On("Make", types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Mode: "invalid",
						},
					},
				}).Return(nil, errors.New("local resource error"))
				return nil
			},
			expectError:  true,
			errorMessage: "local resource error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			resourceMakerMock := resourcesMocks.NewMockMaker(t)

			expectedResources := tt.setupMocks(resourceMakerMock)

			maker := CreateCommonMaker(fndMock, resourceMakerMock)
			result, err := maker.MakeLocalEnvironment(tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.config.Ports.Start, result.Ports.Start)
				assert.Equal(t, tt.config.Ports.End, result.Ports.End)
				assert.Equal(t, expectedResources, result.Resources())
			}

			resourceMakerMock.AssertExpectations(t)
		})
	}
}

func TestCommonMaker_MakeContainerEnvironment(t *testing.T) {
	tests := []struct {
		name         string
		config       *types.ContainerEnvironment
		setupMocks   func(resourceMaker *resourcesMocks.MockMaker)
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful container environment creation",
			config: &types.ContainerEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 3000,
					End:   3500,
				},
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{
						"container_ssl": {
							Certificate: "cert content",
							PrivateKey:  "key content",
						},
					},
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "user",
						Password: "pass",
					},
				},
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) {
				mockResources := &resources.Resources{}
				resourceMaker.On("Make", types.Resources{
					Certificates: map[string]types.Certificate{
						"container_ssl": {
							Certificate: "cert content",
							PrivateKey:  "key content",
						},
					},
				}).Return(mockResources, nil)
			},
		},
		{
			name: "error creating common environment",
			config: &types.ContainerEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 3000,
					End:   3500,
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "user",
						Password: "pass",
					},
				},
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) {
				resourceMaker.On("Make", types.Resources{}).Return(nil, errors.New("common environment creation failed"))
			},
			expectError:  true,
			errorMessage: "common environment creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			resourceMakerMock := resourcesMocks.NewMockMaker(t)

			tt.setupMocks(resourceMakerMock)

			maker := CreateCommonMaker(fndMock, resourceMakerMock)
			result, err := maker.MakeContainerEnvironment(tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.config.Ports.Start, result.Ports.Start)
				assert.Equal(t, tt.config.Ports.End, result.Ports.End)
				assert.Equal(t, tt.config.Registry.Auth.Username, result.Registry.Auth.Username)
				assert.Equal(t, tt.config.Registry.Auth.Password, result.Registry.Auth.Password)
			}

			resourceMakerMock.AssertExpectations(t)
		})
	}
}

func TestCommonMaker_MakeDockerEnvironment(t *testing.T) {
	tests := []struct {
		name         string
		config       *types.DockerEnvironment
		setupMocks   func(resourceMaker *resourcesMocks.MockMaker)
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful docker environment creation",
			config: &types.DockerEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   9000,
				},
				Resources: types.Resources{
					Scripts: map[string]types.Script{
						"docker_init": {
							Content: "#!/bin/bash\necho 'Docker'",
							Path:    "/docker.sh",
						},
					},
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "docker_user",
						Password: "docker_pass",
					},
				},
				NamePrefix: "my-app",
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) {
				mockResources := &resources.Resources{}
				resourceMaker.On("Make", types.Resources{
					Scripts: map[string]types.Script{
						"docker_init": {
							Content: "#!/bin/bash\necho 'Docker'",
							Path:    "/docker.sh",
						},
					},
				}).Return(mockResources, nil)
			},
		},
		{
			name: "docker environment with container creation error",
			config: &types.DockerEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   9000,
				},
				Resources: types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Mode: "invalid",
						},
					},
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "docker_user",
						Password: "docker_pass",
					},
				},
				NamePrefix: "my-app",
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) {
				resourceMaker.On("Make", types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Mode: "invalid",
						},
					},
				}).Return(nil, errors.New("docker container creation failed"))
			},
			expectError:  true,
			errorMessage: "docker container creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			resourceMakerMock := resourcesMocks.NewMockMaker(t)

			tt.setupMocks(resourceMakerMock)

			maker := CreateCommonMaker(fndMock, resourceMakerMock)
			result, err := maker.MakeDockerEnvironment(tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.config.Ports.Start, result.Ports.Start)
				assert.Equal(t, tt.config.Ports.End, result.Ports.End)
				assert.Equal(t, tt.config.NamePrefix, result.NamePrefix)
				assert.Equal(t, tt.config.Registry.Auth.Username, result.Registry.Auth.Username)
			}

			resourceMakerMock.AssertExpectations(t)
		})
	}
}

func TestCommonMaker_MakeKubernetesEnvironment(t *testing.T) {
	tests := []struct {
		name         string
		config       *types.KubernetesEnvironment
		setupMocks   func(resourceMaker *resourcesMocks.MockMaker)
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful kubernetes environment creation",
			config: &types.KubernetesEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 9000,
					End:   10000,
				},
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{
						"k8s_ssl": {
							Certificate: "k8s cert",
							PrivateKey:  "k8s key",
						},
					},
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "k8s_user",
						Password: "k8s_pass",
					},
				},
				Namespace:  "my-namespace",
				Kubeconfig: "/path/to/kubeconfig",
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) {
				mockResources := &resources.Resources{}
				resourceMaker.On("Make", types.Resources{
					Certificates: map[string]types.Certificate{
						"k8s_ssl": {
							Certificate: "k8s cert",
							PrivateKey:  "k8s key",
						},
					},
				}).Return(mockResources, nil)
			},
		},
		{
			name: "kubernetes environment with container creation error",
			config: &types.KubernetesEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 9000,
					End:   10000,
				},
				Resources: types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Mode: "invalid",
						},
					},
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "k8s_user",
						Password: "k8s_pass",
					},
				},
				Namespace:  "my-namespace",
				Kubeconfig: "/path/to/kubeconfig",
			},
			setupMocks: func(resourceMaker *resourcesMocks.MockMaker) {
				resourceMaker.On("Make", types.Resources{
					Scripts: map[string]types.Script{
						"bad_script": {
							Mode: "invalid",
						},
					},
				}).Return(nil, errors.New("k8s container creation failed"))
			},
			expectError:  true,
			errorMessage: "k8s container creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			resourceMakerMock := resourcesMocks.NewMockMaker(t)

			tt.setupMocks(resourceMakerMock)

			maker := CreateCommonMaker(fndMock, resourceMakerMock)
			result, err := maker.MakeKubernetesEnvironment(tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.config.Ports.Start, result.Ports.Start)
				assert.Equal(t, tt.config.Ports.End, result.Ports.End)
				assert.Equal(t, tt.config.Namespace, result.Namespace)
				assert.Equal(t, tt.config.Kubeconfig, result.Kubeconfig)
				assert.Equal(t, tt.config.Registry.Auth.Username, result.Registry.Auth.Username)
			}

			resourceMakerMock.AssertExpectations(t)
		})
	}
}

// Test merging functions with 100% coverage
func TestMergePorts(t *testing.T) {
	tests := []struct {
		name     string
		base     types.EnvironmentPorts
		override types.EnvironmentPorts
		expected types.EnvironmentPorts
	}{
		{
			name:     "override both ports",
			base:     types.EnvironmentPorts{Start: 3000, End: 4000},
			override: types.EnvironmentPorts{Start: 5000, End: 6000},
			expected: types.EnvironmentPorts{Start: 5000, End: 6000},
		},
		{
			name:     "override only start port",
			base:     types.EnvironmentPorts{Start: 3000, End: 4000},
			override: types.EnvironmentPorts{Start: 5000, End: 0},
			expected: types.EnvironmentPorts{Start: 5000, End: 4000},
		},
		{
			name:     "override only end port",
			base:     types.EnvironmentPorts{Start: 3000, End: 4000},
			override: types.EnvironmentPorts{Start: 0, End: 6000},
			expected: types.EnvironmentPorts{Start: 3000, End: 6000},
		},
		{
			name:     "no override",
			base:     types.EnvironmentPorts{Start: 3000, End: 4000},
			override: types.EnvironmentPorts{Start: 0, End: 0},
			expected: types.EnvironmentPorts{Start: 3000, End: 4000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergePorts(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeResources(t *testing.T) {
	tests := []struct {
		name     string
		base     types.Resources
		override types.Resources
		expected types.Resources
	}{
		{
			name: "merge certificates and scripts",
			base: types.Resources{
				Certificates: map[string]types.Certificate{
					"base_cert": {Certificate: "base", PrivateKey: "base_key"},
				},
				Scripts: map[string]types.Script{
					"base_script": {Content: "base content"},
				},
			},
			override: types.Resources{
				Certificates: map[string]types.Certificate{
					"override_cert": {Certificate: "override", PrivateKey: "override_key"},
				},
				Scripts: map[string]types.Script{
					"override_script": {Content: "override content"},
				},
			},
			expected: types.Resources{
				Certificates: map[string]types.Certificate{
					"base_cert":     {Certificate: "base", PrivateKey: "base_key"},
					"override_cert": {Certificate: "override", PrivateKey: "override_key"},
				},
				Scripts: map[string]types.Script{
					"base_script":     {Content: "base content"},
					"override_script": {Content: "override content"},
				},
			},
		},
		{
			name: "override same name",
			base: types.Resources{
				Scripts: map[string]types.Script{
					"same_script": {Content: "base content"},
				},
			},
			override: types.Resources{
				Scripts: map[string]types.Script{
					"same_script": {Content: "override content"},
				},
			},
			expected: types.Resources{
				Certificates: map[string]types.Certificate{},
				Scripts: map[string]types.Script{
					"same_script": {Content: "override content"},
				},
			},
		},
		{
			name: "nil base certificates",
			base: types.Resources{
				Certificates: nil,
				Scripts: map[string]types.Script{
					"base_script": {Content: "base"},
				},
			},
			override: types.Resources{
				Certificates: map[string]types.Certificate{
					"new_cert": {Certificate: "new"},
				},
				Scripts: nil,
			},
			expected: types.Resources{
				Certificates: map[string]types.Certificate{
					"new_cert": {Certificate: "new"},
				},
				Scripts: map[string]types.Script{
					"base_script": {Content: "base"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeResources(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeContainerRegistry(t *testing.T) {
	tests := []struct {
		name     string
		base     types.ContainerRegistry
		override types.ContainerRegistry
		expected types.ContainerRegistry
	}{
		{
			name: "override both username and password",
			base: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "base_user", Password: "base_pass"},
			},
			override: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "override_user", Password: "override_pass"},
			},
			expected: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "override_user", Password: "override_pass"},
			},
		},
		{
			name: "override only username",
			base: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "base_user", Password: "base_pass"},
			},
			override: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "override_user", Password: ""},
			},
			expected: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "override_user", Password: "base_pass"},
			},
		},
		{
			name: "override only password",
			base: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "base_user", Password: "base_pass"},
			},
			override: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "", Password: "override_pass"},
			},
			expected: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "base_user", Password: "override_pass"},
			},
		},
		{
			name: "no override",
			base: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "base_user", Password: "base_pass"},
			},
			override: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "", Password: ""},
			},
			expected: types.ContainerRegistry{
				Auth: types.ContainerRegistryAuth{Username: "base_user", Password: "base_pass"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeContainerRegistry(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test all merge environment methods with 100% coverage
func TestCommonMaker_MergeCommonEnvironments(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	resourceMakerMock := resourcesMocks.NewMockMaker(t)
	maker := CreateCommonMaker(fndMock, resourceMakerMock)

	tests := []struct {
		name     string
		base     *types.CommonEnvironment
		override *types.CommonEnvironment
		expected *types.CommonEnvironment
	}{
		{
			name: "merge with both environments",
			base: &types.CommonEnvironment{
				Ports:     types.EnvironmentPorts{Start: 3000, End: 4000},
				Resources: types.Resources{Scripts: map[string]types.Script{"base": {Content: "base"}}},
			},
			override: &types.CommonEnvironment{
				Ports:     types.EnvironmentPorts{Start: 5000, End: 0},
				Resources: types.Resources{Scripts: map[string]types.Script{"override": {Content: "override"}}},
			},
			expected: &types.CommonEnvironment{
				Ports: types.EnvironmentPorts{Start: 5000, End: 4000},
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{},
					Scripts: map[string]types.Script{
						"base":     {Content: "base"},
						"override": {Content: "override"},
					},
				},
			},
		},
		{
			name:     "nil base",
			base:     nil,
			override: &types.CommonEnvironment{Ports: types.EnvironmentPorts{Start: 5000}},
			expected: &types.CommonEnvironment{Ports: types.EnvironmentPorts{Start: 5000}},
		},
		{
			name:     "nil override",
			base:     &types.CommonEnvironment{Ports: types.EnvironmentPorts{Start: 3000}},
			override: nil,
			expected: &types.CommonEnvironment{Ports: types.EnvironmentPorts{Start: 3000}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maker.MergeCommonEnvironments(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommonMaker_MergeLocalEnvironments(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	resourceMakerMock := resourcesMocks.NewMockMaker(t)
	maker := CreateCommonMaker(fndMock, resourceMakerMock)

	tests := []struct {
		name     string
		base     *types.LocalEnvironment
		override *types.LocalEnvironment
		expected *types.LocalEnvironment
	}{
		{
			name: "merge with both environments",
			base: &types.LocalEnvironment{
				Ports:     types.EnvironmentPorts{Start: 3000, End: 4000},
				Resources: types.Resources{Scripts: map[string]types.Script{"base": {Content: "base"}}},
			},
			override: &types.LocalEnvironment{
				Ports:     types.EnvironmentPorts{Start: 5000, End: 0},
				Resources: types.Resources{Scripts: map[string]types.Script{"override": {Content: "override"}}},
			},
			expected: &types.LocalEnvironment{
				Ports: types.EnvironmentPorts{Start: 5000, End: 4000},
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{},
					Scripts: map[string]types.Script{
						"base":     {Content: "base"},
						"override": {Content: "override"},
					},
				},
			},
		},
		{
			name:     "nil base",
			base:     nil,
			override: &types.LocalEnvironment{Ports: types.EnvironmentPorts{Start: 5000}},
			expected: &types.LocalEnvironment{Ports: types.EnvironmentPorts{Start: 5000}},
		},
		{
			name:     "nil override",
			base:     &types.LocalEnvironment{Ports: types.EnvironmentPorts{Start: 3000}},
			override: nil,
			expected: &types.LocalEnvironment{Ports: types.EnvironmentPorts{Start: 3000}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maker.MergeLocalEnvironments(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommonMaker_MergeContainerEnvironments(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	resourceMakerMock := resourcesMocks.NewMockMaker(t)
	maker := CreateCommonMaker(fndMock, resourceMakerMock)

	tests := []struct {
		name     string
		base     *types.ContainerEnvironment
		override *types.ContainerEnvironment
		expected *types.ContainerEnvironment
	}{
		{
			name: "merge with both environments",
			base: &types.ContainerEnvironment{
				Ports:     types.EnvironmentPorts{Start: 3000, End: 4000},
				Resources: types.Resources{Scripts: map[string]types.Script{"base": {Content: "base"}}},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{Username: "base_user", Password: "base_pass"},
				},
			},
			override: &types.ContainerEnvironment{
				Ports:     types.EnvironmentPorts{Start: 5000, End: 0},
				Resources: types.Resources{Scripts: map[string]types.Script{"override": {Content: "override"}}},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{Username: "override_user", Password: ""},
				},
			},
			expected: &types.ContainerEnvironment{
				Ports: types.EnvironmentPorts{Start: 5000, End: 4000},
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{},
					Scripts: map[string]types.Script{
						"base":     {Content: "base"},
						"override": {Content: "override"},
					},
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{Username: "override_user", Password: "base_pass"},
				},
			},
		},
		{
			name:     "nil base",
			base:     nil,
			override: &types.ContainerEnvironment{Ports: types.EnvironmentPorts{Start: 5000}},
			expected: &types.ContainerEnvironment{Ports: types.EnvironmentPorts{Start: 5000}},
		},
		{
			name:     "nil override",
			base:     &types.ContainerEnvironment{Ports: types.EnvironmentPorts{Start: 3000}},
			override: nil,
			expected: &types.ContainerEnvironment{Ports: types.EnvironmentPorts{Start: 3000}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maker.MergeContainerEnvironments(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommonMaker_MergeDockerEnvironments(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	resourceMakerMock := resourcesMocks.NewMockMaker(t)
	maker := CreateCommonMaker(fndMock, resourceMakerMock)

	tests := []struct {
		name     string
		base     *types.DockerEnvironment
		override *types.DockerEnvironment
		expected *types.DockerEnvironment
	}{
		{
			name: "merge with both environments and override name prefix",
			base: &types.DockerEnvironment{
				Ports:      types.EnvironmentPorts{Start: 3000, End: 4000},
				Resources:  types.Resources{Scripts: map[string]types.Script{"base": {Content: "base"}}},
				Registry:   types.ContainerRegistry{Auth: types.ContainerRegistryAuth{Username: "base_user", Password: "base_pass"}},
				NamePrefix: "base_prefix",
			},
			override: &types.DockerEnvironment{
				Ports:      types.EnvironmentPorts{Start: 5000, End: 0},
				Resources:  types.Resources{Scripts: map[string]types.Script{"override": {Content: "override"}}},
				Registry:   types.ContainerRegistry{Auth: types.ContainerRegistryAuth{Username: "override_user", Password: ""}},
				NamePrefix: "override_prefix",
			},
			expected: &types.DockerEnvironment{
				Ports: types.EnvironmentPorts{Start: 5000, End: 4000},
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{},
					Scripts: map[string]types.Script{
						"base":     {Content: "base"},
						"override": {Content: "override"},
					},
				},
				Registry:   types.ContainerRegistry{Auth: types.ContainerRegistryAuth{Username: "override_user", Password: "base_pass"}},
				NamePrefix: "override_prefix",
			},
		},
		{
			name: "merge with empty override name prefix",
			base: &types.DockerEnvironment{
				Ports:      types.EnvironmentPorts{Start: 3000, End: 4000},
				NamePrefix: "base_prefix",
			},
			override: &types.DockerEnvironment{
				Ports:      types.EnvironmentPorts{Start: 5000, End: 0},
				NamePrefix: "",
			},
			expected: &types.DockerEnvironment{
				Ports: types.EnvironmentPorts{Start: 5000, End: 4000},
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{},
					Scripts:      map[string]types.Script{},
				},
				Registry:   types.ContainerRegistry{},
				NamePrefix: "base_prefix",
			},
		},
		{
			name:     "nil base",
			base:     nil,
			override: &types.DockerEnvironment{Ports: types.EnvironmentPorts{Start: 5000}, NamePrefix: "override"},
			expected: &types.DockerEnvironment{Ports: types.EnvironmentPorts{Start: 5000}, NamePrefix: "override"},
		},
		{
			name:     "nil override",
			base:     &types.DockerEnvironment{Ports: types.EnvironmentPorts{Start: 3000}, NamePrefix: "base"},
			override: nil,
			expected: &types.DockerEnvironment{Ports: types.EnvironmentPorts{Start: 3000}, NamePrefix: "base"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maker.MergeDockerEnvironments(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommonMaker_MergeKubernetesEnvironments(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	resourceMakerMock := resourcesMocks.NewMockMaker(t)
	maker := CreateCommonMaker(fndMock, resourceMakerMock)

	tests := []struct {
		name     string
		base     *types.KubernetesEnvironment
		override *types.KubernetesEnvironment
		expected *types.KubernetesEnvironment
	}{
		{
			name: "merge with both environments and override namespace and kubeconfig",
			base: &types.KubernetesEnvironment{
				Ports:      types.EnvironmentPorts{Start: 3000, End: 4000},
				Resources:  types.Resources{Scripts: map[string]types.Script{"base": {Content: "base"}}},
				Registry:   types.ContainerRegistry{Auth: types.ContainerRegistryAuth{Username: "base_user", Password: "base_pass"}},
				Namespace:  "base_namespace",
				Kubeconfig: "/base/kubeconfig",
			},
			override: &types.KubernetesEnvironment{
				Ports:      types.EnvironmentPorts{Start: 5000, End: 0},
				Resources:  types.Resources{Scripts: map[string]types.Script{"override": {Content: "override"}}},
				Registry:   types.ContainerRegistry{Auth: types.ContainerRegistryAuth{Username: "override_user", Password: ""}},
				Namespace:  "override_namespace",
				Kubeconfig: "/override/kubeconfig",
			},
			expected: &types.KubernetesEnvironment{
				Ports: types.EnvironmentPorts{Start: 5000, End: 4000},
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{},
					Scripts: map[string]types.Script{
						"base":     {Content: "base"},
						"override": {Content: "override"},
					},
				},
				Registry:   types.ContainerRegistry{Auth: types.ContainerRegistryAuth{Username: "override_user", Password: "base_pass"}},
				Namespace:  "override_namespace",
				Kubeconfig: "/override/kubeconfig",
			},
		},
		{
			name: "merge with empty override namespace and kubeconfig",
			base: &types.KubernetesEnvironment{
				Ports:      types.EnvironmentPorts{Start: 3000, End: 4000},
				Namespace:  "base_namespace",
				Kubeconfig: "/base/kubeconfig",
			},
			override: &types.KubernetesEnvironment{
				Ports:      types.EnvironmentPorts{Start: 5000, End: 0},
				Namespace:  "",
				Kubeconfig: "",
			},
			expected: &types.KubernetesEnvironment{
				Ports: types.EnvironmentPorts{Start: 5000, End: 4000},
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{},
					Scripts:      map[string]types.Script{},
				},
				Registry:   types.ContainerRegistry{},
				Namespace:  "base_namespace",
				Kubeconfig: "/base/kubeconfig",
			},
		},
		{
			name: "override only namespace",
			base: &types.KubernetesEnvironment{
				Namespace:  "base_namespace",
				Kubeconfig: "/base/kubeconfig",
			},
			override: &types.KubernetesEnvironment{
				Namespace:  "override_namespace",
				Kubeconfig: "",
			},
			expected: &types.KubernetesEnvironment{
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{},
					Scripts:      map[string]types.Script{},
				},
				Registry:   types.ContainerRegistry{},
				Namespace:  "override_namespace",
				Kubeconfig: "/base/kubeconfig",
			},
		},
		{
			name: "override only kubeconfig",
			base: &types.KubernetesEnvironment{
				Namespace:  "base_namespace",
				Kubeconfig: "/base/kubeconfig",
			},
			override: &types.KubernetesEnvironment{
				Namespace:  "",
				Kubeconfig: "/override/kubeconfig",
			},
			expected: &types.KubernetesEnvironment{
				Resources: types.Resources{
					Certificates: map[string]types.Certificate{},
					Scripts:      map[string]types.Script{},
				},
				Registry:   types.ContainerRegistry{},
				Namespace:  "base_namespace",
				Kubeconfig: "/override/kubeconfig",
			},
		},
		{
			name:     "nil base",
			base:     nil,
			override: &types.KubernetesEnvironment{Ports: types.EnvironmentPorts{Start: 5000}, Namespace: "override", Kubeconfig: "/override"},
			expected: &types.KubernetesEnvironment{Ports: types.EnvironmentPorts{Start: 5000}, Namespace: "override", Kubeconfig: "/override"},
		},
		{
			name:     "nil override",
			base:     &types.KubernetesEnvironment{Ports: types.EnvironmentPorts{Start: 3000}, Namespace: "base", Kubeconfig: "/base"},
			override: nil,
			expected: &types.KubernetesEnvironment{Ports: types.EnvironmentPorts{Start: 3000}, Namespace: "base", Kubeconfig: "/base"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maker.MergeKubernetesEnvironments(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Additional tests for CommonEnvironment methods to achieve 100% coverage
func TestCommonEnvironment_Methods(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	outputMaker := outputMocks.NewMockMaker(t)
	mockResources := &resources.Resources{}

	env := &CommonEnvironment{
		Fnd:         fndMock,
		OutputMaker: outputMaker,
		Used:        false,
		Ports: Ports{
			Start: 3000,
			Used:  3000,
			End:   4000,
		},
		EnvResources: mockResources,
	}

	// Test initial state
	assert.False(t, env.IsUsed())
	assert.Equal(t, int32(3000), env.PortsStart())
	assert.Equal(t, int32(4000), env.PortsEnd())
	assert.Equal(t, mockResources, env.Resources())
	assert.Nil(t, env.ContainerRegistry())

	// Test MarkUsed
	env.MarkUsed()
	assert.True(t, env.IsUsed())

	// Test ReservePort
	port1 := env.ReservePort()
	assert.Equal(t, int32(3000), port1)
	port2 := env.ReservePort()
	assert.Equal(t, int32(3001), port2)
	assert.Equal(t, int32(3002), env.Ports.Used)
}

func TestContainerEnvironment_ContainerRegistry(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	outputMaker := outputMocks.NewMockMaker(t)
	mockResources := &resources.Resources{}

	env := &ContainerEnvironment{
		CommonEnvironment: CommonEnvironment{
			Fnd:          fndMock,
			OutputMaker:  outputMaker,
			EnvResources: mockResources,
		},
		Registry: ContainerRegistry{
			Auth: ContainerRegistryAuth{
				Username: "test_user",
				Password: "test_pass",
			},
		},
	}

	registry := env.ContainerRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, "test_user", registry.Auth.Username)
	assert.Equal(t, "test_pass", registry.Auth.Password)
}

// Test CreateCommonMaker function for full coverage
func TestCreateCommonMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	resourceMakerMock := resourcesMocks.NewMockMaker(t)

	maker := CreateCommonMaker(fndMock, resourceMakerMock)

	assert.NotNil(t, maker)
	assert.Equal(t, fndMock, maker.Fnd)
	assert.Equal(t, resourceMakerMock, maker.ResourcesMaker)
	assert.NotNil(t, maker.OutputMaker)
}
