package hooks

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	environmentMocks "github.com/wstool/wst/mocks/generated/run/environments/environment"
	taskMocks "github.com/wstool/wst/mocks/generated/run/environments/task"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters/parameter"
	templateMocks "github.com/wstool/wst/mocks/generated/run/services/template"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/task"
	"github.com/wstool/wst/run/parameters"
	"os"
	"syscall"
	"testing"
)

func Test_nativeMaker_MakeHooks(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]types.SandboxHook
		expectedHooks Hooks
		expectedError string
	}{
		{
			name: "successful hook creation",
			config: map[string]types.SandboxHook{
				"start":   &types.SandboxHookNative{Enabled: true},
				"stop":    &types.SandboxHookSignal{IsString: true, StringValue: "SIGTERM"},
				"restart": &types.SandboxHookSignal{IsString: false, IntValue: int(syscall.SIGKILL)},
				"reload":  &types.SandboxHookSignal{IsString: false, IntValue: int(syscall.SIGUSR2)},
			},
			expectedHooks: Hooks{
				StartHookType:   &HookNative{BaseHook: BaseHook{Enabled: true, Type: StartHookType}},
				StopHookType:    &HookSignal{BaseHook: BaseHook{Enabled: true, Type: StopHookType}, Signal: syscall.SIGTERM},
				RestartHookType: &HookSignal{BaseHook: BaseHook{Enabled: true, Type: RestartHookType}, Signal: syscall.SIGKILL},
				ReloadHookType:  &HookSignal{BaseHook: BaseHook{Enabled: true, Type: ReloadHookType}, Signal: syscall.SIGUSR2},
			},
		},
		{
			name: "error on unsupported hook type",
			config: map[string]types.SandboxHook{
				"invalid": &types.SandboxHookNative{Enabled: true},
			},
			expectedError: "invalid hook type invalid",
		},
		{
			name: "error on unsupported signal value",
			config: map[string]types.SandboxHook{
				"stop": &types.SandboxHookSignal{IsString: true, StringValue: "INVALID"},
			},
			expectedError: "unsupported string signal value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			maker := CreateMaker(fndMock)

			result, err := maker.MakeHooks(tt.config)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHooks, result)
				for k, v := range result {
					expectedHook := tt.expectedHooks[k]
					assert.IsType(t, expectedHook, v)
				}
			}
		})
	}
}

func Test_nativeMaker_MakeHook(t *testing.T) {
	tests := []struct {
		name          string
		config        types.SandboxHook
		hookType      HookType
		expectedHook  Hook
		expectedError string
	}{
		{
			name:     "successful native hook creation",
			config:   &types.SandboxHookNative{Enabled: true},
			hookType: StartHookType,
			expectedHook: &HookNative{
				BaseHook: BaseHook{Enabled: true, Type: StartHookType},
			},
		},
		{
			name:     "successful shell command hook creation",
			config:   &types.SandboxHookShellCommand{Command: "echo 'hello'", Shell: "/bin/bash"},
			hookType: RestartHookType,
			expectedHook: &HookShellCommand{
				BaseHook: BaseHook{Enabled: true, Type: RestartHookType},
				Command:  "echo 'hello'",
				Shell:    "/bin/bash",
			},
		},
		{
			name:     "successful args command hook creation",
			config:   &types.SandboxHookArgsCommand{Executable: "ls", Args: []string{"-la"}},
			hookType: ReloadHookType,
			expectedHook: &HookArgsCommand{
				BaseHook:   BaseHook{Enabled: true, Type: ReloadHookType},
				Executable: "ls",
				Args:       []string{"-la"},
			},
		},
		{
			name:     "successful signal hook creation with string",
			config:   &types.SandboxHookSignal{IsString: true, StringValue: "SIGTERM"},
			hookType: StopHookType,
			expectedHook: &HookSignal{
				BaseHook: BaseHook{Enabled: true, Type: StopHookType},
				Signal:   syscall.SIGTERM,
			},
		},
		{
			name:     "successful signal hook creation with int",
			config:   &types.SandboxHookSignal{IsString: false, IntValue: int(syscall.SIGKILL)},
			hookType: StopHookType,
			expectedHook: &HookSignal{
				BaseHook: BaseHook{Enabled: true, Type: StopHookType},
				Signal:   syscall.SIGKILL,
			},
		},
		{
			name:          "unsupported string signal value",
			config:        &types.SandboxHookSignal{IsString: true, StringValue: "INVALID"},
			hookType:      StopHookType,
			expectedError: "unsupported string signal value",
		},
		{
			name:          "unsupported int signal value",
			config:        &types.SandboxHookSignal{IsString: false, IntValue: 9999},
			hookType:      StopHookType,
			expectedError: "unsupported int signal value",
		},
		{
			name:          "unsupported hook type",
			config:        mock.Anything,
			hookType:      StartHookType,
			expectedError: "unsupported hook type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			maker := CreateMaker(fndMock)

			result, err := maker.MakeHook(tt.config, tt.hookType)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHook, result)
				// Check if the type of the hook matches as expected
				assert.IsType(t, tt.expectedHook, result)
			}
		})
	}
}

