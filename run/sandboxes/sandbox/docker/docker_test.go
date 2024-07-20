package docker

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	containerMocks "github.com/bukka/wst/mocks/generated/run/sandboxes/sandbox/container"
	"github.com/bukka/wst/run/sandboxes/containers"
	"github.com/bukka/wst/run/sandboxes/dir"
	"github.com/bukka/wst/run/sandboxes/hooks"
	"github.com/bukka/wst/run/sandboxes/sandbox/common"
	"github.com/bukka/wst/run/sandboxes/sandbox/container"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_nativeMaker_MakeSandbox(t *testing.T) {
	tests := []struct {
		name             string
		inputConfig      *types.DockerSandbox
		mockSetup        func(*containerMocks.MockMaker) *container.Sandbox
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful sandbox creation",
			inputConfig: &types.DockerSandbox{
				Available: true,
				Dirs: map[string]string{
					"conf": "/etc/app",
					"run":  "/var/run/app",
				},
				Hooks: map[string]types.SandboxHook{
					"start": &types.SandboxHookNative{Enabled: true},
				},
				Image: types.ContainerImage{
					Name: "nginx",
					Tag:  "latest",
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "user",
						Password: "pass",
					},
				},
			},
			mockSetup: func(m *containerMocks.MockMaker) *container.Sandbox {
				containerSandbox := container.CreateSandbox(
					common.CreateSandbox(true, make(map[dir.DirType]string), make(hooks.Hooks)),
					&containers.ContainerConfig{
						ImageName:        "nginx",
						ImageTag:         "latest",
						RegistryUsername: "user",
						RegistryPassword: "pass",
					},
				)
				m.On("MakeSandbox", &types.ContainerSandbox{
					Available: true,
					Dirs: map[string]string{
						"conf": "/etc/app",
						"run":  "/var/run/app",
					},
					Hooks: map[string]types.SandboxHook{
						"start": &types.SandboxHookNative{Enabled: true},
					},
					Image: types.ContainerImage{
						Name: "nginx",
						Tag:  "latest",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "user",
							Password: "pass",
						},
					},
				}).Return(containerSandbox, nil)
				return containerSandbox
			},
		},
		{
			name: "failed sandbox creation due to container maker error",

			inputConfig: &types.DockerSandbox{
				Available: true,
				Dirs: map[string]string{
					"conf": "/etc/app",
					"run":  "/var/run/app",
				},
				Hooks: map[string]types.SandboxHook{
					"start": &types.SandboxHookNative{Enabled: true},
				},
				Image: types.ContainerImage{
					Name: "nginx",
					Tag:  "latest",
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "user",
						Password: "pass",
					},
				},
			},
			mockSetup: func(m *containerMocks.MockMaker) *container.Sandbox {
				m.On("MakeSandbox", &types.ContainerSandbox{
					Available: true,
					Dirs: map[string]string{
						"conf": "/etc/app",
						"run":  "/var/run/app",
					},
					Hooks: map[string]types.SandboxHook{
						"start": &types.SandboxHookNative{Enabled: true},
					},
					Image: types.ContainerImage{
						Name: "nginx",
						Tag:  "latest",
					},
					Registry: types.ContainerRegistry{
						Auth: types.ContainerRegistryAuth{
							Username: "user",
							Password: "pass",
						},
					},
				}).Return(nil, errors.New("make fail"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "make fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			containerMakerMock := containerMocks.NewMockMaker(t)
			containerSandbox := tt.mockSetup(containerMakerMock)

			maker := CreateMaker(fndMock, containerMakerMock)
			result, err := maker.MakeSandbox(tt.inputConfig)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				expectedSandbox := &Sandbox{
					Sandbox: *containerSandbox,
				}
				assert.Equal(t, expectedSandbox, result)
			}

			containerMakerMock.AssertExpectations(t)
		})
	}
}
