package parameters

import (
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/generated/app"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name         string
		inputConfig  types.Parameters
		setupMocks   func(t *testing.T, pm *parameterMocks.MockMaker) Parameters
		expectError  bool
		errorMessage string
	}{
		{
			name: "Successful parameter creation",
			inputConfig: types.Parameters{
				"test": "string value",
				"num":  123,
			},
			setupMocks: func(t *testing.T, pm *parameterMocks.MockMaker) Parameters {
				params := Parameters{
					"test": parameterMocks.NewMockParameter(t),
					"num":  parameterMocks.NewMockParameter(t),
				}
				pm.On("Make", "string value").Return(params["test"], nil)
				pm.On("Make", 123).Return(params["num"], nil)
				return params
			},
		},
		{
			name: "Error during parameter creation",
			inputConfig: types.Parameters{
				"test": "string value",
			},
			setupMocks: func(t *testing.T, pm *parameterMocks.MockMaker) Parameters {
				pm.On("Make", "string value").Return(
					nil,
					errors.New("param error"),
				)
				return nil
			},
			expectError:  true,
			errorMessage: "param error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := app.NewMockFoundation(t)
			parameterMakerMock := parameterMocks.NewMockMaker(t)
			params := tt.setupMocks(t, parameterMakerMock)

			maker := CreateMaker(fndMock)
			nm := maker.(*nativeMaker)
			nm.parameterMaker = parameterMakerMock

			result, err := nm.Make(tt.inputConfig)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, params, result)
			}
			parameterMakerMock.AssertExpectations(t)
		})
	}
}

func TestParameters_Inherit(t *testing.T) {
	tests := []struct {
		name           string
		originalParams Parameters
		newParams      Parameters
		expectedResult Parameters
	}{
		{
			name: "Inherit with non-overlapping keys",
			originalParams: Parameters{
				"key1": parameterMocks.NewMockParameter(t),
			},
			newParams: Parameters{
				"key2": parameterMocks.NewMockParameter(t),
			},
			expectedResult: Parameters{
				"key1": parameterMocks.NewMockParameter(t), // unchanged
				"key2": parameterMocks.NewMockParameter(t), // added
			},
		},
		{
			name: "Inherit with overlapping keys",
			originalParams: Parameters{
				"key1": parameterMocks.NewMockParameter(t),
				"key2": parameterMocks.NewMockParameter(t), // this will remain unchanged
			},
			newParams: Parameters{
				"key2": parameterMocks.NewMockParameter(t), // should not replace original
				"key3": parameterMocks.NewMockParameter(t), // new key, should be added
			},
			expectedResult: Parameters{
				"key1": parameterMocks.NewMockParameter(t), // unchanged
				"key2": parameterMocks.NewMockParameter(t), // unchanged
				"key3": parameterMocks.NewMockParameter(t), // added
			},
		},
		{
			name:           "Inherit into empty parameters",
			originalParams: Parameters{},
			newParams: Parameters{
				"key1": parameterMocks.NewMockParameter(t),
			},
			expectedResult: Parameters{
				"key1": parameterMocks.NewMockParameter(t),
			},
		},
		{
			name: "Inherit empty parameters",
			originalParams: Parameters{
				"key1": parameterMocks.NewMockParameter(t),
			},
			newParams: Parameters{},
			expectedResult: Parameters{
				"key1": parameterMocks.NewMockParameter(t),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.originalParams.Inherit(tt.newParams)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
