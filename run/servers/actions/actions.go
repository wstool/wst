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

package actions

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/expect"
	"github.com/bukka/wst/run/parameters"
)

type Actions struct {
	Expect map[string]ExpectAction
}

type Maker struct {
	fnd             app.Foundation
	parametersMaker *parameters.Maker
	expectMaker     *expect.ExpectationActionMaker
}

func CreateMaker(fnd app.Foundation, parametersMaker *parameters.Maker) *Maker {
	return &Maker{
		fnd:             fnd,
		parametersMaker: parametersMaker,
		expectMaker:     expect.CreateExpectationActionMaker(fnd),
	}
}

func (m *Maker) Make(configActions *types.ServerActions) (*Actions, error) {
	expectActions, err := m.makeExpectActions(configActions.Expect)
	if err != nil {
		return nil, err
	}

	return &Actions{
		Expect: expectActions,
	}, nil
}

type nativeConfig struct {
	file       string
	parameters types.Parameters
}
