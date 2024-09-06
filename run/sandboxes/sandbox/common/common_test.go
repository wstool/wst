package common

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	hooksMocks "github.com/wstool/wst/mocks/generated/run/sandboxes/hooks"
	sandboxMocks "github.com/wstool/wst/mocks/generated/run/sandboxes/sandbox"
	"github.com/wstool/wst/run/sandboxes/dir"
	"github.com/wstool/wst/run/sandboxes/hooks"
	"testing"
)

func Test_nativeMaker_MakeSandbox(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.CommonSandbox
		setupMocks  func(*hooksMocks.MockMaker)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful sandbox creation",
			config: &types.CommonSandbox{
				Available: true,
				Hooks:     map[string]types.SandboxHook{"start": types.SandboxHookNative{Enabled: true}},
				Dirs:      map[string]string{"conf": "/path/to/conf", "run": "/path/to/run"},
			},
			setupMocks: func(hm *hooksMocks.MockMaker) {
				hm.On(
					"MakeHooks",
					map[string]types.SandboxHook{"start": types.SandboxHookNative{Enabled: true}},
				).Return(hooks.Hooks{"start": &hooks.HookNative{}}, nil)
			},
		},
		{
			name: "hook creation error",
			config: &types.CommonSandbox{
				Hooks: map[string]types.SandboxHook{"start": types.SandboxHookNative{Enabled: true}},
			},
			setupMocks: func(hm *hooksMocks.MockMaker) {
				hm.On(
					"MakeHooks",
					map[string]types.SandboxHook{"start": types.SandboxHookNative{Enabled: true}},
				).Return(nil, errors.New("hook error"))
			},
			expectError: true,
			errorMsg:    "hook error",
		},
		{
			name: "invalid dir type",
			config: &types.CommonSandbox{
				Dirs:  map[string]string{"invalid": "/path/to/invalid"},
				Hooks: map[string]types.SandboxHook{"start": types.SandboxHookNative{Enabled: true}},
			},
			setupMocks: func(hm *hooksMocks.MockMaker) {
				hm.On(
					"MakeHooks",
					map[string]types.SandboxHook{"start": types.SandboxHookNative{Enabled: true}},
				).Return(hooks.Hooks{}, nil)
			},
			expectError: true,
			errorMsg:    "invalid dir type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			hooksMakerMock := hooksMocks.NewMockMaker(t)
			tt.setupMocks(hooksMakerMock)

			maker := CreateMaker(fndMock, hooksMakerMock)
			result, err := maker.MakeSandbox(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.config.Available, result.available)
				assert.Len(t, result.dirs, len(tt.config.Dirs))
				assert.Len(t, result.hooks, len(tt.config.Hooks))
			}

			hooksMakerMock.AssertExpectations(t)
		})
	}
}

func TestSandbox_Available(t *testing.T) {
	s := Sandbox{available: true}
	assert.True(t, s.Available(), "Sandbox should be available")
}

func TestSandbox_ContainerConfig(t *testing.T) {
	s := Sandbox{available: true}
	assert.Nil(t, s.ContainerConfig(), "Sandbox should have nil container config")
}

func TestSandbox_Dirs(t *testing.T) {
	expectedDirs := map[dir.DirType]string{
		dir.ConfDirType: "/path/to/conf",
		dir.RunDirType:  "/path/to/run",
	}
	s := Sandbox{dirs: expectedDirs}
	assert.Equal(t, expectedDirs, s.Dirs(), "Directories should match expected values")
}

func TestSandbox_Dir(t *testing.T) {
	expectedDirs := map[dir.DirType]string{
		dir.ConfDirType: "/path/to/conf",
	}
	s := Sandbox{dirs: expectedDirs}

	d, err := s.Dir(dir.ConfDirType)
	assert.NoError(t, err, "Should not have an error getting conf dir")
	assert.Equal(t, "/path/to/conf", d, "Directory should match expected value")

	_, err = s.Dir(dir.RunDirType)
	assert.Error(t, err, "Should error on non-existent dir type")
}

func TestSandbox_Hooks(t *testing.T) {
	expectedHooks := hooks.Hooks{
		hooks.StartHookType: &hooks.HookNative{}, // Example hook
	}
	s := Sandbox{hooks: expectedHooks}
	assert.Equal(t, expectedHooks, s.Hooks(), "Hooks should match expected values")
}

