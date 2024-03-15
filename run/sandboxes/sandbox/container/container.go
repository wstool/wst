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

func (m *Maker) MakeSandbox(config *types.ContainerSandbox) (*Sandbox, error) {
	commonSandbox, err := m.commonMaker.MakeSandbox(&config.CommonSandbox)
	if err != nil {
		return nil, err
	}

	sandbox := &Sandbox{
		Sandbox: *commonSandbox,
		config: sandbox.ContainerConfig{
			ImageName:        config.Image.Name,
			ImageTag:         config.Image.Tag,
			RegistryUsername: config.Registry.Auth.Username,
			RegistryPassword: config.Registry.Auth.Password,
		},
	}

	return sandbox, nil
}

type Sandbox struct {
	common.Sandbox
	config sandbox.ContainerConfig
}

func (s *Sandbox) ContainerConfig() (*sandbox.ContainerConfig, error) {
	return &s.config, nil
}

func (s *Sandbox) Inherit(parentSandbox sandbox.Sandbox) error {
	err := s.Sandbox.Inherit(parentSandbox)
	if err != nil {
		return err
	}
	containerConfig, err := parentSandbox.ContainerConfig()
	if err != nil {
		return err
	}

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

	return nil
}
