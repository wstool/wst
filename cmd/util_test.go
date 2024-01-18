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
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"reflect"
	"testing"
)

type MockEnv struct {
	MockLogger  *MockLogger
	EnvVars     map[string]string
	MemFs       afero.Fs
	UserHomeDir string
}

func (e *MockEnv) Logger() *zap.SugaredLogger {
	return e.MockLogger.SugaredLogger
}

func (e *MockEnv) LookupEnvVar(key string) (string, bool) {
	val, ok := e.EnvVars[key]
	return val, ok
}

func (e *MockEnv) Fs() afero.Fs {
	return e.MemFs
}

func (e *MockEnv) GetUserHomeDir() (string, error) {
	return e.UserHomeDir, nil
}

type MockLogger struct {
	SugaredLogger *zap.SugaredLogger
	ObservedLogs  *observer.ObservedLogs
}

func NewMockLogger() *MockLogger {
	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	return &MockLogger{
		SugaredLogger: logger.Sugar(),
		ObservedLogs:  observedLogs,
	}
}

func (l *MockLogger) Messages() []string {
	var messages []string
	for _, log := range l.ObservedLogs.All() {
		messages = append(messages, log.Message)
	}
	return messages
}

func Test_getOverwrites(t *testing.T) {
	type args struct {
		overwriteValues []string
		noEnvs          bool
		env             app.Env
	}
	tests := []struct {
		name      string
		args      args
		envVars   map[string]string
		want      map[string]string
		wantWarns []string
	}{
		{
			name: "no overwrite values, no environment variables",
			args: args{
				overwriteValues: []string{},
				noEnvs:          true,
			},
			envVars:   map[string]string{},
			want:      map[string]string{},
			wantWarns: nil,
		},
		{
			name: "one valid overwrite value, no environment variables",
			args: args{
				overwriteValues: []string{"key=value"},
				noEnvs:          true,
			},
			envVars:   map[string]string{},
			want:      map[string]string{"key": "value"},
			wantWarns: nil,
		},
		{
			name: "one invalid overwrite value, no environment variables",
			args: args{
				overwriteValues: []string{"keyvalue"},
				noEnvs:          true,
			},
			envVars:   map[string]string{},
			want:      map[string]string{},
			wantWarns: []string{"Invalid key-value pair: keyvalue"},
		},
		{
			name: "one valid and one invalid overwrite value, no environment variables",
			args: args{
				overwriteValues: []string{"key=value", "keyvalue"},
				noEnvs:          true,
			},
			envVars:   map[string]string{},
			want:      map[string]string{"key": "value"},
			wantWarns: []string{"Invalid key-value pair: keyvalue"},
		},
		{
			name: "environment variables considered",
			args: args{
				overwriteValues: []string{"key=value"},
				noEnvs:          false,
			},
			envVars:   map[string]string{"WST_OVERWRITE": "key2=value2"},
			want:      map[string]string{"key": "value", "key2": "value2"},
			wantWarns: nil,
		},
		{
			name: "environment variables with multiple key values",
			args: args{
				overwriteValues: []string{"key=value"},
				noEnvs:          false,
			},
			envVars:   map[string]string{"WST_OVERWRITE": "key2=value2,key3=value3,key4=value4"},
			want:      map[string]string{"key": "value", "key2": "value2", "key3": "value3", "key4": "value4"},
			wantWarns: nil,
		},
		{
			name: "environment variables with invalid pairs considered",
			args: args{
				overwriteValues: []string{"key=value"},
				noEnvs:          false,
			},
			envVars:   map[string]string{"WST_OVERWRITE": "key2value2"},
			want:      map[string]string{"key": "value"},
			wantWarns: []string{"Invalid environment key-value pair: key2value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockLogger := NewMockLogger()

			mockEnv := &MockEnv{
				MockLogger:  mockLogger,
				EnvVars:     tt.envVars,
				MemFs:       afero.NewMemMapFs(),
				UserHomeDir: "/home/mockuser",
			}

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
