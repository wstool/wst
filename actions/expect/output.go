package expect

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/instances/runtime"
	"github.com/bukka/wst/services"
)

type OutputAction struct {
	ExpectationAction
	Order          OrderType
	Match          MatchType
	Type           OutputType
	Messages       []string
	RenderTemplate bool
}

type OutputExpectationActionMaker struct {
	env app.Env
}

func CreateOutputExpectationActionMaker(env app.Env) *OutputExpectationActionMaker {
	return &OutputExpectationActionMaker{
		env: env,
	}
}

func (m *OutputExpectationActionMaker) MakeAction(
	config *types.OutputExpectationAction,
	svcs services.Services,
) (*OutputAction, error) {
	order := OrderType(config.Output.Order)
	if order != OrderTypeFixed && order != OrderTypeRandom {
		return nil, fmt.Errorf("invalid OrderType: %v", config.Output.Order)
	}

	match := MatchType(config.Output.Match)
	if match != MatchTypeExact && match != MatchTypeRegexp {
		return nil, fmt.Errorf("invalid MatchType: %v", config.Output.Match)
	}

	outputType := OutputType(config.Output.Type)
	if outputType != OutputTypeStdout && outputType != OutputTypeStderr && outputType != OutputTypeAny {
		return nil, fmt.Errorf("invalid OutputType: %v", config.Output.Type)
	}

	svc, err := svcs.GetService(config.Service)
	if err != nil {
		return nil, err
	}

	return &OutputAction{
		ExpectationAction: ExpectationAction{
			Service: svc,
		},
		Order:          order,
		Match:          match,
		Type:           outputType,
		Messages:       config.Output.Messages,
		RenderTemplate: config.Output.RenderTemplate,
	}, nil
}

func (a OutputAction) Execute(runData *runtime.Data) error {
	// implementation here
	// use runData.Store(key, value) to store data.
	// and value, ok := runData.Load(key) to retrieve data.
	return nil
}
