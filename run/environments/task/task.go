package task

import "github.com/bukka/wst/run/environments/environment/providers"

type Task interface {
	Type() providers.Type
	BaseUrl() string
}
