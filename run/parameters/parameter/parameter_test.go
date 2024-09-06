package parameter

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/generated/app"
	"testing"
)

func Test_nativeMaker_Make(t *testing.T) {
	fndMock := app.NewMockFoundation(t)
	maker := CreateMaker(fndMock)

	tests := []struct {
		name           string
		config         interface{}
		expectedParam  *parameter
		expectedType   Type
		expectingError bool
	}{
		{
			name:   "Bool Type",
			config: true,
			expectedParam: &parameter{
				parameterType: BoolType,
				boolValue:     true,
			},
			expectedType:   BoolType,
			expectingError: false,
		},
		{
			name:   "Int Type",
			config: 42,
			expectedParam: &parameter{
				parameterType: IntType,
				intValue:      42,
			},
			expectedType:   IntType,
			expectingError: false,
		},
		{
			name:   "Float Type",
			config: 3.14,
			expectedParam: &parameter{
				parameterType: FloatType,
				floatValue:    3.14,
			},
			expectedType:   FloatType,
			expectingError: false,
		},
		{
			name:   "String Type",
			config: "test",
			expectedParam: &parameter{
				parameterType: StringType,
				stringValue:   "test",
			},
			expectedType:   StringType,
			expectingError: false,
		},
		{
			name:   "Array Type",
			config: []interface{}{1, 2, 3},
			expectedParam: &parameter{
				parameterType: ArrayType,
				arrayValue: []Parameter{
					&parameter{parameterType: IntType, intValue: 1},
					&parameter{parameterType: IntType, intValue: 2},
					&parameter{parameterType: IntType, intValue: 3},
				},
			},
			expectedType:   ArrayType,
			expectingError: false,
		},
		{
			name:   "Map Type",
			config: map[string]interface{}{"key": "value"},
			expectedParam: &parameter{
				parameterType: MapType,
				mapValue: map[string]Parameter{
					"key": &parameter{parameterType: StringType, stringValue: "value"},
				},
			},
			expectedType:   MapType,
			expectingError: false,
		},
		{
			name:   "Parameters Type",
			config: types.Parameters{"key": "value"},
			expectedParam: &parameter{
				parameterType: MapType,
				mapValue: map[string]Parameter{
					"key": &parameter{parameterType: StringType, stringValue: "value"},
				},
			},
			expectedType:   MapType,
			expectingError: false,
		},
		{
			name:           "Unsupported Type",
			config:         struct{}{},
			expectedParam:  nil,
			expectedType:   NilType,
			expectingError: true,
		},
		{
			name:           "Unsupported Array Type",
			config:         []interface{}{struct{}{}},
			expectedParam:  nil,
			expectedType:   NilType,
			expectingError: true,
		},
		{
			name:           "Unsupported Map Type",
			config:         map[string]interface{}{"key": struct{}{}},
			expectedParam:  nil,
			expectedType:   NilType,
			expectingError: true,
		},
		{
			name:           "Unsupported Parameters Type",
			config:         types.Parameters{"key": struct{}{}},
			expectedParam:  nil,
			expectedType:   NilType,
			expectingError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			param, err := maker.Make(test.config)
			if test.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedType, param.Type())

				// Cast the result to the expected type and compare
				actualParam, ok := param.(*parameter)
				if assert.True(t, ok) {
					assert.Equal(t, test.expectedParam, actualParam)
				}
			}
		})
	}
}

