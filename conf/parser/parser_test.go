// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"errors"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/loader"
	"github.com/bukka/wst/conf/parser/factory"
	appMocks "github.com/bukka/wst/mocks/app"
	loaderMocks "github.com/bukka/wst/mocks/conf/loader"
	factoryMocks "github.com/bukka/wst/mocks/conf/parser/factory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func Test_isValidParam(t *testing.T) {
	tests := []struct {
		name  string
		param ConfigParam
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
			name:  "valid parameter - path",
			param: paramPath,
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
			assert.Equalf(t, tt.want, isValidParam(string(tt.param)), "isValidParam(%v)", tt.param)
		})
	}
}

func TestConfigParser_ParseTag(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   map[ConfigParam]string
		errMsg string
	}{
		{
			name: "Testing ParseTag - With all valid params and implicit name",
			tag:  "tagname,default=value1,enum",
			want: map[ConfigParam]string{
				"name":    "tagname",
				"default": "value1",
				"enum":    "true",
			},
			errMsg: "",
		},
		{
			name: "Testing ParseTag - With all valid params and explicit name",
			tag:  "name=tagname,default=value1,enum",
			want: map[ConfigParam]string{
				"name":    "tagname",
				"default": "value1",
				"enum":    "true",
			},
			errMsg: "",
		},
		{
			name:   "Testing ParseTag - invalid parameter key",
			tag:    "invalid=key",
			want:   map[ConfigParam]string{},
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

type MockFactoryFunc func(data interface{}, fieldValue reflect.Value, path string) error

func TestConfigParser_processFactoryParam(t *testing.T) {
	factories := &factoryMocks.MockFunctions{}

	// Set up mock expectations
	factories.On("GetFactoryFunc", "mockFactory").Return(
		factory.Func(func(data interface{}, fieldValue reflect.Value, path string) error {
			return nil
		}))
	factories.On("GetFactoryFunc", "invalidFactory").Return(nil)
	factories.On("GetFactoryFunc", "errorMockFactory").Return(
		factory.Func(func(data interface{}, fieldValue reflect.Value, path string) error {
			return fmt.Errorf("forced error")
		}))

	p := ConfigParser{
		fnd:       nil, // replace with necessary mock if necessary
		factories: factories,
	}

	path := "/var/www/ws"
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
			if err := p.processFactoryParam(tt.factory, tt.data, tt.fieldValue, path); (err != nil) != tt.wantErr {
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
	mockLoadedConfig := &loaderMocks.MockLoadedConfig{}
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
			fieldValue: reflect.ValueOf(map[string]interface{}{}),
			want: map[string]interface{}{"/configs/test.json": map[string]interface{}{
				"key":      "value",
				"wst/path": "/configs/test.json",
			}},
			wantErr: false,
		},
		{
			name:       "Field value kind is slice",
			data:       "*.json",
			fieldValue: reflect.ValueOf([]map[string]interface{}{}),
			want: []map[string]interface{}{{
				"key":      "value",
				"wst/path": "/configs/test.json",
			}},
			wantErr: false,
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
			fieldValue: reflect.ValueOf(map[string]interface{}{}),
			want:       nil,
			wantErr:    true,
			errMsg:     "loading configs: forced GlobConfigs error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoader := &loaderMocks.MockLoader{}
			if tt.name != "Error from GlobConfigs" {
				mockLoader.On("GlobConfigs", tt.data.(string)).Return([]loader.LoadedConfig{mockLoadedConfig}, nil)
			} else {
				mockLoader.On("GlobConfigs", tt.data.(string)).Return(nil, errors.New("forced GlobConfigs error"))
			}

			p := ConfigParser{
				fnd:    nil, // replace with necessary mock if necessary
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

func TestProcessPathParam(t *testing.T) {
	mockFs := afero.NewMemMapFs()

	// mock app.Foundation
	mockFnd := &appMocks.MockFoundation{}
	mockFnd.On("Fs").Return(mockFs)

	fieldName := "file"

	tests := []struct {
		name          string
		fieldValue    reflect.Value
		data          interface{}
		configPath    string
		wantErr       bool
		expectedErr   string
		expectedValue string
	}{
		{
			name:          "Valid relative path",
			fieldValue:    reflect.Indirect(reflect.ValueOf(new(string))),
			data:          "test/path",
			configPath:    "/opt/config/wst.yaml",
			wantErr:       false,
			expectedValue: "/opt/config/test/path",
		},
		{
			name:          "Valid absolute path",
			fieldValue:    reflect.Indirect(reflect.ValueOf(new(string))),
			data:          "/test/path",
			configPath:    "/opt/config/wst.yaml",
			wantErr:       false,
			expectedValue: "/test/path",
		},
		{
			name:        "Non-existent path",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(string))),
			data:        "test/non-existent",
			configPath:  "config/",
			wantErr:     true,
			expectedErr: "file path config/test/non-existent does not exist",
		},
		{
			name:        "Data is not a string",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(string))),
			data:        123,
			configPath:  "config/",
			wantErr:     true,
			expectedErr: "unexpected type int for data, expected string",
		},
		{
			name:          "Empty path",
			fieldValue:    reflect.Indirect(reflect.ValueOf(new(string))),
			data:          "",
			configPath:    "/opt/config/wst.yam",
			wantErr:       false,
			expectedValue: "/opt/config",
		},
		{
			name:        "Null data",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(string))),
			data:        nil,
			configPath:  "config/",
			wantErr:     true,
			expectedErr: "unexpected type <nil> for data, expected string",
		},
		{
			name:        "Invalid configPath",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(string))),
			data:        "test/path",
			configPath:  "invalid/config/path",
			wantErr:     true,
			expectedErr: "file path invalid/config/test/path does not exist",
		},
		{
			name:        "Field value cannot be set",
			fieldValue:  reflect.ValueOf(new(string)),
			data:        "test/path",
			configPath:  "/opt/config/wst.yaml",
			wantErr:     true,
			expectedErr: "field file is not settable",
		},
		{
			name:        "Invalid field value type",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(int))),
			data:        "test/path",
			configPath:  "/opt/config/wst.yaml",
			wantErr:     true,
			expectedErr: "field file is not of type string",
		},
	}

	// Create test files
	err := afero.WriteFile(mockFs, "/opt/config/test/path", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/test/path", []byte(`{"key": "value2"}`), 0644)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := ConfigParser{fnd: mockFnd}
			err := parser.processPathParam(tt.data, tt.fieldValue, fieldName, tt.configPath)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.expectedErr, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedValue, tt.fieldValue.String())
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
	p := ConfigParser{fnd: nil} // you may need to initialize this with suitable fields based on your implementation

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
			_, err := p.processStringParam(tt.fieldName, tt.data, tt.fieldValue, "/var/www/config.yaml")
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
	p := ConfigParser{fnd: nil} // Initialize appropriately

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
			err := p.assignField(tt.data, fieldValue, tt.fieldName, "/var/www/config.yaml")
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
	tests := []struct {
		name               string
		fieldName          string
		data               interface{}
		params             map[ConfigParam]string
		expectedFieldValue interface{}
		configsCalled      bool
		configsData        []ParseFieldConfigData
		factoryFound       bool
		wantErr            bool
	}{
		{
			name:      "parse field with factory param found",
			fieldName: "A",
			data:      map[string]interface{}{"a": "NestedTest", "b": 1},
			params: map[ConfigParam]string{
				"factory": "test",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       true,
			expectedFieldValue: &ParseFieldTestStruct{A: "test_data"},
			wantErr:            false,
		},
		{
			name:      "parse field with factory param not found",
			fieldName: "A",
			data:      map[string]interface{}{"a": "NestedTest", "b": 1},
			params: map[ConfigParam]string{
				"factory": "test",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
		{
			name:      "parse field with string param and value that is string",
			fieldName: "C",
			data:      "data",
			params: map[ConfigParam]string{
				"string": "Value",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{C: ParseFieldInnerTestStruct{Value: "data"}},
			wantErr:            false,
		},
		{
			name:      "parse field with string param and value that is not string",
			fieldName: "C",
			data: map[string]interface{}{
				"val": "data2",
			},
			params: map[ConfigParam]string{
				"string": "Value",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{C: ParseFieldInnerTestStruct{Value: "data2"}},
			wantErr:            false,
		},
		{
			name:      "parse field with string param that points to invalid filed",
			fieldName: "C",
			data:      "data",
			params: map[ConfigParam]string{
				"string": "NotFound",
			},
			configsCalled:      true,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
		{
			name:      "parse array field with loadable param and single path string data",
			fieldName: "D",
			data:      "services/test.yaml",
			params: map[ConfigParam]string{
				"loadable": "true",
			},
			configsCalled: true,
			configsData: []ParseFieldConfigData{
				{Path: "services/test.yaml", Data: map[string]interface{}{"val": "test"}},
			},
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{D: []ParseFieldInnerTestStruct{{Value: "test"}}},
			wantErr:            false,
		},
		{
			name:      "parse array field with loadable param and multiple paths string data",
			fieldName: "D",
			data:      "services/test*.yaml",
			params: map[ConfigParam]string{
				"loadable": "true",
			},
			configsCalled: true,
			configsData: []ParseFieldConfigData{
				{Path: "services/test1.yaml", Data: map[string]interface{}{"val": "test1"}},
				{Path: "services/test2.yaml", Data: map[string]interface{}{"val": "test2"}},
			},
			factoryFound: false,
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
			params: map[ConfigParam]string{
				"loadable": "true",
			},
			configsCalled: true,
			configsData: []ParseFieldConfigData{
				{Path: "services/test1.yaml", Data: map[string]interface{}{"val": "test1"}},
				{Path: "services/test2.yaml", Data: map[string]interface{}{"val": "test2"}},
			},
			factoryFound: false,
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
			params: map[ConfigParam]string{
				"loadable": "true",
			},
			configsCalled:      true,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
		{
			name:      "parse map field with enum when found",
			fieldName: "A",
			data:      "value2",
			params: map[ConfigParam]string{
				"enum": "value1|value2|value3",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{A: "value2"},
			wantErr:            false,
		},
		{
			name:      "parse map field with enum when not found",
			fieldName: "A",
			data:      "value5",
			params: map[ConfigParam]string{
				"enum": "value1|value2|value3",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
		{
			name:      "parse map field with keys when found",
			fieldName: "F",
			data:      map[string]interface{}{"key0": 1, "key1": 2},
			params: map[ConfigParam]string{
				"keys": "key1|key2",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{F: map[string]int{"key0": 1, "key1": 2}},
			wantErr:            false,
		},
		{
			name:      "parse map field with keys when not found",
			fieldName: "F",
			data:      map[string]interface{}{"key0": 1},
			params: map[ConfigParam]string{
				"keys": "key1|key2",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
		{
			name:      "parse map field with keys when not found",
			fieldName: "A",
			data:      "test/path",
			params: map[ConfigParam]string{
				"path": "true",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{A: "/opt/config/test/path"},
			wantErr:            false,
		},
		{
			name:      "parse map field with keys when not found",
			fieldName: "A",
			data:      "test/invalid",
			params: map[ConfigParam]string{
				"path": "true",
			},
			configsCalled:      false,
			configsData:        nil,
			factoryFound:       false,
			expectedFieldValue: &ParseFieldTestStruct{},
			wantErr:            true,
		},
	}

	mockFs := afero.NewMemMapFs()
	// mock app.Foundation
	mockFnd := &appMocks.MockFoundation{}
	mockFnd.On("Fs").Return(mockFs)

	err := afero.WriteFile(mockFs, "/opt/config/test/path", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commonFieldValue := &ParseFieldTestStruct{}
			fieldValue := reflect.ValueOf(commonFieldValue).Elem().FieldByName(tt.fieldName)

			mockLoader := &loaderMocks.MockLoader{}
			if tt.configsCalled {
				if tt.configsData != nil {
					var mockConfigs []loader.LoadedConfig
					for _, configData := range tt.configsData {
						mockLoadedConfig := &loaderMocks.MockLoadedConfig{}
						mockLoadedConfig.On("Path").Return(configData.Path)
						mockLoadedConfig.On("Data").Return(configData.Data)
						mockConfigs = append(mockConfigs, mockLoadedConfig)
					}
					mockLoader.On("GlobConfigs", tt.data.(string)).Return(mockConfigs, nil)
				} else {
					mockLoader.On("GlobConfigs", tt.data.(string)).Return(
						nil, errors.New("forced GlobConfigs error"))
				}
			}

			mockFactories := &factoryMocks.MockFunctions{}
			if tt.factoryFound {
				mockFactories.On("GetFactoryFunc", "test").Return(
					factory.Func(func(data interface{}, fieldValue reflect.Value, path string) error {
						fieldValue.SetString("test_data")
						return nil
					}))
			} else {
				mockFactories.On("GetFactoryFunc", "test").Return(nil)
			}

			// Create a new ConfigParser for each test case
			p := &ConfigParser{
				fnd:       mockFnd,
				loader:    mockLoader,
				factories: mockFactories,
			}

			err := p.parseField(tt.data, fieldValue, tt.fieldName, tt.params, "/opt/config/wst.yaml")

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

func Test_ConfigParser_ParseStruct(t *testing.T) {
	type parseValidTestStruct struct {
		A           int    `wst:"name=a,default=5"`
		B           string `wst:"name=b,default=default_str"`
		C           int    `wst:""`
		D           bool   `wst:"d"`
		F           string
		P           string `wst:"p,path"`
		UnexportedE bool   `wst:"name=e"`
	}
	type parseInvalidTagTestStruct struct {
		A int `wst:"a,unknown=1"`
	}
	type parseInvalidFactoryTestStruct struct {
		A parseValidTestStruct `wst:"a,factory=test"`
	}
	type parseInvalidDefaultTestStruct struct {
		A parseValidTestStruct `wst:"a,default=data"`
	}

	mockLoader := &loaderMocks.MockLoader{}

	mockFactories := &factoryMocks.MockFunctions{}
	mockFactories.On("GetFactoryFunc", "test").Return(nil)

	mockFs := afero.NewMemMapFs()
	// mock app.Foundation
	mockFnd := &appMocks.MockFoundation{}
	mockFnd.On("Fs").Return(mockFs)

	err := afero.WriteFile(mockFs, "/opt/config/test/path", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)

	p := &ConfigParser{
		fnd:       mockFnd,
		loader:    mockLoader,
		factories: mockFactories,
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
			name:           "Test path setting with new path",
			data:           map[string]interface{}{"p": "test/path", "wst/path": "/opt/config/wst.yaml"},
			testStruct:     &parseValidTestStruct{},
			expectedStruct: &parseValidTestStruct{A: 5, B: "default_str", C: 0, D: false, P: "/opt/config/test/path"},
			errMsg:         "",
		},
		{
			name:           "Test path setting with invalid new path type",
			data:           map[string]interface{}{"p": "test/path", "wst/path": 12},
			testStruct:     &parseValidTestStruct{},
			expectedStruct: &parseValidTestStruct{},
			errMsg:         "unexpected type int for path",
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
			errMsg:         "factory function test not found",
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
			err := p.ParseStruct(tt.data, tt.testStruct, "/var/www/config.yaml")

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
		fnd    app.Foundation
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
			assert.Equalf(t, tt.want, CreateParser(tt.args.fnd, tt.args.loader), "CreateParser(%v, %v)", tt.args.fnd, tt.args.loader)
		})
	}
}
