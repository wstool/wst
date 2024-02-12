package sandboxes

import "github.com/bukka/wst/conf/types"

type Sandbox interface {
}

type Sandboxes map[string]Sandbox

func MakeSandboxes(config *types.Config) (Sandboxes, error) {
	//TODO implement me
	panic("implement me")
}
