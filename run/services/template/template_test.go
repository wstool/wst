package template

import (
	"bytes"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	templatesMocks "github.com/bukka/wst/mocks/generated/run/servers/templates"
	serviceMocks "github.com/bukka/wst/mocks/generated/run/services/template/service"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/parameters/parameter"
	"github.com/bukka/wst/run/servers/templates"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_nativeMaker_Make(t *testing.T) {
	// Mock foundation, services, and templates
	fndMock := appMocks.NewMockFoundation(t)
	serviceMock := serviceMocks.NewMockTemplateService(t)
	servicesMock := Services{
		"testService": serviceMock,
	}
	serverTemplates := templates.Templates{
		"t1": templatesMocks.NewMockTemplate(t),
	}

	// Create an instance of the maker
	maker := CreateMaker(fndMock)

	// Use the maker to create a template
	resultTemplate := maker.Make(serviceMock, servicesMock, serverTemplates)

	// Cast the result to nativeTemplate to inspect its fields
	resultNativeTemplate, ok := resultTemplate.(*nativeTemplate)
	assert.True(t, ok, "The result should be of type *nativeTemplate")
	assert.NotNil(t, resultNativeTemplate, "The resulting template should not be nil")
	assert.Equal(t, fndMock, resultNativeTemplate.fnd, "Foundation should match the input")
	assert.Equal(t, serviceMock, resultNativeTemplate.service, "Service should match the input")
	assert.Equal(t, servicesMock, resultNativeTemplate.services, "Services should match the input")
	assert.Equal(t, serverTemplates, resultNativeTemplate.serverTemplates, "ServerTemplates should match the input")

}

func Test_nativeTemplate_RenderToWriter(t *testing.T) {
	tests := []struct {
		name        string
		templateStr string
		setupFunc   func(sm *serviceMocks.MockTemplateService) parameters.Parameters
		expectErr   bool
		errMsg      string
		expected    string
	}{
		{
			name:        "Valid template with string param",
			templateStr: "Hello, {{.Parameters.GetString \"key\"}}!",
			setupFunc: func(sm *serviceMocks.MockTemplateService) parameters.Parameters {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				pm := parameterMocks.NewMockParameter(t)
				pm.On("StringValue").Return("World", nil)
				return parameters.Parameters{
					"key": pm,
				}
			},
			expected: "Hello, World!",
		},
		{
			name:        "Valid template with nested string params",
			templateStr: "{{ .Parameters.GetObjectString \"k1\" \"k1_1\" }}",
			setupFunc: func(sm *serviceMocks.MockTemplateService) parameters.Parameters {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				pm1_1 := parameterMocks.NewMockParameter(t)
				pm1_1.On("StringValue").Return("v1_1", nil)
				pm1 := parameterMocks.NewMockParameter(t)
				pm1.On("MapValue").Return(map[string]parameter.Parameter{
					"k1_1": pm1_1,
				})
				return parameters.Parameters{
					"k1": pm1,
				}
			},
			expected: "v1_1",
		},
		{
			name: "Valid template with nested object iteration",
			templateStr: "{{ $obj := .Parameters.GetObject \"k2\" }}{{ $obj.GetString \"k2_2\" }};" +
				"{{ range $k, $v := .Parameters }}{{ $k }}:{{ if $v.IsObject }}[" +
				"{{ range $ik, $iv := $v.ToObject }}{{ $ik }}:{{ $iv.ToString }},{{ end }}" +
				"]{{ else }}{{ $v.ToString }}{{ end }},{{ end }}",
			setupFunc: func(sm *serviceMocks.MockTemplateService) parameters.Parameters {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				pm1 := parameterMocks.NewMockParameter(t)
				pm1.On("Type").Return(parameter.StringType)
				pm1.On("StringValue").Return("v1", nil)
				pm2_1 := parameterMocks.NewMockParameter(t)
				pm2_1.On("StringValue").Return("v2_1", nil)
				pm2_2 := parameterMocks.NewMockParameter(t)
				pm2_2.On("StringValue").Return("v2_2", nil)
				pm2 := parameterMocks.NewMockParameter(t)
				pm2.On("Type").Return(parameter.MapType)
				pm2.On("MapValue").Return(map[string]parameter.Parameter{
					"k2_1": pm2_1,
					"k2_2": pm2_2,
				})
				pm3 := parameterMocks.NewMockParameter(t)
				pm3.On("Type").Return(parameter.StringType)
				pm3.On("StringValue").Return("v3", nil)
				return parameters.Parameters{
					"k1": pm1,
					"k2": pm2,
					"k3": pm3,
				}
			},
			expected: "v2_2;k1:v1,k2:[k2_1:v2_1,k2_2:v2_2,],k3:v3,",
		},
		{
			name:        "Execution error",
			templateStr: "Hello, {{ .Wrong }}",
			setupFunc: func(sm *serviceMocks.MockTemplateService) parameters.Parameters {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				return nil
			},
			expectErr: true,
			errMsg:    "can't evaluate field Wrong in type *template.Data",
		},
		{
			name:        "Config error",
			templateStr: "Hello",
			setupFunc: func(sm *serviceMocks.MockTemplateService) parameters.Parameters {
				sm.On("EnvironmentConfigPaths").Return(nil)
				return nil
			},
			expectErr: true,
			errMsg:    "configs are not set",
		},
		{
			name:        "Parsing error",
			templateStr: "Hello, {{ .Wrong",
			expectErr:   true,
			errMsg:      "unclosed action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			serviceMock := serviceMocks.NewMockTemplateService(t)
			var params parameters.Parameters

			if tt.setupFunc != nil {
				params = tt.setupFunc(serviceMock)
			}

			nativeTmpl := &nativeTemplate{
				fnd:             fndMock,
				service:         serviceMock,
				serverTemplates: make(templates.Templates),
			}
			buffer := &bytes.Buffer{}
			err := nativeTmpl.RenderToWriter(tt.templateStr, params, buffer)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, buffer.String())
			}
		})
	}
}
