package actions

import (
	"github.com/bukka/wst/instances/runtime"
)

type Action interface {
	Execute(runData *runtime.Data) error
}