func TestHookNative_Execute(t *testing.T) {
	tests := []struct {
		name       string
		hookType   HookType
		startTask  task.Task
		setupMocks func(
			t *testing.T,
			ctx context.Context,
			ss *environment.ServiceSettings,
			envMock *environmentMocks.MockEnvironment,
			taskMock *taskMocks.MockTask,
		)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "Start hook with no existing task",
			hookType:  StartHookType,
			startTask: nil,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
			) {
				var cmd *environment.Command
				envMock.On("RunTask", ctx, ss, cmd).Return(taskMock, nil)
			},
		},
		{
			name:      "Start hook with no existing task that fails",
			hookType:  StartHookType,
			startTask: nil,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
			) {
				var cmd *environment.Command
				envMock.On("RunTask", ctx, ss, cmd).Return(nil, errors.New("run fail"))
			},
			expectError: true,
			errorMsg:    "run fail",
		},
		{
			name:        "Start hook with existing task",
			hookType:    StartHookType,
			startTask:   taskMocks.NewMockTask(t),
			expectError: true,
			errorMsg:    "task has already been created which is likely because start already done",
		},
		{
			name:        "Non-start hook with no task",
			hookType:    StopHookType, // Example of non-start hook
			startTask:   nil,
			expectError: true,
			errorMsg:    "task has not been created which is likely because start is not done",
		},
		{
			name:      "Non-start hook with existing task",
			hookType:  StopHookType,
			startTask: taskMocks.NewMockTask(t),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ss := &environment.ServiceSettings{}
			tmpl := templateMocks.NewMockTemplate(t)
			env := environmentMocks.NewMockEnvironment(t)
			newTask := taskMocks.NewMockTask(t)
			if tt.setupMocks != nil {
				tt.setupMocks(t, ctx, ss, env, newTask)
			}

			hook := HookNative{
				BaseHook: BaseHook{
					Type:    tt.hookType,
					Enabled: true,
				},
			}

			resultTask, err := hook.Execute(ctx, ss, tmpl, env, tt.startTask)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, resultTask)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resultTask)
			}
		})
	}
}