func TestSandbox_Hook(t *testing.T) {
	expectedHooks := map[hooks.HookType]hooks.Hook{
		hooks.StartHookType: &hooks.HookNative{}, // Example hook
	}
	s := Sandbox{hooks: expectedHooks}

	hook, err := s.Hook(hooks.StartHookType)
	assert.NoError(t, err, "Should not have an error getting start hook")
	assert.Equal(t, expectedHooks[hooks.StartHookType], hook, "Hook should match expected hook")

	_, err = s.Hook(hooks.StopHookType)
	assert.Error(t, err, "Should error on non-existent hook type")
}

func TestSandbox_Inherit(t *testing.T) {
	tests := []struct {
		name             string
		childSandbox     *Sandbox
		parentHooks      hooks.Hooks
		parentDirs       map[dir.DirType]string
		expectedHooks    hooks.Hooks
		expectedDirs     map[dir.DirType]string
		setupParentMocks func(
			*sandboxMocks.MockSandbox,
			hooks.Hooks,
			map[dir.DirType]string,
		)
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:         "Unset child sandbox inheritance",
			childSandbox: CreateSandbox(true, make(map[dir.DirType]string), make(hooks.Hooks)),
			parentHooks: hooks.Hooks{
				hooks.StartHookType: hooksMocks.NewMockHook(t),
				hooks.StopHookType:  hooksMocks.NewMockHook(t),
			},
			parentDirs: map[dir.DirType]string{
				dir.ConfDirType: "/etc/app",
				dir.RunDirType:  "/var/run/app",
			},
			expectedHooks: hooks.Hooks{
				hooks.StartHookType: hooksMocks.NewMockHook(t),
				hooks.StopHookType:  hooksMocks.NewMockHook(t),
			},
			expectedDirs: map[dir.DirType]string{
				dir.ConfDirType: "/etc/app",
				dir.RunDirType:  "/var/run/app",
			},
			setupParentMocks: func(
				parent *sandboxMocks.MockSandbox,
				hooks hooks.Hooks,
				dirs map[dir.DirType]string,
			) {
				parent.On("Hooks").Return(hooks)
				parent.On("Dirs").Return(dirs)
			},
		},
		{
			name: "Set child sandbox inheritance",
			childSandbox: CreateSandbox(
				true,
				map[dir.DirType]string{
					dir.ScriptDirType: "/usr/local/bin",
				},
				hooks.Hooks{
					hooks.RestartHookType: hooksMocks.NewMockHook(t),
				},
			),
			parentHooks: hooks.Hooks{
				hooks.StartHookType: hooksMocks.NewMockHook(t),
				hooks.StopHookType:  hooksMocks.NewMockHook(t),
			},
			parentDirs: map[dir.DirType]string{
				dir.ConfDirType: "/etc/app",
				dir.RunDirType:  "/var/run/app",
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
			setupParentMocks: func(
				parent *sandboxMocks.MockSandbox,
				hooks hooks.Hooks,
				dirs map[dir.DirType]string,
			) {
				parent.On("Hooks").Return(hooks)
				parent.On("Dirs").Return(dirs)
			},
		},
		{
			name: "Errors on no hooks",
			childSandbox: CreateSandbox(
				true,
				map[dir.DirType]string{
					dir.ScriptDirType: "/usr/local/bin",
				},
				nil,
			),
			parentHooks: hooks.Hooks{
				hooks.StartHookType: hooksMocks.NewMockHook(t),
				hooks.StopHookType:  hooksMocks.NewMockHook(t),
			},
			parentDirs: map[dir.DirType]string{
				dir.ConfDirType: "/etc/app",
				dir.RunDirType:  "/var/run/app",
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
			expectedError:    true,
			expectedErrorMsg: "sandbox hooks not set",
		},
		{
			name: "Errors on no dirs",
			childSandbox: CreateSandbox(
				true,
				nil,
				hooks.Hooks{
					hooks.RestartHookType: hooksMocks.NewMockHook(t),
				},
			),
			parentHooks: hooks.Hooks{
				hooks.StartHookType: hooksMocks.NewMockHook(t),
				hooks.StopHookType:  hooksMocks.NewMockHook(t),
			},
			parentDirs: map[dir.DirType]string{
				dir.ConfDirType: "/etc/app",
				dir.RunDirType:  "/var/run/app",
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
			expectedError:    true,
			expectedErrorMsg: "sandbox dirs not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentSandbox := sandboxMocks.NewMockSandbox(t)
			if tt.setupParentMocks != nil {
				tt.setupParentMocks(parentSandbox, tt.parentHooks, tt.parentDirs)
			}

			err := tt.childSandbox.Inherit(parentSandbox)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHooks, tt.childSandbox.Hooks())
				assert.Equal(t, tt.expectedDirs, tt.childSandbox.Dirs())
			}
		})
	}
}
