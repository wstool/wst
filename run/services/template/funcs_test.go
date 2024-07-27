package template

import (
	"fmt"
	"github.com/bukka/wst/app"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	templatesMocks "github.com/bukka/wst/mocks/generated/run/servers/templates"
	"github.com/bukka/wst/run/servers/templates"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_nativeTemplate_include(t *testing.T) {
	tests := []struct {
		name             string
		tmplName         string
		includedTmplName string
		setupMocks       func(*templatesMocks.MockTemplate)
		setupFs          func() app.Fs
		data             interface{}
		expected         string
		expectError      bool
		expectedErrMsg   string
	}{
		{
			name:     "Valid template execution",
			tmplName: "validTemplate",
			setupMocks: func(mt *templatesMocks.MockTemplate) {
				mt.On("FilePath").Return("/validTemplate.tpl")
			},
			setupFs: func() app.Fs {
				memMapFs := afero.NewMemMapFs()
				fileContent := "Hello, {{.Name}}!"
				_ = afero.WriteFile(memMapFs, "/validTemplate.tpl", []byte(fileContent), 0644)
				return memMapFs
			},
			data:     struct{ Name string }{Name: "World"},
			expected: "Hello, World!",
		},
		{
			name:             "Template not found",
			tmplName:         "nonExistentTemplate",
			includedTmplName: "missingFileTemplate",
			expectError:      true,
			expectedErrMsg:   "failed to find template",
		},
		{
			name:     "Template file open error",
			tmplName: "fileTemplate",
			setupMocks: func(mt *templatesMocks.MockTemplate) {
				mt.On("FilePath").Return("/missingFile.tpl")
			},
			setupFs: func() app.Fs {
				return afero.NewMemMapFs()
			},
			data:           nil,
			expectError:    true,
			expectedErrMsg: "failed to open template file",
		},
		{
			name:     "Template parsing error",
			tmplName: "invalidTemplate",
			setupMocks: func(mt *templatesMocks.MockTemplate) {
				mt.On("FilePath").Return("/invalidTemplate.tpl")
			},
			setupFs: func() app.Fs {
				memMapFs := afero.NewMemMapFs()
				_ = afero.WriteFile(memMapFs, "/invalidTemplate.tpl", []byte("{{ .Name }"), 0644)
				return memMapFs
			},
			data:           nil,
			expectError:    true,
			expectedErrMsg: "failed to parse template",
		},
		{
			name:     "Template read error",
			tmplName: "readErrorTemplate",
			setupMocks: func(mt *templatesMocks.MockTemplate) {
				mt.On("FilePath").Return("/readError.tpl")
			},
			setupFs: func() app.Fs {
				mockFile := appMocks.NewMockFile(t)
				mockFile.On("Read", mock.Anything).Return(0, fmt.Errorf("read error"))
				mockFile.On("Close").Return(nil)
				mockFs := appMocks.NewMockFs(t)
				mockFs.On("Open", "/readError.tpl").Return(mockFile, nil)
				return mockFs
			},
			data:           struct{}{},
			expectError:    true,
			expectedErrMsg: "failed to read template file",
		},
		{
			name:     "Template execution error",
			tmplName: "executionErrorTemplate",
			setupMocks: func(mt *templatesMocks.MockTemplate) {
				mt.On("FilePath").Return("/executionError.tpl")
			},
			setupFs: func() app.Fs {
				memMapFs := afero.NewMemMapFs()
				_ = afero.WriteFile(memMapFs, "/executionError.tpl", []byte("{{ .Undefined }}"), 0644)
				return memMapFs
			},
			data:           struct{}{},
			expectError:    true,
			expectedErrMsg: "failed to execute template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTemplate := templatesMocks.NewMockTemplate(t)
			fndMock := appMocks.NewMockFoundation(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mockTemplate)
			}
			if tt.setupFs != nil {
				fs := tt.setupFs()
				fndMock.On("Fs").Return(fs)
			}

			nativeTmpl := &nativeTemplate{
				fnd: fndMock,
				serverTemplates: templates.Templates{
					tt.tmplName: mockTemplate,
				},
			}

			includedTmplName := tt.includedTmplName
			if includedTmplName == "" {
				includedTmplName = tt.tmplName
			}
			result, err := nativeTmpl.include(includedTmplName, tt.data)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func Test_nativeTemplate_funcs(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)

	nativeTmpl := &nativeTemplate{
		fnd: fndMock,
	}

	funcMap := nativeTmpl.funcs()

	_, ok := funcMap["include"]
	require.True(t, ok, "include function should be present in funcMap")
}
