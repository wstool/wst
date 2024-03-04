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
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/sandbox/common"
)

type Maker struct {
	fnd         app.Foundation
	commonMaker *common.Maker
}

func CreateMaker(fnd app.Foundation, commonMaker *common.Maker) *Maker {
	return &Maker{
		fnd:         fnd,
		commonMaker: commonMaker,
	}
}

func (m *Maker) MakeSandbox(config *types.LocalSandbox) (*Sandbox, error) {
	commonSandbox, err := m.commonMaker.MakeSandbox(&config.CommonSandbox)
	if err != nil {
		return nil, err
	}

	sandbox := &Sandbox{
		Sandbox: *commonSandbox,
	}

	return sandbox, nil
}

type Sandbox struct {
	common.Sandbox
}
