package instances

import (
	"github.com/bukka/wst/actions"
	"github.com/bukka/wst/instances/runtime"
)

type Instance struct {
	actions []actions.Action
}

func (i *Instance) ExecuteActions(runData *runtime.Data) error {
	for _, action := range i.actions {
		err := action.Execute(runData)
		if err != nil {
			return err
		}
	}
	return nil
}
