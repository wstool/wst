package local

import (
	"context"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	taskMocks "github.com/bukka/wst/mocks/generated/run/environments/task"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/environments/task"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	tests := []struct {
		name string
		fnd  app.Foundation
	}{
		{
			name: "create maker",
			fnd:  fndMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateMaker(tt.fnd)
			maker, ok := got.(*localMaker)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, maker.Fnd)
		})
	}
}

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name              string
		config            *types.LocalEnvironment
		instanceWorkspace string
		getExpectedEnv    func(fndMock *appMocks.MockFoundation) *localEnvironment
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "successful kubernetes environment maker creation",
			config: &types.LocalEnvironment{
				Ports: types.EnvironmentPorts{
					Start: 8000,
					End:   8500,
				},
			},
			instanceWorkspace: "/tmp/ws",
			getExpectedEnv: func(
				fndMock *appMocks.MockFoundation,
			) *localEnvironment {
				return &localEnvironment{
					CommonEnvironment: environment.CommonEnvironment{
						Fnd:  fndMock,
						Used: false,
						Ports: environment.Ports{
							Start: 8000,
							Used:  8000,
							End:   8500,
						},
					},
					tasks:     make(map[string]*localTask),
					workspace: "/tmp/ws/envs/local",
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			m := &localMaker{
				CommonMaker: &environment.CommonMaker{
					Fnd: fndMock,
				},
			}

			got, err := m.Make(tt.config, tt.instanceWorkspace)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				actualEnv, ok := got.(*localEnvironment)
				assert.True(t, ok)
				expectedEnv := tt.getExpectedEnv(fndMock)
				assert.Equal(t, expectedEnv, actualEnv)
			}
		})
	}
}

func Test_localEnvironment_RootPath(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	l := &localEnvironment{
		CommonEnvironment: environment.CommonEnvironment{
			Fnd:  fndMock,
			Used: false,
			Ports: environment.Ports{
				Start: 8000,
				Used:  8000,
				End:   8500,
			},
		},
		tasks:     make(map[string]*localTask),
		workspace: "/tmp/ws/envs/local",
	}
	assert.Equal(t, "/tmp/ws/svc", l.RootPath("/tmp/ws/svc"))
}

func Test_localEnvironment_Init(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*testing.T, *appMocks.MockFs)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful initialization",
			setupMocks: func(t *testing.T, mockFs *appMocks.MockFs) {
				mockFs.On("MkdirAll", "/fake/path", os.FileMode(0644)).Return(nil)
			},
		},
		{
			name: "error on directory creation",
			setupMocks: func(t *testing.T, mockFs *appMocks.MockFs) {
				mockFs.On("MkdirAll", "/fake/path", os.FileMode(0644)).Return(os.ErrPermission)
			},
			expectError:    true,
			expectedErrMsg: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFs := appMocks.NewMockFs(t)
			tt.setupMocks(t, mockFs)
			fs := &afero.Afero{Fs: mockFs}

			fndMock := appMocks.NewMockFoundation(t)
			fndMock.On("Fs").Return(fs)

			env := &localEnvironment{
				CommonEnvironment: environment.CommonEnvironment{Fnd: fndMock},
				workspace:         "/fake/path",
				initialized:       false,
			}

			err := env.Init(context.Background())

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.True(t, env.initialized)
			}
		})
	}
}

