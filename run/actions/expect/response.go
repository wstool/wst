package expect

import (
	"errors"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/request"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"regexp"
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

func (a *ResponseAction) Execute(runData runtime.Data, dryRun bool) (bool, error) { // Retrieve ResponseData from runData using the Request string as the key.
	data, ok := runData.Load(a.Request)
	if !ok {
		return false, errors.New("response data not found")
	}

	responseData, ok := data.(request.ResponseData)
	if !ok {
		return false, errors.New("invalid response data type")
	}

	// Compare headers.
	for key, expectedValue := range a.Headers {
		value, ok := responseData.Headers[key]
		if !ok || (len(value) > 0 && value[0] != expectedValue) {
			return false, errors.New("header mismatch")
		}
	}

	content, err := a.getBodyContent()
	if err != nil {
		return false, err
	}

	// Compare body content based on BodyMatch.
	switch a.BodyMatch {
	case MatchTypeExact:
		if responseData.Body != content {
			return false, errors.New("body content mismatch")
		}
	case MatchTypeRegexp:
		matched, err := regexp.MatchString(content, responseData.Body)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, errors.New("body content does not match regexp")
		}
	}

	return true, nil
}

func (a *ResponseAction) getBodyContent() (string, error) {
	if a.BodyRenderTemplate {
		content, err := a.Service.RenderTemplate(a.BodyContent)
		if err != nil {
			return "", err
		}
		return content, nil
	}

	return a.BodyContent, nil
}
