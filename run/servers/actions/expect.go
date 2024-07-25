package actions

import (
	"fmt"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/parameters"
)

type ExpectAction interface {
	Parameters() parameters.Parameters
	OutputExpectation() *expectations.OutputExpectation
	ResponseExpectation() *expectations.ResponseExpectation
}

func (m *nativeMaker) makeExpectAction(configAction types.ServerExpectationAction) (ExpectAction, error) {
	switch action := configAction.(type) {
	case *types.ServerOutputExpectation:
		params, err := m.parametersMaker.Make(action.Parameters)
		if err != nil {
			return nil, err
		}
		outputExpectation, err := m.expectationsMaker.MakeOutputExpectation(&action.Output)
		if err != nil {
			return nil, err
		}
		return &expectOutputAction{
			parameters:        params,
			outputExpectation: outputExpectation,
		}, nil
	case *types.ServerResponseExpectation:
		params, err := m.parametersMaker.Make(action.Parameters)
		if err != nil {
			return nil, err
		}
		responseExpectation, err := m.expectationsMaker.MakeResponseExpectation(&action.Response)
		if err != nil {
			return nil, err
		}
		return &expectResponseAction{
			parameters:          params,
			responseExpectation: responseExpectation,
		}, nil
	default:
		return nil, fmt.Errorf("invalid server expectation type %T", configAction)
	}
}

func (m *nativeMaker) makeExpectActions(configActions map[string]types.ServerExpectationAction) (map[string]ExpectAction, error) {
	expectActions := make(map[string]ExpectAction, len(configActions))
	var err error
	for key, configAction := range configActions {
		expectActions[key], err = m.makeExpectAction(configAction)
		if err != nil {
			return nil, err
		}
	}
	return expectActions, nil
}

type expectOutputAction struct {
	parameters        parameters.Parameters
	outputExpectation *expectations.OutputExpectation
}

func (a *expectOutputAction) OutputExpectation() *expectations.OutputExpectation {
	return a.outputExpectation
}

func (a *expectOutputAction) ResponseExpectation() *expectations.ResponseExpectation {
	return nil
}

func (a *expectOutputAction) Parameters() parameters.Parameters {
	return a.parameters
}

type expectResponseAction struct {
	parameters          parameters.Parameters
	responseExpectation *expectations.ResponseExpectation
}

func (a *expectResponseAction) OutputExpectation() *expectations.OutputExpectation {
	return nil
}

func (a *expectResponseAction) ResponseExpectation() *expectations.ResponseExpectation {
	return a.responseExpectation
}

func (a *expectResponseAction) Parameters() parameters.Parameters {
	return a.parameters
}
