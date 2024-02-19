// This is an abstract provider

package container

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/providers/common"
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

func (m *Maker) MakeSandbox(config *types.ContainerSandbox) (*Sandbox, error) {
	panic("implement")
}

type Sandbox struct {
	common.Sandbox
}
