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
	"fmt"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"regexp"
	"time"
)

func (m *ExpectationActionMaker) MakeOutputAction(
	config *types.OutputExpectationAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	commonExpectation, err := m.MakeCommonExpectation(sl, config.Service, config.Timeout, defaultTimeout)
	if err != nil {
		return nil, err
	}

	outputExpectation, err := m.expectationsMaker.MakeOutputExpectation(&config.Output)
	if err != nil {
		return nil, err
	}

	return &outputAction{
		CommonExpectation: commonExpectation,
		OutputExpectation: outputExpectation,
		parameters:        parameters.Parameters{},
	}, nil
}

type outputAction struct {
	*CommonExpectation
	*expectations.OutputExpectation
	parameters parameters.Parameters
}

func (a *outputAction) Timeout() time.Duration {
	return a.timeout
}

func (a *outputAction) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing expectation output action")
	outputType, err := a.getServiceOutputType(a.OutputType)
	if err != nil {
		return false, err
	}
	messages, err := a.renderMessages(a.Messages)
	if err != nil {
		return false, err
	}
	scanner, err := a.service.OutputScanner(ctx, outputType)
	if err != nil {
		return false, err
	}
	for scanner.Scan() {
		line := scanner.Text()
		messages, err = a.matchMessages(line, messages)
		if err != nil {
			return false, err
		}
		if len(messages) == 0 {
			return true, nil
		}
	}
	if scanner.Err() != nil {
		return false, scanner.Err()
	}

	if a.fnd.DryRun() {
		return true, nil
	}

	return false, nil
}

func (a *outputAction) getServiceOutputType(outputType expectations.OutputType) (output.Type, error) {
	switch outputType {
	case expectations.OutputTypeStdout:
		return output.Stdout, nil
	case expectations.OutputTypeStderr:
		return output.Stderr, nil
	case expectations.OutputTypeAny:
		return output.Any, nil
	default:
		return output.Any, fmt.Errorf("unknow output type %s", string(outputType))
	}
}

func (a *outputAction) renderMessages(messages []string) ([]string, error) {
	if !a.RenderTemplate {
		return messages, nil
	}
	var renderedMessages []string
	for _, message := range messages {
		renderedMessage, err := a.service.RenderTemplate(message, a.parameters)
		if err != nil {
			return nil, err
		}
		renderedMessages = append(renderedMessages, renderedMessage)
	}

	return renderedMessages, nil
}

func (a *outputAction) matchMessages(line string, messages []string) ([]string, error) {
	if a.OrderType == expectations.OrderTypeFixed {
		if len(messages) > 0 {
			matched, err := a.matchMessage(line, messages[0])
			if err != nil {
				return nil, err
			}
			if matched {
				return messages[1:], nil
			}
		}
	} else if a.OrderType == expectations.OrderTypeRandom {
		for index, message := range messages {
			matched, err := a.matchMessage(line, message)
			if err != nil {
				return nil, err
			}
			if matched {
				return append(messages[:index], messages[index+1:]...), nil
			}
		}
	} else {
		return nil, fmt.Errorf("unknown order type %s", string(a.OrderType))
	}
	return messages, nil
}

func (a *outputAction) matchMessage(line, message string) (bool, error) {
	if a.MatchType == expectations.MatchTypeExact {
		a.fnd.Logger().Debugf("Matching message '%s' against line: %s", message, line)
		return line == message, nil
	} else if a.MatchType == expectations.MatchTypeRegexp {
		a.fnd.Logger().Debugf("Matching pattern '%s' against line: %s", message, line)
		return regexp.MatchString(message, line)
	} else {
		return false, fmt.Errorf("unknown match type %s", string(a.MatchType))
	}
}
