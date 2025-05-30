// Copyright 2024-2025 Jakub Zelenka and The WST Authors
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
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/parameters"
)

type Actions struct {
	Expect     map[string]ExpectAction
	Sequential map[string]SequentialAction
}

func (a *Actions) Inherit(parentActions *Actions) {
	for expectationName, expectation := range parentActions.Expect {
		_, ok := a.Expect[expectationName]
		if !ok {
			a.Expect[expectationName] = expectation
		}
	}
	for sequentialName, sequential := range parentActions.Sequential {
		_, ok := a.Sequential[sequentialName]
		if !ok {
			a.Sequential[sequentialName] = sequential
		}
	}
}

type Maker interface {
	Make(configActions *types.ServerActions) (*Actions, error)
}

type nativeMaker struct {
	fnd               app.Foundation
	expectationsMaker expectations.Maker
	parametersMaker   parameters.Maker
}

func CreateMaker(
	fnd app.Foundation,
	expectationsMaker expectations.Maker,
	parametersMaker parameters.Maker,
) Maker {
	return &nativeMaker{
		fnd:               fnd,
		parametersMaker:   parametersMaker,
		expectationsMaker: expectationsMaker,
	}
}

func (m *nativeMaker) Make(configActions *types.ServerActions) (*Actions, error) {
	expectActions, err := m.makeExpectActions(configActions.Expect)
	if err != nil {
		return nil, err
	}

	return &Actions{
		Expect:     expectActions,
		Sequential: m.makeSequentialActions(configActions.Sequential),
	}, nil
}
