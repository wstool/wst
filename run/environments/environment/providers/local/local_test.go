package local

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	outputMocks "github.com/wstool/wst/mocks/generated/run/environments/environment/output"
	taskMocks "github.com/wstool/wst/mocks/generated/run/environments/task"
	"github.com/wstool/wst/run/environments/environment"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/environments/environment/providers"
	"github.com/wstool/wst/run/environments/task"
	"io"
	"os"
	"strings"
	"testing"
	"time"
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
	assert.Equal(t, "/tmp/ws/svc/_env", l.RootPath("/tmp/ws/svc"))
}

func Test_localEnvironment_Mkdir(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	fsMock := appMocks.NewMockFs(t)
	fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
	fndMock.On("Fs").Return(fsMock)
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
	assert.Nil(t, l.Mkdir("svc", "/fake/path", os.FileMode(0755)))
}

func Test_localEnvironment_ServiceLocalAddress(t *testing.T) {
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
	assert.Equal(t, "127.0.0.1:1234", l.ServiceLocalAddress("svc", 1234, 80))
}

func Test_localEnvironment_ServicePrivateAddress(t *testing.T) {
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
	assert.Equal(t, "127.0.0.1:1234", l.ServicePrivateAddress("svc", 1234, 80))
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
			expectedErrMsg: "failure when creating workspace directory: permission denied",
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

			ctx := context.Background()
			err := env.Init(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.True(t, env.initialized)
				assert.Equal(t, ctx, env.ctx)
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
		name        string
		workspace   string
		initialized bool
		setupMocks  func(
			*testing.T,
			context.Context,
			context.Context,
			*appMocks.MockFoundation,
			*outputMocks.MockMaker,
		) (*appMocks.MockCommand, chan struct{})
		updateServiceSetting func(ss *environment.ServiceSettings)
		expectError          bool
		expectedErrMsg       string
		expectedLogs         []string
		expectTask           bool
		uuid                 string // UUID for each task
	}{
		{
			name:      "successfully runs task with configs and scripts",
			workspace: "/fake/path",
			updateServiceSetting: func(ss *environment.ServiceSettings) {
				ss.WorkspaceConfigPaths = map[string]string{
					"cfg": "/ws/path/c/cfg.json",
				}
				ss.EnvironmentConfigPaths = map[string]string{
					"cfg": "/env/path/c/cfg.json",
				}
				ss.WorkspaceScriptPaths = map[string]string{
					"script": "/ws/path/s/script.sh",
				}
				ss.EnvironmentScriptPaths = map[string]string{
					"script": "/env/path/s/script.sh",
				}
			},
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fsMock.On("MkdirAll", "/env/path/c", os.FileMode(0755)).Return(nil)
				fsMock.On("MkdirAll", "/env/path/s", os.FileMode(0755)).Return(nil)

				srcConfigFile := appMocks.NewMockFile(t)
				srcConfigFile.On("Read", mock.Anything).Return(len("content of the file"), nil).Once()
				srcConfigFile.On("Read", mock.Anything).Return(0, io.EOF)
				srcConfigFile.On("Close").Return(nil).Once()

				dstConfigFile := appMocks.NewMockFile(t)
				dstConfigFile.On("Write", mock.Anything).Return(len("content of the file"), nil).Once()
				dstConfigFile.On("Close").Return(nil).Once()

				srcScriptFile := appMocks.NewMockFile(t)
				srcScriptFile.On("Read", mock.Anything).Return(len("content of the file"), nil).Once()
				srcScriptFile.On("Read", mock.Anything).Return(0, io.EOF)
				srcScriptFile.On("Close").Return(nil).Once()

				dstScriptFile := appMocks.NewMockFile(t)
				dstScriptFile.On("Write", mock.Anything).Return(len("content of the file"), nil).Once()
				dstScriptFile.On("Close").Return(nil).Once()

				fsMock.On("Open", "/ws/path/c/cfg.json").Return(srcConfigFile, nil)
				fsMock.On("Create", "/env/path/c/cfg.json").Return(dstConfigFile, nil)
				fsMock.On("Open", "/ws/path/s/script.sh").Return(srcScriptFile, nil)
				fsMock.On("Create", "/env/path/s/script.sh").Return(dstScriptFile, nil)

				fndMock.On("Fs").Return(fsMock)
				mockCommand := appMocks.NewMockCommand(t)
				mockCommand.On("String").Return("test-command arg1")
				fndMock.On("ExecCommand", actionCtx, "test-command", []string{"arg1"}).Return(mockCommand)

				stdoutWriter := &bytes.Buffer{}
				stderrWriter := &bytes.Buffer{}
				stdoutWriter.Write([]byte("Stdout!"))
				stderrWriter.Write([]byte("Stderr!"))
				collectorMock := outputMocks.NewMockCollector(t)
				outMakerMock.On("MakeCollector", "uuid-123").Return(collectorMock)
				collectorMock.On("StdoutWriter").Return(stdoutWriter)
				collectorMock.On("StderrWriter").Return(stderrWriter)

				// Channel to signal the completion of awaitTask
				taskFinishedChan := make(chan struct{})

				mockCommand.On("SetStdout", stdoutWriter)
				mockCommand.On("SetStderr", stderrWriter)
				fndMock.On("GenerateUuid").Return("uuid-123")

				mockCommand.On("Start").Return(nil)
				mockCommand.On("Wait").Return(nil).Run(func(args mock.Arguments) {
					// Simulate command execution time
					time.Sleep(50 * time.Millisecond)
				})

				// Mock collector close and signal completion
				collectorMock.On("Close").Return(nil).Run(func(args mock.Arguments) {
					close(taskFinishedChan) // Signal task completion
				}).Once()

				return mockCommand, taskFinishedChan
			},
			expectTask: true,
			expectedLogs: []string{
				"Initializing local environment before running task",
				"Creating command: test-command arg1",
				"Task uuid-123 started for command: test-command",
				"Task uuid-123 command finished",
			},
			uuid: "uuid-123",
		},
		{
			name:        "successfully runs initialized task",
			workspace:   "/fake/path",
			initialized: true,
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fndMock.On("Fs").Return(fsMock)

				mockCommand := appMocks.NewMockCommand(t)
				mockCommand.On("String").Return("test-command arg1")
				fndMock.On("ExecCommand", envCtx, "test-command", []string{"arg1"}).Return(mockCommand)

				stdoutWriter := &bytes.Buffer{}
				stderrWriter := &bytes.Buffer{}
				stdoutWriter.Write([]byte("Stdout!"))
				stderrWriter.Write([]byte("Stderr!"))
				collectorMock := outputMocks.NewMockCollector(t)
				outMakerMock.On("MakeCollector", "uuid-123").Return(collectorMock)
				collectorMock.On("StdoutWriter").Return(stdoutWriter)
				collectorMock.On("StderrWriter").Return(stderrWriter)

				// Channel to signal the completion of awaitTask
				taskFinishedChan := make(chan struct{})

				mockCommand.On("SetStdout", stdoutWriter)
				mockCommand.On("SetStderr", stderrWriter)
				fndMock.On("GenerateUuid").Return("uuid-123")

				mockCommand.On("Start").Return(nil)
				mockCommand.On("Wait").Return(nil).Run(func(args mock.Arguments) {
					// Simulate command execution time
					time.Sleep(50 * time.Millisecond)
				})

				// Mock collector close and signal completion
				collectorMock.On("Close").Return(nil).Run(func(args mock.Arguments) {
					close(taskFinishedChan) // Signal task completion
				}).Once()

				return mockCommand, taskFinishedChan
			},
			expectTask: true,
			expectedLogs: []string{
				"Creating command: test-command arg1",
				"Task uuid-123 started for command: test-command",
				"Task uuid-123 command finished",
			},
			uuid: "uuid-123",
		},
		{
			name:      "successfully start but failed wait and close",
			workspace: "/fake/path",
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fndMock.On("Fs").Return(fsMock)
				mockCommand := appMocks.NewMockCommand(t)
				mockCommand.On("String").Return("test-command arg1")
				fndMock.On("ExecCommand", actionCtx, "test-command", []string{"arg1"}).Return(mockCommand)

				stdoutWriter := &bytes.Buffer{}
				stderrWriter := &bytes.Buffer{}
				stdoutWriter.Write([]byte("Stdout!"))
				stderrWriter.Write([]byte("Stderr!"))
				collectorMock := outputMocks.NewMockCollector(t)
				outMakerMock.On("MakeCollector", "uuid-123").Return(collectorMock)
				collectorMock.On("StdoutWriter").Return(stdoutWriter)
				collectorMock.On("StderrWriter").Return(stderrWriter)

				// Channel to signal the completion of awaitTask
				taskFinishedChan := make(chan struct{})

				mockCommand.On("SetStdout", stdoutWriter)
				mockCommand.On("SetStderr", stderrWriter)
				fndMock.On("GenerateUuid").Return("uuid-123")

				mockCommand.On("Start").Return(nil)
				mockCommand.On("Wait").Return(errors.New("wait fail")).Run(func(args mock.Arguments) {
					// Simulate command execution time
					time.Sleep(50 * time.Millisecond)
				})

				// Mock collector close, log output and signal completion
				collectorMock.On("Close").Return(errors.New("close fail")).Once()
				collectorMock.On("LogOutput").Run(func(args mock.Arguments) {
					close(taskFinishedChan) // Signal task completion
				}).Once()

				return mockCommand, taskFinishedChan
			},
			expectTask: true,
			expectedLogs: []string{
				"Initializing local environment before running task",
				"Creating command: test-command arg1",
				"Task uuid-123 started for command: test-command",
				"Waiting for local task uuid-123 failed: wait fail",
				"Closing output collector for local task uuid-123 failed: close fail",
			},
			uuid: "uuid-123",
		},
		{
			name:      "command start error",
			workspace: "/fake/path",
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fndMock.On("Fs").Return(fsMock)
				mockCommand := appMocks.NewMockCommand(t)
				mockCommand.On("String").Return("test-command arg1")
				fndMock.On("ExecCommand", actionCtx, "test-command", []string{"arg1"}).Return(mockCommand)
				stdoutWriter := &bytes.Buffer{}
				stderrWriter := &bytes.Buffer{}
				stdoutWriter.Write([]byte("Stdout!"))
				stderrWriter.Write([]byte("Stderr!"))
				collectorMock := outputMocks.NewMockCollector(t)
				outMakerMock.On("MakeCollector", "uuid-123").Return(collectorMock)
				collectorMock.On("StdoutWriter").Return(stdoutWriter)
				collectorMock.On("StderrWriter").Return(stderrWriter)
				mockCommand.On("SetStdout", stdoutWriter)
				mockCommand.On("SetStderr", stderrWriter)
				fndMock.On("GenerateUuid").Return("uuid-123")
				mockCommand.On("Start").Return(fmt.Errorf("command start error"))
				fndMock.On("GenerateUuid").Return("uuid-123")

				// No need for taskFinishedChan in this error scenario
				return mockCommand, nil
			},
			expectError:    true,
			expectedErrMsg: "command start error",
		},
		{
			name:      "dst write error",
			workspace: "/fake/path",
			updateServiceSetting: func(ss *environment.ServiceSettings) {
				ss.WorkspaceConfigPaths = map[string]string{
					"cfg": "/ws/path/c/cfg.json",
				}
				ss.EnvironmentConfigPaths = map[string]string{
					"cfg": "/env/path/c/cfg.json",
				}
				ss.WorkspaceScriptPaths = map[string]string{
					"script": "/ws/path/s/script.sh",
				}
				ss.EnvironmentScriptPaths = map[string]string{
					"script": "/env/path/s/script.sh",
				}
			},
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fsMock.On("MkdirAll", "/env/path/c", os.FileMode(0755)).Return(nil)
				fsMock.On("MkdirAll", "/env/path/s", os.FileMode(0755)).Return(nil)

				srcConfigFile := appMocks.NewMockFile(t)
				srcConfigFile.On("Read", mock.Anything).Return(len("content of the file"), nil).Once()
				srcConfigFile.On("Read", mock.Anything).Return(0, io.EOF)
				srcConfigFile.On("Close").Return(nil).Once()

				dstConfigFile := appMocks.NewMockFile(t)
				dstConfigFile.On("Write", mock.Anything).Return(len("content of the file"), nil).Once()
				dstConfigFile.On("Close").Return(nil).Once()

				srcScriptFile := appMocks.NewMockFile(t)
				srcScriptFile.On("Read", mock.Anything).Return(len("content of the file"), nil).Once()
				srcScriptFile.On("Read", mock.Anything).Return(0, io.EOF)
				srcScriptFile.On("Close").Return(nil).Once()

				dstScriptFile := appMocks.NewMockFile(t)
				dstScriptFile.On("Write", mock.Anything).Return(0, errors.New("dst write error")).Once()
				dstScriptFile.On("Close").Return(nil).Once()

				fsMock.On("Open", "/ws/path/c/cfg.json").Return(srcConfigFile, nil)
				fsMock.On("Create", "/env/path/c/cfg.json").Return(dstConfigFile, nil)
				fsMock.On("Open", "/ws/path/s/script.sh").Return(srcScriptFile, nil)
				fsMock.On("Create", "/env/path/s/script.sh").Return(dstScriptFile, nil)

				fndMock.On("Fs").Return(fsMock)

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "dst write error",
		},
		{
			name:      "src read error",
			workspace: "/fake/path",
			updateServiceSetting: func(ss *environment.ServiceSettings) {
				ss.WorkspaceConfigPaths = map[string]string{
					"cfg": "/ws/path/c/cfg.json",
				}
				ss.EnvironmentConfigPaths = map[string]string{
					"cfg": "/env/path/c/cfg.json",
				}
				ss.WorkspaceScriptPaths = map[string]string{
					"script": "/ws/path/s/script.sh",
				}
				ss.EnvironmentScriptPaths = map[string]string{
					"script": "/env/path/s/script.sh",
				}
			},
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fsMock.On("MkdirAll", "/env/path/c", os.FileMode(0755)).Return(nil)

				srcConfigFile := appMocks.NewMockFile(t)
				srcConfigFile.On("Read", mock.Anything).Return(0, errors.New("src read error"))
				srcConfigFile.On("Close").Return(nil).Once()

				dstConfigFile := appMocks.NewMockFile(t)
				dstConfigFile.On("Close").Return(nil).Once()

				fsMock.On("Open", "/ws/path/c/cfg.json").Return(srcConfigFile, nil)
				fsMock.On("Create", "/env/path/c/cfg.json").Return(dstConfigFile, nil)

				fndMock.On("Fs").Return(fsMock)

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "src read error",
		},
		{
			name:      "dst create error",
			workspace: "/fake/path",
			updateServiceSetting: func(ss *environment.ServiceSettings) {
				ss.WorkspaceConfigPaths = map[string]string{
					"cfg": "/ws/path/c/cfg.json",
				}
				ss.EnvironmentConfigPaths = map[string]string{
					"cfg": "/env/path/c/cfg.json",
				}
			},
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fsMock.On("MkdirAll", "/env/path/c", os.FileMode(0755)).Return(nil)

				srcConfigFile := appMocks.NewMockFile(t)
				srcConfigFile.On("Close").Return(nil).Once()

				fsMock.On("Open", "/ws/path/c/cfg.json").Return(srcConfigFile, nil)
				fsMock.On("Create", "/env/path/c/cfg.json").Return(
					nil, errors.New("dst create error"))

				fndMock.On("Fs").Return(fsMock)

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "dst create error",
		},
		{
			name:      "src open error",
			workspace: "/fake/path",
			updateServiceSetting: func(ss *environment.ServiceSettings) {
				ss.WorkspaceConfigPaths = map[string]string{
					"cfg": "/ws/path/c/cfg.json",
				}
				ss.EnvironmentConfigPaths = map[string]string{
					"cfg": "/env/path/c/cfg.json",
				}
			},
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fsMock.On("MkdirAll", "/env/path/c", os.FileMode(0755)).Return(nil)

				fsMock.On("Open", "/ws/path/c/cfg.json").Return(
					nil, errors.New("src open error"))

				fndMock.On("Fs").Return(fsMock)

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "src open error",
		},
		{
			name:      "dst mkdir error",
			workspace: "/fake/path",
			updateServiceSetting: func(ss *environment.ServiceSettings) {
				ss.WorkspaceConfigPaths = map[string]string{
					"cfg": "/ws/path/c/cfg.json",
				}
				ss.EnvironmentConfigPaths = map[string]string{
					"cfg": "/env/path/c/cfg.json",
				}
			},
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fsMock.On("MkdirAll", "/env/path/c", os.FileMode(0755)).Return(
					errors.New("dst mkdir error"))

				fndMock.On("Fs").Return(fsMock)

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "dst mkdir error",
		},
		{
			name:      "paths mismatch error",
			workspace: "/fake/path",
			updateServiceSetting: func(ss *environment.ServiceSettings) {
				ss.WorkspaceConfigPaths = map[string]string{
					"cfg": "/ws/path/c/cfg.json",
				}
				ss.EnvironmentConfigPaths = map[string]string{
					"unknown": "/env/path/c/cfg.json",
				}
			},
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(nil)
				fndMock.On("Fs").Return(fsMock)

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "configs environment path not found for cfg",
		},
		{
			name:      "mkdir error",
			workspace: "/fake/path",
			setupMocks: func(
				t *testing.T,
				actionCtx context.Context,
				envCtx context.Context,
				fndMock *appMocks.MockFoundation,
				outMakerMock *outputMocks.MockMaker,
			) (*appMocks.MockCommand, chan struct{}) {
				fsMock := appMocks.NewMockFs(t)
				fndMock.On("Fs").Return(fsMock)
				fsMock.On("MkdirAll", "/fake/path", os.FileMode(0755)).Return(fmt.Errorf("mkdir error"))

				return nil, nil
			},
			expectError:    true,
			expectedErrMsg: "mkdir error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			outputMakerMock := outputMocks.NewMockMaker(t)
			actionCtx := context.WithValue(context.Background(), "type", "action")
			envCtx := context.WithValue(context.Background(), "type", "env")

			logger := external.NewMockLogger()
			fndMock.On("Logger").Return(logger.SugaredLogger)

			mockCommand, taskFinishedChan := tt.setupMocks(t, actionCtx, envCtx, fndMock, outputMakerMock)

			env := &localEnvironment{
				CommonEnvironment: environment.CommonEnvironment{Fnd: fndMock, OutputMaker: outputMakerMock},
				workspace:         tt.workspace,
				initialized:       tt.initialized,
				ctx:               envCtx,
				tasks:             make(map[string]*localTask),
			}

			ss := &environment.ServiceSettings{
				Name: "test-service",
				Port: 8080,
			}
			if tt.updateServiceSetting != nil {
				tt.updateServiceSetting(ss)
			}
			cmd := &environment.Command{
				Name: "test-command",
				Args: []string{"arg1"},
			}

			resultTask, err := env.RunTask(actionCtx, ss, cmd)

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

				// Wait for the task to finish
				<-taskFinishedChan
				time.Sleep(50 * time.Millisecond)
				if tt.expectedLogs != nil {
					assert.Equal(t, tt.expectedLogs, logger.Messages())
				}

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
				lt := &localTask{
					serviceName: "local-service",
					cmd:         &app.ExecCommand{},
				}
				lt.serviceRunning.Store(true)
				lt.cmd = &app.ExecCommand{} // Stubbing the command to be returned
				return lt
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
				lt := &localTask{
					serviceName: "local-service",
					cmd:         &app.ExecCommand{}, // Using a valid local task type
				}
				lt.serviceRunning.Store(true)
				lt.cmd = &app.ExecCommand{} // Stubbing the command to be returned
				return lt
			},
			command: &environment.Command{
				Name: "echo",
				Args: []string{"error"},
			},
			expectError:      true,
			expectedErrorMsg: "execution failed",
		},
		{
			name: "error task not running",
			target: func() task.Task {
				lt := &localTask{
					id:          "tid",
					serviceName: "local-service",
					cmd:         &app.ExecCommand{}, // Using a valid local task type
				}
				lt.serviceRunning.Store(false)
				lt.cmd = &app.ExecCommand{} // Stubbing the command to be returned
				return lt
			},
			command: &environment.Command{
				Name: "echo",
				Args: []string{"hello"},
			},
			expectError:      true,
			expectedErrorMsg: "task tid is not running",
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
				lt := &localTask{
					serviceName: "local-service",
					cmd:         cmdMock,
				}
				lt.serviceRunning.Store(true)
				return lt
			},
			signal:      os.Interrupt,
			expectError: false,
		},
		{
			name: "error during signal execution",
			target: func(t *testing.T) task.Task {
				cmdMock := appMocks.NewMockCommand(t)
				cmdMock.On("ProcessSignal", os.Kill).Return(fmt.Errorf("failed to send signal"))
				lt := &localTask{
					serviceName: "local-service",
					cmd:         cmdMock,
				}
				lt.serviceRunning.Store(true)
				return lt
			},
			signal:           os.Kill,
			expectError:      true,
			expectedErrorMsg: "failed to send signal",
		},
		{
			name: "task is not running",
			target: func(t *testing.T) task.Task {
				cmdMock := appMocks.NewMockCommand(t)
				lt := &localTask{
					id:          "uuid-tid",
					serviceName: "local-service",
					cmd:         cmdMock,
				}
				return lt
			},
			signal:           os.Kill,
			expectError:      true,
			expectedErrorMsg: "task uuid-tid is not running",
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
		setupMocks       func(*testing.T, context.Context, *outputMocks.MockCollector)
		nilTask          bool
		expectError      bool
		expectedOutput   string
		expectedErrorMsg string
	}{
		{
			name:       "successful stdout output collection",
			outputType: output.Stdout,
			setupMocks: func(t *testing.T, ctx context.Context, om *outputMocks.MockCollector) {
				stdout := io.NopCloser(strings.NewReader("Hello, stdout!"))
				om.On("StdoutReader", ctx).Return(stdout)
			},
			expectedOutput: "Hello, stdout!",
		},
		{
			name:       "successful stderr output collection",
			outputType: output.Stderr,
			setupMocks: func(t *testing.T, ctx context.Context, om *outputMocks.MockCollector) {
				stderr := io.NopCloser(strings.NewReader("Hello, stderr!"))
				om.On("StderrReader", ctx).Return(stderr)
			},
			expectedOutput: "Hello, stderr!",
		},
		{
			name:       "successful any output collection",
			outputType: output.Any,
			setupMocks: func(t *testing.T, ctx context.Context, om *outputMocks.MockCollector) {
				anyout := io.NopCloser(strings.NewReader("outout"))
				om.On("AnyReader", ctx).Return(anyout)
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
				tt.setupMocks(t, ctx, ocMock)
			}

			var testTask *localTask = nil
			if !tt.nilTask {
				testTask = &localTask{
					outputCollector: ocMock,
				}
				testTask.serviceRunning.Store(true)
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
	lt := &localTask{
		id:          "lid",
		executable:  "ep",
		cmd:         cmdMock,
		serviceName: "lids",
		serviceUrl:  "http://localhost:1234",
	}
	lt.serviceRunning.Store(true)
	return lt
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

func Test_localTask_IsRunning(t *testing.T) {
	assert.True(t, getTestTask(t).IsRunning())
}

func Test_localTask_Type(t *testing.T) {
	assert.Equal(t, providers.LocalType, getTestTask(t).Type())
}
