package servers

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/servers/configs"
	"github.com/bukka/wst/run/servers/templates"
	"strings"
)

type Server interface {
	GetConfig(name string) (configs.Config, bool)
	GetSandbox(name sandbox.Type) (sandbox.Sandbox, bool)
}

type Servers map[string]map[string]Server

func splitFullName(fullName string) (string, string) {
	parts := strings.Split(fullName, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return fullName, "default"
}

func (s Servers) GetServer(fullName string) (Server, bool) {
	name, tag := splitFullName(fullName)
	server, ok := s[name][tag]
	return server, ok
}

type Maker struct {
	env            app.Env
	configsMaker   *configs.Maker
	sandboxesMaker *sandboxes.Maker
	templatesMaker *templates.Maker
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env:            env,
		configsMaker:   configs.CreateMaker(env),
		sandboxesMaker: sandboxes.CreateMaker(env),
		templatesMaker: templates.CreateMaker(env),
	}
}

func (m *Maker) Make(config *types.Config) (Servers, error) {
	srvs := make(map[string]map[string]Server)
	for _, server := range config.Servers {
		name, tag := splitFullName(config.Name)
		serverConfigs, err := m.configsMaker.Make(server.Configs)
		if err != nil {
			return nil, err
		}

		serverTemplates, err := m.templatesMaker.Make(server.Templates)
		if err != nil {
			return nil, err
		}

		serverSandboxes, err := m.sandboxesMaker.MakeSandboxes(config.Sandboxes, server.Sandboxes)
		if err != nil {
			return nil, err
		}

		srvs[name][tag] = &nativeServer{
			name:       name,
			tag:        tag,
			parentName: server.Extends,
			configs:    serverConfigs,
			templates:  serverTemplates,
			parameters: server.Parameters,
			sandboxes:  serverSandboxes,
		}
	}

	// set parents
	for name, nameServers := range srvs {
		for tag, srv := range nameServers {
			nativeSrv, ok := srv.(nativeServer)
			if !ok {
				// this should never happen so something went seriously wrong
				panic("conversion to nativeServer failed")
			}
			if nativeSrv.parentName != "" {
				parentName, parentTag := splitFullName(nativeSrv.parentName)
				parent, ok := srvs[parentName][parentTag]
				if !ok {
					return nil, fmt.Errorf(
						"parent %s/%s for server %s/%s not found",
						parentName,
						parentTag,
						name,
						tag,
					)
				}
				nativeSrv.parent = &parent
			}
		}
	}

	return srvs, nil
}

type nativeServer struct {
	name       string
	tag        string
	parentName string
	parent     *Server
	configs    configs.Configs
	templates  templates.Templates
	parameters types.Parameters
	sandboxes  sandboxes.Sandboxes
}

func (s nativeServer) GetConfig(name string) (configs.Config, bool) {
	config, ok := s.configs[name]
	return config, ok
}

func (s nativeServer) GetSandbox(name sandbox.Type) (sandbox.Sandbox, bool) {
	sandbox, ok := s.sandboxes[name]
	return sandbox, ok
}
