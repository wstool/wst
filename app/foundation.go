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
	"os/user"
)

type Foundation interface {
	Logger() *zap.SugaredLogger
	Fs() afero.Fs
	CurrentUser() (*user.User, error)
	DryRun() bool
	User(username string) (*user.User, error)
	UserGroup(u *user.User) (*user.Group, error)
	UserHomeDir() (string, error)
	LookupEnvVar(key string) (string, bool)
}

type DefaultFoundation struct {
	logger *zap.SugaredLogger
	fs     afero.Fs
	dryRun bool
}

func CreateFoundation(logger *zap.SugaredLogger, fs afero.Fs, dryRun bool) Foundation {
	return &DefaultFoundation{
		logger: logger,
		fs:     fs,
		dryRun: dryRun,
	}
}

func (f *DefaultFoundation) DryRun() bool {
	return f.dryRun
}

func (f *DefaultFoundation) Logger() *zap.SugaredLogger {
	return f.logger
}

func (f *DefaultFoundation) Fs() afero.Fs {
	return f.fs
}

func (f *DefaultFoundation) CurrentUser() (*user.User, error) {
	return user.Current()
}

func (f *DefaultFoundation) User(username string) (*user.User, error) {
	return user.Lookup(username)
}

func (f *DefaultFoundation) UserGroup(u *user.User) (*user.Group, error) {
	return user.LookupGroupId(u.Gid)
}

func (f *DefaultFoundation) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (f *DefaultFoundation) LookupEnvVar(key string) (string, bool) {
	return os.LookupEnv(key)
}
