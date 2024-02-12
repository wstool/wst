package servers

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/sandboxes"
)

type Server interface {
}

type Servers map[string]map[string]Server

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(config *types.Config, sandboxes sandboxes.Sandboxes) (Servers, error) {
	//TODO implement me
	panic("implement me")
}
