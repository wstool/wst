package sandboxes

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	sandboxMocks "github.com/bukka/wst/mocks/generated/run/sandboxes/sandbox"
	dockerMocks "github.com/bukka/wst/mocks/generated/run/sandboxes/sandbox/docker"
	kubernetesMocks "github.com/bukka/wst/mocks/generated/run/sandboxes/sandbox/kubernetes"
	localMocks "github.com/bukka/wst/mocks/generated/run/sandboxes/sandbox/local"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/sandboxes/hooks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSandboxes_Inherit(t *testing.T) {
	tests := []struct {
		name              string
		childSandboxes    Sandboxes
		parentSandboxes   Sandboxes
		setupExpectations func(childMocks, parentMocks map[providers.Type]*sandboxMocks.MockSandbox)
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "successful inheritance",
			childSandboxes: Sandboxes{
				providers.LocalType:  sandboxMocks.NewMockSandbox(t),
				providers.DockerType: sandboxMocks.NewMockSandbox(t),
			},
			parentSandboxes: Sandboxes{
				providers.LocalType:      sandboxMocks.NewMockSandbox(t),
				providers.KubernetesType: sandboxMocks.NewMockSandbox(t),
			},
			setupExpectations: func(childMocks, parentMocks map[providers.Type]*sandboxMocks.MockSandbox) {
				// Set expectations for inheritance
				childMocks[providers.LocalType].On("Inherit", parentMocks[providers.LocalType]).Return(nil)
				// As Kubernetes is not in child, it should not set any expectations
				// Docker does not have a parent counterpart, so no inheritance should be called either
			},
		},
		{
			name: "error during inheritance",
			childSandboxes: Sandboxes{
				providers.LocalType: sandboxMocks.NewMockSandbox(t),
			},
			parentSandboxes: Sandboxes{
				providers.LocalType: sandboxMocks.NewMockSandbox(t),
			},
			setupExpectations: func(childMocks, parentMocks map[providers.Type]*sandboxMocks.MockSandbox) {
				childMocks[providers.LocalType].On("Inherit", parentMocks[providers.LocalType]).Return(errors.New("inheritance failed"))
			},
			expectError:      true,
			expectedErrorMsg: "inheritance failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			childMocks := map[providers.Type]*sandboxMocks.MockSandbox{}
			parentMocks := map[providers.Type]*sandboxMocks.MockSandbox{}
			for typ, sb := range tt.childSandboxes {
				childMocks[typ] = sb.(*sandboxMocks.MockSandbox)
			}
			for typ, sb := range tt.parentSandboxes {
				parentMocks[typ] = sb.(*sandboxMocks.MockSandbox)
			}
			tt.setupExpectations(childMocks, parentMocks)

			// Execute the inherit method
			err := tt.childSandboxes.Inherit(tt.parentSandboxes)

			// Assert expectations and check for errors
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Validate all mocks' expectations
			for _, m := range childMocks {
				m.AssertExpectations(t)
			}
			for _, m := range parentMocks {
				m.AssertExpectations(t)
			}
		})
	}
}

