package conf

import "reflect"

type factoryFunc func(data interface{}, fieldValue reflect.Value) error

func createActions(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func createContainerImage(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func createExpectations(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func createHooks(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func createParameters(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func createSandboxes(data interface{}, fieldValue reflect.Value) error {
	return nil
}

func getFactories() map[string]factoryFunc {
	return map[string]factoryFunc{
		"createActions":        createActions,
		"createContainerImage": createContainerImage,
		"createExpectations":   createExpectations,
		"createHooks":          createHooks,
		"createParameters":     createParameters,
		"createSandboxes":      createSandboxes,
	}
}
