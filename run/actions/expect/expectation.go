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
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"time"
)

type ExpectationActionMaker struct {
	fnd             app.Foundation
	parametersMaker *parameters.Maker
}

func CreateExpectationActionMaker(fnd app.Foundation, parametersMaker *parameters.Maker) *ExpectationActionMaker {
	return &ExpectationActionMaker{
		fnd:             fnd,
		parametersMaker: parametersMaker,
	}
}

func (m *ExpectationActionMaker) MakeCommonExpectation(
	svcs services.Services,
	serviceName string,
	timeout,
	defaultTimeout int,
) (*CommonExpectation, error) {
	svc, err := svcs.FindService(serviceName)
	if err != nil {
		return nil, err
	}

	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &CommonExpectation{
		service: svc,
		timeout: time.Duration(timeout),
	}, nil
}

type CommonExpectation struct {
	service services.Service
	timeout time.Duration
}
