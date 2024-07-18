package common

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	hooksMocks "github.com/bukka/wst/mocks/generated/run/sandboxes/hooks"
	"github.com/bukka/wst/run/sandboxes/dir"
	"github.com/bukka/wst/run/sandboxes/hooks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	expectedHooks := map[hooks.HookType]hooks.Hook{
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
	parentDirs := map[dir.DirType]string{
		dir.ConfDirType: "/path/to/conf",
		dir.RunDirType:  "/path/to/parent/run",
	}
	parentHooks := map[hooks.HookType]hooks.Hook{
		hooks.StartHookType: &hooks.HookNative{},
	}
	parentSandbox := &Sandbox{
		dirs:  parentDirs,
		hooks: parentHooks,
	}

	childDirs := map[dir.DirType]string{
		dir.ScriptDirType: "/path/to/scripts",
	}
	childHooks := make(map[hooks.HookType]hooks.Hook)
	childSandbox := &Sandbox{
		dirs:  childDirs,
		hooks: childHooks,
	}

	err := childSandbox.Inherit(parentSandbox)
	require.NoError(t, err, "Inheriting from parent sandbox should not produce an error")

	assert.Equal(t, parentDirs[dir.ConfDirType], childSandbox.dirs[dir.ConfDirType], "Child should inherit conf dir from parent")
	assert.Equal(t, childDirs[dir.ScriptDirType], childSandbox.dirs[dir.ScriptDirType], "Child should keep its script dir")
	assert.Equal(t, parentHooks[hooks.StartHookType], childSandbox.hooks[hooks.StartHookType], "Child should inherit start hook from parent")
}
