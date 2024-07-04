package local

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
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

func Test_localEnvironment_ExecTaskCommand(t *testing.T) {

}

func Test_localEnvironment_ExecTaskSignal(t *testing.T) {

}

func Test_localEnvironment_RunTask(t *testing.T) {

}

func Test_localEnvironment_Output(t *testing.T) {

}

func Test_localTask_Id(t1 *testing.T) {
}

func Test_localTask_Name(t1 *testing.T) {
}

func Test_localTask_Pid(t1 *testing.T) {
}

func Test_localTask_PrivateUrl(t1 *testing.T) {
}

func Test_localTask_PublicUrl(t1 *testing.T) {
}

func Test_localTask_Type(t1 *testing.T) {
}
