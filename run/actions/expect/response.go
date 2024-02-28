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
	"context"
	"errors"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/actions/request"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"regexp"
	"time"
)

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
) (actions.Action, error) {
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

	return &responseAction{
		expectationAction: expectationAction{
			service: svc,
			timeout: time.Duration(config.Timeout),
		},
		request:            config.Response.Request,
		headers:            config.Response.Headers,
		bodyContent:        config.Response.Body.Content,
		bodyMatch:          match,
		bodyRenderTemplate: config.Response.Body.RenderTemplate,
	}, nil
}

type responseAction struct {
	expectationAction
	request            string
	headers            types.Headers
	bodyContent        string
	bodyMatch          MatchType
	bodyRenderTemplate bool
}

func (a *responseAction) Timeout() time.Duration {
	return a.timeout
}

func (a *responseAction) Execute(ctx context.Context, runData runtime.Data, dryRun bool) (bool, error) {
	data, ok := runData.Load(a.request)
	if !ok {
		return false, errors.New("response data not found")
	}

	responseData, ok := data.(request.ResponseData)
	if !ok {
		return false, errors.New("invalid response data type")
	}

	// Compare headers.
	for key, expectedValue := range a.headers {
		value, ok := responseData.Headers[key]
		if !ok || (len(value) > 0 && value[0] != expectedValue) {
			return false, errors.New("header mismatch")
		}
	}

	content, err := a.renderBodyContent(ctx)
	if err != nil {
		return false, err
	}

	// Compare body content based on bodyMatch.
	switch a.bodyMatch {
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
			return false, errors.New("body content does not match type regexp")
		}
	}

	return true, nil
}

func (a *responseAction) renderBodyContent(ctx context.Context) (string, error) {
	if a.bodyRenderTemplate {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			content, err := a.service.RenderTemplate(a.bodyContent)
			if err != nil {
				return "", err
			}
			return content, nil
		}
	}

	return a.bodyContent, nil
}
