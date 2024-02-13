package actions

import (
	"fmt"
	"github.com/bukka/wst/actions/expect"
	"github.com/bukka/wst/actions/not"
	"github.com/bukka/wst/actions/parallel"
	"github.com/bukka/wst/actions/request"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/instances/runtime"
	"github.com/bukka/wst/services"
)

type Action interface {
	Execute(runData *runtime.Data) error
}

type ActionMaker struct {
	env                 app.Env
	expectOutputMaker   *expect.OutputExpectationActionMaker
	expectResponseMaker *expect.ResponseExpectationActionMaker
	notMaker            *not.ActionMaker
	parallelMaker       *parallel.ActionMaker
	requestMaker        *request.ActionMaker
}

func CreateActionMaker(env app.Env) *ActionMaker {
	return &ActionMaker{
		env:                 env,
		expectOutputMaker:   expect.CreateOutputExpectationActionMaker(env),
		expectResponseMaker: expect.CreateResponseExpectationActionMaker(env),
		notMaker:            not.CreateActionMaker(env),
		parallelMaker:       parallel.CreateActionMaker(env),
		requestMaker:        request.CreateActionMaker(env),
	}
}

func (m *ActionMaker) MakeAction(config types.Action, svcs services.Services) (Action, error) {
	switch action := config.(type) {
	case *types.OutputExpectationAction:
		return m.expectOutputMaker.MakeAction(action, svcs)
	case *types.ResponseExpectationAction:
		return m.expectResponseMaker.MakeAction(action, svcs)
	case *types.NotAction:
		return m.notMaker.Make(action, svcs, m)
	case *types.ParallelAction:
		return m.parallelMaker.Make(action, svcs, m)
	case *types.RequestAction:
		return m.requestMaker.Make(action, svcs)
	default:
		return nil, fmt.Errorf("unsupported action type: %T", config)
	}
}
