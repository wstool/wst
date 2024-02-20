package services

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/resources/scripts"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/servers/configs"
)

type Service interface {
	GetSandbox() sandbox.Sandbox
	RenderTemplate(text string) (string, error)
}

type Services map[string]Service

func (s Services) GetService(name string) (Service, error) {
	svc, ok := s[name]
	if !ok {
		return svc, fmt.Errorf("service %s not found", name)
	}
	return svc, nil
}

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(
	config map[string]types.Service,
	scriptResources scripts.Scripts,
	srvs servers.Servers,
) (Services, error) {
	svcs := make(Services)
	for serviceName, serviceConfig := range config {
		var includedScripts scripts.Scripts

		if serviceConfig.Resources.Scripts.IncludeAll {
			includedScripts = scriptResources
		} else {
			includedScripts = make(scripts.Scripts)
			for _, scriptName := range serviceConfig.Resources.Scripts.IncludeList {
				script, ok := scriptResources[scriptName]
				if !ok {
					return nil, fmt.Errorf("script %s not found for service %s", scriptName, serviceName)
				}
				includedScripts[scriptName] = script
			}
		}

		server, ok := srvs.GetServer(serviceConfig.Server)
		if !ok {
			return nil, fmt.Errorf("server %s not found for service %s", serviceConfig.Server, serviceName)
		}

		sandbox, ok := server.GetSandbox(serviceConfig.Sandbox)
		if !ok {
			return nil, fmt.Errorf("sandbox %s not found for service %s", serviceConfig.Sandbox, serviceName)
		}

		nativeConfigs := make(map[string]nativeServiceConfig)

		for configName, serviceConfig := range serviceConfig.Configs {
			config, found := server.GetConfig(configName)
			if !found {
				return nil, fmt.Errorf("server config %s not found for service %s", configName, serviceName)
			}
			nativeConfigs[configName] = nativeServiceConfig{
				parameters:          serviceConfig.Parameters,
				overwriteParameters: serviceConfig.OverwriteParameters,
				config:              config,
			}
		}

		service := &nativeService{
			scripts: includedScripts,
			server:  server,
			sandbox: sandbox,
			configs: nativeConfigs,
		}

		svcs[serviceName] = service
	}
	return svcs, nil
}

type nativeServiceConfig struct {
	parameters          types.Parameters
	overwriteParameters bool
	config              configs.Config
}

type nativeService struct {
	scripts scripts.Scripts
	server  servers.Server
	sandbox sandbox.Sandbox
	configs map[string]nativeServiceConfig
}

func (s *nativeService) RenderTemplate(text string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *nativeService) GetSandbox() sandbox.Sandbox {
	return s.sandbox
}
