package task

import "github.com/bukka/wst/run/environments/environment/providers"

type Task interface {
	Name() string
	Executable() string
	Id() string
	Pid() int
	Type() providers.Type
	PrivateUrl() string
	PublicUrl() string
}