func Test_localEnvironment_Destroy(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*testing.T, *appMocks.MockFs, *appMocks.MockCommand)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful destruction with running tasks",
			setupMocks: func(t *testing.T, mockFs *appMocks.MockFs, mockCmd *appMocks.MockCommand) {
				mockCmd.On("IsRunning").Return(true)
				mockCmd.On("ProcessSignal", os.Kill).Return(nil)
				mockFs.On("RemoveAll", "/fake/path").Return(nil)
			},
		},
		{
			name: "failure to kill running task",
			setupMocks: func(t *testing.T, mockFs *appMocks.MockFs, mockCmd *appMocks.MockCommand) {
				mockCmd.On("IsRunning").Return(true)
				mockCmd.On("ProcessSignal", os.Kill).Return(os.ErrPermission)
				mockFs.On("RemoveAll", "/fake/path").Return(nil)
			},
			expectError:    true,
			expectedErrMsg: "failed to kill local environment tasks",
		},
		{
			name: "error on directory removal",
			setupMocks: func(t *testing.T, mockFs *appMocks.MockFs, mockCmd *appMocks.MockCommand) {
				mockCmd.On("IsRunning").Return(false) // No running tasks
				mockFs.On("RemoveAll", "/fake/path").Return(os.ErrPermission)
			},
			expectError:    true,
			expectedErrMsg: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFs := appMocks.NewMockFs(t)
			mockCmd := appMocks.NewMockCommand(t)
			tt.setupMocks(t, mockFs, mockCmd)

			fs := &afero.Afero{Fs: mockFs}
			fndMock := appMocks.NewMockFoundation(t)
			fndMock.On("Fs").Return(fs)

			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Maybe().Return(mockLogger.SugaredLogger)

			env := &localEnvironment{
				CommonEnvironment: environment.CommonEnvironment{Fnd: fndMock},
				workspace:         "/fake/path",
				initialized:       true,
				tasks: map[string]*localTask{
					"task1": {cmd: mockCmd},
				},
			}

			err := env.Destroy(context.Background())

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}

			mockCmd.AssertExpectations(t)
			mockFs.AssertExpectations(t)
		})
	}
}

func Test_localEnvironment_RunTask(t *testing.T) {
	tests := []struct {
		name              string
		workspace         string
		mkdirAllError     error
		commandStartError error
		expectedError     error
		expectTask        bool
		uuid              string // UUID for each task
	}{
		{
			name:          "successfully runs task",
			workspace:     "/fake/path",
			expectedError: nil,
			expectTask:    true,
			uuid:          "uuid-123",
		},
		{
			name:          "initialization error due to filesystem",
			workspace:     "/fake/path",
			mkdirAllError: fmt.Errorf("filesystem error"),
			expectedError: fmt.Errorf("filesystem error"),
			expectTask:    false,
			uuid:          "uuid-456",
		},
		{
			name:              "command start error",
			workspace:         "/fake/path",
			commandStartError: fmt.Errorf("command start error"),
			expectedError:     fmt.Errorf("command start error"),
			expectTask:        false,
			uuid:              "uuid-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			fsMock := appMocks.NewMockFs(t)
			mockCommand := appMocks.NewMockCommand(t)
			ctx := context.Background()

			fsMock.On("MkdirAll", tt.workspace, os.FileMode(0644)).Return(tt.mkdirAllError)
			fndMock.On("Fs").Return(fsMock)

			if tt.mkdirAllError == nil {
				fndMock.On("ExecCommand", ctx, "test-command", []string{"arg1"}).Return(mockCommand)
				mockCommand.On("Start").Return(tt.commandStartError)
				fndMock.On("GenerateUuid").Maybe().Return(tt.uuid)
			}

			env := &localEnvironment{
				CommonEnvironment: environment.CommonEnvironment{Fnd: fndMock},
				workspace:         tt.workspace,
				initialized:       false,
				tasks:             make(map[string]*localTask),
			}

			ss := &environment.ServiceSettings{
				Name: "test-service",
				Port: 8080,
			}
			cmd := &environment.Command{
				Name: "test-command",
				Args: []string{"arg1"},
			}

			resultTask, err := env.RunTask(ctx, ss, cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, resultTask)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resultTask)
				locTask, ok := resultTask.(*localTask)
				assert.True(t, ok)
				assert.Equal(t, "test-service", locTask.serviceName)
				assert.Equal(t, "test-command", locTask.executable)
				assert.Equal(t, fmt.Sprintf("http://localhost:%d", ss.Port), locTask.serviceUrl)
				assert.Equal(t, mockCommand, locTask.cmd)
				assert.Equal(t, tt.uuid, locTask.id) // Checking the UUID
			}

			fndMock.AssertExpectations(t)
			fsMock.AssertExpectations(t)
			if tt.mkdirAllError == nil && mockCommand.AssertNumberOfCalls(t, "Start", 1) {
				mockCommand.AssertExpectations(t)
			}
		})
	}
}

