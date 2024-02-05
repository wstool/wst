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

package cmd

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/mocks/appMocks"
	"github.com/bukka/wst/mocks/externalMocks"
	"reflect"
	"testing"
)

func Test_getOverwrites(t *testing.T) {
	type args struct {
		overwriteValues []string
		noEnvs          bool
		env             app.Env
	}
	tests := []struct {
		name           string
		args           args
		overwrite      string
		overwriteFound bool
		want           map[string]string
		wantWarns      []string
	}{
		{
			name: "no overwrite values, no environment variables",
			args: args{
				overwriteValues: []string{},
				noEnvs:          true,
			},
			overwrite:      "",
			overwriteFound: false,
			want:           map[string]string{},
			wantWarns:      nil,
		},
		{
			name: "one valid overwrite value, no environment variables",
			args: args{
				overwriteValues: []string{"key=value"},
				noEnvs:          true,
			},
			overwrite:      "",
			overwriteFound: false,
			want:           map[string]string{"key": "value"},
			wantWarns:      nil,
		},
		{
			name: "one invalid overwrite value, no environment variables",
			args: args{
				overwriteValues: []string{"keyvalue"},
				noEnvs:          true,
			},
			overwrite: "",
			want:      map[string]string{},
			wantWarns: []string{"Invalid key-value pair: keyvalue"},
		},
		{
			name: "one valid and one invalid overwrite value, no environment variables",
			args: args{
				overwriteValues: []string{"key=value", "keyvalue"},
				noEnvs:          true,
			},
			overwrite:      "",
			overwriteFound: false,
			want:           map[string]string{"key": "value"},
			wantWarns:      []string{"Invalid key-value pair: keyvalue"},
		},
		{
			name: "environment variables considered",
			args: args{
				overwriteValues: []string{"key=value"},
				noEnvs:          false,
			},
			overwrite:      "key2=value2",
			overwriteFound: true,
			want:           map[string]string{"key": "value", "key2": "value2"},
			wantWarns:      nil,
		},
		{
			name: "environment variables with multiple key values",
			args: args{
				overwriteValues: []string{"key=value"},
				noEnvs:          false,
			},
			overwrite:      "key2=value2:key3=value3:key4=value4",
			overwriteFound: true,
			want:           map[string]string{"key": "value", "key2": "value2", "key3": "value3", "key4": "value4"},
			wantWarns:      nil,
		},
		{
			name: "environment variables with invalid pairs considered",
			args: args{
				overwriteValues: []string{"key=value"},
				noEnvs:          false,
			},
			overwrite:      "key2value2",
			overwriteFound: true,
			want:           map[string]string{"key": "value"},
			wantWarns:      []string{"Invalid environment key-value pair: key2value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockLogger := externalMocks.NewMockLogger()

			mockEnv := &appMocks.MockEnv{}
			mockEnv.On("Logger").Return(mockLogger.SugaredLogger)
			mockEnv.On("LookupEnvVar", "WST_OVERWRITE").Return(tt.overwrite, tt.overwriteFound)

			tt.args.env = mockEnv // Assign mockEnv to your test case

			// Run test case
			if got := getOverwrites(tt.args.overwriteValues, tt.args.noEnvs, tt.args.env); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOverwrites() = %v, want %v", got, tt.want)
			}

			// Validate log messages
			messages := mockLogger.Messages()
			if !reflect.DeepEqual(messages, tt.wantWarns) {
				t.Errorf("logger.Warn() calls = %v, want %v", mockLogger.Messages(), tt.wantWarns)
			}
		})
	}
}
