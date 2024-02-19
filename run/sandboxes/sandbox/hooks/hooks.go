package hooks

import (
	"errors"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"os"
	"syscall"
)

type HookNative struct {
	Type string
}

func (h HookNative) Execute(sandbox *sandbox.Sandbox) error {
	//TODO implement me
	panic("implement me")
}

type HookShellCommand struct {
	Command string
	Shell   string
}

func (h HookShellCommand) Execute(sandbox *sandbox.Sandbox) error {
	//TODO implement me
	panic("implement me")
}

type HookCommand struct {
	Executable string
	Args       []string
}

func (h HookCommand) Execute(sandbox *sandbox.Sandbox) error {
	//TODO implement me
	panic("implement me")
}

type HookSignal struct {
	Signal os.Signal
}

func (h HookSignal) Execute(sandbox *sandbox.Sandbox) error {
	//TODO implement me
	panic("implement me")
}

type Hook interface {
	Execute(sandbox *sandbox.Sandbox) error
}

type HookType string

const (
	StartHookType  HookType = "start"
	StopHookType            = "stop"
	ReloadHookType          = "reload"
)

type Hooks map[HookType]Hook

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) MakeHooks(config map[string]types.SandboxHook) (Hooks, error) {
	hooks := make(Hooks)
	for hookType, hookConfig := range config {
		hook, err := m.MakeHook(hookConfig)
		if err != nil {
			return nil, err
		}
		switch types.SandboxHookType(hookType) {
		case types.StartSandboxHookType:
			hooks[StartHookType] = hook
		case types.StopSandboxHookType:
			hooks[StopHookType] = hook
		case types.ReloadSandboxHookType:
			hooks[ReloadHookType] = hook
		}
	}
	return hooks, nil
}

// Define mapping for string signal names to os.Signal
var stringToSignalMap = map[string]os.Signal{
	"SIGTERM": syscall.SIGTERM,
	"SIGKILL": syscall.SIGKILL,
	"SIGINT":  syscall.SIGINT,
	"SIGQUIT": syscall.SIGQUIT,
}

// Define mapping for integer signal numbers to os.Signal
// This is more straightforward since os.Signal is an alias for syscall.Signal, which is int
var intToSignalMap = map[int]os.Signal{
	int(syscall.SIGTERM): syscall.SIGTERM,
	int(syscall.SIGKILL): syscall.SIGKILL,
	int(syscall.SIGINT):  syscall.SIGINT,
	int(syscall.SIGQUIT): syscall.SIGQUIT,
}

func (m *Maker) MakeHook(config types.SandboxHook) (Hook, error) {
	switch hook := config.(type) {
	case *types.SandboxHookNative:
		return &HookNative{Type: hook.Type}, nil
	case *types.SandboxHookShellCommand:
		return &HookShellCommand{Command: hook.Command, Shell: hook.Shell}, nil
	case *types.SandboxHookCommand:
		return &HookCommand{Executable: hook.Executable, Args: hook.Args}, nil
	case *types.SandboxHookSignal:
		if hook.IsString {
			signal, ok := stringToSignalMap[hook.StringValue]
			if !ok {
				return nil, errors.New("unsupported string signal value")
			}
			return &HookSignal{signal}, nil
		} else {
			signal, ok := intToSignalMap[hook.IntValue]
			if !ok {
				return nil, errors.New("unsupported int signal value")
			}
			return &HookSignal{signal}, nil
		}
	default:
		return nil, errors.New("unsupported hook type")
	}
}