func Test_localEnvironment_ExecTaskCommand(t *testing.T) {
	tests := []struct {
		name             string
		setupMocks       func(*testing.T, *appMocks.MockFoundation, *appMocks.MockCommand)
		target           func() task.Task // Using a function allows setup of the task per test case
		command          *environment.Command
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful command execution",
			setupMocks: func(t *testing.T, fnd *appMocks.MockFoundation, cmd *appMocks.MockCommand) {
				cmd.On("Run").Return(nil) // Simulating a successful command run
				fnd.On("ExecCommand", mock.Anything, "echo", []string{"hello"}).Return(cmd)
			},
			target: func() task.Task {
				localTask := &localTask{
					serviceName: "local-service",
					cmd:         &app.ExecCommand{}, // Using a valid local task type
				}
				localTask.cmd = &app.ExecCommand{} // Stubbing the command to be returned
				return localTask
			},
			command: &environment.Command{
				Name: "echo",
				Args: []string{"hello"},
			},
			expectError: false,
		},
		{
			name: "error during command execution",
			setupMocks: func(t *testing.T, fnd *appMocks.MockFoundation, cmd *appMocks.MockCommand) {
				cmd.On("Run").Return(fmt.Errorf("execution failed")) // Simulating a command failure
				fnd.On("ExecCommand", mock.Anything, "echo", []string{"error"}).Return(cmd)
			},
			target: func() task.Task {
				localTask := &localTask{
					serviceName: "local-service",
					cmd:         &app.ExecCommand{}, // Using a valid local task type
				}
				localTask.cmd = &app.ExecCommand{} // Stubbing the command to be returned
				return localTask
			},
			command: &environment.Command{
				Name: "echo",
				Args: []string{"error"},
			},
			expectError:      true,
			expectedErrorMsg: "execution failed",
		},
		{
			name: "error wrong task type",
			target: func() task.Task {
				wrongTypeTask := &taskMocks.MockTask{}
				wrongTypeTask.On("Type").Return(providers.DockerType) // Incorrect task type
				return wrongTypeTask
			},
			command: &environment.Command{
				Name: "echo",
				Args: []string{"hello"},
			},
			expectError:      true,
			expectedErrorMsg: "local environment can process only local task",
		},
		{
			name: "error invalid task casting",
			target: func() task.Task {
				wrongTypeTask := &taskMocks.MockTask{}
				wrongTypeTask.On("Type").Return(providers.LocalType) // Incorrect task type
				return wrongTypeTask
			},
			command: &environment.Command{
				Name: "echo",
				Args: []string{"hello"},
			},
			expectError:      true,
			expectedErrorMsg: "target task is not of type *localTask",
		},
		{
			name: "error for nil task",
			target: func() task.Task {
				return nil
			},
			command: &environment.Command{
				Name: "echo",
				Args: []string{"hello"},
			},
			expectError:      true,
			expectedErrorMsg: "target task is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			cmdMock := appMocks.NewMockCommand(t)
			if tt.setupMocks != nil {
				tt.setupMocks(t, fndMock, cmdMock)
			}
			ctx := context.Background()
			targetTask := tt.target()

			env := &localEnvironment{
				CommonEnvironment: environment.CommonEnvironment{Fnd: fndMock},
			}

			err := env.ExecTaskCommand(ctx, &environment.ServiceSettings{}, targetTask, tt.command)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}

			fndMock.AssertExpectations(t)
			cmdMock.AssertExpectations(t)
		})
	}
}

