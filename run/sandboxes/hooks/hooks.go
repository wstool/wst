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

package hooks

import (
	"context"
	"errors"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment"
	"github.com/bukka/wst/run/environments/task"
	"github.com/bukka/wst/run/services"
	"os"
	"syscall"
)

type HookType string

const (
	StartHookType   HookType = "start"
	StopHookType             = "stop"
	ReloadHookType           = "reload"
	RestartHookType          = "restart"
)

type Hooks map[HookType]Hook

type Maker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd: fnd,
	}
}

func (m *Maker) MakeHooks(config map[string]types.SandboxHook) (Hooks, error) {
	hooks := make(Hooks)
	for configHookTypeStr, hookConfig := range config {
		configHookType := types.SandboxHookType(configHookTypeStr)
		var hookType HookType
		switch configHookType {
		case types.StartSandboxHookType:
			hookType = StartHookType
		case types.StopSandboxHookType:
			hookType = StopHookType
		case types.RestartSandboxHookType:
			hookType = RestartHookType
		case types.ReloadSandboxHookType:
			hookType = ReloadHookType
		default:
			return nil, fmt.Errorf("invalid hook type %s", configHookTypeStr)
		}
		hook, err := m.MakeHook(hookConfig, hookType)
		if err != nil {
			return nil, err
		}
		hooks[hookType] = hook
	}
	return hooks, nil
}

// Define mapping for string signal names to os.Signal
var stringToSignalMap = map[string]os.Signal{
	"SIGTERM": syscall.SIGTERM,
	"SIGKILL": syscall.SIGKILL,
	"SIGINT":  syscall.SIGINT,
	"SIGQUIT": syscall.SIGQUIT,
	"SIGHUP":  syscall.SIGHUP,
	"SIGUSR1": syscall.SIGUSR1,
	"SIGUSR2": syscall.SIGUSR2,
}

// Define mapping for integer signal numbers to os.Signal
// This is more straightforward since os.Signal is an alias for syscall.Signal, which is int
var intToSignalMap = map[int]os.Signal{
	int(syscall.SIGTERM): syscall.SIGTERM,
	int(syscall.SIGKILL): syscall.SIGKILL,
	int(syscall.SIGINT):  syscall.SIGINT,
	int(syscall.SIGQUIT): syscall.SIGQUIT,
	int(syscall.SIGHUP):  syscall.SIGHUP,
	int(syscall.SIGUSR1): syscall.SIGUSR1,
	int(syscall.SIGUSR2): syscall.SIGUSR2,
}

func createBaseHook(enabled bool, hookType HookType) *BaseHook {
	return &BaseHook{Enabled: enabled, Type: hookType}
}

func (m *Maker) MakeHook(config types.SandboxHook, hookType HookType) (Hook, error) {
	var resultHook Hook
	switch hook := config.(type) {
	case *types.SandboxHookNative:
		resultHook = &HookNative{
			BaseHook: *createBaseHook(hook.Enabled, hookType),
		}
	case *types.SandboxHookShellCommand:
		resultHook = &HookShellCommand{
			BaseHook: *createBaseHook(true, hookType),
			Command:  hook.Command,
			Shell:    hook.Shell,
		}
	case *types.SandboxHookArgsCommand:
		resultHook = &HookArgsCommand{
			BaseHook:   *createBaseHook(true, hookType),
			Executable: hook.Executable,
			Args:       hook.Args,
		}
	case *types.SandboxHookSignal:
		baseHook := createBaseHook(true, hookType)
		if hook.IsString {
			signal, ok := stringToSignalMap[hook.StringValue]
			if !ok {
				return nil, errors.New("unsupported string signal value")
			}
			resultHook = &HookSignal{BaseHook: *baseHook, Signal: signal}
		} else {
			signal, ok := intToSignalMap[hook.IntValue]
			if !ok {
				return nil, errors.New("unsupported int signal value")
			}
			resultHook = &HookSignal{BaseHook: *baseHook, Signal: signal}
		}
	default:
		return nil, errors.New("unsupported hook type")
	}

	return resultHook, nil
}

type Hook interface {
	Execute(ctx context.Context, service services.Service) (task.Task, error)
}

type BaseHook struct {
	Type    HookType
	Enabled bool
}

type HookNative struct {
	BaseHook
}

func (h *HookNative) Execute(ctx context.Context, service services.Service) (task.Task, error) {
	serviceTask := service.Task()
	var err error
	if h.Type == StartHookType {
		if serviceTask != nil {
			return nil, errors.New("task has already been created which is likely because start already done")
		}
		serviceTask, err = service.Environment().RunTask(ctx, service, nil)
	} else {
		if serviceTask == nil {
			return nil, errors.New("task has not been created which is likely because start is not done")
		}
	}

	if err != nil {
		return nil, err
	}

	return serviceTask, nil
}

type HookArgsCommand struct {
	BaseHook
	Executable string
	Args       []string
}

func (h *HookArgsCommand) Execute(ctx context.Context, service services.Service) (task.Task, error) {
	serviceTask := service.Task()
	var err error
	if h.Type == StartHookType {
		if serviceTask != nil {
			return nil, errors.New("task has already been created which is likely because start already done")
		}
		serviceTask, err = service.Environment().RunTask(ctx, service, &environment.Command{
			Name: h.Executable,
			Args: h.Args,
		})
	} else {
		if serviceTask == nil {
			return nil, errors.New("task has not been created which is likely because start is not done")
		}
		err = service.Environment().ExecTaskCommand(ctx, service, serviceTask, &environment.Command{
			Name: h.Executable,
			Args: h.Args,
		})
	}

	if err != nil {
		return nil, err
	}

	return serviceTask, nil
}

type HookShellCommand struct {
	BaseHook
	Command string
	Shell   string
}

func (h *HookShellCommand) Execute(ctx context.Context, service services.Service) (task.Task, error) {
	argsCommand := &HookArgsCommand{
		BaseHook:   h.BaseHook,
		Executable: h.Shell,
		Args:       []string{"-c", h.Command},
	}
	return argsCommand.Execute(ctx, service)
}

type HookSignal struct {
	BaseHook
	Signal os.Signal
}

func (h *HookSignal) Execute(ctx context.Context, service services.Service) (task.Task, error) {
	if h.Type == StartHookType {
		return nil, fmt.Errorf("signal hook cannot be executed on start")
	}
	task := service.Task()
	if task == nil {
		return nil, errors.New("task does not exist for signal hook to execute")
	}

	err := service.Environment().ExecTaskSignal(ctx, service, task, h.Signal)
	if err != nil {
		return nil, err
	}

	return task, nil
}
