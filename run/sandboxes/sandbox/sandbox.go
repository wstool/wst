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

package sandbox

import (
	"github.com/bukka/wst/run/sandboxes/hooks"
)

type DirType string

const (
	ConfDirType   DirType = "conf"
	RunDirType            = "run"
	ScriptDirType         = "script"
)

type ContainerConfig struct {
	ImageName        string
	ImageTag         string
	RegistryUsername string
	RegistryPassword string
}

func (c *ContainerConfig) Image() string {
	return c.ImageName + ":" + c.ImageTag
}

type Sandbox interface {
	Available() bool
	Dirs() map[DirType]string
	Dir(dirType DirType) (string, error)
	Hooks() map[hooks.HookType]hooks.Hook
	Hook(hookType hooks.HookType) (hooks.Hook, error)
	ContainerConfig() (*ContainerConfig, error)
	Inherit(parentSandbox Sandbox) error
}