func Test_localEnvironment_ExecTaskSignal(t *testing.T) {
	tests := []struct {
		name             string
		target           func(*testing.T) task.Task
		signal           os.Signal
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful signal execution",
			target: func(t *testing.T) task.Task {
				cmdMock := appMocks.NewMockCommand(t)
				cmdMock.On("ProcessSignal", os.Interrupt).Return(nil)
				return &localTask{
					serviceName: "local-service",
					cmd:         cmdMock,
				}
			},
			signal:      os.Interrupt,
			expectError: false,
		},
		{
			name: "error during signal execution",
			target: func(t *testing.T) task.Task {
				cmdMock := appMocks.NewMockCommand(t)
				cmdMock.On("ProcessSignal", os.Kill).Return(fmt.Errorf("failed to send signal"))
				return &localTask{
					serviceName: "local-service",
					cmd:         cmdMock,
				}
			},
			signal:           os.Kill,
			expectError:      true,
			expectedErrorMsg: "failed to send signal",
		},
		{
			name: "task is nil",
			target: func(t *testing.T) task.Task {
				return nil
			},
			signal:           os.Interrupt,
			expectError:      true,
			expectedErrorMsg: "target task is not set",
		},
		{
			name: "task type mismatch",
			target: func(t *testing.T) task.Task {
				wrongTask := &taskMocks.MockTask{}
				wrongTask.On("Type").Return(providers.DockerType) // Return an unexpected task type
				return wrongTask
			},
			signal:           os.Interrupt,
			expectError:      true,
			expectedErrorMsg: "local environment can process only local task",
		},
		{
			name: "casting error",
			target: func(t *testing.T) task.Task {
				// Correct type but wrong implementation
				wrongTask := &taskMocks.MockTask{}
				wrongTask.On("Type").Return(providers.LocalType)
				return wrongTask
			},
			signal:           os.Interrupt,
			expectError:      true,
			expectedErrorMsg: "target task is not of type *localTask",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			targetTask := tt.target(t)

			env := &localEnvironment{
				CommonEnvironment: environment.CommonEnvironment{Fnd: appMocks.NewMockFoundation(t)},
			}

			err := env.ExecTaskSignal(ctx, &environment.ServiceSettings{}, targetTask, tt.signal)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}

			// Assert expectations on the command mock within the task, if it is a valid local task
			if convertedTask, ok := targetTask.(*localTask); ok {
				cmdMock, _ := convertedTask.cmd.(*appMocks.MockCommand)
				cmdMock.AssertExpectations(t)
			}
		})
	}
}

