package merger

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
)

type Merger interface {
	MergeConfigs(configs []*types.Config, overwrites map[string]string) (*types.Config, error)
}

type nativeMerger struct {
	env app.Env
}

func CreateMerger(env app.Env) Merger {
	return &nativeMerger{
		env: env,
	}
}

func (n *nativeMerger) MergeConfigs(configs []*types.Config, overwrites map[string]string) (*types.Config, error) {
	//TODO implement me
	panic("implement me")
}
