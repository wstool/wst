package scripts

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	parametersMocks "github.com/bukka/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	"github.com/bukka/wst/run/parameters"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_nativeMaker_Make(t *testing.T) {
	tests := []struct {
		name               string
		config             map[string]types.Script
		setupMocks         func(t *testing.T, pm *parametersMocks.MockMaker) parameters.Parameters
		getExpectedScripts func(params parameters.Parameters) Scripts
		expectError        bool
		errorMessage       string
	}{
		{
			name: "successful script creation",
			config: map[string]types.Script{
				"script1": {
					Content: "echo 'Hello, World!'",
					Path:    "/path/to/script1",
					Mode:    "0755",
					Parameters: types.Parameters{
						"env": "production",
					},
				},
			},
			setupMocks: func(t *testing.T, pm *parametersMocks.MockMaker) parameters.Parameters {
				params := parameters.Parameters{"env": parameterMocks.NewMockParameter(t)}
				pm.On("Make", types.Parameters{"env": "production"}).Return(params, nil)
				return params
			},
			getExpectedScripts: func(params parameters.Parameters) Scripts {
				return Scripts{
					"script1": &nativeScript{
						content:    "echo 'Hello, World!'",
						path:       "/path/to/script1",
						mode:       os.FileMode(0755),
						parameters: params,
					},
				}
			},
		},
		{
			name: "error parsing file mode",
			config: map[string]types.Script{
				"script1": {
					Content: "echo 'Hello, World!'",
					Path:    "/path/to/script1",
					Mode:    "invalid",
					Parameters: types.Parameters{
						"env": "production",
					},
				},
			},
			expectError:  true,
			errorMessage: "error parsing file mode for script script1",
		},
		{
			name: "error parsing parameters",
			config: map[string]types.Script{
				"script1": {
					Content: "echo 'Hello, World!'",
					Path:    "/path/to/script1",
					Mode:    "0755",
					Parameters: map[string]interface{}{
						"env": "production",
					},
				},
			},
			setupMocks: func(t *testing.T, pm *parametersMocks.MockMaker) parameters.Parameters {
				pm.On("Make", types.Parameters{"env": "production"}).Return(nil, errors.New("make err"))
				return nil
			},
			expectError:  true,
			errorMessage: "make err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			parametersMakerMock := parametersMocks.NewMockMaker(t)
			var params parameters.Parameters
			if tt.setupMocks != nil {
				params = tt.setupMocks(t, parametersMakerMock)
			}
			maker := CreateMaker(fndMock, parametersMakerMock)

			actualScripts, err := maker.Make(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
				expectedScripts := tt.getExpectedScripts(params)
				assert.Equal(t, expectedScripts, actualScripts)
			}
			parametersMakerMock.AssertExpectations(t)
		})
	}
}

// Helper function to create a nativeScript instance for testing
func getTestScript(t *testing.T) *nativeScript {
	return &nativeScript{
		content:    "echo Hello, World!",
		path:       "/path/to/script",
		mode:       os.FileMode(0755),
		parameters: parameters.Parameters{"env": parameterMocks.NewMockParameter(t)},
	}
}

func TestNativeScript_Path(t *testing.T) {
	script := getTestScript(t)
	assert.Equal(t, "/path/to/script", script.Path(), "Path method should return the correct script path")
}

func TestNativeScript_Content(t *testing.T) {
	script := getTestScript(t)
	assert.Equal(t, "echo Hello, World!", script.Content(), "Content method should return the correct script content")
}

func TestNativeScript_Mode(t *testing.T) {
	script := getTestScript(t)
	assert.Equal(t, os.FileMode(0755), script.Mode(), "Mode method should return the correct script file mode")
}

func TestNativeScript_Parameters(t *testing.T) {
	script := getTestScript(t)
	assert.Equal(t, script.parameters, script.Parameters(), "Parameters method should return the correct script parameters")
}
