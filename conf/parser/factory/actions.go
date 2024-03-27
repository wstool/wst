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

package factory

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"strings"
)

type ActionsFactory interface {
	ParseActions(actions []interface{}, path string) ([]types.Action, error)
}

type ActionMeta struct {
	actionName  string
	serviceName string
	customName  string
}

type NativeActionsFactory struct {
	fnd          app.Foundation
	structParser StructParser
}

func CreateActionsFactory(fnd app.Foundation, structParser StructParser) ActionsFactory {
	return &NativeActionsFactory{
		fnd:          fnd,
		structParser: structParser,
	}
}

func (f *NativeActionsFactory) parseActionString(actionString string) (*ActionMeta, error) {
	if actionString == "" {
		return nil, fmt.Errorf("action string cannot be empty")
	}
	elements := strings.Split(actionString, "/")
	actionName := elements[0]
	var serviceName, customName string
	if len(elements) > 1 {
		serviceName = elements[1]
		if len(elements) > 2 {
			customName = elements[2]
			if len(elements) > 3 {
				return nil, fmt.Errorf("action string cannot be composed of more than three elements")
			}
		}
	}

	return &ActionMeta{
		actionName:  actionName,
		serviceName: serviceName,
		customName:  customName,
	}, nil
}

func (f *NativeActionsFactory) parseExpectationAction(
	meta *ActionMeta,
	data map[string]interface{},
	path string,
) (types.Action, error) {
	if meta.customName != "" {
		return &types.CustomExpectationAction{
			Service:    meta.serviceName,
			Name:       meta.customName,
			Parameters: data,
		}, nil
	}
	if _, ok := data["custom"]; ok {
		customAction := types.CustomExpectationAction{Service: meta.serviceName}
		err := f.structParser(data, &customAction, path)
		if err != nil {
			return nil, err
		}
		return &customAction, nil
	}
	if _, ok := data["metrics"]; ok {
		metricsAction := types.MetricsExpectationAction{Service: meta.serviceName}
		err := f.structParser(data, &metricsAction, path)
		if err != nil {
			return nil, err
		}
		return &metricsAction, nil
	}
	if _, ok := data["output"]; ok {
		outputAction := types.OutputExpectationAction{Service: meta.serviceName}
		err := f.structParser(data, &outputAction, path)
		if err != nil {
			return nil, err
		}
		return &outputAction, nil
	}
	if _, ok := data["response"]; ok {
		customAction := types.ResponseExpectationAction{Service: meta.serviceName}
		err := f.structParser(data, &customAction, path)
		if err != nil {
			return nil, err
		}
		return &customAction, nil
	}
	return nil, fmt.Errorf("invalid expectation action - no expectation type defined")
}

func (f *NativeActionsFactory) parseAction(
	actionString string,
	data map[string]interface{},
	path string,
) (types.Action, error) {
	meta, err := f.parseActionString(actionString)
	if err != nil {
		return nil, err
	}
	customNameAllowed := false
	serviceNameAllowed := true
	var action types.Action
	switch meta.actionName {
	case "bench":
		benchAction := types.BenchAction{Service: meta.serviceName}
		err = f.structParser(data, benchAction, path)
		action = &benchAction
	case "expect":
		customNameAllowed = true
		action, err = f.parseExpectationAction(meta, data, path)
	case "not":
		serviceNameAllowed = false
		notAction := types.NotAction{}
		err = f.structParser(data, notAction, path)
		action = &notAction
	case "parallel":
		serviceNameAllowed = false
		parallelAction := types.ParallelAction{}
		err = f.structParser(data, parallelAction, path)
		action = &parallelAction
	case "reload":
		reloadAction := types.ReloadAction{Service: meta.serviceName}
		err = f.structParser(data, reloadAction, path)
		action = &reloadAction
	case "request":
		requestAction := types.RequestAction{Service: meta.serviceName}
		err = f.structParser(data, requestAction, path)
		action = &requestAction
	case "restart":
		restartAction := types.RestartAction{Service: meta.serviceName}
		err = f.structParser(data, restartAction, path)
		action = &restartAction
	case "start":
		startAction := types.StartAction{Service: meta.serviceName}
		err = f.structParser(data, startAction, path)
		action = &startAction
	case "stop":
		stopAction := types.StopAction{Service: meta.serviceName}
		err = f.structParser(data, stopAction, path)
		action = &stopAction
	default:
		return nil, fmt.Errorf("unknown action %s", meta.actionName)
	}

	if err != nil {
		return nil, err
	}
	if meta.customName != "" && !customNameAllowed {
		return nil, fmt.Errorf("custom name not allowed for action %s", meta.actionName)
	}
	if meta.serviceName != "" && !serviceNameAllowed {
		return nil, fmt.Errorf("service name not allowed for action %s", meta.actionName)
	}

	return action, nil
}

func (f *NativeActionsFactory) parseActionFromMap(action map[string]interface{}, path string) (types.Action, error) {
	if len(action) > 1 {
		return nil, fmt.Errorf("invalid action format - exactly one elelemnt in object is required")
	}
	for name, value := range action {
		valueMap, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid action format - action value must be an object")
		}
		return f.parseAction(name, valueMap, path)
	}
	return nil, fmt.Errorf("invalid action format - empty object is not valid action")
}

func (f *NativeActionsFactory) ParseActions(actions []interface{}, path string) ([]types.Action, error) {
	var parsedActions []types.Action
	for _, untypedAction := range actions {
		switch action := untypedAction.(type) {
		case string:
			parsedAction, err := f.parseAction(action, map[string]interface{}{}, path)
			if err != nil {
				return nil, err
			}
			parsedActions = append(parsedActions, parsedAction)
		case map[string]interface{}:
			parsedAction, err := f.parseActionFromMap(action, path)
			if err != nil {
				return nil, err
			}
			parsedActions = append(parsedActions, parsedAction)
		default:
			return nil, fmt.Errorf("unsupported action type")
		}
	}
	return parsedActions, nil
}
