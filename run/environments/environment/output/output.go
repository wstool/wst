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

package output

import (
	"github.com/wstool/wst/app"
)

type Type int

const (
	Stdout Type = 1
	Stderr      = 2
	Any         = 3
)

type Maker interface {
	MakeCollector(tid string) Collector
}

type nativeMaker struct {
	fnd app.Foundation
}

func (m *nativeMaker) MakeCollector(tid string) Collector {
	return NewBufferedCollector(m.fnd, tid)
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{fnd: fnd}
}
