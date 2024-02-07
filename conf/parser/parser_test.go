package parser

import (
	"errors"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/loader"
	"github.com/bukka/wst/mocks/confMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func Test_isValidParam(t *testing.T) {
	tests := []struct {
		name  string
		param string
		want  bool
	}{
		{
			name:  "valid parameter - name",
			param: paramName,
			want:  true,
		},
		{
			name:  "valid parameter - loadable",
			param: paramLoadable,
			want:  true,
		},
		{
			name:  "valid parameter - default",
			param: paramDefault,
			want:  true,
		},
		{
			name:  "valid parameter - factory",
			param: paramFactory,
			want:  true,
		},
		{
			name:  "valid parameter - enum",
			param: paramEnum,
			want:  true,
		},
		{
			name:  "valid parameter - keys",
			param: paramKeys,
			want:  true,
		},
		{
			name:  "valid parameter - string",
			param: paramString,
			want:  true,
		},
		{
			name:  "invalid parameter - non-existent",
			param: "nonexistentparam",
			want:  false,
		},
		{
			name:  "invalid parameter - empty string",
			param: "",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isValidParam(tt.param), "isValidParam(%v)", tt.param)
		})
	}
}

func TestConfigParser_ParseTag(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   map[string]string
		errMsg string
	}{
		{
			name: "Testing ParseTag - With all valid params and implicit name",
			tag:  "tagname,default=value1,enum",
			want: map[string]string{
				"name":    "tagname",
				"default": "value1",
				"enum":    "true",
			},
			errMsg: "",
		},
		{
			name: "Testing ParseTag - With all valid params and explicit name",
			tag:  "name=tagname,default=value1,enum",
			want: map[string]string{
				"name":    "tagname",
				"default": "value1",
				"enum":    "true",
			},
			errMsg: "",
		},
		{
			name:   "Testing ParseTag - invalid parameter key",
			tag:    "invalid=key",
			want:   map[string]string{},
			errMsg: "invalid parameter key: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := ConfigParser{}
			got, err := parser.parseTag(tt.tag)

			// if an error is expected
			if tt.errMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg) // check that the error message is as expected
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_ConfigParser_processDefaultParam(t *testing.T) {
	type defaultTestStruct struct {
		A int    `wst:"name:a"`
		B bool   `wst:"name:b"`
		C string `wst:"name:c"`
	}

	p := &ConfigParser{}
	testCase := defaultTestStruct{}

	tests := []struct {
		name        string
		defaultVal  string
		fieldVal    reflect.Value
		expectedVal interface{}
		wantErr     bool
	}{
		{
			name:        "successfully process integer default value",
			defaultVal:  "5",
			fieldVal:    reflect.ValueOf(&testCase.A).Elem(),
			expectedVal: 5,
			wantErr:     false,
		},
		{
			name:        "successfully process boolean default value",
			defaultVal:  "true",
			fieldVal:    reflect.ValueOf(&testCase.B).Elem(),
			expectedVal: true,
			wantErr:     false,
		},
		{
			name:        "successfully process string default value",
			defaultVal:  "hello",
			fieldVal:    reflect.ValueOf(&testCase.C).Elem(),
			expectedVal: "hello",
			wantErr:     false,
		},
		{
			name:        "error on invalid integer default value",
			defaultVal:  "not_integer",
			fieldVal:    reflect.ValueOf(&testCase.A).Elem(),
			expectedVal: 0,
			wantErr:     true,
		},
		{
			name:        "error on invalid boolean default value",
			defaultVal:  "not_boolean",
			fieldVal:    reflect.ValueOf(&testCase.B).Elem(),
			expectedVal: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.processDefaultParam(tt.name, tt.defaultVal, tt.fieldVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("processDefaultParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(tt.fieldVal.Interface(), tt.expectedVal) {
				t.Errorf("processDefaultParam() value = %v, want %v", tt.fieldVal.Interface(), tt.expectedVal)
			}
		})
	}
}

type MockFactoryFunc func(data interface{}, fieldValue reflect.Value) error

func TestConfigParser_processFactoryParam(t *testing.T) {
	// Mock Factory Function
	mockFactoryFunc := func(data interface{}, fieldValue reflect.Value) error {
		return nil
	}

	// Error Mock Factory Function
	errorMockFactoryFunc := func(data interface{}, fieldValue reflect.Value) error {
		return fmt.Errorf("forced factory function error")
	}

	p := ConfigParser{
		env: nil, // replace with necessary mock if necessary
		factories: map[string]factoryFunc{
			"mockFactory":      mockFactoryFunc,
			"errorMockFactory": errorMockFactoryFunc,
		},
	}

	fieldValue := reflect.ValueOf("fieldValue")

	tests := []struct {
		name       string
		factory    string
		data       interface{}
		fieldValue reflect.Value
		wantErr    bool
	}{
		{
			name:       "Valid factory",
			factory:    "mockFactory",
			data:       "mockData",
			fieldValue: fieldValue,
			wantErr:    false,
		},
		{
			name:       "Invalid factory",
			factory:    "invalidFactory",
			data:       "mockData",
			fieldValue: fieldValue,
			wantErr:    true,
		},
		{
			name:       "Forced factory function error",
			factory:    "errorMockFactory",
			data:       "mockData",
			fieldValue: fieldValue,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.processFactoryParam(tt.factory, tt.data, tt.fieldValue); (err != nil) != tt.wantErr {
				t.Errorf("processFactoryParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ConfigParser_processEnumParam(t *testing.T) {
	p := ConfigParser{}

	tests := []struct {
		name      string
		enums     string
		data      interface{}
		fieldName string
		wantErr   bool
	}{
		{
			name:      "Value found in enum list",
			enums:     "enum1|enum2|enum3",
			data:      "enum2",
			fieldName: "field",
			wantErr:   false, // No error because data is in enum list
		},
		{
			name:      "Value not found - should trigger error",
			enums:     "enum1|enum2|enum3",
			data:      "enum4",
			fieldName: "field",
			wantErr:   true, // Error because data is not in enum list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.processEnumParam(tt.enums, tt.data, tt.fieldName); (err != nil) != tt.wantErr {
				t.Errorf("processEnumParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ConfigParser_processKeysParam(t *testing.T) {
	p := ConfigParser{}

	tests := []struct {
		name      string
		keys      string
		data      interface{}
		fieldName string
		wantErr   bool
	}{
		{
			name:      "Key found in data",
			keys:      "key1|key2|key3",
			data:      map[string]interface{}{"key2": "value"},
			fieldName: "field",
			wantErr:   false, // No error because key is in data
		},
		{
			name:      "Key not found - should trigger error",
			keys:      "key1|key2|key3",
			data:      map[string]interface{}{"key4": "value"},
			fieldName: "field",
			wantErr:   true, // Error because key is not in data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.processKeysParam(tt.keys, tt.data, tt.fieldName); (err != nil) != tt.wantErr {
				t.Errorf("processKeysParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func Test_ConfigParser_processLoadableParam(t *testing.T) {
	mockLoadedConfig := &confMocks.MockLoadedConfig{}
	mockLoadedConfig.On("Path").Return("/configs/test.json")
	mockLoadedConfig.On("Data").Return(map[string]interface{}{"key": "value"})

	tests := []struct {
		name       string
		data       interface{}
		fieldValue reflect.Value
		want       interface{}
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Field value kind is map",
			data:       "*.json",
			fieldValue: reflect.ValueOf(map[string]map[string]interface{}{}),
			want:       map[string]map[string]interface{}{"/configs/test.json": {"key": "value"}},
			wantErr:    false,
		},
		{
			name:       "Field value kind is slice",
			data:       "*.json",
			fieldValue: reflect.ValueOf([]map[string]interface{}{}),
			want:       []map[string]interface{}{{"key": "value"}},
			wantErr:    false,
		},
		{
			name:       "Field value kind is string (unsupported kind) - should trigger error",
			data:       "*.json",
			fieldValue: reflect.ValueOf("string"), // unsupported kind
			want:       nil,
			wantErr:    true,
			errMsg:     "type of field is neither map nor slice (kind=string)",
		},
		{
			name:       "Error from GlobConfigs",
			data:       "*.json",
			fieldValue: reflect.ValueOf(map[string]map[string]interface{}{}),
			want:       nil,
			wantErr:    true,
			errMsg:     "loading configs: forced GlobConfigs error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoader := &confMocks.MockLoader{}
			if tt.name != "Error from GlobConfigs" {
				mockLoader.On("GlobConfigs", tt.data.(string)).Return([]loader.LoadedConfig{mockLoadedConfig}, nil)
			} else {
				mockLoader.On("GlobConfigs", tt.data.(string)).Return(nil, errors.New("forced GlobConfigs error"))
			}

			p := ConfigParser{
				env:    nil, // replace with necessary mock if necessary
				loader: mockLoader,
			}
			got, err := p.processLoadableParam(tt.data, tt.fieldValue)

			// if an error is expected
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg) // check that the error message is as expected
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

type StringParamTestStruct struct {
	StringField string
}

type StringParamParentStruct struct {
	Child StringParamTestStruct
}

type StringParamTestMapStruct struct {
	MapField map[string]StringParamTestStruct
}

func Test_ConfigParser_processStringParam(t *testing.T) {
	// Prepare ConfigParser
	p := ConfigParser{env: nil} // you may need to initialize this with suitable fields based on your implementation

	// Testing data setup
	dataVal := "stringValue"

	var structVal StringParamParentStruct

	// Testing data setup for map
	mapDataVal := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	var mapStructVal StringParamTestMapStruct

	tests := []struct {
		name       string
		fieldName  string
		data       interface{}
		fieldValue reflect.Value
		wantErr    bool
		want       interface{}
	}{
		{
			name:       "process string param in struct field",
			fieldName:  "StringField",
			data:       dataVal,
			fieldValue: reflect.ValueOf(&structVal.Child),
			wantErr:    false,
			want:       StringParamTestStruct{StringField: dataVal},
		},
		{
			name:       "process map param",
			fieldName:  "StringField",
			data:       mapDataVal,
			fieldValue: reflect.ValueOf(&mapStructVal.MapField),
			wantErr:    false,
			want: map[string]StringParamTestStruct{
				"key1": StringParamTestStruct{StringField: "value1"},
				"key2": StringParamTestStruct{StringField: "value2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.processStringParam(tt.fieldName, tt.data, tt.fieldValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("processStringParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got := tt.fieldValue.Elem().Interface(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processStringParam() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type AssignFieldTestStruct struct {
	A string
	B int
	C []string
	D map[string]int
}

type AssignFieldAnotherStruct struct {
	E string `wst:"e"`
	F bool   `wst:"f"`
}

type AssignFieldParentStruct struct {
	Child AssignFieldAnotherStruct
}

func Test_ConfigParser_assignField(t *testing.T) {
	p := ConfigParser{env: nil} // Initialize appropriately

	tests := []struct {
		name          string
		fieldName     string
		data          interface{}
		value         interface{}
		wantErr       bool
		expectedValue interface{}
	}{
		{
			name:      "assign struct string field",
			fieldName: "A",
			data:      "TestA",
			value:     &AssignFieldTestStruct{},
			wantErr:   false,
			expectedValue: &AssignFieldTestStruct{
				A: "TestA",
			},
		},
		{
			name:      "assign struct int field",
			fieldName: "B",
			data:      12,
			value:     &AssignFieldTestStruct{},
			wantErr:   false,
			expectedValue: &AssignFieldTestStruct{
				B: 12,
			},
		},
		{
			name:      "assign array field",
			fieldName: "C",
			data:      []interface{}{"TestA", "TestB"},
			value:     &AssignFieldTestStruct{},
			wantErr:   false,
			expectedValue: &AssignFieldTestStruct{
				C: []string{"TestA", "TestB"},
			},
		},
		{
			name:      "assign tuple in map field",
			fieldName: "D",
			data:      map[string]interface{}{"test": 7},
			value:     &AssignFieldTestStruct{},
			wantErr:   false,
			expectedValue: &AssignFieldTestStruct{
				D: map[string]int{"test": 7},
			},
		},
		{
			name:          "assign struct field with mismatched type should error",
			fieldName:     "A",
			data:          5,
			value:         &AssignFieldTestStruct{},
			wantErr:       true,
			expectedValue: &AssignFieldTestStruct{},
		},
		{
			name:      "assign to nested struct field",
			fieldName: "Child",
			data:      map[string]interface{}{"e": "NestedTest", "f": true},
			value:     &AssignFieldParentStruct{},
			wantErr:   false,
			expectedValue: &AssignFieldParentStruct{
				Child: AssignFieldAnotherStruct{
					E: "NestedTest",
					F: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldValue := reflect.ValueOf(tt.value).Elem().FieldByName(tt.fieldName)
			err := p.assignField(tt.data, fieldValue, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("assignField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Add a check to make sure value got updated to what was expected.
			if !reflect.DeepEqual(tt.value, tt.expectedValue) {
				t.Errorf("expected updated struct value to be %v, got %v", tt.expectedValue, tt.value)
			}
		})
	}
}

type ParseFieldInnerTestStruct struct {
	Value string `wst:"val"`
}

type ParseFieldTestStruct struct {
	A string                               `wst:"a"`
	B int                                  `wst:"b"`
	C ParseFieldInnerTestStruct            `wst:"c"`
	D []ParseFieldInnerTestStruct          `wst:"d"`
	E map[string]ParseFieldInnerTestStruct `wst:"e"`
	F map[string]int                       `wst:"f"`
}

type ParseFieldConfigData struct {
	Path string
	Data map[string]interface{}
}

func Test_ConfigParser_parseField(t *testing.T) {
	// Insert your test cases here, I'm providing one sample case
	tests := []struct {
		name               string
		fieldName          string
		data               interface{}
		params             map[string]string
		expectedFieldValue interface{}
		configsCalled      bool
		configsData        []ParseFieldConfigData
		factories          map[string]factoryFunc
		wantErr            bool
	}{
		{
			name:      "parse field with factory param found",
			fieldName: "A",
			data:      map[string]interface{}{"a": "NestedTest", "b": 1},
			params: map[string]string{
				"factory": "test",
			},
			configsCalled: false,
			configsData:   nil,
			factories: map[string]factoryFunc{
				"test": func(_ interface{}, fieldValue reflect.Value) error {
					// Mimic factory behavior here
					fieldValue.SetString("test_data")
					return nil
				},
			},
			expectedFieldValue: &ParseFieldTestStruct{A: "test_data"},
			wantErr:            false,
		},
		{
			name:      "parse field with factory param not found",
			fieldName: "A",
			data:      map[string]interface{}{"a": "NestedTest", "b": 1},
			params: map[string]string{
				"factory": "invalid",
			},
			configsCalled: false,
			configsData:   nil,
			factories: map[string]factoryFunc{
				"test": func(_ interface{}, fieldValue reflect.Value) error {
					// Mimic factory behavior here
					fieldValue.SetString("test_data")
					return nil
				},
			},
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
		{
			name:      "parse field with string param and value that is string",
			fieldName: "C",
			data:      "data",
			params: map[string]string{
				"string": "Value",
			},
			configsCalled:      false,
			configsData:        nil,
			factories:          nil,
			expectedFieldValue: &ParseFieldTestStruct{C: ParseFieldInnerTestStruct{Value: "data"}},
			wantErr:            false,
		},
		{
			name:      "parse field with string param and value that is not string",
			fieldName: "C",
			data: map[string]interface{}{
				"val": "data2",
			},
			params: map[string]string{
				"string": "Value",
			},
			configsCalled:      false,
			configsData:        nil,
			factories:          nil,
			expectedFieldValue: &ParseFieldTestStruct{C: ParseFieldInnerTestStruct{Value: "data2"}},
			wantErr:            false,
		},
		{
			name:      "parse field with string param that points to invalid filed",
			fieldName: "C",
			data:      "data",
			params: map[string]string{
				"string": "NotFound",
			},
			configsCalled:      true,
			configsData:        nil,
			factories:          nil,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
		{
			name:      "parse array field with loadable param and single path string data",
			fieldName: "D",
			data:      "services/test.yaml",
			params: map[string]string{
				"loadable": "true",
			},
			configsCalled: true,
			configsData: []ParseFieldConfigData{
				{Path: "services/test.yaml", Data: map[string]interface{}{"val": "test"}},
			},
			factories:          nil,
			expectedFieldValue: &ParseFieldTestStruct{D: []ParseFieldInnerTestStruct{{Value: "test"}}},
			wantErr:            false,
		},
		{
			name:      "parse array field with loadable param and multiple paths string data",
			fieldName: "D",
			data:      "services/test*.yaml",
			params: map[string]string{
				"loadable": "true",
			},
			configsCalled: true,
			configsData: []ParseFieldConfigData{
				{Path: "services/test1.yaml", Data: map[string]interface{}{"val": "test1"}},
				{Path: "services/test2.yaml", Data: map[string]interface{}{"val": "test2"}},
			},
			factories: nil,
			expectedFieldValue: &ParseFieldTestStruct{D: []ParseFieldInnerTestStruct{
				{Value: "test1"},
				{Value: "test2"},
			}},
			wantErr: false,
		},
		{
			name:      "parse map field with loadable param and multiple paths string data",
			fieldName: "E",
			data:      "services/test*.yaml",
			params: map[string]string{
				"loadable": "true",
			},
			configsCalled: true,
			configsData: []ParseFieldConfigData{
				{Path: "services/test1.yaml", Data: map[string]interface{}{"val": "test1"}},
				{Path: "services/test2.yaml", Data: map[string]interface{}{"val": "test2"}},
			},
			factories: nil,
			expectedFieldValue: &ParseFieldTestStruct{E: map[string]ParseFieldInnerTestStruct{
				"services/test1.yaml": {Value: "test1"},
				"services/test2.yaml": {Value: "test2"},
			}},
			wantErr: false,
		},
		{
			name:      "parse map field with loadable param and failed loading",
			fieldName: "E",
			data:      "services/test*.yaml",
			params: map[string]string{
				"loadable": "true",
			},
			configsCalled:      true,
			configsData:        nil,
			factories:          nil,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
		{
			name:      "parse map field with enum when found",
			fieldName: "A",
			data:      "value2",
			params: map[string]string{
				"enum": "value1|value2|value3",
			},
			configsCalled:      false,
			configsData:        nil,
			factories:          nil,
			expectedFieldValue: &ParseFieldTestStruct{A: "value2"},
			wantErr:            false,
		},
		{
			name:      "parse map field with enum when not found",
			fieldName: "A",
			data:      "value5",
			params: map[string]string{
				"enum": "value1|value2|value3",
			},
			configsCalled:      false,
			configsData:        nil,
			factories:          nil,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
		{
			name:      "parse map field with keys when found",
			fieldName: "F",
			data:      map[string]interface{}{"key0": 1, "key1": 2},
			params: map[string]string{
				"keys": "key1|key2",
			},
			configsCalled:      false,
			configsData:        nil,
			factories:          nil,
			expectedFieldValue: &ParseFieldTestStruct{F: map[string]int{"key0": 1, "key1": 2}},
			wantErr:            false,
		},
		{
			name:      "parse map field with keys when not found",
			fieldName: "F",
			data:      map[string]interface{}{"key0": 1},
			params: map[string]string{
				"keys": "key1|key2",
			},
			configsCalled:      false,
			configsData:        nil,
			factories:          nil,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commonFieldValue := &ParseFieldTestStruct{}
			fieldValue := reflect.ValueOf(commonFieldValue).Elem().FieldByName(tt.fieldName)

			mockLoader := &confMocks.MockLoader{}
			if tt.configsCalled {
				if tt.configsData != nil {
					var mockConfigs []loader.LoadedConfig
					for _, configData := range tt.configsData {
						mockLoadedConfig := &confMocks.MockLoadedConfig{}
						mockLoadedConfig.On("Path").Return(configData.Path)
						mockLoadedConfig.On("Data").Return(configData.Data)
						mockConfigs = append(mockConfigs, mockLoadedConfig)
					}
					mockLoader.On("GlobConfigs", tt.data.(string)).Return(mockConfigs, nil)
				} else {
					mockLoader.On("GlobConfigs", tt.data.(string)).Return(nil, errors.New("forced GlobConfigs error"))
				}
			}

			// Create a new ConfigParser for each test case
			p := &ConfigParser{
				env:       nil,
				loader:    mockLoader,
				factories: tt.factories,
			}

			err := p.parseField(tt.data, fieldValue, tt.fieldName, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && !reflect.DeepEqual(commonFieldValue, tt.expectedFieldValue) {
				t.Errorf("unexpected value: got %v, want %v", commonFieldValue, tt.expectedFieldValue)
			}
		})
	}
}

func Test_ConfigParser_parseStruct(t *testing.T) {
	type parseValidTestStruct struct {
		A           int    `wst:"name=a,default=5"`
		B           string `wst:"name=b,default=default_str"`
		C           int    `wst:""`
		D           bool   `wst:"d"`
		F           string
		UnexportedE bool `wst:"name=e"`
	}
	type parseInvalidTagTestStruct struct {
		A int `wst:"a,unknown=1"`
	}
	type parseInvalidFactoryTestStruct struct {
		A parseValidTestStruct `wst:"a,factory=incorrect"`
	}
	type parseInvalidDefaultTestStruct struct {
		A parseValidTestStruct `wst:"a,default=data"`
	}

	mockLoader := &confMocks.MockLoader{}
	p := &ConfigParser{
		env:       nil,
		loader:    mockLoader,
		factories: map[string]factoryFunc{},
	}

	tests := []struct {
		name           string
		data           map[string]interface{}
		testStruct     interface{}
		expectedStruct *parseValidTestStruct
		errMsg         string
	}{
		{
			name:           "Test valid default data",
			data:           map[string]interface{}{},
			testStruct:     &parseValidTestStruct{},
			expectedStruct: &parseValidTestStruct{A: 5, B: "default_str", C: 0, D: false},
			errMsg:         "",
		},
		{
			name:           "Test invalid data",
			data:           map[string]interface{}{},
			testStruct:     "data",
			expectedStruct: nil,
			errMsg:         "expected a pointer to a struct, got string",
		},
		{
			name:           "Test invalid tag",
			data:           map[string]interface{}{},
			testStruct:     &parseInvalidTagTestStruct{},
			expectedStruct: nil,
			errMsg:         "invalid parameter key: unknown",
		},
		{
			name:           "Test failing field due to invalid factory",
			data:           map[string]interface{}{"a": "data"},
			testStruct:     &parseInvalidFactoryTestStruct{},
			expectedStruct: nil,
			errMsg:         "factory function incorrect not found",
		},
		{
			name:           "Test invalid default",
			data:           map[string]interface{}{},
			testStruct:     &parseInvalidDefaultTestStruct{},
			expectedStruct: nil,
			errMsg:         "default value data for field a cannot be converted to type parser.parseValidTestStruct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.parseStruct(tt.data, tt.testStruct)

			if tt.errMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if !reflect.DeepEqual(tt.testStruct, tt.expectedStruct) {
					t.Errorf("unexpected structure content: got %v, want %v", tt.testStruct, tt.expectedStruct)
				}
			}
		})
	}
}

func TestCreateParser(t *testing.T) {
	type args struct {
		env    app.Env
		loader loader.Loader
	}
	tests := []struct {
		name string
		args args
		want Parser
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CreateParser(tt.args.env, tt.args.loader), "CreateParser(%v, %v)", tt.args.env, tt.args.loader)
		})
	}
}
