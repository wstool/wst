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

package expectations

import (
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/parameters"
)

type Maker interface {
	MakeMetricsExpectation(config *types.MetricsExpectation) (*MetricsExpectation, error)
	MakeOutputExpectation(config *types.OutputExpectation) (*OutputExpectation, error)
	MakeResponseExpectation(config *types.ResponseExpectation) (*ResponseExpectation, error)
}

type nativeMaker struct {
	fnd             app.Foundation
	parametersMaker parameters.Maker
}

func CreateMaker(fnd app.Foundation, parametersMaker parameters.Maker) Maker {
	return &nativeMaker{
		fnd:             fnd,
		parametersMaker: parametersMaker,
	}
}
