package servers

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes"
	"github.com/bukka/wst/run/servers/configs"
	"strings"
)

type Server interface {
	GetConfig(name string) (configs.Config, bool)
	GetSandbox(name string) (sandboxes.Sandbox, bool)
}

type Servers map[string]map[string]Server

func (s Servers) GetServer(name string) (Server, bool) {
	parts := strings.Split(name, "/")

	if len(parts) == 2 {
		server, ok := s[parts[0]][parts[1]]
		return server, ok
	}

	server, ok := s[name]["default"]
	return server, ok
}

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