func TestHookArgsCommand_Execute(t *testing.T) {
	tests := []struct {
		name       string
		hookType   HookType
		startTask  task.Task
		setupMocks func(
			t *testing.T,
			ctx context.Context,
			ss *environment.ServiceSettings,
			envMock *environmentMocks.MockEnvironment,
			taskMock *taskMocks.MockTask,
			tmplMock *templateMocks.MockTemplate,
		)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "Start hook with command execution success",
			hookType:  StartHookType,
			startTask: nil,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "executable-path", ss.ServerParameters).Return("executed-path", nil)
				tmplMock.On("RenderToString", "arg1", ss.ServerParameters).Return("argument1", nil)
				cmd := &environment.Command{Name: "executed-path", Args: []string{"argument1"}}
				envMock.On("RunTask", ctx, ss, cmd).Return(taskMock, nil)
			},
		},
		{
			name:      "Start hook with command execution failure",
			hookType:  StartHookType,
			startTask: nil,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "executable-path", ss.ServerParameters).Return("executed-path", nil)
				tmplMock.On("RenderToString", "arg1", ss.ServerParameters).Return("argument1", nil)
				cmd := &environment.Command{Name: "executed-path", Args: []string{"argument1"}}
				envMock.On("RunTask", ctx, ss, cmd).Return(nil, errors.New("run fail"))
			},
			expectError: true,
			errorMsg:    "run fail",
		},
		{
			name:      "Start hook with command creation failure for new task",
			hookType:  StartHookType,
			startTask: nil,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "executable-path", ss.ServerParameters).Return("executed-path", nil)
				tmplMock.On("RenderToString", "arg1", ss.ServerParameters).Return("", errors.New("template render error"))
			},
			expectError: true,
			errorMsg:    "template render error",
		},
		{
			name:      "Start hook with command creation failure for already existing task",
			hookType:  StopHookType,
			startTask: taskMocks.NewMockTask(t),
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "executable-path", ss.ServerParameters).Return("", errors.New("template render error"))
			},
			expectError: true,
			errorMsg:    "template render error",
		},
		{
			name:        "Start hook with task already exists",
			hookType:    StartHookType,
			startTask:   taskMocks.NewMockTask(t),
			expectError: true,
			errorMsg:    "task has already been created which is likely because start already done",
		},
		{
			name:        "Non-start hook with task not created",
			hookType:    StopHookType, // Example of non-start hook
			startTask:   nil,
			expectError: true,
			errorMsg:    "task has not been created which is likely because start is not done",
		},
		{
			name:      "Non-start hook with task exists",
			hookType:  StopHookType,
			startTask: taskMocks.NewMockTask(t),
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "executable-path", ss.ServerParameters).Return("executed-path", nil)
				tmplMock.On("RenderToString", "arg1", ss.ServerParameters).Return("argument1", nil)
				cmd := &environment.Command{Name: "executed-path", Args: []string{"argument1"}}
				envMock.On("ExecTaskCommand", mock.Anything, ss, mock.AnythingOfType("*task.MockTask"), cmd).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ss := &environment.ServiceSettings{
				ServerParameters: parameters.Parameters{
					"p": parameterMocks.NewMockParameter(t),
				},
			}
			tmplMock := templateMocks.NewMockTemplate(t)
			envMock := environmentMocks.NewMockEnvironment(t)
			taskMock := taskMocks.NewMockTask(t)

			if tt.setupMocks != nil {
				tt.setupMocks(t, ctx, ss, envMock, taskMock, tmplMock)
			}

			hook := HookArgsCommand{
				BaseHook: BaseHook{
					Type:    tt.hookType,
					Enabled: true,
				},
				Executable: "executable-path",
				Args:       []string{"arg1"},
			}

			resultTask, err := hook.Execute(ctx, ss, tmplMock, envMock, tt.startTask)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, resultTask)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resultTask)
			}

			tmplMock.AssertExpectations(t)
			envMock.AssertExpectations(t)
		})
	}
}

func TestHookShellCommand_Execute(t *testing.T) {
	tests := []struct {
		name       string
		hookType   HookType
		startTask  task.Task
		setupMocks func(
			t *testing.T,
			ctx context.Context,
			ss *environment.ServiceSettings,
			envMock *environmentMocks.MockEnvironment,
			taskMock *taskMocks.MockTask,
			tmplMock *templateMocks.MockTemplate,
		)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "Start hook with command execution success",
			hookType:  StartHookType,
			startTask: nil,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "/bin/sh", ss.ServerParameters).Return("/bin/sh", nil)
				tmplMock.On("RenderToString", "-c", ss.ServerParameters).Return("-c", nil)
				tmplMock.On("RenderToString", "cat file", ss.ServerParameters).Return("cat file", nil)
				cmd := &environment.Command{Name: "/bin/sh", Args: []string{"-c", "cat file"}}
				envMock.On("RunTask", ctx, ss, cmd).Return(taskMock, nil)
			},
		},
		{
			name:      "Start hook with command execution failure",
			hookType:  StartHookType,
			startTask: nil,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "/bin/sh", ss.ServerParameters).Return("/bin/sh", nil)
				tmplMock.On("RenderToString", "-c", ss.ServerParameters).Return("-c", nil)
				tmplMock.On("RenderToString", "cat file", ss.ServerParameters).Return("cat file", nil)
				cmd := &environment.Command{Name: "/bin/sh", Args: []string{"-c", "cat file"}}
				envMock.On("RunTask", ctx, ss, cmd).Return(nil, errors.New("run fail"))
			},
			expectError: true,
			errorMsg:    "run fail",
		},
		{
			name:      "Start hook with command creation failure for new task",
			hookType:  StartHookType,
			startTask: nil,
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "/bin/sh", ss.ServerParameters).Return("", errors.New("template render error"))
			},
			expectError: true,
			errorMsg:    "template render error",
		},
		{
			name:      "Start hook with command creation failure for already existing task",
			hookType:  StopHookType,
			startTask: taskMocks.NewMockTask(t),
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "/bin/sh", ss.ServerParameters).Return("", errors.New("template render error"))
			},
			expectError: true,
			errorMsg:    "template render error",
		},
		{
			name:        "Start hook with task already exists",
			hookType:    StartHookType,
			startTask:   taskMocks.NewMockTask(t),
			expectError: true,
			errorMsg:    "task has already been created which is likely because start already done",
		},
		{
			name:        "Non-start hook with task not created",
			hookType:    StopHookType, // Example of non-start hook
			startTask:   nil,
			expectError: true,
			errorMsg:    "task has not been created which is likely because start is not done",
		},
		{
			name:      "Non-start hook with task exists",
			hookType:  StopHookType,
			startTask: taskMocks.NewMockTask(t),
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				ss *environment.ServiceSettings,
				envMock *environmentMocks.MockEnvironment,
				taskMock *taskMocks.MockTask,
				tmplMock *templateMocks.MockTemplate,
			) {
				tmplMock.On("RenderToString", "/bin/sh", ss.ServerParameters).Return("/bin/sh", nil)
				tmplMock.On("RenderToString", "-c", ss.ServerParameters).Return("-c", nil)
				tmplMock.On("RenderToString", "cat file", ss.ServerParameters).Return("cat file", nil)
				cmd := &environment.Command{Name: "/bin/sh", Args: []string{"-c", "cat file"}}
				envMock.On("ExecTaskCommand", mock.Anything, ss, mock.AnythingOfType("*task.MockTask"), cmd).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ss := &environment.ServiceSettings{
				ServerParameters: parameters.Parameters{
					"p": parameterMocks.NewMockParameter(t),
				},
			}
			tmplMock := templateMocks.NewMockTemplate(t)
			envMock := environmentMocks.NewMockEnvironment(t)
			taskMock := taskMocks.NewMockTask(t)

			if tt.setupMocks != nil {
				tt.setupMocks(t, ctx, ss, envMock, taskMock, tmplMock)
			}

			hook := HookShellCommand{
				BaseHook: BaseHook{
					Type:    tt.hookType,
					Enabled: true,
				},
				Shell:   "/bin/sh",
				Command: "cat file",
			}

			resultTask, err := hook.Execute(ctx, ss, tmplMock, envMock, tt.startTask)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, resultTask)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resultTask)
			}

			tmplMock.AssertExpectations(t)
			envMock.AssertExpectations(t)
		})
	}
}

