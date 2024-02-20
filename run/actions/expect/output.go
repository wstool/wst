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
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/services"
	"regexp"
)

type OutputAction struct {
	ExpectationAction
	Order          OrderType
	Match          MatchType
	Type           OutputType
	Messages       []string
	RenderTemplate bool
}

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
) (*OutputAction, error) {
	order := OrderType(config.Output.Order)
	if order != OrderTypeFixed && order != OrderTypeRandom {
		return nil, fmt.Errorf("invalid OrderType: %v", config.Output.Order)
	}

	match := MatchType(config.Output.Match)
	if match != MatchTypeExact && match != MatchTypeRegexp {
		return nil, fmt.Errorf("invalid MatchType: %v", config.Output.Match)
	}

	outputType := OutputType(config.Output.Type)
	if outputType != OutputTypeStdout && outputType != OutputTypeStderr {
		return nil, fmt.Errorf("invalid OutputType: %v", config.Output.Type)
	}

	svc, err := svcs.GetService(config.Service)
	if err != nil {
		return nil, err
	}

	return &OutputAction{
		ExpectationAction: ExpectationAction{
			Service: svc,
		},
		Order:          order,
		Match:          match,
		Type:           outputType,
		Messages:       config.Output.Messages,
		RenderTemplate: config.Output.RenderTemplate,
	}, nil
}

func (a *OutputAction) Execute(runData runtime.Data, dryRun bool) (bool, error) {
	sandbox := a.Service.GetSandbox()
	outputType, err := a.getSandboxOutputType(a.Type)
	if err != nil {
		return false, err
	}
	messages, err := a.renderMessages(a.Messages)
	if err != nil {
		return false, err
	}
	scanner := sandbox.GetOutputScanner(outputType)
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

func (a *OutputAction) getSandboxOutputType(outputType OutputType) (sandbox.OutputType, error) {
	switch outputType {
	case OutputTypeStdout:
		return sandbox.StdoutOutputType, nil
	case OutputTypeStderr:
		return sandbox.StderrOutputType, nil
	default:
		return sandbox.StderrOutputType, fmt.Errorf("unknow output type %s", string(outputType))
	}
}

func (a *OutputAction) renderMessages(messages []string) ([]string, error) {
	if !a.RenderTemplate {
		return messages, nil
	}
	var renderedMessages []string
	for _, message := range messages {
		renderedMessage, err := a.Service.RenderTemplate(message)
		if err != nil {
			return nil, err
		}
		renderedMessages = append(renderedMessages, renderedMessage)
	}

	return renderedMessages, nil
}

func (a *OutputAction) matchMessages(line string, messages []string) ([]string, error) {
	if a.Order == OrderTypeFixed {
		if len(messages) > 0 {
			matched, err := a.matchMessage(line, messages[0])
			if err != nil {
				return nil, err
			}
			if matched {
				return messages[1:], nil
			}
		}
	} else if a.Order == OrderTypeRandom {
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
		return nil, fmt.Errorf("unknown order %s", string(a.Order))
	}
	return messages, nil
}

func (a *OutputAction) matchMessage(line, message string) (bool, error) {
	if a.Match == MatchTypeExact {
		return line == message, nil
	} else if a.Match == MatchTypeRegexp {
		return regexp.MatchString(message, line)
	} else {
		return false, fmt.Errorf("unknown match type %s", string(a.Match))
	}
}
