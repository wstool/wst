package not

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
)

type Action struct {
	Action actions.Action
}

type ActionMaker struct {
	env app.Env
}

func CreateActionMaker(env app.Env) *ActionMaker {
	return &ActionMaker{
		env: env,
	}
}

func (m *ActionMaker) Make(
	config *types.NotAction,
	svcs services.Services,
	actionMaker *actions.ActionMaker,
) (*Action, error) {
	action, err := actionMaker.MakeAction(config.Action, svcs)
	if err != nil {
		return nil, err
	}
	return &Action{
		Action: action,
	}, nil
}

func (a Action) Execute(runData runtime.Data) (bool, error) {
	success, err := a.Action.Execute(runData)
	if err != nil {
		return false, err
	}
	return !success, nil
}
