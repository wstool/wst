package factory

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
)

type ActionsFactory interface {
	ParseActions(actions []interface{}) ([]types.Action, error)
}

type NativeActionsFactory struct {
	env app.Env
}

func CreateActionsFactory(env app.Env) ActionsFactory {
	return &NativeActionsFactory{
		env: env,
	}
}

func (f *NativeActionsFactory) parseActionFromMap(action map[string]interface{}) (types.Action, error) {
	return nil, nil
}

func (f *NativeActionsFactory) parseActionFromString(actionString string) (types.Action, error) {
	return nil, nil
}

func (f *NativeActionsFactory) ParseActions(actions []interface{}) ([]types.Action, error) {
	var parsedActions []types.Action
	for _, action := range actions {
		switch action := action.(type) {
		case string:
			parsedAction, err := f.parseActionFromString(action)
			if err != nil {
				return nil, err
			}
			parsedActions = append(parsedActions, parsedAction)
		case map[string]interface{}:
			parsedAction, err := f.parseActionFromMap(action)
			if err != nil {
				return nil, err
			}
			parsedActions = append(parsedActions, parsedAction)
		default:
			return nil, fmt.Errorf("unsupported action type")
		}
	}
	return parsedActions, nil
}
