package template

import (
	appMocks "github.com/bukka/wst/mocks/generated/app"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	serviceMocks "github.com/bukka/wst/mocks/generated/run/services/template/service"
	"github.com/bukka/wst/run/parameters/parameter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParameters_GetString(t *testing.T) {
	mockService := serviceMocks.NewMockTemplateService(t)
	mockService.On("EnvironmentConfigPaths").Return(map[string]string{"c": "value"})
	mockParam := parameterMocks.NewMockParameter(t)
	mockParam.On("StringValue").Return("content")
	fndMock := appMocks.NewMockFoundation(t)
	fndMock.TestData().Set("t", 1)
	tmpl := &nativeTemplate{fnd: fndMock, service: mockService}

	params := Parameters{
		"key1": NewParameter(mockParam, nil, tmpl),
	}

	// Test retrieving existing object
	result, err := params.GetString("key1")
	assert.NoError(t, err)
	assert.Equal(t, "content", result)

	// Test missing key
	result, err = params.GetString("key2")
	assert.Empty(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "object string for key [key2] not found")
}

func TestParameters_GetObject(t *testing.T) {
	mockInnerParam := parameterMocks.NewMockParameter(t)
	mockInnerParam.TestData().Set("inner", true)
	mockOuterParam := parameterMocks.NewMockParameter(t)
	mockOuterParam.TestData().Set("outer", true)
	mockMap := map[string]parameter.Parameter{
		"subkey1": mockInnerParam,
	}
	mockOuterParam.On("MapValue").Return(mockMap)
	fndMock := appMocks.NewMockFoundation(t)
	fndMock.TestData().Set("t", 1)
	tmpl := &nativeTemplate{fnd: fndMock}

	params := Parameters{
		"key1": NewParameter(mockOuterParam, nil, tmpl),
	}

	// Test retrieving existing object
	result := params.GetObject("key1")
	expectedResult := Parameters{
		"subkey1": {param: mockInnerParam, params: mockMap, tmpl: tmpl},
	}
	assert.Equal(t, expectedResult, result)

	// Test missing key
	result = params.GetObject("key2")
	assert.Empty(t, result)
}

func TestParameters_GetObjectString(t *testing.T) {
	mockService := serviceMocks.NewMockTemplateService(t)
	mockService.On("EnvironmentConfigPaths").Return(map[string]string{"c": "value"})
	mockInnerParam := parameterMocks.NewMockParameter(t)
	mockInnerParam.TestData().Set("inner", true)
	mockInnerParam.On("StringValue").Return("content")
	mockOuterParam := parameterMocks.NewMockParameter(t)
	mockOuterParam.TestData().Set("outer", true)
	mockMap := map[string]parameter.Parameter{
		"subkey1": mockInnerParam,
	}
	mockOuterParam.On("MapValue").Return(mockMap)
	fndMock := appMocks.NewMockFoundation(t)
	fndMock.TestData().Set("t", 1)
	tmpl := &nativeTemplate{fnd: fndMock, service: mockService}

	params := Parameters{
		"key1": NewParameter(mockOuterParam, nil, tmpl),
	}

	// Test retrieving existing object
	result, err := params.GetObjectString("key1", "subkey1")
	assert.Nil(t, err)
	assert.Equal(t, "content", result)

	// Test missing key
	result, err = params.GetObjectString("key2", "x")
	assert.Empty(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "object string for key [key2][x] not found")
}

func TestParameter_IsNumber(t *testing.T) {
	tests := []struct {
		name      string
		paramType parameter.Type
		expected  bool
	}{
		{"Is Int", parameter.IntType, true},
		{"Is Float", parameter.FloatType, true},
		{"Is String", parameter.StringType, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockParam := parameterMocks.NewMockParameter(t)
			mockParam.On("Type").Return(tt.paramType)
			param := &Parameter{param: mockParam}
			assert.Equal(t, tt.expected, param.IsNumber())
		})
	}
}

func TestParameter_ToNumber(t *testing.T) {
	mockParam := parameterMocks.NewMockParameter(t)
	mockParam.On("FloatValue").Return(3.14)
	param := &Parameter{param: mockParam}

	result := param.ToNumber()
	assert.Equal(t, 3.14, result)

	// Test with nil param
	nilParam := &Parameter{}
	result = nilParam.ToNumber()
	assert.Equal(t, 0.0, result)
}

func TestParameter_ToObject(t *testing.T) {
	mockParam := parameterMocks.NewMockParameter(t)
	mockInnerParam := parameterMocks.NewMockParameter(t)
	mockMap := map[string]parameter.Parameter{
		"inner": mockInnerParam,
	}
	mockParam.On("MapValue").Return(mockMap)
	mockTemplate := &nativeTemplate{}

	param := &Parameter{param: mockParam, tmpl: mockTemplate}
	result := param.ToObject()

	assert.NotEmpty(t, result)
	assert.Contains(t, result, "inner")

	// Test with nil param
	nilParam := &Parameter{}
	result = nilParam.ToObject()
	assert.Empty(t, result)
}

func TestParameter_ToString(t *testing.T) {
	mockService := serviceMocks.NewMockTemplateService(t)
	mockService.On("EnvironmentConfigPaths").Return(map[string]string{"c": "value"})
	fndMock := appMocks.NewMockFoundation(t)
	tmpl := &nativeTemplate{fnd: fndMock, service: mockService}

	tests := []struct {
		name           string
		setupParam     func(*parameterMocks.MockParameter)
		emptyParam     bool
		expectError    bool
		expected       string
		expectedErrMsg string
	}{
		{
			name: "Successful rendering",
			setupParam: func(mp *parameterMocks.MockParameter) {
				mp.On("StringValue").Return("Hello, World").Once()
			},
			expectError: false,
			expected:    "Hello, World",
		},
		{
			name: "Rendering error",
			setupParam: func(mp *parameterMocks.MockParameter) {
				mp.On("StringValue").Return("{{.xx}} Hello, World")
			},
			expectError:    true,
			expectedErrMsg: "can't evaluate field xx",
		},
		{
			name: "Recursive rendering detected",
			setupParam: func(mp *parameterMocks.MockParameter) {
				mp.On("StringValue").Return("recursive")
			},
			expectError:    true,
			expectedErrMsg: "recursive rendering of parameter",
		},
		{
			name:           "Parameter not set",
			setupParam:     func(mp *parameterMocks.MockParameter) {},
			emptyParam:     true,
			expectError:    true,
			expectedErrMsg: "trying to render not set parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockParam := parameterMocks.NewMockParameter(t)
			if tt.setupParam != nil {
				tt.setupParam(mockParam)
			}
			if tt.emptyParam {
				mockParam = nil
			}

			param := NewParameter(mockParam, nil, tmpl)
			if tt.name == "Recursive rendering detected" {
				param.startedRendering = true
			}

			result, err := param.ToString()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
				// Try again to see if rendered value works
				result, err = param.ToString()
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)

			}
		})
	}
}
