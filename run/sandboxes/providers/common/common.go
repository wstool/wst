// This is an abstract provider

package common

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/sandboxes/sandbox/hooks"
)

type Maker struct {
	env        app.Env
	hooksMaker *hooks.Maker
}

func CreateMaker(env app.Env, hooksMaker *hooks.Maker) *Maker {
	return &Maker{
		env:        env,
		hooksMaker: hooksMaker,
	}
}

func (m *Maker) MakeSandbox(config *types.CommonSandbox) (*Sandbox, error) {
	panic("implement")
}

type Sandbox struct {
	Dirs  map[string]string
	Hooks map[string]hooks.Hook
}
