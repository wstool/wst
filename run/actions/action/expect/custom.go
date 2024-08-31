package expect

import (
	"context"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"github.com/pkg/errors"
	"time"
)

func (m *ExpectationActionMaker) MakeCustomAction(
	config *types.CustomExpectationAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	commonExpectation, err := m.MakeCommonExpectation(sl, config.Service, config.Timeout, defaultTimeout, config.When)
	if err != nil {
		return nil, err
	}

	server := commonExpectation.service.Server()
	customName := config.Custom.Name
	expectation, ok := server.ExpectAction(customName)
	if !ok {
		return nil, errors.Errorf("expectation action %s not found", customName)
	}

	configParameters, err := m.parametersMaker.Make(config.Custom.Parameters)
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
	*expectations.OutputExpectation
	*expectations.ResponseExpectation
	parameters parameters.Parameters
}

func (a *customAction) When() action.When {
	return a.when
}

func (a *customAction) Timeout() time.Duration {
	return a.timeout
}

func (a *customAction) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	if a.OutputExpectation != nil {
		oa := outputAction{
			CommonExpectation: a.CommonExpectation,
			OutputExpectation: a.OutputExpectation,
			parameters:        a.parameters,
		}
		return oa.Execute(ctx, runData)
	}
	if a.ResponseExpectation != nil {
		ra := responseAction{
			CommonExpectation:   a.CommonExpectation,
			ResponseExpectation: a.ResponseExpectation,
			parameters:          a.parameters,
		}
		return ra.Execute(ctx, runData)
	}
	return false, errors.Errorf("no expectation set")
}
