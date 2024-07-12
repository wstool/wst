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

package parameters

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/parameters/parameter"
)

type Parameters map[string]parameter.Parameter

type Maker interface {
	Make(config types.Parameters) (Parameters, error)
}

type nativeMaker struct {
	fnd            app.Foundation
	parameterMaker parameter.Maker
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{
		fnd:            fnd,
		parameterMaker: parameter.CreateMaker(fnd),
	}
}

func (m *nativeMaker) Make(config types.Parameters) (Parameters, error) {
	params := make(map[string]parameter.Parameter)
	for key, elem := range config {
		if paramElem, err := m.parameterMaker.Make(elem); err == nil {
			params[key] = paramElem
		} else {
			return nil, err
		}
	}

	return params, nil
}

func (p Parameters) Inherit(parameters Parameters) Parameters {
	for key, value := range parameters {
		if _, ok := p[key]; !ok {
			p[key] = value
		}
	}
	return p
}