func Test_parameter_BoolValue(t *testing.T) {
	tests := []struct {
		name      string
		parameter *parameter
		expected  bool
	}{
		{
			name: "Bool True",
			parameter: &parameter{
				parameterType: BoolType,
				boolValue:     true,
			},
			expected: true,
		},
		{
			name: "Bool False",
			parameter: &parameter{
				parameterType: BoolType,
				boolValue:     false,
			},
			expected: false,
		},
		{
			name: "Int Non-zero",
			parameter: &parameter{
				parameterType: IntType,
				intValue:      1,
			},
			expected: true,
		},
		{
			name: "Int Zero",
			parameter: &parameter{
				parameterType: IntType,
				intValue:      0,
			},
			expected: false,
		},
		{
			name: "Float Non-zero",
			parameter: &parameter{
				parameterType: FloatType,
				floatValue:    1.5,
			},
			expected: true,
		},
		{
			name: "Float Zero",
			parameter: &parameter{
				parameterType: FloatType,
				floatValue:    0.0,
			},
			expected: false,
		},
		{
			name: "Non-empty String",
			parameter: &parameter{
				parameterType: StringType,
				stringValue:   "hello",
			},
			expected: true,
		},
		{
			name: "Empty String",
			parameter: &parameter{
				parameterType: StringType,
				stringValue:   "",
			},
			expected: false,
		},
		{
			name: "Non-empty Array",
			parameter: &parameter{
				parameterType: ArrayType,
				arrayValue:    []Parameter{&parameter{parameterType: IntType, intValue: 1}},
			},
			expected: true,
		},
		{
			name: "Empty Array",
			parameter: &parameter{
				parameterType: ArrayType,
				arrayValue:    []Parameter{},
			},
			expected: false,
		},
		{
			name: "Non-empty Map",
			parameter: &parameter{
				parameterType: MapType,
				mapValue:      map[string]Parameter{"key": &parameter{parameterType: StringType, stringValue: "value"}},
			},
			expected: true,
		},
		{
			name: "Empty Map",
			parameter: &parameter{
				parameterType: MapType,
				mapValue:      map[string]Parameter{},
			},
			expected: false,
		},
		{
			name: "Nil Type",
			parameter: &parameter{
				parameterType: NilType,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parameter.BoolValue()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_parameter_IntValue(t *testing.T) {
	tests := []struct {
		name      string
		parameter *parameter
		expected  int
	}{
		{
			name: "Int Type",
			parameter: &parameter{
				parameterType: IntType,
				intValue:      42,
			},
			expected: 42,
		},
		{
			name: "Bool True",
			parameter: &parameter{
				parameterType: BoolType,
				boolValue:     true,
			},
			expected: 1,
		},
		{
			name: "Bool False",
			parameter: &parameter{
				parameterType: BoolType,
				boolValue:     false,
			},
			expected: 0,
		},
		{
			name: "Float Type With Truncation",
			parameter: &parameter{
				parameterType: FloatType,
				floatValue:    42.99,
			},
			expected: 42,
		},
		{
			name: "Valid Integer String",
			parameter: &parameter{
				parameterType: StringType,
				stringValue:   "123",
			},
			expected: 123,
		},
		{
			name: "Invalid Integer String",
			parameter: &parameter{
				parameterType: StringType,
				stringValue:   "hello",
			},
			expected: 0,
		},
		{
			name: "Array Type Count",
			parameter: &parameter{
				parameterType: ArrayType,
				arrayValue:    []Parameter{&parameter{parameterType: IntType, intValue: 1}, &parameter{parameterType: IntType, intValue: 2}},
			},
			expected: 2,
		},
		{
			name: "Empty Array",
			parameter: &parameter{
				parameterType: ArrayType,
				arrayValue:    []Parameter{},
			},
			expected: 0,
		},
		{
			name: "Map Type Count",
			parameter: &parameter{
				parameterType: MapType,
				mapValue: map[string]Parameter{
					"one": &parameter{parameterType: IntType, intValue: 1},
					"two": &parameter{parameterType: IntType, intValue: 2},
				},
			},
			expected: 2,
		},
		{
			name: "Empty Map",
			parameter: &parameter{
				parameterType: MapType,
				mapValue:      map[string]Parameter{},
			},
			expected: 0,
		},
		{
			name: "Nil Type",
			parameter: &parameter{
				parameterType: NilType,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parameter.IntValue()
			assert.Equal(t, tt.expected, result, "The IntValue does not match the expected value.")
		})
	}
}

func Test_parameter_FloatValue(t *testing.T) {
	tests := []struct {
		name      string
		parameter *parameter
		expected  float64
	}{
		{
			name: "Float Type",
			parameter: &parameter{
				parameterType: FloatType,
				floatValue:    3.14159,
			},
			expected: 3.14159,
		},
		{
			name: "Int Type",
			parameter: &parameter{
				parameterType: IntType,
				intValue:      42,
			},
			expected: 42.0,
		},
		{
			name: "Bool True",
			parameter: &parameter{
				parameterType: BoolType,
				boolValue:     true,
			},
			expected: 1.0,
		},
		{
			name: "Bool False",
			parameter: &parameter{
				parameterType: BoolType,
				boolValue:     false,
			},
			expected: 0.0,
		},
		{
			name: "Array Type Count",
			parameter: &parameter{
				parameterType: ArrayType,
				arrayValue:    []Parameter{&parameter{}, &parameter{}},
			},
			expected: 2.0,
		},
		{
			name: "Map Type Count",
			parameter: &parameter{
				parameterType: MapType,
				mapValue: map[string]Parameter{
					"one": &parameter{},
					"two": &parameter{},
				},
			},
			expected: 2.0,
		},
		{
			name: "Nil Type",
			parameter: &parameter{
				parameterType: NilType,
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parameter.FloatValue()
			assert.Equal(t, tt.expected, result, "The FloatValue does not match the expected value.")
		})
	}
}

func Test_parameter_StringValue(t *testing.T) {
	tests := []struct {
		name      string
		parameter *parameter
		expected  string
	}{
		{
			name: "String Type",
			parameter: &parameter{
				parameterType: StringType,
				stringValue:   "hello world",
			},
			expected: "hello world",
		},
		{
			name: "Bool Type True",
			parameter: &parameter{
				parameterType: BoolType,
				boolValue:     true,
			},
			expected: "true",
		},
		{
			name: "Bool Type False",
			parameter: &parameter{
				parameterType: BoolType,
				boolValue:     false,
			},
			expected: "false",
		},
		{
			name: "Int Type",
			parameter: &parameter{
				parameterType: IntType,
				intValue:      123,
			},
			expected: "123",
		},
		{
			name: "Float Type",
			parameter: &parameter{
				parameterType: FloatType,
				floatValue:    45.67,
			},
			expected: "45.67",
		},
		{
			name: "Default Type",
			parameter: &parameter{
				parameterType: NilType, // Or any other type that isn't explicitly handled
			},
			expected: fmt.Sprintf("%v", &parameter{parameterType: NilType}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parameter.StringValue()
			assert.Equal(t, tt.expected, result, "The StringValue does not match the expected value.")
		})
	}
}

func Test_parameter_ArrayValue(t *testing.T) {
	testParam := &parameter{stringValue: "test"}

	tests := []struct {
		name      string
		parameter *parameter
		expected  []Parameter
	}{
		{
			name: "Array Type",
			parameter: &parameter{
				parameterType: ArrayType,
				arrayValue:    []Parameter{testParam},
			},
			expected: []Parameter{testParam},
		},
		{
			name: "Map Type",
			parameter: &parameter{
				parameterType: MapType,
				mapValue:      map[string]Parameter{"key": testParam},
			},
			expected: []Parameter{testParam},
		},
		{
			name: "Other Type",
			parameter: &parameter{
				parameterType: IntType,
				intValue:      42,
			},
			expected: []Parameter{&parameter{
				parameterType: IntType,
				intValue:      42,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parameter.ArrayValue()
			assert.Equal(t, tt.expected, result, "The ArrayValue does not match the expected value.")
		})
	}
}

func Test_parameter_MapValue(t *testing.T) {
	testParam := &parameter{stringValue: "test"}

	tests := []struct {
		name      string
		parameter *parameter
		expected  map[string]Parameter
	}{
		{
			name: "Map Type",
			parameter: &parameter{
				parameterType: MapType,
				mapValue:      map[string]Parameter{"key": testParam},
			},
			expected: map[string]Parameter{"key": testParam},
		},
		{
			name: "Array Type",
			parameter: &parameter{
				parameterType: ArrayType,
				arrayValue:    []Parameter{testParam, testParam},
			},
			expected: map[string]Parameter{"0": testParam, "1": testParam},
		},
		{
			name: "Other Type",
			parameter: &parameter{
				parameterType: StringType,
				stringValue:   "hello",
			},
			expected: map[string]Parameter{"0": &parameter{
				parameterType: StringType,
				stringValue:   "hello",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parameter.MapValue()
			assert.Equal(t, tt.expected, result, "The MapValue does not match the expected map.")
		})
	}
}

func Test_parameter_Type(t *testing.T) {
	param := &parameter{
		parameterType: FloatType,
		floatValue:    45.67,
	}
	assert.Equal(t, FloatType, param.Type())
}
