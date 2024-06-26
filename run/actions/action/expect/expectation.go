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
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"time"
)

type Maker interface {
	MakeCustomAction(
		config *types.CustomExpectationAction,
		sl services.ServiceLocator,
		defaultTimeout int,
	) (action.Action, error)
	MakeMetricsAction(
		config *types.MetricsExpectationAction,
		sl services.ServiceLocator,
		defaultTimeout int,
	) (action.Action, error)
	MakeOutputAction(
		config *types.OutputExpectationAction,
		sl services.ServiceLocator,
		defaultTimeout int,
	) (action.Action, error)
	MakeResponseAction(
		config *types.ResponseExpectationAction,
		sl services.ServiceLocator,
		defaultTimeout int,
	) (action.Action, error)
}

type ExpectationActionMaker struct {
	fnd               app.Foundation
	expectationsMaker expectations.Maker
	parametersMaker   parameters.Maker
}

func CreateExpectationActionMaker(
	fnd app.Foundation,
	expectationsMaker expectations.Maker,
	parametersMaker parameters.Maker,
) *ExpectationActionMaker {
	return &ExpectationActionMaker{
		fnd:               fnd,
		parametersMaker:   parametersMaker,
		expectationsMaker: expectationsMaker,
	}
}

func (m *ExpectationActionMaker) MakeCommonExpectation(
	sl services.ServiceLocator,
	serviceName string,
	timeout,
	defaultTimeout int,
) (*CommonExpectation, error) {
	svc, err := sl.Find(serviceName)
	if err != nil {
		return nil, err
	}

	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &CommonExpectation{
		fnd:     m.fnd,
		service: svc,
		timeout: time.Duration(timeout * 1e6),
	}, nil
}

type CommonExpectation struct {
	fnd     app.Foundation
	service services.Service
	timeout time.Duration
}
