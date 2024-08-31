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
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/loader"
	"github.com/bukka/wst/conf/parser/factory"
	"github.com/bukka/wst/conf/parser/location"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	loaderMocks "github.com/bukka/wst/mocks/generated/conf/loader"
	factoryMocks "github.com/bukka/wst/mocks/generated/conf/parser/factory"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"os"
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
			param: ConfigParamName,
			want:  true,
		},
		{
			name:  "valid parameter - loadable",
			param: ConfigParamLoadable,
			want:  true,
		},
		{
			name:  "valid parameter - default",
			param: ConfigParamDefault,
			want:  true,
		},
		{
			name:  "valid parameter - factory",
			param: ConfigParamFactory,
			want:  true,
		},
		{
			name:  "valid parameter - enum",
			param: ConfigParamEnum,
			want:  true,
		},
		{
			name:  "valid parameter - keys",
			param: ConfigParamKeys,
			want:  true,
		},
		{
			name:  "valid parameter - path",
			param: ConfigParamPath,
			want:  true,
		},
		{
			name:  "valid parameter - string",
			param: ConfigParamString,
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
			parser := ConfigParser{loc: location.CreateLocation()}
			got, err := parser.ParseTag(tt.tag)

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
			p := &ConfigParser{loc: location.CreateLocation()}
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
		}), nil)
	factories.On("GetFactoryFunc", "invalidFactory").Return(
		nil,
		errors.New("unknwon factory"),
	)
	factories.On("GetFactoryFunc", "errorMockFactory").Return(
		factory.Func(func(data interface{}, fieldValue reflect.Value, path string) error {
			return fmt.Errorf("forced error")
		}), nil)

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
			p := ConfigParser{loc: location.CreateLocation()}
			if err := p.processEnumParam(tt.enums, tt.data, tt.fieldName); (err != nil) != tt.wantErr {
				t.Errorf("processEnumParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ConfigParser_processKeysParam(t *testing.T) {
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
			p := ConfigParser{loc: location.CreateLocation()}
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
			want: map[string]interface{}{
				"key":      "value",
				"wst/path": "/configs/test.json",
			},
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
			errMsg:     "type of field f1 is neither map nor slice (kind=string)",
		},
		{
			name:       "Error from GlobConfigs",
			data:       "*.json",
			fieldValue: reflect.ValueOf(map[string]interface{}{}),
			want:       nil,
			wantErr:    true,
			errMsg:     "loading configs for field f1: forced GlobConfigs error",
		},
		{
			name:       "Skip if data is not string",
			data:       1,
			fieldValue: reflect.ValueOf(map[string]interface{}{}),
			want:       1,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoader := &loaderMocks.MockLoader{}
			if tt.name != "Error from GlobConfigs" {
				stringData, isString := tt.data.(string)
				if isString {
					mockLoader.On("GlobConfigs", stringData, "/var/www").Return([]loader.LoadedConfig{mockLoadedConfig}, nil)
				}
			} else {
				mockLoader.On("GlobConfigs", tt.data.(string), "/var/www").Return(nil, errors.New("forced GlobConfigs error"))
			}

			p := ConfigParser{
				fnd:    nil, // replace with necessary mock if necessary
				loader: mockLoader,
				loc:    location.CreateLocation(),
			}
			p.loc.StartObject()
			p.loc.SetField("f1")
			got, err := p.processLoadableParam(tt.data, tt.fieldValue, "/var/www/wst.yaml")

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

type ErrorFs struct {
	afero.Fs
	ShouldError bool
}

func (efs *ErrorFs) Stat(name string) (os.FileInfo, error) {
	if efs.ShouldError {
		return nil, fmt.Errorf("injected error for Exists call")
	}
	return efs.Fs.Stat(name)
}

func TestProcessPathParam(t *testing.T) {
	mockFs := &ErrorFs{
		Fs: afero.NewMemMapFs(),
	}

	// mock app.Foundation
	mockFnd := &appMocks.MockFoundation{}
	mockFnd.On("Fs").Return(mockFs)

	tests := []struct {
		name           string
		fieldValue     reflect.Value
		data           interface{}
		configPath     string
		configPathType string
		fsExistsErr    bool
		wantErr        bool
		expectedErr    string
		expectedValue  string
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
			name:           "Valid virtual path",
			fieldValue:     reflect.Indirect(reflect.ValueOf(new(string))),
			data:           "test/path",
			configPath:     "/home/wst.yaml",
			configPathType: "virtual",
			wantErr:        false,
			expectedValue:  "/home/test/path",
		},
		{
			name:        "Existence path check error",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(string))),
			data:        "/test/path",
			configPath:  "/opt/config/wst.yaml",
			fsExistsErr: true,
			wantErr:     true,
			expectedErr: "injected error for Exists call",
		},
		{
			name:        "Non-existent path",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(string))),
			data:        "test/non-existent",
			configPath:  "config/",
			wantErr:     true,
			expectedErr: "file path config/test/non-existent for field unknown does not exist",
		},
		{
			name:        "Data is not a string",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(string))),
			data:        123,
			configPath:  "config/",
			wantErr:     true,
			expectedErr: "unexpected type int for data in field unknown, expected string",
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
			expectedErr: "unexpected type <nil> for data in field unknown, expected string",
		},
		{
			name:        "Invalid configPath",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(string))),
			data:        "test/path",
			configPath:  "invalid/config/path",
			wantErr:     true,
			expectedErr: "file path invalid/config/test/path for field unknown does not exist",
		},
		{
			name:        "Field value cannot be set",
			fieldValue:  reflect.ValueOf(new(string)),
			data:        "test/path",
			configPath:  "/opt/config/wst.yaml",
			wantErr:     true,
			expectedErr: "field unknown is not settable",
		},
		{
			name:        "Invalid field value type",
			fieldValue:  reflect.Indirect(reflect.ValueOf(new(int))),
			data:        "test/path",
			configPath:  "/opt/config/wst.yaml",
			wantErr:     true,
			expectedErr: "field unknown is not of type string",
		},
	}

	// Create test files
	err := afero.WriteFile(mockFs, "/opt/config/test/path", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/test/path", []byte(`{"key": "value2"}`), 0644)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFs.ShouldError = tt.fsExistsErr
			parser := ConfigParser{fnd: mockFnd, loc: location.CreateLocation()}
			err := parser.processPathParam(tt.data, tt.fieldValue, tt.configPath, tt.configPathType)

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
	StringField string `wst:"string_field"`
	IntField    int    `wst:"int_field"`
}

