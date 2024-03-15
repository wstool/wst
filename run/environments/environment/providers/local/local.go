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
	"github.com/bukka/wst/run/instances"
	"github.com/bukka/wst/run/services"
	"io"
	"os"
	"os/exec"
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
	instanceWorkspace string,
) (environment.Environment, error) {
	return &localEnvironment{
		fnd:         m.fnd,
		portsStart:  config.Ports.From,
		portsEnd:    config.Ports.To,
		workspace:   filepath.Join(instanceWorkspace, "envs", "local"),
		initialized: false,
	}, nil
}

type localEnvironment struct {
	fnd         app.Foundation
	portsStart  int16
	portsEnd    int16
	instance    instances.Instance
	workspace   string
	initialized bool
}

func (l *localEnvironment) RootPath(service services.Service) string {
	return service.Workspace()
}

func (l *localEnvironment) Init(ctx context.Context) error {
	fs := l.fnd.Fs()
	err := fs.MkdirAll(l.workspace, 0644)
	if err != nil {
		return err
	}
	l.initialized = true

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
	if !l.initialized {
		err := l.Init(ctx)
		if err != nil {
			return nil, err
		}
	}

	command := exec.CommandContext(ctx, cmd.Name, cmd.Args...)

	if err := command.Start(); err != nil {
		return nil, err
	}

	return &localTask{
		cmd:         command,
		serviceName: service.Name(),
	}, nil
}

func convertTask(target task.Task) (*localTask, error) {
	if target == nil {
		return nil, fmt.Errorf("target task is not set")
	}
	if target.Type() != providers.LocalType {
		return nil, fmt.Errorf("local environment can process only local task")
	}
	localTask, ok := target.(*localTask)
	if !ok {
		// this should not happen
		return nil, fmt.Errorf("target task is not of type *localTask")
	}
	return localTask, nil
}

func (l *localEnvironment) ExecTaskCommand(ctx context.Context, service services.Service, target task.Task, cmd *environment.Command) error {
	localTask, err := convertTask(target)
	if err != nil {
		return err
	}

	// Render with pid
	_ = localTask.cmd.Process.Pid

	return exec.CommandContext(ctx, cmd.Name, cmd.Args...).Run()
}

func (l *localEnvironment) ExecTaskSignal(ctx context.Context, service services.Service, target task.Task, signal os.Signal) error {
	localTask, err := convertTask(target)
	if err != nil {
		return err
	}

	err = localTask.cmd.Process.Signal(signal)
	if err != nil {
		return err
	}

	return nil
}

func (l *localEnvironment) Output(ctx context.Context, target task.Task, outputType output.Type) (io.Reader, error) {
	localTask, err := convertTask(target)
	if err != nil {
		return nil, err
	}

	outputCollector := NewBufferedOutputCollector()
	switch outputType {
	case output.Stdout:
		stdoutPipe, err := localTask.cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		outputCollector.collectOutput(stdoutPipe)
	case output.Stderr:
		stderrPipe, err := localTask.cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		outputCollector.collectOutput(stderrPipe)
	case output.Any:
		stdoutPipe, err := localTask.cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		stderrPipe, err := localTask.cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		outputCollector.collectOutput(stdoutPipe, stderrPipe)
	default:
		return nil, fmt.Errorf("unsupported output type")
	}

	return outputCollector, nil
}

type localTask struct {
	cmd         *exec.Cmd
	serviceName string
}

func (t *localTask) Id() string {
	return t.serviceName
}

func (t *localTask) Name() string {
	return t.serviceName
}

func (t *localTask) BaseUrl() string {
	return ""
}

func (t *localTask) Type() providers.Type {
	return providers.LocalType
}
