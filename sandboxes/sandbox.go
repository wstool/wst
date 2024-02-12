package sandboxes

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
)

type Sandbox interface {
}

type Sandboxes map[string]Sandbox

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
