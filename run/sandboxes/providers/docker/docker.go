package docker

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/providers/container"
	"github.com/bukka/wst/run/sandboxes/sandbox/hooks"
)

type Maker struct {
	env            app.Env
	containerMaker *container.Maker
}

func CreateMaker(env app.Env, containerMaker *container.Maker) *Maker {
	return &Maker{
		env:            env,
		containerMaker: containerMaker,
	}
}

func (m *Maker) MakeSandbox(config *types.DockerSandbox) (*Sandbox, error) {
	panic("implement")
}

type Sandbox struct {
	container.Sandbox
}

func (s Sandbox) ExecuteCommand(command *hooks.HookCommand) error {
	//TODO implement me
	panic("implement me")
}

func (s Sandbox) ExecuteSignal(signal *hooks.HookSignal) error {
	//TODO implement me
	panic("implement me")
}
