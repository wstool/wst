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

package processor

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
)

type Processor interface {
	ProcessConfig(config *types.Config) error
}

type nativeProcessor struct {
	env app.Foundation
}

func CreateProcessor(env app.Foundation) Processor {
	return &nativeProcessor{
		env: env,
	}
}

func (p *nativeProcessor) ProcessConfig(config *types.Config) error {
	//TODO implement me
	panic("implement me")
}
