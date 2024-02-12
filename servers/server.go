package servers

import (
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/sandboxes"
)

type Server interface {
}

type Servers map[string]map[string]Server

func MakeServers(config *types.Config, sandboxes sandboxes.Sandboxes) (Servers, error) {
	//TODO implement me
	panic("implement me")
}
