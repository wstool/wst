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
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
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
		setupFunc   func(
			fm *appMocks.MockFoundation,
			sm *serviceMocks.MockTemplateService,
			st templates.Templates,
		) (parameters.Parameters, Services)
		expectErr bool
		errMsg    string
		expected  string
	}{
		{
			name:        "Valid template with string param",
			templateStr: "Hello, {{.Parameters.GetString \"key\"}}!",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				pm := parameterMocks.NewMockParameter(t)
				pm.On("StringValue").Return("World", nil)
				return parameters.Parameters{
					"key": pm,
				}, nil
			},
			expected: "Hello, World!",
		},
		{
			name:        "Valid template with nested string params",
			templateStr: "{{ .Parameters.GetObjectString \"k1\" \"k1_1\" }}",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				pm1_1 := parameterMocks.NewMockParameter(t)
				pm1_1.On("StringValue").Return("v1_1", nil)
				pm1 := parameterMocks.NewMockParameter(t)
				pm1.On("MapValue").Return(map[string]parameter.Parameter{
					"k1_1": pm1_1,
				})
				return parameters.Parameters{
					"k1": pm1,
				}, nil
			},
			expected: "v1_1",
		},
		{
			name: "Valid template with nested object params iteration",
			templateStr: "{{ $obj := .Parameters.GetObject \"k2\" }}{{ $obj.GetString \"k2_2\" }};" +
				"{{ range $k, $v := .Parameters }}{{ $k }}:{{ if $v.IsObject }}[" +
				"{{ range $ik, $iv := $v.ToObject }}{{ $ik }}:{{ $iv.ToString }},{{ end }}" +
				"]{{ else }}{{ $v.ToString }}{{ end }},{{ end }}",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
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
				}, nil
			},
			expected: "v2_2;k1:v1,k2:[k2_1:v2_1,k2_2:v2_2,],k3:v3,",
		},
		{
			name:        "Valid template with services find",
			templateStr: "{{ (.Services.Find \"s1\").Pid }}",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("Pid").Return(12, nil)
				sm.On("EnvironmentConfigPaths").Return(map[string]string{
					"p1": "/config/path1",
					"p2": "/config/path2",
				})
				svcs := Services{
					"s1": sm,
					"s2": serviceMocks.NewMockTemplateService(t),
				}
				return parameters.Parameters{
					"k": parameterMocks.NewMockParameter(t),
				}, svcs
			},
			expected: "12",
		},
		{
			name: "Valid template with service values",
			templateStr: "{{ .Service.PrivateUrl }};{{ .Service.Pid }};{{ .Service.User }};{{ .Service.Group }};" +
				"conf:{{ .Service.ConfDir }},run:{{ .Service.RunDir }},script:{{ .Service.ScriptDir }};" +
				"{{ range $k, $v := .Service.EnvironmentConfigPaths }}{{ $k }}:{{ $v }},{{ end }};" +
				"{{ range $k, $v := .Configs }}{{ $k }}:{{ $v }},{{ end }}",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("PrivateUrl").Return("http://svc", nil)
				sm.On("Pid").Return(1234, nil)
				sm.On("User").Return("grp")
				sm.On("Group").Return("usr")
				sm.On("ConfDir").Return("/etc", nil)
				sm.On("RunDir").Return("/var/run", nil)
				sm.On("ScriptDir").Return("/var/www", nil)
				sm.On("EnvironmentConfigPaths").Return(map[string]string{
					"p1": "/var/p1",
					"p2": "/var/p2",
				})
				return parameters.Parameters{
					"k": parameterMocks.NewMockParameter(t),
				}, nil
			},
			expected: "http://svc;1234;grp;usr;conf:/etc,run:/var/run,script:/var/www;p1:/var/p1,p2:/var/p2,;p1:/var/p1,p2:/var/p2,",
		},
		{
			name:        "Valid template with includes",
			templateStr: "Hello, {{ include \"t.tpl\" . }}!",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				pm := parameterMocks.NewMockParameter(t)
				pm.On("StringValue").Return("pv", nil)

				memMapFs := afero.NewMemMapFs()
				filePath := "/var/www/t.tpl"
				fileContent := "pk:{{ .Parameters.GetString \"key\" }}"
				_ = afero.WriteFile(memMapFs, filePath, []byte(fileContent), 0644)
				fm.On("Fs").Return(memMapFs)

				tm := templatesMocks.NewMockTemplate(t)
				tm.On("FilePath").Return(filePath)

				svcs := Services{"svc": sm}
				st["t.tpl"] = tm

				return parameters.Parameters{
					"key": pm,
				}, svcs
			},
			expected: "Hello, pk:pv!",
		},
		{
			name:        "Valid template with nested includes",
			templateStr: "Hello, {{ include \"t1.tpl\" . }}!",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				pm := parameterMocks.NewMockParameter(t)
				pm.On("StringValue").Return("pv", nil)

				memMapFs := afero.NewMemMapFs()
				firstFilePath := "/var/www/t1.tpl"
				firstFileContent := "inc:{{ include \"t2.tpl\" .}}"
				secondFilePath := "/var/www/t2.tpl"
				secondFileContent := "pk:{{ .Parameters.GetString \"key\" }}"

				_ = afero.WriteFile(memMapFs, firstFilePath, []byte(firstFileContent), 0644)
				_ = afero.WriteFile(memMapFs, secondFilePath, []byte(secondFileContent), 0644)
				fm.On("Fs").Return(memMapFs)

				tm1 := templatesMocks.NewMockTemplate(t)
				tm1.On("FilePath").Return(firstFilePath)
				tm2 := templatesMocks.NewMockTemplate(t)
				tm2.On("FilePath").Return(secondFilePath)

				svcs := Services{"svc": sm}
				st["t1.tpl"] = tm1
				st["t2.tpl"] = tm2

				return parameters.Parameters{
					"key": pm,
				}, svcs
			},
			expected: "Hello, inc:pk:pv!",
		},
		{
			name:        "Error due to service private url",
			templateStr: "{{ .Service.PrivateUrl }};{{ .Service.Pid }}",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				sm.On("PrivateUrl").Return("", errors.New("priv url err"))
				return parameters.Parameters{
					"k": parameterMocks.NewMockParameter(t),
				}, nil
			},
			expectErr: true,
			errMsg:    "priv url err",
		},
		{
			name:        "Execution error",
			templateStr: "Hello, {{ .Wrong }}",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				return nil, nil
			},
			expectErr: true,
			errMsg:    "can't evaluate field Wrong in type *template.Data",
		},
		{
			name:        "Config error",
			templateStr: "Hello",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("EnvironmentConfigPaths").Return(nil)
				return nil, nil
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
			var svcs Services
			serverTemplates := make(templates.Templates)

			if tt.setupFunc != nil {
				params, svcs = tt.setupFunc(fndMock, serviceMock, serverTemplates)
			}

			nativeTmpl := &nativeTemplate{
				fnd:             fndMock,
				service:         serviceMock,
				services:        svcs,
				serverTemplates: serverTemplates,
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

func Test_nativeTemplate_RenderToFile(t *testing.T) {
	tests := []struct {
		name        string
		templateStr string
		filePath    string
		perm        os.FileMode
		setupFunc   func(
			fm *appMocks.MockFoundation,
			sm *serviceMocks.MockTemplateService,
			st templates.Templates,
		) (parameters.Parameters, Services)
		expectErr bool
		errMsg    string
		expected  string
	}{
		{
			name:        "Valid template with string param",
			templateStr: "Hello, {{.Parameters.GetString \"key\"}}!",
			filePath:    "/var/www/t.txt",
			perm:        0644,
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				fileMock := appMocks.NewMockFile(t)
				fileMock.On("Write", []byte("Hello, ")).Return(7, nil)
				fileMock.On("Write", []byte("World")).Return(5, nil)
				fileMock.On("Write", []byte("!")).Return(1, nil)
				fileMock.On("Close").Return(nil)

				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/var/www", fs.FileMode(0755)).Return(nil)
				fsMock.On("OpenFile", "/var/www/t.txt", os.O_RDWR|os.O_CREATE, fs.FileMode(0644)).Return(fileMock, nil)

				fm.On("Fs").Return(fsMock)

				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				pm := parameterMocks.NewMockParameter(t)
				pm.On("StringValue").Return("World", nil)
				return parameters.Parameters{
					"key": pm,
				}, nil
			},
			expected: "Hello, World!",
		},
		{
			name:        "Parsing error",
			templateStr: "{{ .Wrong",
			filePath:    "/var/www/t.txt",
			perm:        0644,
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				fileMock := appMocks.NewMockFile(t)
				fileMock.On("Close").Return(nil)

				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/var/www", fs.FileMode(0755)).Return(nil)
				fsMock.On("OpenFile", "/var/www/t.txt", os.O_RDWR|os.O_CREATE, fs.FileMode(0644)).Return(fileMock, nil)

				fm.On("Fs").Return(fsMock)
				return nil, nil
			},
			expectErr: true,
			errMsg:    "unclosed action",
		},
		{
			name:        "Open file error",
			templateStr: "content",
			filePath:    "/var/www/t.txt",
			perm:        0644,
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/var/www", fs.FileMode(0755)).Return(nil)
				fsMock.On("OpenFile", "/var/www/t.txt", os.O_RDWR|os.O_CREATE, fs.FileMode(0644)).Return(
					nil,
					errors.New("open fail"),
				)

				fm.On("Fs").Return(fsMock)
				return nil, nil
			},
			expectErr: true,
			errMsg:    "open fail",
		},
		{
			name:        "Make dir error",
			templateStr: "content",
			filePath:    "/var/www/t.txt",
			perm:        0644,
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				fsMock := appMocks.NewMockFs(t)
				fsMock.On("MkdirAll", "/var/www", fs.FileMode(0755)).Return(errors.New("mkdir fail"))

				fm.On("Fs").Return(fsMock)
				return nil, nil
			},
			expectErr: true,
			errMsg:    "mkdir fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			serviceMock := serviceMocks.NewMockTemplateService(t)
			var params parameters.Parameters
			var svcs Services
			serverTemplates := make(templates.Templates)

			if tt.setupFunc != nil {
				params, svcs = tt.setupFunc(fndMock, serviceMock, serverTemplates)
			}

			nativeTmpl := &nativeTemplate{
				fnd:             fndMock,
				service:         serviceMock,
				services:        svcs,
				serverTemplates: serverTemplates,
			}
			err := nativeTmpl.RenderToFile(tt.templateStr, params, tt.filePath, tt.perm)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_nativeTemplate_RenderToString(t *testing.T) {
	tests := []struct {
		name        string
		templateStr string
		setupFunc   func(
			fm *appMocks.MockFoundation,
			sm *serviceMocks.MockTemplateService,
			st templates.Templates,
		) (parameters.Parameters, Services)
		expectErr bool
		errMsg    string
		expected  string
	}{
		{
			name:        "Valid template with string param",
			templateStr: "Hello, {{.Parameters.GetString \"key\"}}!",
			setupFunc: func(
				fm *appMocks.MockFoundation,
				sm *serviceMocks.MockTemplateService,
				st templates.Templates,
			) (parameters.Parameters, Services) {
				sm.On("EnvironmentConfigPaths").Return(map[string]string{"path1": "/config/path1"})
				pm := parameterMocks.NewMockParameter(t)
				pm.On("StringValue").Return("World", nil)
				return parameters.Parameters{
					"key": pm,
				}, nil
			},
			expected: "Hello, World!",
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
			var svcs Services
			serverTemplates := make(templates.Templates)

			if tt.setupFunc != nil {
				params, svcs = tt.setupFunc(fndMock, serviceMock, serverTemplates)
			}

			nativeTmpl := &nativeTemplate{
				fnd:             fndMock,
				service:         serviceMock,
				services:        svcs,
				serverTemplates: serverTemplates,
			}
			result, err := nativeTmpl.RenderToString(tt.templateStr, params)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
