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
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"regexp"
	"time"
)

type OutputExpectationActionMaker struct {
	env app.Env
}

func CreateOutputExpectationActionMaker(env app.Env) *OutputExpectationActionMaker {
	return &OutputExpectationActionMaker{
		env: env,
	}
}

func (m *OutputExpectationActionMaker) MakeAction(
	config *types.OutputExpectationAction,
	svcs services.Services,
	defaultTimeout int,
) (actions.Action, error) {
	order := OrderType(config.Output.Order)
	if order != OrderTypeFixed && order != OrderTypeRandom {
		return nil, fmt.Errorf("invalid order type: %v", config.Output.Order)
	}

	match := MatchType(config.Output.Match)
	if match != MatchTypeExact && match != MatchTypeRegexp {
		return nil, fmt.Errorf("invalid match type: %v", config.Output.Match)
	}

	outputType := OutputType(config.Output.Type)
	if outputType != OutputTypeAny && outputType != OutputTypeStdout && outputType != OutputTypeStderr {
		return nil, fmt.Errorf("invalid output type: %v", config.Output.Type)
	}

	svc, err := svcs.FindService(config.Service)
	if err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return &outputAction{
		expectationAction: expectationAction{
			service: svc,
			timeout: time.Duration(config.Timeout),
		},
		orderType:      order,
		matchType:      match,
		outputType:     outputType,
		messages:       config.Output.Messages,
		renderTemplate: config.Output.RenderTemplate,
	}, nil
}

type outputAction struct {
	expectationAction
	orderType      OrderType
	matchType      MatchType
	outputType     OutputType
	messages       []string
	renderTemplate bool
}

func (a *outputAction) Timeout() time.Duration {
	return a.timeout
}

func (a *outputAction) Execute(ctx context.Context, runData runtime.Data, dryRun bool) (bool, error) {
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
		renderedMessage, err := a.service.RenderTemplate(message)
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
