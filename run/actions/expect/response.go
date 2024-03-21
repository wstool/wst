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
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/actions/request"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"regexp"
	"time"
)

func (m *ExpectationActionMaker) MakeResponseExpectation(
	config *types.ResponseExpectation,
) (*ResponseExpectation, error) {
	match := MatchType(config.Body.Match)
	if match != MatchTypeExact && match != MatchTypeRegexp {
		return nil, fmt.Errorf("invalid MatchType: %v", config.Body.Match)
	}

	return &ResponseExpectation{
		request:            config.Request,
		headers:            config.Headers,
		bodyContent:        config.Body.Content,
		bodyMatch:          match,
		bodyRenderTemplate: config.Body.RenderTemplate,
	}, nil
}

type ResponseExpectation struct {
	request            string
	headers            types.Headers
	bodyContent        string
	bodyMatch          MatchType
	bodyRenderTemplate bool
}

func (m *ExpectationActionMaker) MakeResponseAction(
	config *types.ResponseExpectationAction,
	svcs services.Services,
	defaultTimeout int,
) (actions.Action, error) {
	commonExpectation, err := m.MakeCommonExpectation(svcs, config.Service, config.Timeout, defaultTimeout)
	if err != nil {
		return nil, err
	}

	responseExpectation, err := m.MakeResponseExpectation(&config.Response)
	if err != nil {
		return nil, err
	}

	return &responseAction{
		CommonExpectation:   commonExpectation,
		ResponseExpectation: responseExpectation,
		parameters:          parameters.Parameters{},
	}, nil
}

type responseAction struct {
	*CommonExpectation
	*ResponseExpectation
	parameters parameters.Parameters
}

func (a *responseAction) Timeout() time.Duration {
	return a.timeout
}

func (a *responseAction) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
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

	if a.fnd.DryRun() {
		return true, nil
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
			content, err := a.service.RenderTemplate(a.bodyContent, a.parameters)
			if err != nil {
				return "", err
			}
			return content, nil
		}
	}

	return a.bodyContent, nil
}
