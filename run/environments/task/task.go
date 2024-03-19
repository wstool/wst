package task

import "github.com/bukka/wst/run/environments/environment/providers"

type Task interface {
	Name() string
	Id() string
	Type() providers.Type
	PrivateUrl() string
	PublicUrl() string
}
