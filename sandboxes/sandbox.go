package sandboxes

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"os"
)

type SandboxHookNative struct {
	Type string
}

type SandboxHookShellCommand struct {
	Command string
	Shell   string
}

type SandboxHookCommand struct {
	Executable string
	Args       []string
}

type SandboxHookSignal os.Signal

type SandboxHook interface {
	Execute(sandbox *Sandbox) error
}

type Sandbox interface {
	ExecuteCommand(command *SandboxHookCommand) error
	ExecuteSignal(signal *SandboxHookSignal) error
}

type Sandboxes map[string]Sandbox

const (
	CommonSandboxHook     string = "common"
	LocalSandboxHook             = "local"
	ContainerSandboxHook         = "container"
	DockerSandboxHook            = "docker"
	KubernetesSandboxHook        = "kubernetes"
)

const (
	SandboxHookStartType  string = "start"
	SandboxHookStopType          = "stop"
	SandboxHookReloadType        = "reload"
)

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(config *types.Config) (Sandboxes, error) {
	//TODO implement me
	panic("implement me")
}
