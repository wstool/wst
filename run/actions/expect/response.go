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
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"regexp"
	"time"
)

func (m *ExpectationActionMaker) MakeResponseAction(
	config *types.ResponseExpectationAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (actions.Action, error) {
	commonExpectation, err := m.MakeCommonExpectation(sl, config.Service, config.Timeout, defaultTimeout)
	if err != nil {
		return nil, err
	}

	responseExpectation, err := m.expectationsMaker.MakeResponseExpectation(&config.Response)
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
	*expectations.ResponseExpectation
	parameters parameters.Parameters
}

func (a *responseAction) Timeout() time.Duration {
	return a.timeout
}

func (a *responseAction) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing expectation output action")
	data, ok := runData.Load(fmt.Sprintf("response/%s", a.Request))
	if !ok {
		return false, errors.New("response data not found")
	}

	responseData, ok := data.(request.ResponseData)
	if !ok {
		return false, errors.New("invalid response data type")
	}
	a.fnd.Logger().Debugf("Checking response %s data: %v", a.Request, responseData)

	// Compare headers.
	for key, expectedValue := range a.Headers {
		value, ok := responseData.Headers[key]
		a.fnd.Logger().Debugf("Comparing header %s with value %s against expected value %s", key, value, expectedValue)
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
	switch a.BodyMatch {
	case expectations.MatchTypeExact:
		a.fnd.Logger().Debugf("Matching body %s with expected content %s", responseData.Body, content)
		if responseData.Body != content {
			return false, errors.New("body content mismatch")
		}
	case expectations.MatchTypeRegexp:
		a.fnd.Logger().Debugf("Matching body %s with expected pattern %s", responseData.Body, content)
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
	if a.BodyRenderTemplate {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			content, err := a.service.RenderTemplate(a.BodyContent, a.parameters)
			if err != nil {
				return "", err
			}
			return content, nil
		}
	}

	return a.BodyContent, nil
}
