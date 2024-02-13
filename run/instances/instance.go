package instances

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/resources/scripts"
	"github.com/bukka/wst/run/servers"
	"github.com/bukka/wst/run/services"
)

type Instance interface {
	ExecuteActions() error
}

type InstanceMaker struct {
	env           app.Env
	actionMaker   *actions.ActionMaker
	servicesMaker *services.Maker
	scriptsMaker  *scripts.Maker
}

func CreateInstanceMaker(env app.Env) *InstanceMaker {
	return &InstanceMaker{
		env:           env,
		actionMaker:   actions.CreateActionMaker(env),
		servicesMaker: services.CreateMaker(env),
		scriptsMaker:  scripts.CreateMaker(env),
	}
}

func (m *InstanceMaker) Make(config types.Instance, srvs servers.Servers) (Instance, error) {
	scriptResources, err := m.scriptsMaker.Make(config.Resources.Scripts)
	if err != nil {
		return nil, err
	}

	svcs, err := m.servicesMaker.Make(config.Services, scriptResources, srvs)
	if err != nil {
		return nil, err
	}

	actions := make([]actions.Action, len(config.Actions))
	for i, actionConfig := range config.Actions {
		action, err := m.actionMaker.MakeAction(actionConfig, svcs)
		if err != nil {
			return nil, err
		}
		actions[i] = action
	}
	runData := runtime.CreateData()
	return &nativeInstance{
		actions: actions,
		runData: runData,
	}, nil
}

type nativeInstance struct {
	actions []actions.Action
	runData runtime.Data
}

func (i *nativeInstance) ExecuteActions() error {
	for _, action := range i.actions {
		err := action.Execute(i.runData)
		if err != nil {
			return err
		}
	}
	return nil
}
