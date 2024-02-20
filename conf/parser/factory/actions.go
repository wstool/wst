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
)

type ActionsFactory interface {
	ParseActions(actions []interface{}) ([]types.Action, error)
}

type NativeActionsFactory struct {
	env app.Env
}

func CreateActionsFactory(env app.Env) ActionsFactory {
	return &NativeActionsFactory{
		env: env,
	}
}

func (f *NativeActionsFactory) parseActionFromMap(action map[string]interface{}) (types.Action, error) {
	return nil, nil
}

func (f *NativeActionsFactory) parseActionFromString(actionString string) (types.Action, error) {
	return nil, nil
}

func (f *NativeActionsFactory) ParseActions(actions []interface{}) ([]types.Action, error) {
	var parsedActions []types.Action
	for _, action := range actions {
		switch action := action.(type) {
		case string:
			parsedAction, err := f.parseActionFromString(action)
			if err != nil {
				return nil, err
			}
			parsedActions = append(parsedActions, parsedAction)
		case map[string]interface{}:
			parsedAction, err := f.parseActionFromMap(action)
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