func TestHookSignal_Execute(t *testing.T) {
	tests := []struct {
		name        string
		hookType    HookType
		signal      os.Signal
		noTask      bool
		setupMocks  func(t *testing.T, envMock *environmentMocks.MockEnvironment, taskMock *taskMocks.MockTask)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Signal hook on start type",
			hookType:    StartHookType,
			signal:      os.Interrupt,
			expectError: true,
			errorMsg:    "signal hook cannot be executed on start",
		},
		{
			name:        "Signal hook without task",
			hookType:    StopHookType,
			signal:      os.Interrupt,
			noTask:      true,
			expectError: true,
			errorMsg:    "task does not exist for signal hook to execute",
		},
		{
			name:     "Signal hook with existing task",
			hookType: StopHookType,
			signal:   os.Interrupt,
			setupMocks: func(t *testing.T, envMock *environmentMocks.MockEnvironment, taskMock *taskMocks.MockTask) {
				envMock.On("ExecTaskSignal", mock.Anything, mock.Anything, taskMock, os.Interrupt).Return(nil)
			},
		},
		{
			name:     "Signal hook with existing task and failure",
			hookType: StopHookType,
			signal:   os.Interrupt,
			setupMocks: func(t *testing.T, envMock *environmentMocks.MockEnvironment, taskMock *taskMocks.MockTask) {
				envMock.On("ExecTaskSignal", mock.Anything, mock.Anything, taskMock, os.Interrupt).Return(errors.New("signal error"))
			},
			expectError: true,
			errorMsg:    "signal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ss := &environment.ServiceSettings{}
			tmpl := templateMocks.NewMockTemplate(t)
			env := environmentMocks.NewMockEnvironment(t)
			taskMock := taskMocks.NewMockTask(t)

			if tt.noTask {
				taskMock = nil
			}
			if tt.setupMocks != nil {
				tt.setupMocks(t, env, taskMock)
			}

			hook := HookSignal{
				BaseHook: BaseHook{
					Type:    tt.hookType,
					Enabled: true,
				},
				Signal: tt.signal,
			}

			resultTask, err := hook.Execute(ctx, ss, tmpl, env, taskMock)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, resultTask)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, taskMock, resultTask)
			}

			env.AssertExpectations(t)
		})
	}
}
