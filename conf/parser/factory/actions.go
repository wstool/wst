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
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/parser/location"
	"github.com/bukka/wst/conf/types"
	"github.com/pkg/errors"
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
	loc          *location.Location
	structParser StructParser
}

func CreateActionsFactory(fnd app.Foundation, structParser StructParser, loc *location.Location) ActionsFactory {
	return &NativeActionsFactory{
		fnd:          fnd,
		loc:          loc,
		structParser: structParser,
	}
}

func (f *NativeActionsFactory) parseActionString(actionString string) (*ActionMeta, error) {
	if actionString == "" {
		return nil, errors.Errorf("action %s string cannot be empty", f.loc.String())
	}
	elements := strings.Split(actionString, "/")
	actionName := elements[0]
	var serviceName, customName string
	if len(elements) > 1 {
		serviceName = elements[1]
		if len(elements) > 2 {
			customName = elements[2]
			if len(elements) > 3 {
				return nil, errors.Errorf(
					"action %s string cannot be composed of more than three elements",
					f.loc.String(),
				)
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
	var structure interface{}
	parsed := false
	for expKey, _ := range data {
		switch expKey {
		case "service":
			continue
		case "name":
			continue
		case "timeout":
			continue
		case "custom":
			structure = &types.CustomExpectationAction{}
		case "metrics":
			structure = &types.MetricsExpectationAction{}
		case "output":
			structure = &types.OutputExpectationAction{}
		case "response":
			structure = &types.ResponseExpectationAction{}
		default:
			return nil, errors.Errorf("invalid expectation key %s at %s", expKey, f.loc.String())
		}
		if parsed {
			return nil, errors.Errorf(
				"expression cannot have multiple types - additional key %s at %s",
				f.loc.String(),
				expKey,
			)
		}
		parsed = true
	}

	err := f.structParser(data, structure, path)
	if err != nil {
		return nil, err
	}

	return structure, nil
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
		benchAction := &types.BenchAction{Service: meta.serviceName}
		err = f.structParser(data, benchAction, path)
		action = benchAction
	case "expect":
		customNameAllowed = true
		action, err = f.parseExpectationAction(meta, data, path)
	case "not":
		serviceNameAllowed = false
		notAction := &types.NotAction{}
		err = f.structParser(data, notAction, path)
		action = notAction
	case "parallel":
		serviceNameAllowed = false
		parallelAction := &types.ParallelAction{}
		err = f.structParser(data, parallelAction, path)
		action = parallelAction
	case "reload":
		reloadAction := &types.ReloadAction{Service: meta.serviceName}
		err = f.structParser(data, reloadAction, path)
		action = reloadAction
	case "request":
		requestAction := &types.RequestAction{Service: meta.serviceName}
		err = f.structParser(data, requestAction, path)
		action = requestAction
	case "restart":
		restartAction := &types.RestartAction{Service: meta.serviceName}
		err = f.structParser(data, restartAction, path)
		action = restartAction
	case "start":
		startAction := &types.StartAction{Service: meta.serviceName}
		err = f.structParser(data, startAction, path)
		action = startAction
	case "stop":
		stopAction := &types.StopAction{Service: meta.serviceName}
		err = f.structParser(data, stopAction, path)
		action = stopAction
	default:
		return nil, errors.Errorf("unknown action %s at %s", meta.actionName, f.loc.String())
	}

	if err != nil {
		return nil, err
	}
	if meta.customName != "" && !customNameAllowed {
		return nil, errors.Errorf("custom name not allowed for action %s at %s", meta.actionName, f.loc.String())
	}
	if meta.serviceName != "" && !serviceNameAllowed {
		return nil, errors.Errorf("service name not allowed for action %s at %s", meta.actionName, f.loc.String())
	}

	return action, nil
}

func (f *NativeActionsFactory) parseActionFromMap(action map[string]interface{}, path string) (types.Action, error) {
	if len(action) > 1 {
		return nil, errors.Errorf("invalid action %s format - exactly one item in map is required", f.loc.String())
	}
	for name, value := range action {
		f.loc.SetField(name)
		valueMap, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("invalid action %s format - action value must be an object", f.loc.String())
		}
		return f.parseAction(name, valueMap, path)
	}
	return nil, errors.Errorf("invalid action %s format - empty object is not valid action", f.loc.String())
}

func (f *NativeActionsFactory) ParseActions(actions []interface{}, path string) ([]types.Action, error) {
	var parsedActions []types.Action
	f.loc.StartArray()
	for i, untypedAction := range actions {
		f.loc.SetIndex(i)
		f.loc.StartObject()
		switch action := untypedAction.(type) {
		case string:
			f.loc.SetField(action)
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
			return nil, errors.Errorf("unsupported action %s type %T", f.loc.String(), untypedAction)
		}
		f.loc.EndObject()
	}
	f.loc.EndArray()
	return parsedActions, nil
}
