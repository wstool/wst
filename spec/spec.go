package spec

import (
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/servers"
)

type Spec interface {
	ExecuteInstances(filteredInstances []string, dryRun bool) error
}

func MakeSpec(config *types.Config, servers servers.Servers) (Spec, error) {
	//TODO implement me
	panic("implement me")
}