type StringParamParentStruct struct {
	Child StringParamTestStruct
}

type StringParamTestMapStruct struct {
	MapField map[string]StringParamTestStruct
}

type StringParamTestMapInvalidStruct struct {
	MapField map[string]StringParamInvalidTagTestStruct
}

type StringParamInvalidTagTestStruct struct {
	A int `wst:"a,unknown=1"`
}

type StringParamInvalidParentStruct struct {
	Child StringParamInvalidTagTestStruct
}

type StringParamUnexportedFieldStruct struct {
	unexportedField string // Unexported: cannot be set via reflection.
}

type StringParamNonexistentFieldStruct struct {
	ExportedField string // This struct is okay, but we'll try to set a nonexistent field.
}

type StringParamMapFieldStruct struct {
	Field1 string
	Field2 string
}

func Test_ConfigParser_processStringParam(t *testing.T) {
	// Prepare ConfigParser
	p := ConfigParser{fnd: nil, loc: location.CreateLocation()}

	// Testing data setup
	dataVal := "stringValue"

	var structVal StringParamParentStruct

	// Testing data setup for map
	mapDataVal := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	var mapStructVal StringParamTestMapStruct
	var mapInvalidStructVal StringParamTestMapInvalidStruct
	var invalidStruct StringParamInvalidParentStruct

	msVal := reflect.ValueOf(&mapStructVal)
	msField := msVal.Elem().FieldByName("MapField")

	strVal := "test"

	tests := []struct {
		name       string
		fieldName  string
		data       interface{}
		fieldValue reflect.Value
		changed    bool
		want       interface{}
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "process string param in struct field",
			fieldName:  "StringField",
			data:       dataVal,
			fieldValue: reflect.ValueOf(&structVal.Child),
			changed:    true,
			want:       StringParamTestStruct{StringField: dataVal},
		},
		{
			name:       "process string param with invalid data type",
			fieldName:  "StringField",
			data:       false,
			fieldValue: reflect.ValueOf(&structVal.Child),
			changed:    false,
		},
		{
			name:       "process string param with invalid struct",
			fieldName:  "StringField",
			data:       dataVal,
			fieldValue: reflect.ValueOf(&invalidStruct.Child),
			wantErr:    true,
			errMsg:     "error parsing struct in field A for string param: field A invalid parameter key: unknown",
		},
		{
			name:       "attempt to set an unexported field",
			fieldName:  "unexportedField",
			data:       "value",
			fieldValue: reflect.ValueOf(&StringParamUnexportedFieldStruct{}),
			wantErr:    true,
			errMsg:     "cannot set the field: unexportedField",
		},
		{
			name:       "attempt to set a nonexistent field",
			fieldName:  "NonexistentField",
			data:       "value",
			fieldValue: reflect.ValueOf(&StringParamNonexistentFieldStruct{}),
			wantErr:    true,
			errMsg:     "failed to set field A : field A does not have a valid field name: NonexistentField",
		},
		{
			name:      "process map param with shallow fields",
			fieldName: "StringField",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			fieldValue: msField,
			changed:    true,
			want: map[string]StringParamTestStruct{
				"key1": {StringField: "value1"},
				"key2": {StringField: "value2"},
			},
		},
		{
			name:      "process map param with shallow fields",
			fieldName: "StringField",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			fieldValue: msField,
			changed:    true,
			want: map[string]StringParamTestStruct{
				"key1": {StringField: "value1"},
				"key2": {StringField: "value2"},
			},
		},
		{
			name:      "process map param with shallow fields",
			fieldName: "StringField",
			data: map[string]interface{}{
				"key1": map[string]interface{}{
					"string_field": "str",
					"int_field":    2,
				},
				"key2": "value2",
			},
			fieldValue: msField,
			changed:    true,
			want: map[string]StringParamTestStruct{
				"key1": {StringField: "str", IntField: 2},
				"key2": {StringField: "value2"},
			},
		},
		{
			name:      "process map param with invalid map value type",
			fieldName: "StringField",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": 1,
			},
			fieldValue: msField,
			wantErr:    true,
			errMsg:     "invalid map value type for string param - expected string, got int",
		},
		{
			name:       "process map param with non-existent field",
			fieldName:  "NonExistentField",
			data:       map[string]interface{}{"NonExistentField": "someValue"},
			fieldValue: msField,
			wantErr:    true,
			errMsg:     "failed to set field A: field A does not have a valid field name: NonExistentField",
		},
		{
			name:       "process map param with invalid struct field",
			fieldName:  "InvalidField",
			data:       mapDataVal,
			fieldValue: reflect.ValueOf(mapInvalidStructVal.MapField),
			wantErr:    true,
			errMsg:     "field A.A error parsing struct for string param: field A.A invalid parameter key: unknown",
		},
		{
			name:       "process map param with invalid data type",
			fieldName:  "StringField",
			data:       "data",
			fieldValue: msField,
			changed:    false,
		},
		{
			name:       "process string param with string value",
			fieldName:  "StringField",
			data:       dataVal,
			fieldValue: reflect.ValueOf(strVal),
			wantErr:    true,
			errMsg:     "field A.A must be a struct or interface type or a pointer to such",
		},
		{
			name:       "process string param with pointer to string value",
			fieldName:  "StringField",
			data:       dataVal,
			fieldValue: reflect.ValueOf(&strVal),
			wantErr:    true,
			errMsg:     "field A.A value must be a pointer to a struct or a map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			result, err := p.processStringParam(tt.fieldName, tt.data, tt.fieldValue, "/var/www/config.yaml")

			if tt.wantErr {
				assert.Error(err)
				if tt.errMsg != "" {
					assert.ErrorContains(err, tt.errMsg)
				}
			} else {
				assert.NoError(err)
				if !tt.changed {
					assert.False(result)
				} else {
					var actualValue interface{}
					if tt.fieldValue.Kind() == reflect.Ptr && !tt.fieldValue.IsNil() {
						actualValue = tt.fieldValue.Elem().Interface()
					} else {
						actualValue = tt.fieldValue.Interface()
					}
					assert.Equal(tt.want, actualValue)
					assert.True(result)
				}
			}
		})
	}
}

