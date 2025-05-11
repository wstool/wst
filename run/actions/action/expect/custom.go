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
	"github.com/pkg/errors"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/services"
	"time"
)

func (m *ExpectationActionMaker) MakeCustomAction(
	config *types.CustomExpectationAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	commonExpectation, err := m.MakeCommonExpectation(sl, config.Service, config.Timeout, defaultTimeout, config.When)
	if err != nil {
		return nil, err
	}

	server := commonExpectation.service.Server()
	customName := config.Custom.Name
	expectation, ok := server.ExpectAction(customName)
	if !ok {
		return nil, errors.Errorf("expectation action %s not found", customName)
	}

	configParameters, err := m.parametersMaker.Make(config.Custom.Parameters)
	if err != nil {
		return nil, err
	}

	return &customAction{
		CommonExpectation:   commonExpectation,
		OutputExpectation:   expectation.OutputExpectation(),
		ResponseExpectation: expectation.ResponseExpectation(),
		parameters:          configParameters.Inherit(expectation.Parameters()).Inherit(commonExpectation.service.ServerParameters()),
	}, nil
}

type customAction struct {
	*CommonExpectation
	*expectations.OutputExpectation
	*expectations.ResponseExpectation
	parameters parameters.Parameters
}

func (a *customAction) When() action.When {
	return a.when
}

func (a *customAction) Timeout() time.Duration {
	return a.timeout
}

func (a *customAction) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	if a.OutputExpectation != nil {
		oa := outputAction{
			CommonExpectation: a.CommonExpectation,
			OutputExpectation: a.OutputExpectation,
			parameters:        a.parameters,
		}
		return oa.Execute(ctx, runData)
	}
	if a.ResponseExpectation != nil {
		ra := responseAction{
			CommonExpectation:   a.CommonExpectation,
			ResponseExpectation: a.ResponseExpectation,
			parameters:          a.parameters,
		}
		return ra.Execute(ctx, runData)
	}
	return false, errors.Errorf("no expectation set")
}
