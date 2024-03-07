// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package servers

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/environments/environment/providers"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/sandboxes"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/servers/actions"
	"github.com/bukka/wst/run/servers/configs"
	"github.com/bukka/wst/run/servers/templates"
	"strings"
)

type Server interface {
	ExpectAction(name string) (actions.ExpectAction, bool)
	Config(name string) (configs.Config, bool)
	Sandbox(name providers.Type) (sandbox.Sandbox, bool)
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
	fnd             app.Foundation
	actionsMaker    *actions.Maker
	configsMaker    *configs.Maker
	sandboxesMaker  *sandboxes.Maker
	templatesMaker  *templates.Maker
	parametersMaker *parameters.Maker
}

func CreateMaker(fnd app.Foundation, parametersMaker *parameters.Maker) *Maker {
	return &Maker{
		fnd:             fnd,
		actionsMaker:    actions.CreateMaker(fnd, parametersMaker),
		configsMaker:    configs.CreateMaker(fnd),
		sandboxesMaker:  sandboxes.CreateMaker(fnd),
		templatesMaker:  templates.CreateMaker(fnd),
		parametersMaker: parametersMaker,
	}
}

func (m *Maker) Make(config *types.Spec) (Servers, error) {
	srvs := make(map[string]map[string]Server)
	for _, server := range config.Servers {
		name, tag := splitFullName(server.Name)

		serverActions, err := m.actionsMaker.Make(&server.Actions)
		if err != nil {
			return nil, err
		}

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

		serverParameters, err := m.parametersMaker.Make(server.Parameters)
		if err != nil {
			return nil, err
		}

		srvs[name][tag] = &nativeServer{
			name:       name,
			tag:        tag,
			parentName: server.Extends,
			actions:    serverActions,
			configs:    serverConfigs,
			templates:  serverTemplates,
			parameters: serverParameters,
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
	actions    *actions.Actions
	configs    configs.Configs
	templates  templates.Templates
	parameters parameters.Parameters
	sandboxes  sandboxes.Sandboxes
}

func (s nativeServer) ExpectAction(name string) (actions.ExpectAction, bool) {
	act, ok := s.actions.Expect[name]
	return act, ok
}

func (s nativeServer) Config(name string) (configs.Config, bool) {
	cfg, ok := s.configs[name]
	return cfg, ok
}

func (s nativeServer) Sandbox(name providers.Type) (sandbox.Sandbox, bool) {
	sb, ok := s.sandboxes[name]
	return sb, ok
}
