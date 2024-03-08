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
	"errors"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/hooks"
	"github.com/bukka/wst/run/sandboxes/sandbox"
)

type Maker struct {
	fnd        app.Foundation
	hooksMaker *hooks.Maker
}

func CreateMaker(fnd app.Foundation, hooksMaker *hooks.Maker) *Maker {
	return &Maker{
		fnd:        fnd,
		hooksMaker: hooksMaker,
	}
}

func (m *Maker) MakeSandbox(config *types.CommonSandbox) (*Sandbox, error) {
	sandboxHooks, err := m.hooksMaker.MakeHooks(config.Hooks)
	if err != nil {
		return nil, err
	}

	return &Sandbox{
		dirs:  config.Dirs,
		hooks: sandboxHooks,
	}, nil
}

type Sandbox struct {
	dirs  map[string]string
	hooks map[hooks.HookType]hooks.Hook
}

func (s *Sandbox) Dirs() map[string]string {
	return s.dirs
}

func (s *Sandbox) Hooks() map[hooks.HookType]hooks.Hook {
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
