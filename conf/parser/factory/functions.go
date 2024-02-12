package factory

import (
	"fmt"
	"github.com/bukka/wst/app"
	"reflect"
)

type Func func(data interface{}, fieldValue reflect.Value) error

// Functions define an interface for factory functions
type Functions interface {
	GetFactoryFunc(funcName string) Func
}

// FuncProvider contains the implementation of the FactoryFunctions
type FuncProvider struct {
	env            app.Env
	actionsFactory ActionsFactory
}

func CreateFactories(env app.Env) Functions {
	return &FuncProvider{
		env:            env,
		actionsFactory: CreateActionsFactory(env),
	}
}

func (f *FuncProvider) GetFactoryFunc(funcName string) Func {
	switch funcName {
	case "createActions":
		return f.createActions
	case "createContainerImage":
		return f.createContainerImage
	case "createExpectations":
		return f.createExpectations
	case "createHooks":
		return f.createHooks
	case "createParameters":
		return f.createParameters
	case "createSandboxes":
		return f.createSandboxes
	case "createServiceScripts":
		return f.createServiceScripts
	default:
		return nil
	}
}

// Define your factory functions as methods of FactoryFuncProvider struct
func (f *FuncProvider) createActions(data interface{}, fieldValue reflect.Value) error {
	// Check if data is a slice
	dataSlice, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("data must be an array, got %T", data)
	}

	actions, err := f.actionsFactory.ParseActions(dataSlice)
	if err != nil {
		return err
	}

	fieldValue.Set(reflect.ValueOf(actions))

	return nil
}

func (f *FuncProvider) createContainerImage(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func (f *FuncProvider) createExpectations(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func (f *FuncProvider) createHooks(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func (f *FuncProvider) createParameters(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func (f *FuncProvider) createSandboxes(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func (f *FuncProvider) createServiceScripts(data interface{}, fieldValue reflect.Value) error {
	return nil
}
