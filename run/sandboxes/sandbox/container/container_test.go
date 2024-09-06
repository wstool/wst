package container

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/generated/app"
	hooksMocks "github.com/wstool/wst/mocks/generated/run/sandboxes/hooks"
	sandboxMocks "github.com/wstool/wst/mocks/generated/run/sandboxes/sandbox"
	commonMocks "github.com/wstool/wst/mocks/generated/run/sandboxes/sandbox/common"
	"github.com/wstool/wst/run/sandboxes/containers"
	"github.com/wstool/wst/run/sandboxes/dir"
	"github.com/wstool/wst/run/sandboxes/hooks"
	"github.com/wstool/wst/run/sandboxes/sandbox/common"
	"testing"
)

func Test_nativeMaker_MakeSandbox(t *testing.T) {
	tests := []struct {
		name            string
		containerConfig *types.ContainerSandbox
		setupMocks      func(*commonMocks.MockMaker)
		expectError     bool
		expectedErr     string
	}{
		{
			name: "successful sandbox creation",
			containerConfig: &types.ContainerSandbox{
				Available: true,
				Dirs: map[string]string{
					"conf": "/conf",
					"run":  "/run",
				},
				Hooks: map[string]types.SandboxHook{
					"start": &types.SandboxHookShellCommand{Command: "echo 'starting'"},
				},
				Image: types.ContainerImage{
					Name: "nginx",
					Tag:  "latest",
				},
				Registry: types.ContainerRegistry{
					Auth: types.ContainerRegistryAuth{
						Username: "testuser",
						Password: "testpass",
					},
				},
			},
			setupMocks: func(cm *commonMocks.MockMaker) {
				commonSandbox := &common.Sandbox{}
				cm.On("MakeSandbox", &types.CommonSandbox{
					Dirs: map[string]string{
						"conf": "/conf",
						"run":  "/run",
					},
					Hooks: map[string]types.SandboxHook{
						"start": &types.SandboxHookShellCommand{Command: "echo 'starting'"},
					},
					Available: true,
				}).Return(commonSandbox, nil)
			},
		},
		{
			name: "error during common sandbox creation",
			containerConfig: &types.ContainerSandbox{
				Available: true,
			},
			setupMocks: func(cm *commonMocks.MockMaker) {
				cm.On("MakeSandbox", mock.AnythingOfType("*types.CommonSandbox")).Return(
					nil, errors.New("common sandbox creation failed"))
			},
			expectError: true,
			expectedErr: "common sandbox creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := app.NewMockFoundation(t)
			commonMakerMock := commonMocks.NewMockMaker(t)
			tt.setupMocks(commonMakerMock)

			maker := CreateMaker(fndMock, commonMakerMock)
			result, err := maker.MakeSandbox(tt.containerConfig)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				// Check that the container config was correctly propagated to the sandbox.
				assert.Equal(t, tt.containerConfig.Image.Name+":"+tt.containerConfig.Image.Tag, result.config.Image())
				assert.Equal(t, tt.containerConfig.Registry.Auth.Username, result.config.RegistryUsername)
				assert.Equal(t, tt.containerConfig.Registry.Auth.Password, result.config.RegistryPassword)
			}

			commonMakerMock.AssertExpectations(t)
		})
	}
}

func TestSandbox_ContainerConfig(t *testing.T) {
	sandbox := Sandbox{
		config: containers.ContainerConfig{
			ImageName:        "nginx",
			ImageTag:         "latest",
			RegistryUsername: "user",
			RegistryPassword: "pass",
		},
	}

	config := sandbox.ContainerConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "nginx", config.ImageName)
	assert.Equal(t, "latest", config.ImageTag)
	assert.Equal(t, "user", config.RegistryUsername)
	assert.Equal(t, "pass", config.RegistryPassword)
}

