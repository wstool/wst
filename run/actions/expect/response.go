package expect

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
)

type ResponseAction struct {
	ExpectationAction
	Request            string
	Headers            types.Headers
	BodyContent        string
	BodyMatch          MatchType
	BodyRenderTemplate bool
}

type ResponseExpectationActionMaker struct {
	env app.Env
}

func CreateResponseExpectationActionMaker(env app.Env) *ResponseExpectationActionMaker {
	return &ResponseExpectationActionMaker{
		env: env,
	}
}

func (m *ResponseExpectationActionMaker) MakeAction(
	config *types.ResponseExpectationAction,
	svcs services.Services,
) (*ResponseAction, error) {
	match := MatchType(config.Response.Body.Match)
	if match != MatchTypeExact && match != MatchTypeRegexp {
		return nil, fmt.Errorf("invalid MatchType: %v", config.Response.Body.Match)
	}

	svc, err := svcs.GetService(config.Service)
	if err != nil {
		return nil, err
	}

	return &ResponseAction{
		ExpectationAction: ExpectationAction{
			Service: svc,
		},
		Request:            config.Response.Request,
		Headers:            config.Response.Headers,
		BodyContent:        config.Response.Body.Content,
		BodyMatch:          match,
		BodyRenderTemplate: config.Response.Body.RenderTemplate,
	}, nil
}

func (a ResponseAction) Execute(runData runtime.Data) (bool, error) {
	// implementation here
	// use runData.Store(key, value) to store data.
	// and value, ok := runData.Load(key) to retrieve data.
	return true, nil
}
