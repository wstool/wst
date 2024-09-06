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
	"github.com/wstool/wst/run/sandboxes/containers"
	"github.com/wstool/wst/run/sandboxes/dir"
	"github.com/wstool/wst/run/sandboxes/hooks"
)

type Sandbox interface {
	Available() bool
	Dirs() map[dir.DirType]string
	Dir(dirType dir.DirType) (string, error)
	Hooks() hooks.Hooks
	Hook(hookType hooks.HookType) (hooks.Hook, error)
	ContainerConfig() *containers.ContainerConfig
	Inherit(parentSandbox Sandbox) error
}
