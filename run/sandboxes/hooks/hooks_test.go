package hooks

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	"github.com/stretchr/testify/assert"
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
			},
			expectedHooks: Hooks{
				StartHookType:   &HookNative{BaseHook: BaseHook{Enabled: true, Type: StartHookType}},
				StopHookType:    &HookSignal{BaseHook: BaseHook{Enabled: true, Type: StopHookType}, Signal: syscall.SIGTERM},
				RestartHookType: &HookSignal{BaseHook: BaseHook{Enabled: true, Type: RestartHookType}, Signal: syscall.SIGKILL},
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
