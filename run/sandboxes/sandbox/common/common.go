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

package common

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/sandboxes/containers"
	"github.com/wstool/wst/run/sandboxes/dir"
	"github.com/wstool/wst/run/sandboxes/hooks"
	"github.com/wstool/wst/run/sandboxes/sandbox"
)

type Maker interface {
	MakeSandbox(config *types.CommonSandbox) (*Sandbox, error)
}

type nativeMaker struct {
	fnd        app.Foundation
	hooksMaker hooks.Maker
}

func CreateMaker(fnd app.Foundation, hooksMaker hooks.Maker) Maker {
	return &nativeMaker{
		fnd:        fnd,
		hooksMaker: hooksMaker,
	}
}

func (m *nativeMaker) MakeSandbox(config *types.CommonSandbox) (*Sandbox, error) {
	sandboxHooks, err := m.hooksMaker.MakeHooks(config.Hooks)
	if err != nil {
		return nil, err
	}

	sandboxDirs := make(map[dir.DirType]string)
	for dirTypeStr, dirPath := range config.Dirs {
		dirType := dir.DirType(dirTypeStr)
		if dirType != dir.CertDirType && dirType != dir.ConfDirType && dirType != dir.RunDirType && dirType != dir.ScriptDirType {
			return nil, errors.Errorf("invalid dir type: %v", dirType)
		}
		sandboxDirs[dirType] = dirPath
	}

	return CreateSandbox(config.Available, sandboxDirs, sandboxHooks), nil
}

func CreateSandbox(available bool, dirs map[dir.DirType]string, hooks hooks.Hooks) *Sandbox {
	return &Sandbox{
		available: available,
		dirs:      dirs,
		hooks:     hooks,
	}
}

type Sandbox struct {
	available bool
	dirs      map[dir.DirType]string
	hooks     hooks.Hooks
}

func (s *Sandbox) ContainerConfig() *containers.ContainerConfig {
	return nil
}

func (s *Sandbox) Available() bool {
	return s.available
}

func (s *Sandbox) Dirs() map[dir.DirType]string {
	return s.dirs
}

func (s *Sandbox) Dir(dirType dir.DirType) (string, error) {
	dir, ok := s.dirs[dirType]
	if !ok {
		return "", fmt.Errorf("directory not found for dir type: %v", dirType)
	}
	return dir, nil
}

func (s *Sandbox) Hooks() hooks.Hooks {
	return s.hooks
}

func (s *Sandbox) Hook(hookType hooks.HookType) (hooks.Hook, error) {
	hook, ok := s.hooks[hookType]
	if !ok {
		return nil, errors.New("hook not found")
	}
	return hook, nil
}

func (s *Sandbox) Inherit(parentSandbox sandbox.Sandbox) error {
	if s.hooks == nil {
		return errors.New("sandbox hooks not set")
	}
	if s.dirs == nil {
		return errors.New("sandbox dirs not set")
	}
	// Inherit hooks.
	for hookType, parentHook := range parentSandbox.Hooks() {
		_, ok := s.hooks[hookType]
		if !ok {
			s.hooks[hookType] = parentHook
		}
	}
	// Inherit Dirs.
	for dirName, dirPath := range parentSandbox.Dirs() {
		_, ok := s.dirs[dirName]
		if !ok {
			s.dirs[dirName] = dirPath
		}
	}

	return nil
}
