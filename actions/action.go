package actions

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/instances/runtime"
	"github.com/bukka/wst/services"
)

type Action interface {
	Execute(runData *runtime.Data) error
}

type ActionMaker struct {
	env app.Env
}

func CreateActionMaker(env app.Env) *ActionMaker {
	return &ActionMaker{
		env: env,
	}
}

func (m *ActionMaker) MakeAction(config types.Action, svcs services.Services) (Action, error) {
	//TODO implement me
	panic("implement me")
}
