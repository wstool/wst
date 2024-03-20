package expect

import (
	"context"
	"fmt"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"time"
)

func (m *ExpectationActionMaker) MakeCustomAction(
	config *types.CustomExpectationAction,
	svcs services.Services,
	defaultTimeout int,
) (actions.Action, error) {
	commonExpectation, err := m.MakeCommonExpectation(svcs, config.Service, config.Timeout, defaultTimeout)
	if err != nil {
		return nil, err
	}

	server := commonExpectation.service.Server()
	expectation, ok := server.ExpectAction(config.Name)
	if !ok {
		return nil, fmt.Errorf("expectation action %s not found", config.Name)
	}

	configParameters, err := m.parametersMaker.Make(config.Parameters)
	if err != nil {
		return nil, err
	}

	return &customAction{
		CommonExpectation:   commonExpectation,
		OutputExpectation:   expectation.OutputExpectation(),
		ResponseExpectation: expectation.ResponseExpectation(),
		parameters:          configParameters.Inherit(expectation.Parameters()).Inherit(server.Parameters()),
	}, nil
}

type customAction struct {
	*CommonExpectation
	*OutputExpectation
	*ResponseExpectation
	parameters parameters.Parameters
}

func (a *customAction) Timeout() time.Duration {
	return a.timeout
}

func (a *customAction) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	if a.OutputExpectation != nil {
		action := outputAction{
			CommonExpectation: a.CommonExpectation,
			OutputExpectation: a.OutputExpectation,
			parameters:        a.parameters,
		}
		return action.Execute(ctx, runData)
	}
	if a.ResponseExpectation != nil {
		action := responseAction{
			CommonExpectation:   a.CommonExpectation,
			ResponseExpectation: a.ResponseExpectation,
			parameters:          a.parameters,
		}
		return action.Execute(ctx, runData)
	}
	return false, fmt.Errorf("no expectation set")
}
