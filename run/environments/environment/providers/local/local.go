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

package local

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/services"
	"github.com/bukka/wst/run/task"
	"io"
	"os"
	"path/filepath"
)

type Maker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd: fnd,
	}
}

func (m *Maker) Make(
	config *types.LocalEnvironment,
	service services.Service,
) (environment.Environment, error) {
	return &localEnvironment{
		fnd:        m.fnd,
		portsStart: config.Ports.From,
		portsEnd:   config.Ports.To,
		service:    service,
		workspace:  filepath.Join(service.Workspace(), "env"),
	}, nil
}

type localEnvironment struct {
	fnd        app.Foundation
	portsStart int16
	portsEnd   int16
	service    services.Service
	workspace  string
}

func (l *localEnvironment) Init(ctx context.Context) error {
	fs := l.fnd.Fs()
	err := fs.MkdirAll(l.workspace, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (l *localEnvironment) Destroy(ctx context.Context) error {
	fs := l.fnd.Fs()
	err := fs.RemoveAll(l.workspace)
	if err != nil {
		return err
	}

	return nil
}

func (l *localEnvironment) RunTask(ctx context.Context, service services.Service, cmd *environment.Command) (task.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (l *localEnvironment) ExecTaskCommand(ctx context.Context, service services.Service, target task.Task, cmd *environment.Command) error {
	//TODO implement me
	panic("implement me")
}

func (l *localEnvironment) ExecTaskSignal(ctx context.Context, service services.Service, target task.Task, signal os.Signal) error {
	//TODO implement me
	panic("implement me")
}

func (l *localEnvironment) Output(ctx context.Context, outputType output.Type) (io.Reader, error) {
	//TODO implement me
	panic("implement me")
}

type localTask struct {
	id int
}
