package conf

import "reflect"

type factoryFunc func(data interface{}, fieldValue reflect.Value) (interface{}, error)

func createActions(data interface{}, fieldValue reflect.Value) (interface{}, error) {
	return nil, nil
}

func createContainerImage(data interface{}, fieldValue reflect.Value) (interface{}, error) {
	return nil, nil
}

func createExpectations(data interface{}, fieldValue reflect.Value) (interface{}, error) {
	return nil, nil
}

func createHooks(data interface{}, fieldValue reflect.Value) (interface{}, error) {
	return nil, nil
}

func createParameters(data interface{}, fieldValue reflect.Value) (interface{}, error) {
	return nil, nil
}

func createSandboxes(data interface{}, fieldValue reflect.Value) (interface{}, error) {
	return nil, nil
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
