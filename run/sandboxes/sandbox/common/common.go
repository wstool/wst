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
	sandboxHooks := map[hooks.HookType]hooks.Hook{}
	for name, hookConfig := range config.Hooks {
		hook, err := m.hooksMaker.MakeHook(hookConfig)
		if err != nil {
			return nil, err
		}
		sandboxHooks[hooks.HookType(name)] = hook
	}

	return &Sandbox{
		Dirs:  config.Dirs,
		Hooks: sandboxHooks,
	}, nil
}

type Sandbox struct {
	Dirs  map[string]string
	Hooks map[hooks.HookType]hooks.Hook
}

func (s *Sandbox) Hook(hookType hooks.HookType) (hooks.Hook, error) {
	hook, ok := s.Hooks[hookType]
	if !ok {
		return nil, errors.New("hook not found")
	}
	return hook, nil
}
