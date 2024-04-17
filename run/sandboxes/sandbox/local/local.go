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
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/sandbox"
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
	commonSandbox, err := m.commonMaker.MakeSandbox(&types.CommonSandbox{
		Dirs:      config.Dirs,
		Hooks:     config.Hooks,
		Available: config.Available,
	})
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

func (s *Sandbox) ContainerConfig() (*sandbox.ContainerConfig, error) {
	return nil, fmt.Errorf("local sandbox does not have a container config")
}
