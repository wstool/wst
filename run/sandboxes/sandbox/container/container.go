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

// This is an abstract provider

package container

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

func (m *Maker) MakeSandbox(config *types.ContainerSandbox) (*Sandbox, error) {
	commonSandbox, err := m.commonMaker.MakeSandbox(&config.CommonSandbox)
	if err != nil {
		return nil, err
	}

	sandbox := &Sandbox{
		Sandbox:          *commonSandbox,
		ImageName:        config.Image.Name,
		ImageTag:         config.Image.Tag,
		RegistryUsername: config.Registry.Auth.Username,
		RegistryPassword: config.Registry.Auth.Password,
	}

	return sandbox, nil
}

type Sandbox struct {
	common.Sandbox
	ImageName        string
	ImageTag         string
	RegistryUsername string
	RegistryPassword string
}

func (s *Sandbox) InheritContainer(parentSandbox *Sandbox) error {
	s.Sandbox.Inherit(parentSandbox)
	if s.ImageName == "" {
		s.ImageName = parentSandbox.ImageName
	}
	if s.ImageTag == "" {
		s.ImageTag = parentSandbox.ImageTag
	}
	if s.RegistryUsername == "" {
		s.RegistryUsername = parentSandbox.RegistryUsername
	}
	if s.RegistryPassword == "" {
		s.RegistryPassword = parentSandbox.RegistryPassword
	}

	return nil
}
