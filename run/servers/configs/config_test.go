package configs

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	parametersMocks "github.com/bukka/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	"github.com/bukka/wst/run/parameters"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func testParams(t *testing.T, len int) []*parameterMocks.MockParameter {
	params := make([]*parameterMocks.MockParameter, len)
	for i := 0; i < len; i++ {
		param := parameterMocks.NewMockParameter(t)
		// Differentiate params
		param.TestData().Set("id", i)
		params[i] = param
	}
	return params
}

func TestConfigs_Inherit(t *testing.T) {
	params := testParams(t, 2)
	tests := []struct {
		name            string
		childConfigs    Configs
		parentConfigs   Configs
		expectedConfigs Configs
	}{
		{
			name: "inherit new configs",
			childConfigs: Configs{
				"existing": &nativeConfig{
					file:       "/path/existing",
					parameters: parameters.Parameters{"param1": params[0]},
				},
			},
			parentConfigs: Configs{
				"new": &nativeConfig{
					file:       "/path/new",
					parameters: parameters.Parameters{"param2": params[1]},
				},
			},
			expectedConfigs: Configs{
				"existing": &nativeConfig{
					file:       "/path/existing",
					parameters: parameters.Parameters{"param1": params[0]},
				},
				"new": &nativeConfig{
					file:       "/path/new",
					parameters: parameters.Parameters{"param2": params[1]},
				},
			},
		},
		{
			name: "do not override existing configs",
			childConfigs: Configs{
				"common": &nativeConfig{
					file:       "/path/common1",
					parameters: parameters.Parameters{"param1": params[0]},
				},
			},
			parentConfigs: Configs{
				"common": &nativeConfig{
					file:       "/path/common2",
					parameters: parameters.Parameters{"param2": params[1]}},
			},
			expectedConfigs: Configs{
				"common": &nativeConfig{
					file:       "/path/common1",
					parameters: parameters.Parameters{"param1": params[0]},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.childConfigs.Inherit(tt.parentConfigs)
			assert.Equal(t, tt.expectedConfigs, tt.childConfigs, "Configs should be correctly inherited")
		})
	}
}

func Test_nativeMaker_Make(t *testing.T) {
	params := testParams(t, 1)
	tests := []struct {
		name             string
		serverConfigs    map[string]types.ServerConfig
		setupMocks       func(pm *parametersMocks.MockMaker)
		expectedConfigs  Configs
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful config creation",
			serverConfigs: map[string]types.ServerConfig{
				"config1": {File: "/path/config1", Parameters: types.Parameters{"key": "value"}},
			},
			setupMocks: func(pm *parametersMocks.MockMaker) {
				pm.On("Make", types.Parameters{
					"key": "value",
				}).Return(parameters.Parameters{"key": params[0]}, nil)
			},
			expectedConfigs: Configs{
				"config1": &nativeConfig{
					file:       "/path/config1",
					parameters: parameters.Parameters{"key": params[0]},
				},
			},
			expectError: false,
		},
		{
			name: "failure in parameters maker",
			serverConfigs: map[string]types.ServerConfig{
				"config1": {File: "/path/config1", Parameters: types.Parameters{"key": "value"}},
			},
			setupMocks: func(pm *parametersMocks.MockMaker) {
				pm.On("Make", types.Parameters{
					"key": "value",
				}).Return(nil, errors.New("param fail"))
			},
			expectError:      true,
			expectedErrorMsg: "param fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			parametersMock := parametersMocks.NewMockMaker(t)
			tt.setupMocks(parametersMock)
			maker := CreateMaker(fndMock, parametersMock)

			configs, err := maker.Make(tt.serverConfigs)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedConfigs, configs)
			}

			parametersMock.AssertExpectations(t)
		})
	}
}

func Test_nativeConfig_FilePath(t *testing.T) {
	params := testParams(t, 1)
	config := &nativeConfig{
		file:       "/path/config1",
		parameters: parameters.Parameters{"key": params[0]},
	}
	assert.Equal(t, "/path/config1", config.FilePath())
}

func Test_nativeConfig_Parameters(t *testing.T) {
	params := testParams(t, 1)
	config := &nativeConfig{
		file:       "/path/config1",
		parameters: parameters.Parameters{"key": params[0]},
	}
	assert.Equal(t, parameters.Parameters{"key": params[0]}, config.Parameters())
}
