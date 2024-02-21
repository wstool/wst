// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	defaultTimeout int,
) (*ResponseAction, error) {
	match := MatchType(config.Response.Body.Match)
	if match != MatchTypeExact && match != MatchTypeRegexp {
		return nil, fmt.Errorf("invalid MatchType: %v", config.Response.Body.Match)
	}

	svc, err := svcs.FindService(config.Service)
	if err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return &ResponseAction{
		ExpectationAction: ExpectationAction{
			Service: svc,
			Timeout: config.Timeout,
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
