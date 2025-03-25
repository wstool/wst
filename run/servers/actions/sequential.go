// Copyright 2025 Jakub Zelenka and The WST Authors
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

package actions

import (
	"github.com/wstool/wst/conf/types"
)

type SequentialAction interface {
	Actions() []types.Action
}

type nativeSequentialAction struct {
	actions []types.Action
}

func (a *nativeSequentialAction) Actions() []types.Action {
	return a.actions
}

func (m *nativeMaker) makeSequentialActions(configActions map[string]types.ServerSequentialAction) map[string]SequentialAction {
	sequentialActions := make(map[string]SequentialAction, len(configActions))
	for key, configAction := range configActions {
		sequentialActions[key] = &nativeSequentialAction{
			actions: configAction.Actions,
		}
	}
	return sequentialActions
}
