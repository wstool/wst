package services

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/sandboxes"
	"github.com/bukka/wst/servers"
)

type Service interface {
}

type Services map[string]Service

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(config *types.Instance, sandboxes sandboxes.Sandboxes, servers servers.Servers) (Services, error) {
	//TODO implement me
	panic("implement me")
}
