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
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/sandboxes/containers"
	"github.com/wstool/wst/run/sandboxes/sandbox"
	"github.com/wstool/wst/run/sandboxes/sandbox/common"
)

type Maker interface {
	MakeSandbox(config *types.ContainerSandbox) (*Sandbox, error)
}

type nativeMaker struct {
	fnd         app.Foundation
	commonMaker common.Maker
}

func CreateMaker(fnd app.Foundation, commonMaker common.Maker) Maker {
	return &nativeMaker{
		fnd:         fnd,
		commonMaker: commonMaker,
	}
}

func (m *nativeMaker) MakeSandbox(config *types.ContainerSandbox) (*Sandbox, error) {
	commonSandbox, err := m.commonMaker.MakeSandbox(&types.CommonSandbox{
		Dirs:      config.Dirs,
		Hooks:     config.Hooks,
		Available: config.Available,
	})
	if err != nil {
		return nil, err
	}

	sb := CreateSandbox(commonSandbox, &containers.ContainerConfig{
		ImageName:        config.Image.Name,
		ImageTag:         config.Image.Tag,
		RegistryUsername: config.Registry.Auth.Username,
		RegistryPassword: config.Registry.Auth.Password,
	})

	return sb, nil
}

func CreateSandbox(commonSandbox *common.Sandbox, config *containers.ContainerConfig) *Sandbox {
	return &Sandbox{
		Sandbox: *commonSandbox,
		config:  *config,
	}
}

type Sandbox struct {
	common.Sandbox
	config containers.ContainerConfig
}

func (s *Sandbox) ContainerConfig() *containers.ContainerConfig {
	return &s.config
}

func (s *Sandbox) Inherit(parentSandbox sandbox.Sandbox) error {
	err := s.Sandbox.Inherit(parentSandbox)
	if err != nil {
		return err
	}
	containerConfig := parentSandbox.ContainerConfig()

	if containerConfig != nil {
		if s.config.ImageName == "" {
			s.config.ImageName = containerConfig.ImageName
		}
		if s.config.ImageTag == "" {
			s.config.ImageTag = containerConfig.ImageTag
		}
		if s.config.RegistryUsername == "" {
			s.config.RegistryUsername = containerConfig.RegistryUsername
		}
		if s.config.RegistryPassword == "" {
			s.config.RegistryPassword = containerConfig.RegistryPassword
		}
	}

	return nil
}
