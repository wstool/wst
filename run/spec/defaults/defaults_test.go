package defaults

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	parametersMocks "github.com/wstool/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters/parameter"
	"github.com/wstool/wst/run/parameters"
	"testing"
)

func TestCreateMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	parametersMakerMock := parametersMocks.NewMockMaker(t)

	maker := CreateMaker(fndMock, parametersMakerMock)

	nm, ok := maker.(*nativeMaker)
	assert.True(t, ok, "expected maker to be of type *nativeMaker")
	assert.NotNil(t, nm)

	assert.Equal(t, fndMock, nm.fnd)
	assert.Equal(t, parametersMakerMock, nm.parametersMaker)
}

func Test_nativeMaker_Make(t *testing.T) {
	param1 := parameterMocks.NewMockParameter(t)
	tests := []struct {
		name           string
		config         *types.SpecDefaults
		setupMocks     func(*parametersMocks.MockMaker)
		expected       *Defaults
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "successfully makes defaults",
			config: &types.SpecDefaults{
				Service: types.SpecServiceDefaults{
					Sandbox: "docker",
					Server: types.SpecServiceServerDefaults{
						Tag: "debian",
					},
				},
				Timeouts: types.SpecTimeouts{
					Actions: 30,
					Action:  5,
				},
				Parameters: types.Parameters{
					"param1": "value1",
				},
			},
			setupMocks: func(pm *parametersMocks.MockMaker) {
				expectedParameters := parameters.Parameters{
					"param1": param1,
				}
				pm.On("Make", types.Parameters{
					"param1": "value1",
				}).Return(expectedParameters, nil)
			},
			expected: &Defaults{
				Service: ServiceDefaults{
					Sandbox: "docker",
					Server:  ServiceServerDefaults{Tag: "debian"},
				},
				Timeouts: TimeoutsDefaults{
					Actions: 30,
					Action:  5,
				},
				Parameters: parameters.Parameters{"param1": param1},
			},
			expectError: false,
		},
		{
			name: "parameters maker returns error",
			config: &types.SpecDefaults{
				Service: types.SpecServiceDefaults{
					Sandbox: "docker",
					Server: types.SpecServiceServerDefaults{
						Tag: "debian",
					},
				},
				Timeouts: types.SpecTimeouts{
					Actions: 30,
					Action:  5,
				},
				Parameters: types.Parameters{
					"param1": "value1",
				},
			},
			setupMocks: func(pm *parametersMocks.MockMaker) {
				pm.On("Make", types.Parameters{
					"param1": "value1",
				}).Return(nil, errors.New("failed params make"))
			},
			expectError:    true,
			expectedErrMsg: "failed params make",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			parametersMakerMock := parametersMocks.NewMockMaker(t)

			tt.setupMocks(parametersMakerMock)
			maker := CreateMaker(fndMock, parametersMakerMock)
			result, err := maker.Make(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			parametersMakerMock.AssertExpectations(t)
		})
	}
}
