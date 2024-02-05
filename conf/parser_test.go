package conf

import (
	"fmt"
	"github.com/bukka/wst/app"
	appMocks "github.com/bukka/wst/mocks/app"
	externalMocks "github.com/bukka/wst/mocks/external"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
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

func TestConfigParser_processLoadableParam(t *testing.T) {
	type fields struct {
		env       app.Env
		loader    Loader
		factories map[string]factoryFunc
	}
	type args struct {
		data       interface{}
		fieldValue reflect.Value
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
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
			got, err := p.processLoadableParam(tt.args.data, tt.args.fieldValue)
			if !tt.wantErr(t, err, fmt.Sprintf("processLoadableParam(%v, %v)", tt.args.data, tt.args.fieldValue)) {
				return
			}
			assert.Equalf(t, tt.want, got, "processLoadableParam(%v, %v)", tt.args.data, tt.args.fieldValue)
		})
	}
}

func TestConfigParser_processStringParam(t *testing.T) {
	type fields struct {
		env       app.Env
		loader    Loader
		factories map[string]factoryFunc
	}
	type args struct {
		fieldName  string
		data       interface{}
		fieldValue reflect.Value
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
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
			got, err := p.processStringParam(tt.args.fieldName, tt.args.data, tt.args.fieldValue)
			if !tt.wantErr(t, err, fmt.Sprintf("processStringParam(%v, %v, %v)", tt.args.fieldName, tt.args.data, tt.args.fieldValue)) {
				return
			}
			assert.Equalf(t, tt.want, got, "processStringParam(%v, %v, %v)", tt.args.fieldName, tt.args.data, tt.args.fieldValue)
		})
	}
}

func TestConfigParser_ParseConfig(t *testing.T) {
	type fields struct {
		env       app.Env
		loader    Loader
		factories map[string]factoryFunc
	}
	type args struct {
		data   map[string]interface{}
		config *Config
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

func TestConfigParser_assignField(t *testing.T) {
	type fields struct {
		env       app.Env
		loader    Loader
		factories map[string]factoryFunc
	}
	type args struct {
		data       interface{}
		fieldValue reflect.Value
		fieldName  string
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
			tt.wantErr(t, p.assignField(tt.args.data, tt.args.fieldValue, tt.args.fieldName), fmt.Sprintf("assignField(%v, %v, %v)", tt.args.data, tt.args.fieldValue, tt.args.fieldName))
		})
	}
}

func TestConfigParser_parseField(t *testing.T) {
	type fields struct {
		env       app.Env
		loader    Loader
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
		loader    Loader
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
		loader Loader
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

func Test_processMapValue(t *testing.T) {
	type args struct {
		rv        reflect.Value
		fieldName string
		strVal    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, processMapValue(tt.args.rv, tt.args.fieldName, tt.args.strVal), fmt.Sprintf("processMapValue(%v, %v, %v)", tt.args.rv, tt.args.fieldName, tt.args.strVal))
		})
	}
}

func Test_setFieldByName(t *testing.T) {
	type args struct {
		v     interface{}
		name  string
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, setFieldByName(tt.args.v, tt.args.name, tt.args.value), fmt.Sprintf("setFieldByName(%v, %v, %v)", tt.args.v, tt.args.name, tt.args.value))
		})
	}
}
