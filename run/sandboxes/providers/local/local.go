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

package local

import (
	"bufio"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/providers/common"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/sandboxes/sandbox/hooks"
)

type Maker struct {
	env         app.Env
	commonMaker *common.Maker
}

func CreateMaker(env app.Env, commonMaker *common.Maker) *Maker {
	return &Maker{
		env:         env,
		commonMaker: commonMaker,
	}
}

func (m *Maker) MakeSandbox(config *types.LocalSandbox) (*Sandbox, error) {
	panic("implement")
}

type Sandbox struct {
	common.Sandbox
}

func (s Sandbox) OutputScanner(outputType sandbox.OutputType) *bufio.Scanner {
	//TODO implement me
	panic("implement me")
}

func (s Sandbox) ExecuteCommand(command *hooks.HookCommand) error {
	//TODO implement me
	panic("implement me")
}

func (s Sandbox) ExecuteSignal(signal *hooks.HookSignal) error {
	//TODO implement me
	panic("implement me")
}
