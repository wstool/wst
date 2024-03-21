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
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"regexp"
	"time"
)

func (m *ExpectationActionMaker) MakeOutputExpectation(
	config *types.OutputExpectation,
) (*OutputExpectation, error) {
	orderType := OrderType(config.Order)
	if orderType != OrderTypeFixed && orderType != OrderTypeRandom {
		return nil, fmt.Errorf("invalid order type: %v", config.Order)
	}

	matchType := MatchType(config.Match)
	if matchType != MatchTypeExact && matchType != MatchTypeRegexp {
		return nil, fmt.Errorf("invalid match type: %v", config.Match)
	}

	outputType := OutputType(config.Type)
	if outputType != OutputTypeAny && outputType != OutputTypeStdout && outputType != OutputTypeStderr {
		return nil, fmt.Errorf("invalid output type: %v", config.Type)
	}

	return &OutputExpectation{
		orderType:      orderType,
		matchType:      matchType,
		outputType:     outputType,
		messages:       config.Messages,
		renderTemplate: config.RenderTemplate,
	}, nil
}

type OutputExpectation struct {
	orderType      OrderType
	matchType      MatchType
	outputType     OutputType
	messages       []string
	renderTemplate bool
}

func (m *ExpectationActionMaker) MakeOutputAction(
	config *types.OutputExpectationAction,
	svcs services.Services,
	defaultTimeout int,
) (actions.Action, error) {
	commonExpectation, err := m.MakeCommonExpectation(svcs, config.Service, config.Timeout, defaultTimeout)
	if err != nil {
		return nil, err
	}

	outputExpectation, err := m.MakeOutputExpectation(&config.Output)
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
	*OutputExpectation
	parameters parameters.Parameters
}

func (a *outputAction) Timeout() time.Duration {
	return a.timeout
}

func (a *outputAction) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	outputType, err := a.getServiceOutputType(a.outputType)
	if err != nil {
		return false, err
	}
	messages, err := a.renderMessages(a.messages)
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

func (a *outputAction) getServiceOutputType(outputType OutputType) (output.Type, error) {
	switch outputType {
	case OutputTypeStdout:
		return output.Stdout, nil
	case OutputTypeStderr:
		return output.Stderr, nil
	case OutputTypeAny:
		return output.Any, nil
	default:
		return output.Any, fmt.Errorf("unknow output type %s", string(outputType))
	}
}

func (a *outputAction) renderMessages(messages []string) ([]string, error) {
	if !a.renderTemplate {
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
	if a.orderType == OrderTypeFixed {
		if len(messages) > 0 {
			matched, err := a.matchMessage(line, messages[0])
			if err != nil {
				return nil, err
			}
			if matched {
				return messages[1:], nil
			}
		}
	} else if a.orderType == OrderTypeRandom {
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
		return nil, fmt.Errorf("unknown order type %s", string(a.orderType))
	}
	return messages, nil
}

func (a *outputAction) matchMessage(line, message string) (bool, error) {
	if a.matchType == MatchTypeExact {
		return line == message, nil
	} else if a.matchType == MatchTypeRegexp {
		return regexp.MatchString(message, line)
	} else {
		return false, fmt.Errorf("unknown match type %s", string(a.matchType))
	}
}
