package spec

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/servers"
)

type Spec interface {
	ExecuteInstances(filteredInstances []string, dryRun bool) error
}

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(config *types.Config, servers servers.Servers) (Spec, error) {
	//TODO implement me
	panic("implement me")
}