type AssignFieldTestStruct struct {
	A string
	B int
	C []string
	D map[string]int
	E int32
	F int16
	G int8
	H int64
	I []map[string]int
}

type AssignFieldAnotherStruct struct {
	E string `wst:"e"`
	F bool   `wst:"f"`
}

type AssignFieldParentStruct struct {
	Child AssignFieldAnotherStruct
}

func Test_ConfigParser_assignField(t *testing.T) {
	p := ConfigParser{fnd: nil, loc: location.CreateLocation()}

	tests := []struct {
		name          string
		fieldName     string
		data          interface{}
		value         interface{}
		expectedValue interface{}
		wantErr       bool
		errMsg        string
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
			name:      "assign array field from array of string",
			fieldName: "C",
			data:      []interface{}{"TestA", "TestB"},
			value:     &AssignFieldTestStruct{},
			wantErr:   false,
			expectedValue: &AssignFieldTestStruct{
				C: []string{"TestA", "TestB"},
			},
		},
		{
			name:      "assign array field from array of map",
			fieldName: "I",
			data: []map[string]interface{}{
				{
					"a": 1,
				},
				{
					"a": 2,
				},
			},
			value:   &AssignFieldTestStruct{},
			wantErr: false,
			expectedValue: &AssignFieldTestStruct{
				I: []map[string]int{
					{
						"a": 1,
					},
					{
						"a": 2,
					},
				},
			},
		},
		{
			name:      "assign array field from invalid values",
			fieldName: "C",
			data:      []interface{}{"TestA", 1},
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "field [1] is an integer and cannot be converted to string",
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
			name:      "assign invalid data in map field",
			fieldName: "D",
			data:      map[string]interface{}{"test": "xy"},
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "field [1].test could not be set due to type mismatch or non-convertible types",
		},
		{
			name:      "assign invalid data to struct",
			fieldName: "Child",
			data:      "test",
			value:     &AssignFieldParentStruct{},
			wantErr:   true,
			errMsg:    "unable to convert data for field [1].test to map[string]interface{}",
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
		{
			name:          "assign integer field to int64",
			fieldName:     "H",
			data:          123,
			value:         &AssignFieldTestStruct{},
			expectedValue: &AssignFieldTestStruct{H: 123},
		},
		{
			name:          "assign integer field to int32",
			fieldName:     "E",
			data:          int64(123),
			value:         &AssignFieldTestStruct{},
			expectedValue: &AssignFieldTestStruct{E: 123},
		},
		{
			name:      "assign integer field with overflow (int32 from int64)",
			fieldName: "E",
			data:      int64(math.MaxInt32 + 1), // This should overflow int32
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "overflow error",
		},
		{
			name:          "assign integer field to int16",
			fieldName:     "F",
			data:          int64(123),
			value:         &AssignFieldTestStruct{},
			expectedValue: &AssignFieldTestStruct{F: 123},
		},
		{
			name:      "assign integer field with overflow (int16 from int64)",
			fieldName: "F",
			data:      int64(math.MaxInt16 + 1), // This should overflow int16
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "overflow error",
		},
		{
			name:          "assign integer field to int8",
			fieldName:     "G",
			data:          int64(123),
			value:         &AssignFieldTestStruct{},
			expectedValue: &AssignFieldTestStruct{G: 123},
		},
		{
			name:      "assign integer field with overflow (int8 from int64)",
			fieldName: "G",
			data:      int64(math.MaxInt8 + 1), // This should overflow int8
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "overflow error",
		},
		{
			name:      "assign float to integer field with fractional part",
			fieldName: "B",
			data:      12.34, // Has a fractional part, cannot be cleanly converted to int
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "float64 value has a fractional part",
		},
		{
			name:          "assign float to integer field without fractional part",
			fieldName:     "B",
			data:          12.0, // Does not have a fractional part, can be cleanly converted to int
			value:         &AssignFieldTestStruct{},
			expectedValue: &AssignFieldTestStruct{B: 12},
		},
		{
			name:      "assign incompatible type",
			fieldName: "A",            // A is a string in AssignFieldTestStruct
			data:      []int{1, 2, 3}, // Trying to assign a slice of ints to a string
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "type mismatch or non-convertible types",
		},
		{
			name:      "assign field with unconvertible type",
			fieldName: "B", // B is an integer field in AssignFieldTestStruct
			data:      "incompatibleString",
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "could not be set due to type mismatch",
		},
		{
			name:      "assign map field with incompatible key type",
			fieldName: "D",
			data:      map[int]interface{}{1: "one"}, // Assuming D expects map[string]int
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "unable to convert data for field [1].test to map[string]interface{}",
		},
		{
			name:      "assign slice field with single incompatible element",
			fieldName: "C",
			data:      "test", // Mixed types, expecting []string
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
			errMsg:    "unable to convert data for field [1].test to []interface{}",
		},
		{
			name:      "assign non-struct/interface/ptr type",
			fieldName: "NonStructField", // This test implies existence of such a field
			data:      "value",
			value:     new(int),
			wantErr:   true,
			errMsg:    "field [1].test could not be set due to type mismatch or non-convertible types",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			fieldValue := reflect.ValueOf(tt.value)
			if fieldValue.Kind() == reflect.Ptr {
				fieldValue = reflect.ValueOf(tt.value).Elem()
				if fieldValue.Kind() == reflect.Struct {
					fieldValue = fieldValue.FieldByName(tt.fieldName)
				}
				if !fieldValue.IsValid() {
					t.Fatalf("Invalid test setup: field %s does not exist in %T", tt.fieldName, tt.value)
				}
			}

			err := p.assignField(tt.data, fieldValue, tt.fieldName, "/var/www/config.yaml")

			if tt.wantErr {
				assert.Error(err)
				if tt.errMsg != "" {
					assert.ErrorContains(err, tt.errMsg)
				}
			} else {
				assert.Equal(tt.value, tt.expectedValue)
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
		name                string
		fieldName           string
		data                interface{}
		params              map[ConfigParam]string
		expectedFieldValue  interface{}
		configsCalled       bool
		configsData         []ParseFieldConfigData
		factoryFound        bool
		factoryExpectedKind reflect.Kind
		wantErr             bool
		errMsg              string
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
				{Path: "services/test1.yaml", Data: map[string]interface{}{
					"s1": map[string]interface{}{"val": "test1"},
				}},
				{Path: "services/test2.yaml", Data: map[string]interface{}{
					"s2": map[string]interface{}{"val": "test2"},
				}},
			},
			factoryFound: false,
			expectedFieldValue: &ParseFieldTestStruct{E: map[string]ParseFieldInnerTestStruct{
				"s2": {Value: "test2"}, // For map only the last item is added
			}},
			wantErr: false,
		},
		{
			name:      "parse field with loadable and factory param found",
			fieldName: "D",
			data:      "data.yaml",
			params: map[ConfigParam]string{
				"factory":  "test",
				"loadable": "true",
			},
			configsCalled: true,
			configsData: []ParseFieldConfigData{
				{Path: "data.yaml", Data: map[string]interface{}{"val": "test"}},
			},
			factoryFound:        true,
			factoryExpectedKind: reflect.Slice,
			expectedFieldValue: &ParseFieldTestStruct{
				D: []ParseFieldInnerTestStruct{
					{Value: "test_data_1"},
					{Value: "test_data_2"},
				},
			},
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
		{
			name:      "parse field with factory param found",
			fieldName: "A",
			data:      map[string]interface{}{"a": "NestedTest", "b": 1},
			params:    map[ConfigParam]string{},
			wantErr:   true,
			errMsg:    "field unknown could not be set due to type mismatch or non-convertible types",
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
			assert := assert.New(t)
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
					mockLoader.On("GlobConfigs", tt.data.(string), "/opt/config").Return(mockConfigs, nil)
				} else {
					mockLoader.On("GlobConfigs", tt.data.(string), "/opt/config").Return(
						nil, errors.New("forced GlobConfigs error"))
				}
			}

			mockFactories := &factoryMocks.MockFunctions{}
			if tt.factoryFound {
				mockFactories.On("GetFactoryFunc", "test").Return(
					factory.Func(func(data interface{}, fieldValue reflect.Value, path string) error {
						if tt.factoryExpectedKind != reflect.Invalid {
							assert.Equal(tt.factoryExpectedKind.String(), fieldValue.Kind().String())
							if tt.factoryExpectedKind == reflect.Slice {
								// Create a slice of ParseFieldInnerTestStruct
								sliceType := fieldValue.Type().Elem()
								slice := reflect.MakeSlice(reflect.SliceOf(sliceType), 0, 0)

								// Populate the slice with test data
								for i := 0; i < 2; i++ {
									elem := reflect.New(sliceType).Elem()
									elem.FieldByName("Value").SetString(fmt.Sprintf("test_data_%d", i+1))
									slice = reflect.Append(slice, elem)
								}

								fieldValue.Set(slice)
								return nil
							}
						}
						fieldValue.SetString("test_data")
						return nil
					}), nil)
			} else {
				mockFactories.On("GetFactoryFunc", "test").Return(
					nil,
					errors.New("factory func not found"),
				)
			}

			// Create a new ConfigParser for each test case
			p := &ConfigParser{
				fnd:       mockFnd,
				loader:    mockLoader,
				factories: mockFactories,
				loc:       location.CreateLocation(),
			}

			err := p.parseField(tt.data, fieldValue, tt.fieldName, tt.params, "/opt/config/wst.yaml")

			if tt.wantErr {
				assert.Error(err)
				if tt.errMsg != "" {
					assert.ErrorContains(err, tt.errMsg)
				}
			} else {
				assert.NoError(err)
				assert.Equal(tt.expectedFieldValue, commonFieldValue)
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
		E           int    `wst:"default=3"`
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
	mockFactories.On("GetFactoryFunc", "test").Return(nil, errors.New("unknown factory"))

	mockFs := afero.NewMemMapFs()
	// mock app.Foundation
	mockFnd := &appMocks.MockFoundation{}
	mockFnd.On("Fs").Return(mockFs)

	err := afero.WriteFile(mockFs, "/opt/config/test/path", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)

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
			expectedStruct: &parseValidTestStruct{A: 5, B: "default_str", C: 0, D: false, E: 3},
			errMsg:         "",
		},
		{
			name:           "Test path setting with new path",
			data:           map[string]interface{}{"p": "test/path", "wst/path": "/opt/config/wst.yaml", "E": 2},
			testStruct:     &parseValidTestStruct{},
			expectedStruct: &parseValidTestStruct{A: 5, B: "default_str", C: 0, D: false, E: 2, P: "/opt/config/test/path"},
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
			errMsg:         "unknown factory",
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
			p := &ConfigParser{
				fnd:       mockFnd,
				loader:    mockLoader,
				factories: mockFactories,
				loc:       location.CreateLocation(),
			}

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

func Test_ConfigParser_ParseConfig(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]interface{}
		files          map[string]string
		expectedConfig *types.Config
		setupMocks     func(fnd *appMocks.MockFoundation)
		path           string
		errMsg         string
	}{
		{
			name: "successfully parse big compact config",
			data: map[string]interface{}{
				"version":     "1.0",
				"name":        "WST Project",
				"description": "A project to demonstrate JSON Schema representation in Go",
				"spec": map[string]interface{}{
					"environments": map[string]interface{}{
						"common": map[string]interface{}{
							"ports": map[string]interface{}{
								"start": 8000,
								"end":   9000,
							},
						},
						"docker": map[string]interface{}{
							"registry": map[string]interface{}{
								"auth": map[string]interface{}{
									"username": "user",
									"password": "pass",
								},
							},
							"name_prefix": "wst_",
						},
						"kubernetes": map[string]interface{}{
							"ports": map[string]interface{}{
								"start": 9500,
								"end":   9800,
							},
							"namespace":  "default",
							"kubeconfig": "/path/to/kubeconfig",
						},
					},
					"sandboxes": map[string]interface{}{
						"common": map[string]interface{}{
							"available": true,
							"dirs": map[string]interface{}{
								"conf":   "/etc/common",
								"run":    "/var/run/common",
								"script": "/usr/local/bin",
							},
							"hooks": map[string]interface{}{
								"start": map[string]interface{}{
									"command": map[string]interface{}{
										"executable": "/usr/local/bin/start-common",
										"args":       []interface{}{"--config", "/etc/common/config.yaml"},
									},
								},
								"stop": map[string]interface{}{
									"signal": "SIGTERM",
								},
							},
						},
						"local": map[string]interface{}{
							"available": true,
							"dirs": map[string]interface{}{
								"conf":   "/etc/local",
								"run":    "/var/run/local",
								"script": "/usr/local/bin",
							},
						},
						"docker": map[string]interface{}{
							"image": map[string]interface{}{
								"name": "wst/docker-sandbox",
								"tag":  "latest",
							},
							"registry": map[string]interface{}{
								"auth": map[string]interface{}{
									"username": "dockeruser",
									"password": "dockerpass",
								},
							},
							"hooks": map[string]interface{}{
								"restart": map[string]interface{}{
									"native": map[string]interface{}{
										"force": true,
									},
								},
							},
						},
						"kubernetes": map[string]interface{}{
							"image": map[string]interface{}{
								"name": "wst/k8s-sandbox",
								"tag":  "v1.0",
							},
							"registry": map[string]interface{}{
								"auth": map[string]interface{}{
									"username": "k8suser",
									"password": "k8spass",
								},
							},
							"auth": map[string]interface{}{
								"kubeconfig": "/home/user/.kube/config",
							},
						},
					},
					"servers": []map[string]interface{}{
						{
							"name": "web_server",
							"user": "webuser",
							"port": 8080,
							"configs": map[string]interface{}{
								"nginx.conf": map[string]interface{}{
									"file": "/etc/nginx/nginx.conf",
									"parameters": map[string]interface{}{
										"worker_processes": "2",
									},
								},
							},
							"actions": map[string]interface{}{
								"expect": map[string]interface{}{
									"status": map[string]interface{}{
										"parameters": map[string]interface{}{
											"body": "1",
										},
										"response": map[string]interface{}{
											"headers": map[string]interface{}{
												"content-type": "application/json",
											},
											"body": map[string]interface{}{
												"content": "{{ .Parameters.GetString \"body\" }}",
											},
										},
									},
								},
							},
						},
					},
					"instances": []interface{}{
						map[string]interface{}{
							"name": "Instance 1",
							"environments": map[string]interface{}{
								"local": map[string]interface{}{
									"ports": map[string]interface{}{
										"start": 9500,
										"end":   9600,
									},
								},
							},
							"resources": map[string]interface{}{
								"scripts": map[string]interface{}{
									"index_php": map[string]interface{}{
										"content": "<?php echo 1; ?>",
										"path":    "index.php",
									},
								},
							},
							"services": map[string]interface{}{
								"web_service": map[string]interface{}{
									"server": map[string]interface{}{
										"name": "web_server",
									},
									"resources": map[string]interface{}{
										"scripts": []interface{}{
											"init.sh",
										},
									},
								},
							},
							"actions": []interface{}{
								"start/web_service",
								map[string]interface{}{
									"request": map[string]interface{}{
										"service": "web_service",
										"path":    "/api/status",
										"method":  "GET",
									},
								},
								map[string]interface{}{
									"expect/web_service/status": map[string]interface{}{
										"body": "2",
									},
								},
								map[string]interface{}{
									"expect": map[string]interface{}{
										"service": "web_service",
										"custom": map[string]interface{}{
											"name": "status",
											"parameters": map[string]interface{}{
												"body": "3",
											},
										},
									},
								},
								map[string]interface{}{
									"expect": map[string]interface{}{
										"service": "web_service",
										"response": map[string]interface{}{
											"body": map[string]interface{}{
												"content": "OK",
												"match":   "exact",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			files: map[string]string{
				"/path/to/kubeconfig":   "kube: true",
				"/etc/nginx/nginx.conf": "nginx on",
			},
			expectedConfig: &types.Config{
				Version:     "1.0",
				Name:        "WST Project",
				Description: "A project to demonstrate JSON Schema representation in Go",
				Spec: types.Spec{
					Environments: map[string]types.Environment{
						string(types.CommonEnvironmentType): &types.CommonEnvironment{
							Ports: types.EnvironmentPorts{
								Start: 8000,
								End:   9000,
							},
						},
						string(types.DockerEnvironmentType): &types.DockerEnvironment{
							Registry: types.ContainerRegistry{
								Auth: types.ContainerRegistryAuth{
									Username: "user",
									Password: "pass",
								},
							},
							NamePrefix: "wst_",
						},
						string(types.KubernetesEnvironmentType): &types.KubernetesEnvironment{
							Ports: types.EnvironmentPorts{
								Start: 9500,
								End:   9800,
							},
							Namespace:  "default",
							Kubeconfig: "/path/to/kubeconfig",
						},
					},
					Instances: []types.Instance{
						{
							Name: "Instance 1",
							Resources: types.Resources{
								Scripts: map[string]types.Script{
									"index_php": {
										Content:    "<?php echo 1; ?>",
										Path:       "index.php",
										Mode:       "0644",
										Parameters: nil,
									},
								},
							},
							Services: map[string]types.Service{
								"web_service": types.Service{
									Server: types.ServiceServer{
										Name:    "web_server",
										Sandbox: "local",
									},
									Resources: types.ServiceResources{
										Scripts: types.ServiceScripts{
											IncludeList: []string{"init.sh"},
										},
									},
									Requires: nil,
									Public:   false,
								},
							},
							Timeouts: types.InstanceTimeouts{},
							Environments: map[string]types.Environment{
								"local": &types.LocalEnvironment{
									Ports: types.EnvironmentPorts{
										Start: 9500,
										End:   9600,
									},
								},
							},
							Actions: []types.Action{
								&types.StartAction{
									Service:  "web_service",
									Services: nil,
									Timeout:  0,
									When:     "on_success",
								},
								&types.RequestAction{
									Service: "web_service",
									Timeout: 0,
									When:    "on_success",
									Id:      "last",
									Path:    "/api/status",
									Method:  "GET",
								},
								&types.CustomExpectationAction{
									Service: "web_service",
									Timeout: 0,
									When:    "on_success",
									Custom: types.CustomExpectation{
										Name: "status",
										Parameters: types.Parameters{
											"body": "2",
										},
									},
								},
								&types.CustomExpectationAction{
									Service: "web_service",
									Timeout: 0,
									When:    "on_success",
									Custom: types.CustomExpectation{
										Name: "status",
										Parameters: types.Parameters{
											"body": "3",
										},
									},
								},
								&types.ResponseExpectationAction{
									Service: "web_service",
									Timeout: 0,
									When:    "on_success",
									Response: types.ResponseExpectation{
										Request: "last",
										Body: types.ResponseBody{
											Content:        "OK",
											Match:          "exact",
											RenderTemplate: true,
										},
									},
								},
							},
						},
					},
					Sandboxes: map[string]types.Sandbox{
						"common": &types.CommonSandbox{
							Available: true,
							Dirs: map[string]string{
								"conf":   "/etc/common",
								"run":    "/var/run/common",
								"script": "/usr/local/bin",
							},
							Hooks: map[string]types.SandboxHook{
								"start": &types.SandboxHookArgsCommand{
									Executable: "/usr/local/bin/start-common",
									Args:       []string{"--config", "/etc/common/config.yaml"},
								},
								"stop": &types.SandboxHookSignal{
									IsString:    true,
									StringValue: "SIGTERM",
								},
							},
						},
						"local": &types.LocalSandbox{
							Available: true,
							Dirs: map[string]string{
								"conf":   "/etc/local",
								"run":    "/var/run/local",
								"script": "/usr/local/bin",
							},
						},
						"docker": &types.DockerSandbox{
							Available: true,
							Hooks: map[string]types.SandboxHook{
								"restart": &types.SandboxHookNative{
									Enabled: true,
									Force:   true,
								},
							},
							Image: types.ContainerImage{
								Name: "wst/docker-sandbox",
								Tag:  "latest",
							},
							Registry: types.ContainerRegistry{
								Auth: types.ContainerRegistryAuth{
									Username: "dockeruser",
									Password: "dockerpass",
								},
							},
						},
						"kubernetes": &types.KubernetesSandbox{
							Available: true,
							Image: types.ContainerImage{
								Name: "wst/k8s-sandbox",
								Tag:  "v1.0",
							},
							Registry: types.ContainerRegistry{
								Auth: types.ContainerRegistryAuth{
									Username: "k8suser",
									Password: "k8spass",
								},
							},
						},
					},
					Servers: []types.Server{
						{
							Name: "web_server",
							User: "webuser",
							Port: 8080,
							Configs: map[string]types.ServerConfig{
								"nginx.conf": {
									File: "/etc/nginx/nginx.conf",
									Parameters: types.Parameters{ // Assuming Parameters is a map[string]interface{} or similar
										"worker_processes": "2",
									},
								},
							},
							Actions: types.ServerActions{
								Expect: map[string]types.ServerExpectationAction{
									"status": &types.ServerResponseExpectation{
										Parameters: types.Parameters{
											"body": "1",
										},
										Response: types.ResponseExpectation{
											Request: "last",
											Headers: map[string]string{
												"content-type": "application/json",
											},
											Body: types.ResponseBody{
												Content:        "{{ .Parameters.GetString \"body\" }}",
												Match:          "exact",
												RenderTemplate: true,
											},
										},
									},
								},
							},
						},
					},
					Workspace: "",
				},
			},
			setupMocks: func(fnd *appMocks.MockFoundation) {
				fnd.On("Getwd").Return("/home", nil)
				fnd.On("Chdir", "/var/www").Return(nil)
				fnd.On("Chdir", "/home").Return(nil)
			},
			path: "/var/www/wst.yaml",
		},
		{
			name: "failure on changing directory",
			data: map[string]interface{}{
				"version":     "1.0",
				"name":        "WST Project",
				"description": "A project to demonstrate JSON Schema representation in Go",
				"spec":        map[string]interface{}{},
			},
			setupMocks: func(fnd *appMocks.MockFoundation) {
				fnd.On("Getwd").Return("/home", nil)
				fnd.On("Chdir", "/var/www").Return(errors.New("chdir fail"))
			},
			path:   "/var/www/wst.yaml",
			errMsg: "chdir fail",
		},
		{
			name: "failure on getting working directory",
			data: map[string]interface{}{
				"version":     "1.0",
				"name":        "WST Project",
				"description": "A project to demonstrate JSON Schema representation in Go",
				"spec":        map[string]interface{}{},
			},
			setupMocks: func(fnd *appMocks.MockFoundation) {
				fnd.On("Getwd").Return("", errors.New("getwd fail"))
			},
			path:   "/var/www/wst.yaml",
			errMsg: "getwd fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLoader := &loaderMocks.MockLoader{}

			mockFs := afero.NewMemMapFs()
			// mock app.Foundation
			mockFnd := &appMocks.MockFoundation{}
			mockFnd.On("Fs").Return(mockFs)

			for fileName, content := range tt.files {
				err := afero.WriteFile(mockFs, fileName, []byte(content), 0644)
				assert.NoError(t, err)
			}

			tt.setupMocks(mockFnd)

			p := CreateParser(mockFnd, mockLoader)
			config := types.Config{}
			err := p.ParseConfig(tt.data, &config, tt.path)

			if tt.errMsg != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, tt.expectedConfig, &config)
				} else {
					errExtra, ok := err.(stackTracer)
					if ok {
						fmt.Printf("An error occurred: %+v\n", errExtra)
					}
				}
			}
		})
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func TestCreateParser(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	loaderMock := loaderMocks.NewMockLoader(t)

	tests := []struct {
		name   string
		fnd    app.Foundation
		loader loader.Loader
	}{
		{
			name:   "Testing CreateLoader",
			fnd:    fndMock,
			loader: loaderMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateParser(tt.fnd, tt.loader)
			parser, ok := got.(*ConfigParser)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, parser.fnd)
			assert.NotNil(t, parser.factories)
		})
	}
}
