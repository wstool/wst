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

package docker

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/sandboxes/sandbox/container"
)

type Maker struct {
	fnd            app.Foundation
	containerMaker *container.Maker
}

func CreateMaker(fnd app.Foundation, containerMaker *container.Maker) *Maker {
	return &Maker{
		fnd:            fnd,
		containerMaker: containerMaker,
	}
}

func (m *Maker) MakeSandbox(config *types.DockerSandbox) (*Sandbox, error) {
	commonSandbox, err := m.containerMaker.MakeSandbox(&config.ContainerSandbox)
	if err != nil {
		return nil, err
	}

	sandbox := &Sandbox{
		Sandbox: *commonSandbox,
	}

	return sandbox, nil
}

type Sandbox struct {
	container.Sandbox
}

func (s *Sandbox) Inherit(parentSandbox sandbox.Sandbox) error {
	dockerParentSandbox, ok := parentSandbox.(*Sandbox)
	if !ok {
		return fmt.Errorf("invalid docker parent sandbox type %t", parentSandbox)
	}

	return s.Sandbox.InheritContainer(&dockerParentSandbox.Sandbox)
}
