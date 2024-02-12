package processor

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
)

type Processor interface {
	ProcessConfig(config *types.Config) error
}

type nativeProcessor struct {
	env app.Env
}

func CreateProcessor(env app.Env) Processor {
	return &nativeProcessor{
		env: env,
	}
}

func (p *nativeProcessor) ProcessConfig(config *types.Config) error {
	//TODO implement me
	panic("implement me")
}
