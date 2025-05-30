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
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/actions/action/request"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/services"
	"regexp"
	"strings"
)

func (m *ExpectationActionMaker) MakeResponseAction(
	config *types.ResponseExpectationAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	commonExpectation, err := m.MakeCommonExpectation(
		sl, config.Service, config.Timeout, defaultTimeout, config.When, config.OnFailure)
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
		parameters:          commonExpectation.service.ServerParameters(),
	}, nil
}

type responseAction struct {
	*CommonExpectation
	*expectations.ResponseExpectation
	parameters parameters.Parameters
}

func (a *responseAction) Execute(_ context.Context, runData runtime.Data) (bool, error) {
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

	noMatchResult := false
	if a.fnd.DryRun() {
		noMatchResult = true
	}

	// Compare status code.
	if a.StatusCode != 0 {
		a.fnd.Logger().Debugf("Comparing status code %d against expected status code %d",
			responseData.StatusCode, a.StatusCode)
		if responseData.StatusCode != a.StatusCode {
			a.fnd.Logger().Infof("Status code did not match")
			return noMatchResult, nil
		}
	}

	// Compare headers.
	for key, expectedValue := range a.Headers {
		value, ok := responseData.Headers[key]
		a.fnd.Logger().Debugf("Comparing header %s with value %s against expected value %s",
			key, value, expectedValue)
		if !ok || (len(value) > 0 && value[0] != expectedValue) {
			a.fnd.Logger().Infof("Headers did not match")
			return noMatchResult, nil
		}
	}

	content, err := a.renderBodyContent()
	if err != nil {
		return false, err
	}

	// Compare body content based on bodyMatch.
	switch a.BodyMatch {
	case expectations.MatchTypeExact:
		a.fnd.Logger().Debugf("Matching body %s with expected content %s", responseData.Body, content)
		if responseData.Body != content {
			a.fnd.Logger().Infof("Body did not exactly match")
			return noMatchResult, nil
		}
	case expectations.MatchTypeRegexp:
		a.fnd.Logger().Debugf("Matching body %s with expected pattern %s", responseData.Body, content)
		matched, err := regexp.MatchString(content, responseData.Body)
		if err != nil {
			return noMatchResult, err
		}
		if !matched {
			a.fnd.Logger().Infof("Body did not match the pattern")
			return noMatchResult, nil
		}
	case expectations.MatchTypePrefix:
		a.fnd.Logger().Debugf("Matching body %s with expected prefix %s", responseData.Body, content)
		if !strings.HasPrefix(responseData.Body, content) {
			a.fnd.Logger().Infof("Body did not match the prefix")
			return noMatchResult, nil
		}
	case expectations.MatchTypeSuffix:
		a.fnd.Logger().Debugf("Matching body %s with expected suffix %s", responseData.Body, content)
		if !strings.HasSuffix(responseData.Body, content) {
			a.fnd.Logger().Infof("Body did not match the suffix")
			return noMatchResult, nil
		}
	case expectations.MatchTypeInfix:
		a.fnd.Logger().Debugf("Matching body %s with expected infix %s", responseData.Body, content)
		if !strings.Contains(responseData.Body, content) {
			a.fnd.Logger().Infof("Body did not contain the expected content")
			return noMatchResult, nil
		}
	}

	return true, nil
}

func (a *responseAction) renderBodyContent() (string, error) {
	if a.BodyRenderTemplate {
		content, err := a.service.RenderTemplate(a.BodyContent, a.parameters)
		if err != nil {
			return "", err
		}
		return content, nil
	}

	return a.BodyContent, nil
}
