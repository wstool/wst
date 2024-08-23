package local

import (
	"context"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	outputMocks "github.com/bukka/wst/mocks/generated/run/environments/environment/output"
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
				mockFs.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
			},
		},
		{
			name: "error on directory creation",
			setupMocks: func(t *testing.T, mockFs *appMocks.MockFs) {
				mockFs.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(os.ErrPermission)
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
		setupMocks     func(*testing.T, *appMocks.MockFs, *appMocks.MockCommand, *outputMocks.MockCollector)
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successful destruction with running tasks",
			setupMocks: func(
				t *testing.T,
				mockFs *appMocks.MockFs,
				mockCmd *appMocks.MockCommand,
				mockOc *outputMocks.MockCollector,
			) {
				mockCmd.On("IsRunning").Return(true)
				mockCmd.On("ProcessSignal", os.Kill).Return(nil)
				mockOc.On("Close").Return(nil)
				mockFs.On("RemoveAll", "/fake/path").Return(nil)
			},
		},
		{
			name: "failure to kill running task",
			setupMocks: func(
				t *testing.T,
				mockFs *appMocks.MockFs,
				mockCmd *appMocks.MockCommand,
				mockOc *outputMocks.MockCollector,
			) {
				mockCmd.On("IsRunning").Return(true)
				mockCmd.On("ProcessSignal", os.Kill).Return(os.ErrPermission)
				mockOc.On("Close").Return(nil)
				mockFs.On("RemoveAll", "/fake/path").Return(nil)
			},
			expectError:    true,
			expectedErrMsg: "failed to kill local environment tasks",
		},
		{
			name: "error on directory removal",
			setupMocks: func(
				t *testing.T,
				mockFs *appMocks.MockFs,
				mockCmd *appMocks.MockCommand,
				mockOc *outputMocks.MockCollector,
			) {
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
			mockOc := outputMocks.NewMockCollector(t)
			tt.setupMocks(t, mockFs, mockCmd, mockOc)

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
					"task1": {
						cmd:             mockCmd,
						outputCollector: mockOc,
					},
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
		name       string
		workspace  string
		setupMocks func(
			*testing.T,
			context.Context,
			*appMocks.MockFoundation,
			*outputMocks.MockMaker,
		) *appMocks.MockCommand
		expectError    bool
		expectedErrMsg string
		expectTask     bool
		uuid           string // UUID for each task
	}{
		{
			name:      "successfully runs task",
			workspace: "/fake/path",
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) *appMocks.MockCommand {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fndMock.On("Fs").Return(fsMock)
				mockCommand := appMocks.NewMockCommand(t)
				fndMock.On("ExecCommand", ctx, "test-command", []string{"arg1"}).Return(mockCommand)
				stdout := io.NopCloser(strings.NewReader("Hello, stdout!"))
				stderr := io.NopCloser(strings.NewReader("Hello, stderr!"))
				mockCommand.On("StdoutPipe").Return(stdout, nil)
				mockCommand.On("StderrPipe").Return(stderr, nil)
				collectorMock := outputMocks.NewMockCollector(t)
				outMakerMock.On("MakeCollector").Return(collectorMock)
				collectorMock.On("Start", stdout, stderr).Return(nil)
				mockCommand.On("Start").Return(nil)
				fndMock.On("GenerateUuid").Return("uuid-123")
				return mockCommand
			},
			expectTask: true,
			uuid:       "uuid-123",
		},
		{
			name:      "command start error",
			workspace: "/fake/path",
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) *appMocks.MockCommand {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fndMock.On("Fs").Return(fsMock)
				mockCommand := appMocks.NewMockCommand(t)
				fndMock.On("ExecCommand", ctx, "test-command", []string{"arg1"}).Return(mockCommand)
				stdout := io.NopCloser(strings.NewReader("Hello, stdout!"))
				stderr := io.NopCloser(strings.NewReader("Hello, stderr!"))
				mockCommand.On("StdoutPipe").Return(stdout, nil)
				mockCommand.On("StderrPipe").Return(stderr, nil)
				collectorMock := outputMocks.NewMockCollector(t)
				outMakerMock.On("MakeCollector").Return(collectorMock)
				collectorMock.On("Start", stdout, stderr).Return(nil)
				mockCommand.On("Start").Return(fmt.Errorf("command start error"))
				return mockCommand
			},
			expectError:    true,
			expectedErrMsg: "command start error",
		},
		{
			name:      "collector error",
			workspace: "/fake/path",
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) *appMocks.MockCommand {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fndMock.On("Fs").Return(fsMock)
				mockCommand := appMocks.NewMockCommand(t)
				fndMock.On("ExecCommand", ctx, "test-command", []string{"arg1"}).Return(mockCommand)
				stdout := io.NopCloser(strings.NewReader("Hello, stdout!"))
				stderr := io.NopCloser(strings.NewReader("Hello, stderr!"))
				mockCommand.On("StdoutPipe").Return(stdout, nil)
				mockCommand.On("StderrPipe").Return(stderr, nil)
				collectorMock := outputMocks.NewMockCollector(t)
				outMakerMock.On("MakeCollector").Return(collectorMock)
				collectorMock.On("Start", stdout, stderr).Return(fmt.Errorf("collector start error"))
				return mockCommand
			},
			expectError:    true,
			expectedErrMsg: "collector start error",
		},
		{
			name:      "stderr error",
			workspace: "/fake/path",
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) *appMocks.MockCommand {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fndMock.On("Fs").Return(fsMock)
				mockCommand := appMocks.NewMockCommand(t)
				fndMock.On("ExecCommand", ctx, "test-command", []string{"arg1"}).Return(mockCommand)
				stdout := io.NopCloser(strings.NewReader("Hello, stdout!"))
				mockCommand.On("StdoutPipe").Return(stdout, nil)
				mockCommand.On("StderrPipe").Return(nil, fmt.Errorf("stderr error"))
				return mockCommand
			},
			expectError:    true,
			expectedErrMsg: "stderr error",
		},
		{
			name:      "stdout error",
			workspace: "/fake/path",
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) *appMocks.MockCommand {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fndMock.On("Fs").Return(fsMock)
				mockCommand := appMocks.NewMockCommand(t)
				fndMock.On("ExecCommand", ctx, "test-command", []string{"arg1"}).Return(mockCommand)
				mockCommand.On("StdoutPipe").Return(nil, fmt.Errorf("stdout error"))
				return mockCommand
			},
			expectError:    true,
			expectedErrMsg: "stdout error",
		},
		{
			name:      "mkdir error",
			workspace: "/fake/path",
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) *appMocks.MockCommand {
				fsMock := appMocks.NewMockFs(t)
				fndMock.On("Fs").Return(fsMock)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(fmt.Errorf("mkdir error"))
				return nil
			},
			expectError:    true,
			expectedErrMsg: "mkdir error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			outputMakerMock := outputMocks.NewMockMaker(t)
			ctx := context.Background()

			mockCommand := tt.setupMocks(t, ctx, fndMock, outputMakerMock)

			env := &localEnvironment{
				CommonEnvironment: environment.CommonEnvironment{Fnd: fndMock, OutputMaker: outputMakerMock},
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

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
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
				assert.Equal(t, tt.uuid, locTask.id)
			}

			fndMock.AssertExpectations(t)
			if mockCommand != nil {
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
		setupMocks       func(*testing.T, *outputMocks.MockCollector)
		nilTask          bool
		expectError      bool
		expectedOutput   string
		expectedErrorMsg string
	}{
		{
			name:       "successful stdout output collection",
			outputType: output.Stdout,
			setupMocks: func(t *testing.T, om *outputMocks.MockCollector) {
				stdout := io.NopCloser(strings.NewReader("Hello, stdout!"))
				om.On("StdoutReader").Return(stdout)
			},
			expectedOutput: "Hello, stdout!",
		},
		{
			name:       "successful stderr output collection",
			outputType: output.Stderr,
			setupMocks: func(t *testing.T, om *outputMocks.MockCollector) {
				stderr := io.NopCloser(strings.NewReader("Hello, stderr!"))
				om.On("StderrReader").Return(stderr)
			},
			expectedOutput: "Hello, stderr!",
		},
		{
			name:       "successful any output collection",
			outputType: output.Any,
			setupMocks: func(t *testing.T, om *outputMocks.MockCollector) {
				anyout := io.NopCloser(strings.NewReader("outout"))
				om.On("AnyReader").Return(anyout)
			},
			expectedOutput: "outout",
		},
		{
			name:             "unsupported output type",
			outputType:       output.Type(999), // Invalid output type
			expectError:      true,
			expectedErrorMsg: "unsupported output type",
		},
		{
			name:             "nil task",
			outputType:       output.Any, // Invalid output type
			nilTask:          true,
			expectError:      true,
			expectedErrorMsg: "target task is not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ocMock := outputMocks.NewMockCollector(t)
			if tt.setupMocks != nil {
				tt.setupMocks(t, ocMock)
			}

			var testTask task.Task = nil
			if !tt.nilTask {
				testTask = &localTask{
					outputCollector: ocMock,
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
