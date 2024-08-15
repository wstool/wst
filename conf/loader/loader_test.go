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

package loader

import (
	"github.com/bukka/wst/app"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfigLoader_LoadConfig(t *testing.T) {
	// Create and setup an in-memory file system
	mockFs := afero.NewMemMapFs()
	err := afero.WriteFile(mockFs, "/test.json", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/test-invalid.json", []byte(`{"key":`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/test.yaml", []byte("key: value"), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/test.toml", []byte("key = \"value\""), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/test.unknown", []byte("key = \"value\""), 0644)
	assert.NoError(t, err)

	// mock app.Foundation
	MockFnd := &appMocks.MockFoundation{}
	MockFnd.On("Fs").Return(mockFs)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fnd     app.Foundation
		args    args
		want    LoadedConfig
		wantErr bool
	}{
		{
			name:    "Testing LoadConfig - JSON",
			fnd:     MockFnd,
			args:    args{path: "/test.json"},
			want:    LoadedConfigData{path: "/test.json", data: map[string]interface{}{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "Testing LoadConfig - YAML",
			fnd:     MockFnd,
			args:    args{path: "/test.yaml"},
			want:    LoadedConfigData{path: "/test.yaml", data: map[string]interface{}{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "Testing LoadConfig - TOML",
			fnd:     MockFnd,
			args:    args{path: "/test.toml"},
			want:    LoadedConfigData{path: "/test.toml", data: map[string]interface{}{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "Testing LoadConfig - Unsupported file type",
			fnd:     MockFnd,
			args:    args{path: "/test.unknown"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Testing LoadConfig - Not found file type",
			fnd:     MockFnd,
			args:    args{path: "/testx.json"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Testing LoadConfig - Invalid JSON",
			fnd:     MockFnd,
			args:    args{path: "/test-invalid.json"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := ConfigLoader{fnd: tt.fnd}
			got, err := l.LoadConfig(tt.args.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfigLoader_LoadConfigs(t *testing.T) {
	// Create and setup an in-memory file system
	mockFs := afero.NewMemMapFs()
	err := afero.WriteFile(mockFs, "/test.json", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/test2.json", []byte(`{"key": "value2"}`), 0644)
	assert.NoError(t, err)

	// mock app.Foundation
	MockFnd := &appMocks.MockFoundation{}
	MockFnd.On("Fs").Return(mockFs)

	type args struct {
		paths []string
	}
	tests := []struct {
		name    string
		fnd     app.Foundation
		args    args
		want    []LoadedConfig
		wantErr bool
	}{
		{
			name: "Testing LoadConfigs",
			fnd:  MockFnd,
			args: args{paths: []string{"/test.json", "/test2.json"}},
			want: []LoadedConfig{
				LoadedConfigData{
					path: "/test.json",
					data: map[string]interface{}{"key": "value"},
				},
				LoadedConfigData{
					path: "/test2.json",
					data: map[string]interface{}{"key": "value2"},
				},
			},
			wantErr: false,
		},
		{
			name:    "Testing LoadConfigs - Error case - Non existent file",
			fnd:     MockFnd,
			args:    args{paths: []string{"non_existent_file.json"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := ConfigLoader{
				fnd: tt.fnd,
			}
			got, err := l.LoadConfigs(tt.args.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfigLoader_GlobConfigs(t *testing.T) {
	// Create and setup an in-memory file system
	mockFs := afero.NewMemMapFs()
	err := afero.WriteFile(mockFs, "/dir/test.json", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/dir/test2.json", []byte(`{"key": "value2"}`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "dir/test.json", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "dir/test2.json", []byte(`{"key": "value2"}`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/var/wst/dir/test.json", []byte(`{"key": "value"}`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(mockFs, "/var/wst/dir/test2.json", []byte(`{"key": "value2"}`), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name             string
		fnd              app.Foundation
		pattern          string
		wd               string
		setupMocks       func(fnd *appMocks.MockFoundation)
		want             []LoadedConfig
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:    "success on return of absolute paths",
			pattern: "/dir/*.json",
			wd:      "/var/wst",
			setupMocks: func(fnd *appMocks.MockFoundation) {
				fnd.On("Getwd").Return("/home", nil)
				fnd.On("Chdir", "/var/wst").Return(nil)
				fnd.On("Chdir", "/home").Return(nil)
			},
			want: []LoadedConfig{
				LoadedConfigData{
					path: "/dir/test.json",
					data: map[string]interface{}{"key": "value"},
				},
				LoadedConfigData{
					path: "/dir/test2.json",
					data: map[string]interface{}{"key": "value2"},
				},
			},
		},
		{
			name:    "success on return of relative paths",
			pattern: "dir/*.json",
			wd:      "/var/wst",
			setupMocks: func(fnd *appMocks.MockFoundation) {
				fnd.On("Getwd").Return("/home", nil)
				fnd.On("Chdir", "/var/wst").Return(nil)
				fnd.On("Chdir", "/home").Return(nil)
			},
			want: []LoadedConfig{
				LoadedConfigData{
					path: "/var/wst/dir/test.json",
					data: map[string]interface{}{"key": "value"},
				},
				LoadedConfigData{
					path: "/var/wst/dir/test2.json",
					data: map[string]interface{}{"key": "value2"},
				},
			},
		},
		{
			name:    "success on no path found",
			pattern: "/dir/non_matching_pattern.json",
			wd:      "/var/wst",
			setupMocks: func(fnd *appMocks.MockFoundation) {
				fnd.On("Getwd").Return("/home", nil)
				fnd.On("Chdir", "/var/wst").Return(nil)
				fnd.On("Chdir", "/home").Return(nil)
			},
			want: []LoadedConfig{},
		},
		{
			name:    "error on invalid pattern",
			pattern: "/dir/test[.yaml",
			wd:      "/var/wst",
			setupMocks: func(fnd *appMocks.MockFoundation) {
				fnd.On("Getwd").Return("/home", nil)
				fnd.On("Chdir", "/var/wst").Return(nil)
				fnd.On("Chdir", "/home").Return(nil)
			},
			expectError:      true,
			expectedErrorMsg: "syntax error in pattern",
		},
		{
			name:    "error on changing directory",
			pattern: "/dir/test[.yaml",
			wd:      "/var/wst",
			setupMocks: func(fnd *appMocks.MockFoundation) {
				fnd.On("Getwd").Return("/home", nil)
				fnd.On("Chdir", "/var/wst").Return(errors.New("chdir fail"))
			},
			expectError:      true,
			expectedErrorMsg: "chdir fail",
		},
		{
			name:    "error on getting working directory directory",
			pattern: "/dir/test[.yaml",
			wd:      "/var/wst",
			setupMocks: func(fnd *appMocks.MockFoundation) {
				fnd.On("Getwd").Return("/home", errors.New("wd fail"))
			},
			expectError:      true,
			expectedErrorMsg: "wd fail",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			fndMock.On("Fs").Return(mockFs)

			l := ConfigLoader{
				fnd: fndMock,
			}
			tt.setupMocks(fndMock)
			got, err := l.GlobConfigs(tt.pattern, tt.wd)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCreateLoader(t *testing.T) {
	// Create and setup an in-memory file system
	mockFs := afero.NewMemMapFs()

	// mock app.Foundation
	MockFnd := &appMocks.MockFoundation{}
	MockFnd.On("Fs").Return(mockFs)

	tests := []struct {
		name string
		fnd  app.Foundation
		want Loader
	}{
		{
			name: "Testing CreateLoader",
			fnd:  MockFnd,
			want: &ConfigLoader{fnd: MockFnd},
		},
		// TODO: Add more test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateLoader(tt.fnd)
			// Here we use testify library's require package to compare struct pointers by their values
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLoadedConfigData_Path(t *testing.T) {
	type fields struct {
		path string
		data map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test path",
			fields: fields{
				path: "/var/test",
				data: nil,
			},
			want: "/var/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := LoadedConfigData{
				path: tt.fields.path,
				data: tt.fields.data,
			}
			assert.Equalf(t, tt.want, d.Path(), "Path()")
		})
	}
}

func TestLoadedConfigData_Data(t *testing.T) {
	type fields struct {
		path string
		data map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{
			name: "test path",
			fields: fields{
				path: "/var/test",
				data: map[string]interface{}{"test": "val"},
			},
			want: map[string]interface{}{"test": "val"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := LoadedConfigData{
				path: tt.fields.path,
				data: tt.fields.data,
			}
			assert.Equalf(t, tt.want, d.Data(), "Data()")
		})
	}
}
