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

package conf

import (
	"github.com/bukka/wst/app"
	appMocks "github.com/bukka/wst/mocks/app"
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

	// mock app.Env
	mockEnv := &appMocks.MockEnv{}
	mockEnv.On("Fs").Return(mockFs)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		env     app.Env
		args    args
		want    LoadedConfig
		wantErr bool
	}{
		{
			name:    "Testing LoadConfig - JSON",
			env:     mockEnv,
			args:    args{path: "/test.json"},
			want:    LoadedConfig{Path: "/test.json", Data: map[string]interface{}{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "Testing LoadConfig - YAML",
			env:     mockEnv,
			args:    args{path: "/test.yaml"},
			want:    LoadedConfig{Path: "/test.yaml", Data: map[string]interface{}{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "Testing LoadConfig - TOML",
			env:     mockEnv,
			args:    args{path: "/test.toml"},
			want:    LoadedConfig{Path: "/test.toml", Data: map[string]interface{}{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "Testing LoadConfig - Unsupported file type",
			env:     mockEnv,
			args:    args{path: "/test.unknown"},
			want:    LoadedConfig{},
			wantErr: true,
		},
		{
			name:    "Testing LoadConfig - Invalid JSON",
			env:     mockEnv,
			args:    args{path: "/test-invalid.json"},
			want:    LoadedConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := ConfigLoader{env: tt.env}
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

	// mock app.Env
	mockEnv := &appMocks.MockEnv{}
	mockEnv.On("Fs").Return(mockFs)

	type args struct {
		paths []string
	}
	tests := []struct {
		name    string
		env     app.Env
		args    args
		want    []LoadedConfig
		wantErr bool
	}{
		{
			name: "Testing LoadConfigs",
			env:  mockEnv,
			args: args{paths: []string{"/test.json", "/test2.json"}},
			want: []LoadedConfig{
				{
					Path: "/test.json",
					Data: map[string]interface{}{"key": "value"},
				},
				{
					Path: "/test2.json",
					Data: map[string]interface{}{"key": "value2"},
				},
			},
			wantErr: false,
		},
		{
			name:    "Testing LoadConfigs - Error case - Non existent file",
			env:     mockEnv,
			args:    args{paths: []string{"non_existent_file.json"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := ConfigLoader{
				env: tt.env,
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

	// mock app.Env
	mockEnv := &appMocks.MockEnv{}
	mockEnv.On("Fs").Return(mockFs)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		env     app.Env
		args    args
		want    []LoadedConfig
		wantErr bool
	}{
		{
			name: "Testing GlobConfigs",
			env:  mockEnv,
			args: args{path: "/dir/*.json"},
			want: []LoadedConfig{
				{
					Path: "/dir/test.json",
					Data: map[string]interface{}{"key": "value"},
				},
				{
					Path: "/dir/test2.json",
					Data: map[string]interface{}{"key": "value2"},
				},
			},
			wantErr: false,
		},
		{
			name:    "Testing GlobConfigs - Error case - No Matching Files",
			env:     mockEnv,
			args:    args{path: "/dir/non_matching_pattern.json"},
			want:    []LoadedConfig{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := ConfigLoader{
				env: tt.env,
			}
			got, err := l.GlobConfigs(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GlobConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreateLoader(t *testing.T) {
	// Create and setup an in-memory file system
	mockFs := afero.NewMemMapFs()

	// mock app.Env
	mockEnv := &appMocks.MockEnv{}
	mockEnv.On("Fs").Return(mockFs)

	tests := []struct {
		name string
		env  app.Env
		want Loader
	}{
		{
			name: "Testing CreateLoader",
			env:  mockEnv,
			want: &ConfigLoader{env: mockEnv},
		},
		// TODO: Add more test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateLoader(tt.env)
			// Here we use testify library's require package to compare struct pointers by their values
			require.Equal(t, tt.want, got)
		})
	}
}
