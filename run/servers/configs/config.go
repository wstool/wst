package configs

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
)

type Config interface {
}

type Configs map[string]Config

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(config map[string]types.ServerConfig) (Configs, error) {
	configs := make(Configs)
	for name, serverConfig := range config {
		configs[name] = &nativeConfig{
			file:       serverConfig.File,
			parameters: serverConfig.Parameters,
		}
	}
	return configs, nil
}

type nativeConfig struct {
	file       string
	parameters types.Parameters
}
