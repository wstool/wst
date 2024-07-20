package local

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	commonMocks "github.com/bukka/wst/mocks/generated/run/sandboxes/sandbox/common"
	"github.com/bukka/wst/run/sandboxes/dir"
	"github.com/bukka/wst/run/sandboxes/hooks"
	"github.com/bukka/wst/run/sandboxes/sandbox/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_nativeMaker_MakeSandbox(t *testing.T) {
	tests := []struct {
		name             string
		inputConfig      *types.LocalSandbox
		mockSetup        func(*commonMocks.MockMaker) *common.Sandbox
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful sandbox creation",
			inputConfig: &types.LocalSandbox{
				Available: true,
				Dirs: map[string]string{
					"conf": "/etc/app",
					"run":  "/var/run/app",
				},
				Hooks: map[string]types.SandboxHook{
					"start": &types.SandboxHookNative{Enabled: true},
				},
			},
			mockSetup: func(m *commonMocks.MockMaker) *common.Sandbox {
				commonSandbox := common.CreateSandbox(true, make(map[dir.DirType]string), make(hooks.Hooks))
				m.On("MakeSandbox", &types.CommonSandbox{
					Available: true,
					Dirs: map[string]string{
						"conf": "/etc/app",
						"run":  "/var/run/app",
					},
					Hooks: map[string]types.SandboxHook{
						"start": &types.SandboxHookNative{Enabled: true},
					},
				}).Return(commonSandbox, nil)
				return commonSandbox
			},
		},
		{
			name: "failed sandbox creation due to container maker error",

			inputConfig: &types.LocalSandbox{
				Available: true,
				Dirs: map[string]string{
					"conf": "/etc/app",
					"run":  "/var/run/app",
				},
				Hooks: map[string]types.SandboxHook{
					"start": &types.SandboxHookNative{Enabled: true},
				},
			},
			mockSetup: func(m *commonMocks.MockMaker) *common.Sandbox {
				m.On("MakeSandbox", &types.CommonSandbox{
					Available: true,
					Dirs: map[string]string{
						"conf": "/etc/app",
						"run":  "/var/run/app",
					},
					Hooks: map[string]types.SandboxHook{
						"start": &types.SandboxHookNative{Enabled: true},
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
			commonMakerMock := commonMocks.NewMockMaker(t)
			containerSandbox := tt.mockSetup(commonMakerMock)

			maker := CreateMaker(fndMock, commonMakerMock)
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

			commonMakerMock.AssertExpectations(t)
		})
	}
}