func Test_nativeMaker_MakeSandboxes(t *testing.T) {
	tests := []struct {
		name         string
		specConfig   map[string]types.Sandbox
		serverConfig map[string]types.Sandbox
		setupMocks   func(
			*testing.T,
			*localMocks.MockMaker,
			*dockerMocks.MockMaker,
			*kubernetesMocks.MockMaker,
		) Sandboxes
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful merge and sandbox creation",
			specConfig: map[string]types.Sandbox{
				"local": &types.LocalSandbox{
					Available: true,
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/local",
						"run":    "/wst/var/wst/run",
						"script": "/var/www",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "c1"},
					},
				},
				"docker": &types.DockerSandbox{
					Available: true,
					Dirs: map[string]string{
						"conf": "/etc/wst/docker",
						"run":  "/var/wst/run/docker",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "c2"},
					},
				},
				"kubernetes": &types.KubernetesSandbox{
					Available: true,
					Dirs: map[string]string{
						"conf": "/etc/wst/kubernetes",
						"run":  "/var/wst/run/kubernetes",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "c3"},
					},
				},
			},
			serverConfig: map[string]types.Sandbox{
				"local": &types.LocalSandbox{
					Dirs: map[string]string{
						"conf": "/wst/etc/wst/server-local",
						"run":  "/wst/var/wst/server-run",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "cs"},
					},
				},
			},
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Sandboxes {
				// Setup local sandbox maker
				localEnv := sandboxMocks.NewMockSandbox(t)
				localMock.On(
					"MakeSandbox",
					&types.LocalSandbox{
						Available: false,
						Dirs: map[string]string{
							"conf":   "/wst/etc/wst/server-local",
							"run":    "/wst/var/wst/server-run",
							"script": "/var/www",
						},
						Hooks: map[string]types.SandboxHook{
							"start": hooks.HookShellCommand{Command: "cs"},
						},
					},
				).Return(localEnv, nil)

				// Setup Docker sandbox maker
				dockerEnv := sandboxMocks.NewMockSandbox(t)
				dockerMock.On(
					"MakeSandbox",
					&types.DockerSandbox{
						Available: true,
						Dirs: map[string]string{
							"conf": "/etc/wst/docker",
							"run":  "/var/wst/run/docker",
						},
						Hooks: map[string]types.SandboxHook{
							"start": hooks.HookShellCommand{Command: "c2"},
						},
					},
				).Return(dockerEnv, nil)

				// Setup Kubernetes sandbox maker
				kubernetesEnv := sandboxMocks.NewMockSandbox(t)
				kubernetesMock.On(
					"MakeSandbox",
					&types.KubernetesSandbox{
						Available: true,
						Dirs: map[string]string{
							"conf": "/etc/wst/kubernetes",
							"run":  "/var/wst/run/kubernetes",
						},
						Hooks: map[string]types.SandboxHook{
							"start": hooks.HookShellCommand{Command: "c3"},
						},
					},
				).Return(kubernetesEnv, nil)

				return Sandboxes{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubernetesEnv,
				}
			},
		},
		{
			name: "successful merge of more sandboxes",
			specConfig: map[string]types.Sandbox{
				"common": &types.CommonSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/common",
						"run":    "/wst/var/wst/run/common",
						"script": "/var/www",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "com1"},
						"restart": hooks.HookShellCommand{Command: "com2"},
					},
				},
				"container": &types.ContainerSandbox{
					Dirs: map[string]string{
						"conf": "/wst/etc/wst/container",
						"run":  "/wst/var/wst/run/container",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "con1"},
						"stop":  hooks.HookShellCommand{Command: "con2"},
					},
					Image: types.ContainerImage{
						Name: "ci1",
						Tag:  "1.0",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "test",
							Password: "pwd",
						},
					},
				},
				"docker": &types.DockerSandbox{
					Dirs: map[string]string{
						"conf": "/wst/etc/wst/docker",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "d1"},
					},
					Image: types.ContainerImage{
						Tag: "1.2",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "user",
							Password: "1234",
						},
					},
				},
				"kubernetes": &types.KubernetesSandbox{},
			},
			serverConfig: map[string]types.Sandbox{
				"local": &types.LocalSandbox{
					Available: true,
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/local",
						"run":    "/wst/var/wst/run",
						"script": "/var/www/x",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "cs"},
					},
				},
			},
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Sandboxes {
				// Setup local sandbox maker
				localEnv := sandboxMocks.NewMockSandbox(t)
				localMock.On(
					"MakeSandbox",
					&types.LocalSandbox{
						Available: true,
						Dirs: map[string]string{
							"conf":   "/wst/etc/wst/local",
							"run":    "/wst/var/wst/run",
							"script": "/var/www/x",
						},
						Hooks: map[string]types.SandboxHook{
							"start":   hooks.HookShellCommand{Command: "cs"},
							"restart": hooks.HookShellCommand{Command: "com2"},
						},
					},
				).Return(localEnv, nil)

				// Setup Docker sandbox maker
				dockerEnv := sandboxMocks.NewMockSandbox(t)
				dockerMock.On(
					"MakeSandbox",
					&types.DockerSandbox{
						Dirs: map[string]string{
							"conf":   "/wst/etc/wst/docker",
							"run":    "/wst/var/wst/run/container",
							"script": "/var/www",
						},
						Hooks: map[string]types.SandboxHook{
							"start":   hooks.HookShellCommand{Command: "d1"},
							"stop":    hooks.HookShellCommand{Command: "con2"},
							"restart": hooks.HookShellCommand{Command: "com2"},
						},
						Image: types.ContainerImage{
							Name: "ci1",
							Tag:  "1.2",
						},
						Registry: types.ContainerRegistry{
							Auth: types.ContainerRegistryAuth{
								Username: "user",
								Password: "1234",
							},
						},
					},
				).Return(dockerEnv, nil)

				// Setup Kubernetes sandbox maker
				kubernetesEnv := sandboxMocks.NewMockSandbox(t)
				kubernetesMock.On(
					"MakeSandbox",
					&types.KubernetesSandbox{
						Dirs: map[string]string{
							"conf":   "/wst/etc/wst/container",
							"run":    "/wst/var/wst/run/container",
							"script": "/var/www",
						},
						Hooks: map[string]types.SandboxHook{
							"start":   hooks.HookShellCommand{Command: "con1"},
							"stop":    hooks.HookShellCommand{Command: "con2"},
							"restart": hooks.HookShellCommand{Command: "com2"},
						},
						Image: types.ContainerImage{
							Name: "ci1",
							Tag:  "1.0",
						},
						Registry: types.ContainerRegistry{
							Auth: types.ContainerRegistryAuth{
								Username: "test",
								Password: "pwd",
							},
						},
					},
				).Return(kubernetesEnv, nil)

				return Sandboxes{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubernetesEnv,
				}
			},
		},
		{
			name: "successful merge of all sandboxes",
			specConfig: map[string]types.Sandbox{
				"common": &types.CommonSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/common",
						"run":    "/wst/var/wst/run/common",
						"script": "/var/www",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "com1"},
						"restart": hooks.HookShellCommand{Command: "com2"},
					},
				},
				"container": &types.ContainerSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/container",
						"run":    "/wst/var/wst/run/container",
						"script": "/var/www/container",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "con1"},
						"stop":  hooks.HookShellCommand{Command: "con2"},
					},
					Image: types.ContainerImage{
						Name: "ci1",
						Tag:  "1.0",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "test",
							Password: "pwd",
						},
					},
				},
				"docker": &types.DockerSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/docker1",
						"run":    "/wst/run/wst/docker1",
						"script": "/var/www/docker1",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "d1"},
						"stop":    hooks.HookShellCommand{Command: "d2"},
						"restart": hooks.HookShellCommand{Command: "d3"},
					},
					Image: types.ContainerImage{
						Tag: "1.2",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "duser",
							Password: "1234",
						},
					},
				},
				"kubernetes": &types.KubernetesSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/kube",
						"run":    "/wst/run/wst/kube",
						"script": "/var/www/kube",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "d1"},
					},
					Image: types.ContainerImage{
						Tag: "1.2",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "kuser",
							Password: "1234",
						},
					},
				},
				"local": &types.LocalSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/common",
						"run":    "/wst/var/wst/run/common",
						"script": "/var/www",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "com1"},
						"restart": hooks.HookShellCommand{Command: "com2"},
					},
				},
			},
			serverConfig: map[string]types.Sandbox{
				"common": &types.CommonSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/common1",
						"run":    "/wst/var/wst/run/common1",
						"script": "/var/www/1",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "com11"},
						"restart": hooks.HookShellCommand{Command: "com21"},
					},
				},
				"container": &types.ContainerSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/container",
						"run":    "/wst/var/wst/run/container",
						"script": "/var/www/container",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "con11"},
						"stop":  hooks.HookShellCommand{Command: "con21"},
					},
					Image: types.ContainerImage{
						Name: "ci2",
						Tag:  "1.1",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "test",
							Password: "pwd",
						},
					},
				},
				"docker": &types.DockerSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/docker",
						"run":    "/wst/run/wst/docker",
						"script": "/var/www/docker",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "d11"},
						"stop":    hooks.HookShellCommand{Command: "d21"},
						"restart": hooks.HookShellCommand{Command: "d31"},
					},
					Image: types.ContainerImage{
						Tag: "1.3",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "sduser",
							Password: "12345",
						},
					},
				},
				"kubernetes": &types.KubernetesSandbox{
					Dirs: map[string]string{
						"conf":   "/etc/wst/kube",
						"run":    "/var/run/wst/kube",
						"script": "/var/www/server/kube",
					},
					Hooks: map[string]types.SandboxHook{
						"start": hooks.HookShellCommand{Command: "k1"},
					},
					Image: types.ContainerImage{
						Name: "kube",
						Tag:  "1.3",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "skuser",
							Password: "12345",
						},
					},
				},
				"local": &types.LocalSandbox{
					Dirs: map[string]string{
						"conf":   "/etc/wst/local",
						"run":    "/var/wst/run/local",
						"script": "/var/www/local",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "com11"},
						"restart": hooks.HookShellCommand{Command: "com22"},
					},
				},
			},
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Sandboxes {
				// Setup local sandbox maker
				localEnv := sandboxMocks.NewMockSandbox(t)
				localMock.On(
					"MakeSandbox",
					&types.LocalSandbox{
						Dirs: map[string]string{
							"conf":   "/etc/wst/local",
							"run":    "/var/wst/run/local",
							"script": "/var/www/local",
						},
						Hooks: map[string]types.SandboxHook{
							"start":   hooks.HookShellCommand{Command: "com11"},
							"restart": hooks.HookShellCommand{Command: "com22"},
						},
					},
				).Return(localEnv, nil)

				// Setup Docker sandbox maker
				dockerEnv := sandboxMocks.NewMockSandbox(t)
				dockerMock.On(
					"MakeSandbox",
					&types.DockerSandbox{
						Dirs: map[string]string{
							"conf":   "/wst/etc/wst/docker",
							"run":    "/wst/run/wst/docker",
							"script": "/var/www/docker",
						},
						Hooks: map[string]types.SandboxHook{
							"start":   hooks.HookShellCommand{Command: "d11"},
							"stop":    hooks.HookShellCommand{Command: "d21"},
							"restart": hooks.HookShellCommand{Command: "d31"},
						},
						Image: types.ContainerImage{
							Name: "ci2",
							Tag:  "1.3",
						},
						Registry: types.ContainerRegistry{
							Auth: types.ContainerRegistryAuth{
								Username: "sduser",
								Password: "12345",
							},
						},
					},
				).Return(dockerEnv, nil)

				// Setup Kubernetes sandbox maker
				kubernetesEnv := sandboxMocks.NewMockSandbox(t)
				kubernetesMock.On(
					"MakeSandbox",
					&types.KubernetesSandbox{
						Dirs: map[string]string{
							"conf":   "/etc/wst/kube",
							"run":    "/var/run/wst/kube",
							"script": "/var/www/server/kube",
						},
						Hooks: map[string]types.SandboxHook{
							"start":   hooks.HookShellCommand{Command: "k1"},
							"stop":    hooks.HookShellCommand{Command: "con21"},
							"restart": hooks.HookShellCommand{Command: "com21"},
						},
						Image: types.ContainerImage{
							Name: "kube",
							Tag:  "1.3",
						},
						Registry: types.ContainerRegistry{
							Auth: types.ContainerRegistryAuth{
								Username: "skuser",
								Password: "12345",
							},
						},
					},
				).Return(kubernetesEnv, nil)

				return Sandboxes{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubernetesEnv,
				}
			},
		},
		{
			name: "successful merge of common only sandbox",
			specConfig: map[string]types.Sandbox{
				"common": &types.CommonSandbox{
					Dirs: map[string]string{
						"conf":   "/wst/etc/wst/common1",
						"run":    "/wst/var/wst/run/common1",
						"script": "/var/www/1",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "com11"},
						"restart": hooks.HookShellCommand{Command: "com21"},
					},
				},
			},
			serverConfig: map[string]types.Sandbox{},
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Sandboxes {
				// Setup local sandbox maker
				localEnv := sandboxMocks.NewMockSandbox(t)
				localMock.On(
					"MakeSandbox",
					&types.LocalSandbox{
						Dirs: map[string]string{
							"conf":   "/wst/etc/wst/common1",
							"run":    "/wst/var/wst/run/common1",
							"script": "/var/www/1",
						},
						Hooks: map[string]types.SandboxHook{
							"start":   hooks.HookShellCommand{Command: "com11"},
							"restart": hooks.HookShellCommand{Command: "com21"},
						},
					},
				).Return(localEnv, nil)

				// Setup Docker sandbox maker
				dockerEnv := sandboxMocks.NewMockSandbox(t)
				dockerMock.On(
					"MakeSandbox",
					&types.DockerSandbox{
						Dirs: map[string]string{
							"conf":   "/wst/etc/wst/common1",
							"run":    "/wst/var/wst/run/common1",
							"script": "/var/www/1",
						},
						Hooks: map[string]types.SandboxHook{
							"start":   hooks.HookShellCommand{Command: "com11"},
							"restart": hooks.HookShellCommand{Command: "com21"},
						},
					},
				).Return(dockerEnv, nil)

				// Setup Kubernetes sandbox maker
				kubernetesEnv := sandboxMocks.NewMockSandbox(t)
				kubernetesMock.On(
					"MakeSandbox",
					&types.KubernetesSandbox{
						Dirs: map[string]string{
							"conf":   "/wst/etc/wst/common1",
							"run":    "/wst/var/wst/run/common1",
							"script": "/var/www/1",
						},
						Hooks: map[string]types.SandboxHook{
							"start":   hooks.HookShellCommand{Command: "com11"},
							"restart": hooks.HookShellCommand{Command: "com21"},
						},
					},
				).Return(kubernetesEnv, nil)

				return Sandboxes{
					providers.LocalType:      localEnv,
					providers.DockerType:     dockerEnv,
					providers.KubernetesType: kubernetesEnv,
				}
			},
		},
		{
			name: "partial configuration with only local defined",
			specConfig: map[string]types.Sandbox{
				"local": &types.LocalSandbox{
					Dirs: map[string]string{
						"conf":   "/etc/wst/local",
						"run":    "/var/wst/run/local",
						"script": "/var/www/local",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "com11"},
						"restart": hooks.HookShellCommand{Command: "com22"},
					},
				},
			},
			setupMocks: func(t *testing.T, localMock *localMocks.MockMaker, dockerMock *dockerMocks.MockMaker, kubernetesMock *kubernetesMocks.MockMaker) Sandboxes {
				localEnv := sandboxMocks.NewMockSandbox(t)
				localMock.On("MakeSandbox", &types.LocalSandbox{
					Dirs: map[string]string{
						"conf":   "/etc/wst/local",
						"run":    "/var/wst/run/local",
						"script": "/var/www/local",
					},
					Hooks: map[string]types.SandboxHook{
						"start":   hooks.HookShellCommand{Command: "com11"},
						"restart": hooks.HookShellCommand{Command: "com22"},
					},
				}).Return(localEnv, nil)
				return Sandboxes{providers.LocalType: localEnv}
			},
		},
		{
			name: "conflicting dirs resolved by instance config",
			specConfig: map[string]types.Sandbox{
				"local": &types.LocalSandbox{
					Dirs: map[string]string{
						"conf":   "/etc/wst/local1",
						"run":    "/var/wst/run/local1",
						"script": "/var/www/local1",
					},
				},
			},
			serverConfig: map[string]types.Sandbox{
				"local": &types.LocalSandbox{
					Dirs: map[string]string{
						"conf":   "/etc/wst/local",
						"run":    "/var/wst/run/local",
						"script": "/var/www/local",
					},
				},
			},
			setupMocks: func(t *testing.T, localMock *localMocks.MockMaker, dockerMock *dockerMocks.MockMaker, kubernetesMock *kubernetesMocks.MockMaker) Sandboxes {
				localEnv := sandboxMocks.NewMockSandbox(t)
				localMock.On("MakeSandbox", &types.LocalSandbox{
					Dirs: map[string]string{
						"conf":   "/etc/wst/local",
						"run":    "/var/wst/run/local",
						"script": "/var/www/local",
					},
					Hooks: map[string]types.SandboxHook{},
				}).Return(localEnv, nil)
				return Sandboxes{providers.LocalType: localEnv}
			},
		},
		{
			name: "mixed success and errors in docker sandbox creation",
			specConfig: map[string]types.Sandbox{
				"local":  &types.LocalSandbox{},
				"docker": &types.DockerSandbox{},
			},
			setupMocks: func(t *testing.T, localMock *localMocks.MockMaker, dockerMock *dockerMocks.MockMaker, kubernetesMock *kubernetesMocks.MockMaker) Sandboxes {
				localEnv := sandboxMocks.NewMockSandbox(t)
				localMock.On("MakeSandbox", mock.AnythingOfType("*types.LocalSandbox")).Return(localEnv, nil)
				dockerMock.On("MakeSandbox", mock.AnythingOfType("*types.DockerSandbox")).Return(nil, errors.New("docker creation failed"))
				return Sandboxes{providers.LocalType: localEnv}
			},
			expectError:      true,
			expectedErrorMsg: "docker creation failed",
		},
		{
			name: "mixed success and errors in kubernetes sandbox creation",
			specConfig: map[string]types.Sandbox{
				"local":      &types.LocalSandbox{},
				"kubernetes": &types.KubernetesSandbox{},
			},
			setupMocks: func(t *testing.T, localMock *localMocks.MockMaker, dockerMock *dockerMocks.MockMaker, kubernetesMock *kubernetesMocks.MockMaker) Sandboxes {
				localEnv := sandboxMocks.NewMockSandbox(t)
				localMock.On("MakeSandbox", mock.AnythingOfType("*types.LocalSandbox")).Return(localEnv, nil)
				kubernetesMock.On("MakeSandbox", mock.AnythingOfType("*types.KubernetesSandbox")).Return(nil, errors.New("k8s creation failed"))
				return Sandboxes{providers.LocalType: localEnv}
			},
			expectError:      true,
			expectedErrorMsg: "k8s creation failed",
		},
		{
			name: "error during sandbox creation",
			specConfig: map[string]types.Sandbox{
				"local": &types.LocalSandbox{},
			},
			serverConfig: map[string]types.Sandbox{
				"local": &types.LocalSandbox{},
			},
			setupMocks: func(
				t *testing.T,
				localMock *localMocks.MockMaker,
				dockerMock *dockerMocks.MockMaker,
				kubernetesMock *kubernetesMocks.MockMaker,
			) Sandboxes {
				localMock.On(
					"MakeSandbox",
					mock.AnythingOfType("*types.LocalSandbox"),
				).Return(nil, errors.New("local err"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "local err",
		},
		{
			name: "error during common env merging",
			specConfig: map[string]types.Sandbox{
				"common": &types.LocalSandbox{},
			},
			serverConfig: map[string]types.Sandbox{
				"common": &types.LocalSandbox{},
			},
			expectError:      true,
			expectedErrorMsg: "common sandbox is not set in common field",
		},
		{
			name: "error during local env merging",
			specConfig: map[string]types.Sandbox{
				"local": &types.CommonSandbox{},
			},
			serverConfig: map[string]types.Sandbox{
				"local": &types.CommonSandbox{},
			},
			expectError:      true,
			expectedErrorMsg: "local sandbox is not set in local field",
		},
		{
			name: "error during container env merging",
			specConfig: map[string]types.Sandbox{
				"container": &types.LocalSandbox{},
			},
			serverConfig: map[string]types.Sandbox{
				"container": &types.LocalSandbox{},
			},
			expectError:      true,
			expectedErrorMsg: "container sandbox is not set in container field",
		},
		{
			name: "error during docker env merging",
			specConfig: map[string]types.Sandbox{
				"docker": &types.LocalSandbox{},
			},
			serverConfig: map[string]types.Sandbox{
				"docker": &types.LocalSandbox{},
			},
			expectError:      true,
			expectedErrorMsg: "docker sandbox is not set in docker field",
		},
		{
			name: "error during kubernetes env merging",
			specConfig: map[string]types.Sandbox{
				"kubernetes": &types.LocalSandbox{},
			},
			serverConfig: map[string]types.Sandbox{
				"kubernetes": &types.LocalSandbox{},
			},
			expectError:      true,
			expectedErrorMsg: "kubernetes sandbox is not set in kubernetes field",
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

			var expectSandboxes types.Sandbox
			if tt.setupMocks != nil {
				expectSandboxes = tt.setupMocks(t, localMock, dockerMock, kubernetesMock)
			}

			actualSandboxes, err := nm.MakeSandboxes(tt.specConfig, tt.serverConfig)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectSandboxes, actualSandboxes)
			}

			localMock.AssertExpectations(t)
			dockerMock.AssertExpectations(t)
			kubernetesMock.AssertExpectations(t)
		})
	}
}
