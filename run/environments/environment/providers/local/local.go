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
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/environments/task"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
)

type Maker interface {
	Make(config *types.LocalEnvironment, instanceWorkspace string) (environment.Environment, error)
}

type localMaker struct {
	*environment.CommonMaker
}

func CreateMaker(fnd app.Foundation) Maker {
	return &localMaker{
		CommonMaker: environment.CreateCommonMaker(fnd),
	}
}

func (m *localMaker) Make(
	config *types.LocalEnvironment,
	instanceWorkspace string,
) (environment.Environment, error) {
	return &localEnvironment{
		CommonEnvironment: *m.MakeCommonEnvironment(&types.CommonEnvironment{
			Ports: config.Ports,
		}),
		workspace:   filepath.Join(instanceWorkspace, "envs", "local"),
		initialized: false,
		tasks:       make(map[string]*localTask),
	}, nil
}

type localEnvironment struct {
	environment.CommonEnvironment
	workspace   string
	initialized bool
	tasks       map[string]*localTask
}

func (l *localEnvironment) RootPath(workspace string) string {
	return workspace
}

func (l *localEnvironment) Init(ctx context.Context) error {
	fs := l.Fnd.Fs()
	err := fs.MkdirAll(l.workspace, 0755)
	if err != nil {
		return errors.Errorf("Creating workspace directory for local env failed: %v", err)
	}
	l.initialized = true

	return nil
}

func (l *localEnvironment) Destroy(ctx context.Context) error {
	hasError := false
	for _, t := range l.tasks {
		if t.cmd.IsRunning() {
			if err := t.cmd.ProcessSignal(os.Kill); err != nil {
				l.Fnd.Logger().Errorf("Failed to kill process: %v", err)
				hasError = true
			}
			// Ignore errors for now as we do not want to log error on EOF closing
			_ = t.outputCollector.Close()
		}
	}

	fs := l.Fnd.Fs()

	if err := fs.RemoveAll(l.workspace); err != nil {
		return err
	}

	if hasError {
		return errors.New("failed to kill local environment tasks")
	}

	return nil
}

func (l *localEnvironment) RunTask(ctx context.Context, ss *environment.ServiceSettings, cmd *environment.Command) (task.Task, error) {
	if !l.initialized {
		err := l.Init(ctx)
		if err != nil {
			return nil, err
		}
	}

	command := l.Fnd.ExecCommand(ctx, cmd.Name, cmd.Args)

	outputCollector := l.OutputMaker.MakeCollector()
	command.SetStdout(outputCollector.StdoutWriter())
	command.SetStderr(outputCollector.StderrWriter())

	if err := command.Start(); err != nil {
		return nil, err
	}

	t := &localTask{
		id:              l.Fnd.GenerateUuid(),
		cmd:             command,
		outputCollector: outputCollector,
		executable:      cmd.Name,
		serviceName:     ss.Name,
		serviceUrl:      fmt.Sprintf("http://localhost:%d", ss.Port),
	}
	l.tasks[t.id] = t

	return t, nil
}

func convertTask(target task.Task) (*localTask, error) {
	if target == nil {
		return nil, fmt.Errorf("target task is not set")
	}
	if target.Type() != providers.LocalType {
		return nil, fmt.Errorf("local environment can process only local task")
	}
	t, ok := target.(*localTask)
	if !ok {
		// this should not happen
		return nil, fmt.Errorf("target task is not of type *localTask")
	}
	return t, nil
}

func (l *localEnvironment) ExecTaskCommand(ctx context.Context, ss *environment.ServiceSettings, target task.Task, cmd *environment.Command) error {
	_, err := convertTask(target)
	if err != nil {
		return err
	}

	return l.Fnd.ExecCommand(ctx, cmd.Name, cmd.Args).Run()
}

func (l *localEnvironment) ExecTaskSignal(ctx context.Context, ss *environment.ServiceSettings, target task.Task, signal os.Signal) error {
	t, err := convertTask(target)
	if err != nil {
		return err
	}

	err = t.cmd.ProcessSignal(signal)
	if err != nil {
		return err
	}

	return nil
}

func (l *localEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	t, err := convertTask(target)
	if err != nil {
		return nil, err
	}

	switch outputType {
	case output.Stdout:
		return t.outputCollector.StdoutReader(ctx), nil
	case output.Stderr:
		return t.outputCollector.StderrReader(ctx), nil
	case output.Any:
		return t.outputCollector.AnyReader(ctx), nil
	default:
		return nil, fmt.Errorf("unsupported output type")
	}
}

type localTask struct {
	id              string
	cmd             app.Command
	outputCollector output.Collector
	executable      string
	serviceName     string
	serviceUrl      string
}

func (t *localTask) Pid() int {
	return t.cmd.ProcessPid()
}

func (t *localTask) Id() string {
	return t.id
}

func (t *localTask) Executable() string {
	return t.executable
}

func (t *localTask) Name() string {
	return t.serviceName
}

func (t *localTask) PublicUrl() string {
	return t.serviceUrl
}

func (t *localTask) PrivateUrl() string {
	return t.serviceUrl
}

func (t *localTask) Type() providers.Type {
	return providers.LocalType
}