func TestSandbox_Inherit(t *testing.T) {
	tests := []struct {
		name             string
		childSandbox     *Sandbox
		parentHooks      hooks.Hooks
		parentDirs       map[dir.DirType]string
		parentConfig     *containers.ContainerConfig
		expectedHooks    hooks.Hooks
		expectedDirs     map[dir.DirType]string
		expectedConfig   *containers.ContainerConfig
		setupParentMocks func(
			*sandboxMocks.MockSandbox,
			hooks.Hooks,
			map[dir.DirType]string,
			*containers.ContainerConfig,
		)
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "Unset child sandbox inheritance",
			childSandbox: &Sandbox{
				Sandbox: *common.CreateSandbox(true, make(map[dir.DirType]string), make(hooks.Hooks)),
			},
			parentHooks: hooks.Hooks{
				hooks.StartHookType: hooksMocks.NewMockHook(t),
				hooks.StopHookType:  hooksMocks.NewMockHook(t),
			},
			parentDirs: map[dir.DirType]string{
				dir.ConfDirType: "/etc/app",
				dir.RunDirType:  "/var/run/app",
			},
			parentConfig: &containers.ContainerConfig{
				ImageName:        "ubuntu",
				ImageTag:         "latest",
				RegistryUsername: "user",
				RegistryPassword: "pass",
			},
			expectedHooks: hooks.Hooks{
				hooks.StartHookType: hooksMocks.NewMockHook(t),
				hooks.StopHookType:  hooksMocks.NewMockHook(t),
			},
			expectedDirs: map[dir.DirType]string{
				dir.ConfDirType: "/etc/app",
				dir.RunDirType:  "/var/run/app",
			},
			expectedConfig: &containers.ContainerConfig{
				ImageName:        "ubuntu",
				ImageTag:         "latest",
				RegistryUsername: "user",
				RegistryPassword: "pass",
			},
			setupParentMocks: func(
				parent *sandboxMocks.MockSandbox,
				hooks hooks.Hooks,
				dirs map[dir.DirType]string,
				config *containers.ContainerConfig,
			) {
				parent.On("Hooks").Return(hooks)
				parent.On("Dirs").Return(dirs)
				parent.On("ContainerConfig").Return(config)
			},
		},
		{
			name: "Set child sandbox inheritance",
			childSandbox: &Sandbox{
				Sandbox: *common.CreateSandbox(
					true,
					map[dir.DirType]string{
						dir.ScriptDirType: "/usr/local/bin",
					},
					hooks.Hooks{
						hooks.RestartHookType: hooksMocks.NewMockHook(t),
					},
				),
				config: containers.ContainerConfig{
					ImageName:        "img",
					ImageTag:         "1.0",
					RegistryUsername: "u1",
					RegistryPassword: "p1",
				},
			},
			parentHooks: hooks.Hooks{
				hooks.StartHookType: hooksMocks.NewMockHook(t),
				hooks.StopHookType:  hooksMocks.NewMockHook(t),
			},
			parentDirs: map[dir.DirType]string{
				dir.ConfDirType: "/etc/app",
				dir.RunDirType:  "/var/run/app",
			},
			parentConfig: &containers.ContainerConfig{
				ImageName:        "ubuntu",
				ImageTag:         "latest",
				RegistryUsername: "user",
				RegistryPassword: "pass",
			},
			expectedHooks: hooks.Hooks{
				hooks.RestartHookType: hooksMocks.NewMockHook(t),
				hooks.StartHookType:   hooksMocks.NewMockHook(t),
				hooks.StopHookType:    hooksMocks.NewMockHook(t),
			},
			expectedDirs: map[dir.DirType]string{
				dir.ScriptDirType: "/usr/local/bin",
				dir.ConfDirType:   "/etc/app",
				dir.RunDirType:    "/var/run/app",
			},
			expectedConfig: &containers.ContainerConfig{
				ImageName:        "img",
				ImageTag:         "1.0",
				RegistryUsername: "u1",
				RegistryPassword: "p1",
			},
			setupParentMocks: func(
				parent *sandboxMocks.MockSandbox,
				hooks hooks.Hooks,
				dirs map[dir.DirType]string,
				config *containers.ContainerConfig,
			) {
				parent.On("Hooks").Return(hooks)
				parent.On("Dirs").Return(dirs)
				parent.On("ContainerConfig").Return(config)
			},
		},
		{
			name: "Errors on child sandbox inheritance",
			childSandbox: &Sandbox{
				Sandbox: *common.CreateSandbox(
					true,
					nil,
					nil,
				),
			},
			parentHooks: hooks.Hooks{
				hooks.StartHookType: hooksMocks.NewMockHook(t),
				hooks.StopHookType:  hooksMocks.NewMockHook(t),
			},
			parentDirs: map[dir.DirType]string{
				dir.ConfDirType: "/etc/app",
				dir.RunDirType:  "/var/run/app",
			},
			parentConfig: &containers.ContainerConfig{
				ImageName:        "ubuntu",
				ImageTag:         "latest",
				RegistryUsername: "user",
				RegistryPassword: "pass",
			},
			expectedHooks: hooks.Hooks{
				hooks.RestartHookType: hooksMocks.NewMockHook(t),
				hooks.StartHookType:   hooksMocks.NewMockHook(t),
				hooks.StopHookType:    hooksMocks.NewMockHook(t),
			},
			expectedDirs: map[dir.DirType]string{
				dir.ScriptDirType: "/usr/local/bin",
				dir.ConfDirType:   "/etc/app",
				dir.RunDirType:    "/var/run/app",
			},
			expectedConfig: &containers.ContainerConfig{
				ImageName:        "img",
				ImageTag:         "1.0",
				RegistryUsername: "u1",
				RegistryPassword: "p1",
			},
			expectedError:    true,
			expectedErrorMsg: "sandbox hooks not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentSandbox := sandboxMocks.NewMockSandbox(t)
			if tt.setupParentMocks != nil {
				tt.setupParentMocks(parentSandbox, tt.parentHooks, tt.parentDirs, tt.parentConfig)
			}

			err := tt.childSandbox.Inherit(parentSandbox)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHooks, tt.childSandbox.Hooks())
				assert.Equal(t, tt.expectedDirs, tt.childSandbox.Dirs())
				assert.Equal(t, tt.expectedConfig, tt.childSandbox.ContainerConfig())
			}
		})
	}
}
