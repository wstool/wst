package templates

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
)

type Template interface {
}

type Templates map[string]Template

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(config map[string]types.ServerTemplate) (Templates, error) {
	configs := make(Templates)
	for name, serverTemplate := range config {
		configs[name] = &nativeTemplate{
			file: serverTemplate.File,
		}
	}
	return configs, nil
}

type nativeTemplate struct {
	file string
}
