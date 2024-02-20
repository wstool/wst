package parallel

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"sync"
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

func (a Action) Execute(runData runtime.Data, dryRun bool) (bool, error) {
	// Use a WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup
	wg.Add(len(a.Actions))

	// Use an error channel to collect potential errors from actions.
	errs := make(chan error, len(a.Actions))

	for _, action := range a.Actions {
		go func(act actions.Action) {
			defer wg.Done()

			// Execute the action, passing the context.
			success, err := act.Execute(runData, dryRun)
			if err != nil || !success {
				errs <- err
			}
		}(action)
	}

	// Wait for all actions to complete.
	wg.Wait()
	close(errs)

	// Check if there were any errors.
	for err := range errs {
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
