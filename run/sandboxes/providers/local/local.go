package local

import (
	"bufio"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/providers/common"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/sandboxes/sandbox/hooks"
)

type Maker struct {
	env         app.Env
	commonMaker *common.Maker
}

func CreateMaker(env app.Env, commonMaker *common.Maker) *Maker {
	return &Maker{
		env:         env,
		commonMaker: commonMaker,
	}
}

func (m *Maker) MakeSandbox(config *types.LocalSandbox) (*Sandbox, error) {
	panic("implement")
}

type Sandbox struct {
	common.Sandbox
}

func (s Sandbox) GetOutputScanner(outputType sandbox.OutputType) *bufio.Scanner {
	//TODO implement me
	panic("implement me")
}

func (s Sandbox) ExecuteCommand(command *hooks.HookCommand) error {
	//TODO implement me
	panic("implement me")
}

func (s Sandbox) ExecuteSignal(signal *hooks.HookSignal) error {
	//TODO implement me
	panic("implement me")
}
