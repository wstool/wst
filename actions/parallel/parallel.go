package parallel

import (
	"github.com/bukka/wst/actions"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/instances/runtime"
	"github.com/bukka/wst/services"
)

type Action struct {
	Actions []actions.Action
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
	config *types.ParallelAction,
	svcs services.Services,
	actionMaker *actions.ActionMaker,
) (*Action, error) {
	var actions []actions.Action
	for _, configAction := range config.Actions {
		action, err := actionMaker.MakeAction(configAction, svcs)
		if err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	return &Action{
		Actions: actions,
	}, nil
}

func (a Action) Execute(runData *runtime.Data) error {
	// implementation here
	// use runData.Store(key, value) to store data.
	// and value, ok := runData.Load(key) to retrieve data.
	return nil
}
