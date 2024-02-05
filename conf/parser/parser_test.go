package parser

import (
	"errors"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/loader"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/appMocks"
	"github.com/bukka/wst/mocks/confMocks"
	"github.com/bukka/wst/mocks/externalMocks"
	"github.com/spf13/afero"
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
		name       string
		tag        string
		want       map[string]string
		wantErrors []string
	}{
		{
			name: "Testing ParseTag - With all valid params and implicit name",
			tag:  "tagname,default=value1,enum",
			want: map[string]string{
				"name":    "tagname",
				"default": "value1",
				"enum":    "true",
			},
			wantErrors: nil,
		},
		{
			name: "Testing ParseTag - With all valid params and explicit name",
			tag:  "name=tagname,default=value1,enum",
			want: map[string]string{
				"name":    "tagname",
				"default": "value1",
				"enum":    "true",
			},
			wantErrors: nil,
		},
		{
			name:       "Testing ParseTag - invalid parameter key",
			tag:        "invalid=key",
			want:       map[string]string{},
			wantErrors: []string{"Invalid parameter key: invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create the logger
			mockLogger := externalMocks.NewMockLogger()

			// create and setup a mocked environment
			mockEnv := &appMocks.MockEnv{}
			mockEnv.On("Logger").Return(mockLogger.SugaredLogger)
			mockEnv.On("Fs").Return(afero.NewMemMapFs())

			parser := ConfigParser{env: mockEnv}
			got := parser.parseTag(tt.tag)
			assert.Equal(t, tt.want, got)

			// Validate log messages
			messages := mockLogger.Messages()
			if !reflect.DeepEqual(messages, tt.wantErrors) {
				t.Errorf("logger.Warn() calls = %v, want %v", mockLogger.Messages(), tt.wantErrors)
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

type TestStruct struct {
	StringField string
}

type ParentStruct struct {
	Child TestStruct
}

type TestMapStruct struct {
	MapField map[string]TestStruct
}

func Test_ConfigParser_processStringParam(t *testing.T) {
	// Prepare ConfigParser
	p := ConfigParser{env: nil} // you may need to initialize this with suitable fields based on your implementation

	// Testing data setup
	dataVal := "stringValue"

	var structVal ParentStruct

	// Testing data setup for map
	mapDataVal := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	var mapStructVal TestMapStruct

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
			want:       TestStruct{StringField: dataVal},
		},
		{
			name:       "process map param",
			fieldName:  "StringField",
			data:       mapDataVal,
			fieldValue: reflect.ValueOf(&mapStructVal.MapField),
			wantErr:    false,
			want: map[string]TestStruct{
				"key1": TestStruct{StringField: "value1"},
				"key2": TestStruct{StringField: "value2"},
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
	E string
	F bool
}

type AssignFieldParentStruct struct {
	Child AssignFieldAnotherStruct
}

func Test_ConfigParser_assignField(t *testing.T) {
	p := ConfigParser{env: nil} // Initialize appropriately

	tests := []struct {
		name      string
		fieldName string
		data      interface{}
		value     interface{}
		wantErr   bool
	}{
		{
			name:      "assign struct field",
			fieldName: "A",
			data:      "TestA",
			value:     &AssignFieldTestStruct{},
			wantErr:   false,
		},
		{
			name:      "assign tuple in map field",
			fieldName: "D",
			data:      map[string]interface{}{"test": 7},
			value:     &AssignFieldTestStruct{},
			wantErr:   false,
		},
		{
			name:      "assign array field",
			fieldName: "C",
			data:      []interface{}{"TestA", "TestB"},
			value:     &AssignFieldTestStruct{},
			wantErr:   false,
		},
		{
			name:      "assign struct field with mismatched type should error",
			fieldName: "A",
			data:      5,
			value:     &AssignFieldTestStruct{},
			wantErr:   true,
		},
		{
			name:      "assign to nested struct field",
			fieldName: "Child",
			data:      map[string]interface{}{"E": "NestedTest", "F": true},
			value:     &AssignFieldParentStruct{},
			wantErr:   false,
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
		})
	}
}

func TestConfigParser_ParseConfig(t *testing.T) {
	type fields struct {
		env       app.Env
		loader    loader.Loader
		factories map[string]factoryFunc
	}
	type args struct {
		data   map[string]interface{}
		config *types.Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ConfigParser{
				env:       tt.fields.env,
				loader:    tt.fields.loader,
				factories: tt.fields.factories,
			}
			tt.wantErr(t, p.ParseConfig(tt.args.data, tt.args.config), fmt.Sprintf("ParseConfig(%v, %v)", tt.args.data, tt.args.config))
		})
	}
}

func TestConfigParser_parseField(t *testing.T) {
	type fields struct {
		env       app.Env
		loader    loader.Loader
		factories map[string]factoryFunc
	}
	type args struct {
		data       interface{}
		fieldValue reflect.Value
		fieldName  string
		params     map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ConfigParser{
				env:       tt.fields.env,
				loader:    tt.fields.loader,
				factories: tt.fields.factories,
			}
			tt.wantErr(t, p.parseField(tt.args.data, tt.args.fieldValue, tt.args.fieldName, tt.args.params), fmt.Sprintf("parseField(%v, %v, %v, %v)", tt.args.data, tt.args.fieldValue, tt.args.fieldName, tt.args.params))
		})
	}
}

func TestConfigParser_parseStruct(t *testing.T) {
	type fields struct {
		env       app.Env
		loader    loader.Loader
		factories map[string]factoryFunc
	}
	type args struct {
		data map[string]interface{}
		s    interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ConfigParser{
				env:       tt.fields.env,
				loader:    tt.fields.loader,
				factories: tt.fields.factories,
			}
			tt.wantErr(t, p.parseStruct(tt.args.data, tt.args.s), fmt.Sprintf("parseStruct(%v, %v)", tt.args.data, tt.args.s))
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
