package request

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
)

type Action struct {
	Service services.Service
	Id      string
	Path    string
	Method  string
	Headers types.Headers
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
	config *types.RequestAction,
	svcs services.Services,
) (*Action, error) {
	svc, err := svcs.GetService(config.Service)
	if err != nil {
		return nil, err
	}

	return &Action{
		Service: svc,
		Id:      config.Id,
		Path:    config.Path,
		Method:  config.Method,
		Headers: config.Headers,
	}, nil
}

func (a Action) Execute(runData runtime.Data) error {
	// implementation here
	// use runData.Store(key, value) to store data.
	// and value, ok := runData.Load(key) to retrieve data.
	return nil
}
