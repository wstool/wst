package instances

import (
	"github.com/bukka/wst/actions"
	"github.com/bukka/wst/instances/runtime"
)

type Instance interface {
	ExecuteActions(runData *runtime.Data) error
}

type nativeInstance struct {
	actions []actions.Action
	runData *runtime.Data
}

func (i *nativeInstance) ExecuteActions() error {
	for _, action := range i.actions {
		err := action.Execute(i.runData)
		if err != nil {
			return err
		}
	}
	return nil
}