func Test_localEnvironment_Output(t *testing.T) {
	tests := []struct {
		name             string
		outputType       output.Type
		setupMocks       func(*testing.T, *appMocks.MockCommand)
		nilTask          bool
		expectError      bool
		expectedOutput   string
		expectedErrorMsg string
	}{
		{
			name:       "successful stdout output collection",
			outputType: output.Stdout,
			setupMocks: func(t *testing.T, cmd *appMocks.MockCommand) {
				stdout := io.NopCloser(strings.NewReader("Hello, stdout!"))
				cmd.On("StdoutPipe").Return(stdout, nil)
			},
			expectedOutput: "Hello, stdout!",
		},
		{
			name:       "successful stderr output collection",
			outputType: output.Stderr,
			setupMocks: func(t *testing.T, cmd *appMocks.MockCommand) {
				stderr := io.NopCloser(strings.NewReader("Hello, stderr!"))
				cmd.On("StderrPipe").Return(stderr, nil)
			},
			expectedOutput: "Hello, stderr!",
		},
		{
			name:       "successful any output collection",
			outputType: output.Any,
			setupMocks: func(t *testing.T, cmd *appMocks.MockCommand) {
				stdout := io.NopCloser(strings.NewReader("out"))
				stderr := io.NopCloser(strings.NewReader("out"))
				cmd.On("StdoutPipe").Return(stdout, nil)
				cmd.On("StderrPipe").Return(stderr, nil)
			},
			expectedOutput: "outout",
		},
		{
			name:       "error on stdout pipe",
			outputType: output.Stdout,
			setupMocks: func(t *testing.T, cmd *appMocks.MockCommand) {
				cmd.On("StdoutPipe").Return(nil, fmt.Errorf("stdout error"))
			},
			expectError:      true,
			expectedErrorMsg: "stdout error",
		},
		{
			name:       "error on stderr pipe",
			outputType: output.Stderr,
			setupMocks: func(t *testing.T, cmd *appMocks.MockCommand) {
				cmd.On("StderrPipe").Return(nil, fmt.Errorf("stderr error"))
			},
			expectError:      true,
			expectedErrorMsg: "stderr error",
		},
		{
			name:       "error on stdout pipe in any type",
			outputType: output.Any,
			setupMocks: func(t *testing.T, cmd *appMocks.MockCommand) {
				cmd.On("StdoutPipe").Return(nil, fmt.Errorf("stdout error"))
			},
			expectError:      true,
			expectedErrorMsg: "stdout error", // Assuming first error encountered is returned
		},
		{
			name:       "error on stderr pipe in any type",
			outputType: output.Any,
			setupMocks: func(t *testing.T, cmd *appMocks.MockCommand) {
				stdout := io.NopCloser(strings.NewReader("out"))
				cmd.On("StdoutPipe").Return(stdout, nil)
				cmd.On("StderrPipe").Return(nil, fmt.Errorf("stderr error"))
			},
			expectError:      true,
			expectedErrorMsg: "stderr error", // Assuming first error encountered is returned
		},
		{
			name:             "unsupported output type",
			outputType:       output.Type(999), // Invalid output type
			setupMocks:       func(t *testing.T, cmd *appMocks.MockCommand) {},
			expectError:      true,
			expectedErrorMsg: "unsupported output type",
		},
		{
			name:             "nil task",
			outputType:       output.Any, // Invalid output type
			setupMocks:       func(t *testing.T, cmd *appMocks.MockCommand) {},
			nilTask:          true,
			expectError:      true,
			expectedErrorMsg: "target task is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cmdMock := appMocks.NewMockCommand(t)
			tt.setupMocks(t, cmdMock)

			var testTask task.Task = nil
			if !tt.nilTask {
				testTask = &localTask{
					cmd: cmdMock,
				}
			}

			env := &localEnvironment{}

			reader, err := env.Output(ctx, testTask, tt.outputType)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				buf := new(strings.Builder)
				_, err = io.Copy(buf, reader)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, buf.String())
			}
		})
	}
}

func getTestTask(t *testing.T) *localTask {
	cmdMock := appMocks.NewMockCommand(t)
	cmdMock.On("ProcessPid").Maybe().Return(22)
	return &localTask{
		id:          "lid",
		executable:  "ep",
		cmd:         cmdMock,
		serviceName: "lids",
		serviceUrl:  "http://localhost:1234",
	}
}

func Test_localTask_Id(t *testing.T) {
	assert.Equal(t, "lid", getTestTask(t).Id())
}

func Test_localTask_Executable(t *testing.T) {
	assert.Equal(t, "ep", getTestTask(t).Executable())
}

func Test_localTask_Name(t *testing.T) {
	assert.Equal(t, "lids", getTestTask(t).Name())
}

func Test_localTask_Pid(t *testing.T) {
	assert.Equal(t, 22, getTestTask(t).Pid())
}

func Test_localTask_PrivateUrl(t *testing.T) {
	assert.Equal(t, "http://localhost:1234", getTestTask(t).PrivateUrl())
}

func Test_localTask_PublicUrl(t *testing.T) {
	assert.Equal(t, "http://localhost:1234", getTestTask(t).PublicUrl())
}

func Test_localTask_Type(t *testing.T) {
	assert.Equal(t, providers.LocalType, getTestTask(t).Type())
}
