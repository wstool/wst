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

package app

import (
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"os"
)

type Foundation interface {
	Logger() *zap.SugaredLogger
	Fs() afero.Fs
	UserHomeDir() (string, error)
	LookupEnvVar(key string) (string, bool)
}

type DefaultEnv struct {
	logger *zap.SugaredLogger
	fs     afero.Fs
}

func CreateEnv(logger *zap.SugaredLogger, fs afero.Fs) Foundation {
	return &DefaultEnv{
		logger: logger,
		fs:     fs,
	}
}

func (e *DefaultEnv) Logger() *zap.SugaredLogger {
	return e.logger
}

func (e *DefaultEnv) Fs() afero.Fs {
	return e.fs
}

func (e *DefaultEnv) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (e *DefaultEnv) LookupEnvVar(key string) (string, bool) {
	return os.LookupEnv(key)
}
