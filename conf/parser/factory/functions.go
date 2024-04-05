// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package factory

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

type Func func(data interface{}, fieldValue reflect.Value, path string) error

type StructParser func(data map[string]interface{}, structure interface{}, path string) error

// Functions define an interface for factory functions
type Functions interface {
	GetFactoryFunc(funcName string) (Func, error)
}

// FuncProvider contains the implementation of the FactoryFunctions
type FuncProvider struct {
	fnd            app.Foundation
	actionsFactory ActionsFactory
	structParser   StructParser
}

func CreateFactories(fnd app.Foundation, structParser StructParser) Functions {
	return &FuncProvider{
		fnd:            fnd,
		actionsFactory: CreateActionsFactory(fnd, structParser),
		structParser:   structParser,
	}
}

func (f *FuncProvider) GetFactoryFunc(funcName string) (Func, error) {
	switch funcName {
	case "createActions":
		return f.createActions, nil
	case "createContainerImage":
		return f.createContainerImage, nil
	case "createEnvironments":
		return f.createEnvironments, nil
	case "createHooks":
		return f.createHooks, nil
	case "createParameters":
		return f.createParameters, nil
	case "createSandboxes":
		return f.createSandboxes, nil
	case "createServerExpectations":
		return f.createServerExpectations, nil
	case "createServiceScripts":
		return f.createServiceScripts, nil
	default:
		return nil, errors.Errorf("unknown function %s", funcName)
	}
}

// Define your factory functions as methods of FactoryFuncProvider struct
func (f *FuncProvider) createActions(data interface{}, fieldValue reflect.Value, path string) error {
	// Check if data is a slice
	dataSlice, ok := data.([]interface{})
	if !ok {
		return errors.Errorf("data must be an array, got %T", data)
	}

	actions, err := f.actionsFactory.ParseActions(dataSlice, path)
	if err != nil {
		return err
	}

	fieldValue.Set(reflect.ValueOf(actions))

	return nil
}

func (f *FuncProvider) createContainerImage(data interface{}, fieldValue reflect.Value, path string) error {
	var img types.ContainerImage
	switch v := data.(type) {
	case string:
		parts := strings.SplitN(v, ":", 2)
		img.Name = parts[0]
		if len(parts) == 2 {
			img.Tag = parts[1]
		} else {
			img.Tag = "latest"
		}
	case map[string]interface{}:
		err := f.structParser(v, img, path)
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("unsupported type for image data")
	}

	fieldValue.Set(reflect.ValueOf(img))

	return nil
}

type typeMapFactory[T any] func(key string) T

func processTypeMap[T any](
	name string,
	data interface{},
	factories map[string]typeMapFactory[T],
	structParser StructParser,
	path string,
) (map[string]T, error) {
	// Check if data is a map
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("data for %s must be a map, got %T", name, data)
	}

	result := make(map[string]T, len(dataMap))
	var factory typeMapFactory[T]
	var valMap map[string]interface{}
	for key, val := range dataMap {
		if factory, ok = factories[key]; ok {
			valMap, ok = val.(map[string]interface{})
			if !ok {
				return nil, errors.Errorf("data for value in %s must be a map, got %T", name, val)
			}
			structure := factory(key)
			if err := structParser(valMap, structure, path); err != nil {
				return nil, err
			}
			result[key] = structure
		} else {
			return nil, errors.Errorf("unknown environment type: %s", key)
		}
	}

	return result, nil
}

func (f *FuncProvider) createEnvironments(data interface{}, fieldValue reflect.Value, path string) error {
	environmentFactories := map[string]typeMapFactory[types.Environment]{
		"common": func(key string) types.Environment {
			return &types.CommonEnvironment{}
		},
		"local": func(key string) types.Environment {
			return &types.LocalEnvironment{}
		},
		"container": func(key string) types.Environment {
			return &types.ContainerEnvironment{}
		},
		"docker": func(key string) types.Environment {
			return &types.DockerEnvironment{}
		},
		"kubernetes": func(key string) types.Environment {
			return &types.KubernetesEnvironment{}
		},
	}
	environments, err := processTypeMap("environments", data, environmentFactories, f.structParser, path)
	if err != nil {
		return err
	}

	fieldValue.Set(reflect.ValueOf(environments))

	return nil
}

type hooksMapFactory func(hook interface{}) (types.SandboxHook, map[string]interface{}, error)

func (f *FuncProvider) createHooks(data interface{}, fieldValue reflect.Value, path string) error {
	// Check if data is a map
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors.Errorf("data for hooks must be a map, got %T", data)
	}

	hooksFactories := map[string]hooksMapFactory{
		"native": func(hook interface{}) (types.SandboxHook, map[string]interface{}, error) {
			hookNative, ok := hook.(map[string]interface{})
			if !ok {
				return nil, nil, errors.Errorf("native hook must be a map, got %T", hook)
			}
			return &types.SandboxHookNative{}, hookNative, nil
		},
		"command": func(hook interface{}) (types.SandboxHook, map[string]interface{}, error) {
			hookMap, ok := hook.(map[string]interface{})
			if !ok {
				return nil, nil, errors.Errorf("command hooks must be a map, got %T", hook)
			}

			if _, ok := hookMap["command"]; ok {
				return &types.SandboxHookShellCommand{}, hookMap, nil
			} else if _, ok := hookMap["executable"]; ok {
				return &types.SandboxHookArgsCommand{}, hookMap, nil
			} else {
				return nil, nil, errors.Errorf("command hooks data is invalid")
			}
		},
		"signal": func(hook interface{}) (types.SandboxHook, map[string]interface{}, error) {
			if strHook, ok := hook.(string); ok {
				return &types.SandboxHookSignal{
					IsString:    true,
					StringValue: strHook,
				}, nil, nil
			}
			if intHook, ok := hook.(int); ok {
				return &types.SandboxHookSignal{
					IsString: false,
					IntValue: intHook,
				}, nil, nil
			}
			return nil, nil, errors.Errorf("invalid signal hook type %t, only string and int is allowed", hook)
		},
	}

	hooks := make(map[string]types.SandboxHook, len(dataMap))
	for hookEvent, hookData := range dataMap {
		hookMap, ok := hookData.(map[string]interface{})
		if !ok {
			return errors.Errorf("hook data must be a map, got %T", hookData)
		}
		if hookMap == nil || len(hookMap) == 0 {
			return errors.New("hook data cannot be an empty map")
		}
		if len(hookMap) > 1 {
			return errors.New("hook data must have only one element")
		}
		for hookType, hook := range hookMap {
			factory, ok := hooksFactories[hookType]
			if ok {
				hookStructure, hookTypeData, err := factory(hook)
				if hookTypeData != nil {
					if err := f.structParser(hookTypeData, hookStructure, path); err != nil {
						return err
					}
				}
				if err != nil {
					return err
				}
				hooks[hookEvent] = hookStructure
			} else {
				return errors.Errorf("unknown hook type: %s", hookType)
			}
		}
	}

	fieldValue.Set(reflect.ValueOf(hooks))

	return nil

}

func convertParams(data map[string]interface{}) types.Parameters {
	params := make(types.Parameters)
	for key, val := range data {
		switch v := val.(type) {
		case map[string]interface{}:
			// Recursively process nested maps.
			params[key] = convertParams(v)
		case []interface{}:
			// Process each element in the slice.
			var slice []interface{}
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					// Recursively process nested maps within slices.
					slice = append(slice, convertParams(itemMap))
				} else {
					// Directly append other types.
					slice = append(slice, item)
				}
			}
			params[key] = slice
		default:
			// Directly assign all other types.
			params[key] = val
		}
	}
	return params
}

func (f *FuncProvider) createParameters(data interface{}, fieldValue reflect.Value, path string) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors.Errorf("data for parameters must be a map, got %T", data)
	}

	params := convertParams(dataMap)

	fieldValue.Set(reflect.ValueOf(params))

	return nil
}

func (f *FuncProvider) createSandboxes(data interface{}, fieldValue reflect.Value, path string) error {
	sandboxFactories := map[string]typeMapFactory[types.Sandbox]{
		"common": func(key string) types.Sandbox {
			return &types.CommonSandbox{}
		},
		"local": func(key string) types.Sandbox {
			return &types.LocalSandbox{}
		},
		"container": func(key string) types.Sandbox {
			return &types.ContainerSandbox{}
		},
		"docker": func(key string) types.Sandbox {
			return &types.DockerSandbox{}
		},
		"kubernetes": func(key string) types.Sandbox {
			return &types.KubernetesSandbox{}
		},
	}
	sandboxes, err := processTypeMap("sandboxes", data, sandboxFactories, f.structParser, path)
	if err != nil {
		return err
	}

	fieldValue.Set(reflect.ValueOf(sandboxes))

	return nil
}

func (f *FuncProvider) createServerExpectations(data interface{}, fieldValue reflect.Value, path string) error {
	// Check if data is a map
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors.Errorf("data for server action expectations must be a map, got %T", data)
	}
	expectations := make(map[string]types.ServerExpectationAction)
	var structure interface{}
	for key, val := range dataMap {
		expData, ok := val.(map[string]interface{})
		if !ok {
			return errors.Errorf("data for value in server action expectations must be a map, got %T", val)
		}
		parsed := false
		for expKey, _ := range expData {
			switch expKey {
			case "parameters":
				continue
			case "metrics":
				structure = &types.ServerMetricsExpectation{}
			case "output":
				structure = &types.ServerOutputExpectation{}
			case "response":
				structure = &types.ServerResponseExpectation{}
			default:
				return errors.Errorf("invalid server expectation key %s", expKey)
			}
			if parsed {
				return errors.Errorf("expectation cannot have multiple types - additional key %s", expKey)
			}
			if err := f.structParser(expData, structure, path); err != nil {
				return err
			}
			expectations[key] = structure
			parsed = true
		}
	}

	fieldValue.Set(reflect.ValueOf(expectations))

	return nil
}

func (f *FuncProvider) createServiceScripts(data interface{}, fieldValue reflect.Value, path string) error {
	serviceScripts := types.ServiceScripts{}
	boolVal, ok := data.(bool)
	if ok {
		serviceScripts.IncludeAll = boolVal
	} else {
		arrVal, ok := data.([]interface{})
		if !ok {
			return errors.Errorf("invalid services scripts type, expected bool or string array but got %T", data)
		}
		for idx, item := range arrVal {
			strVal, ok := item.(string)
			if !ok {
				return errors.Errorf("invalid services scripts item type at index %d, expected string but got %T", idx, item)
			}
			serviceScripts.IncludeList = append(serviceScripts.IncludeList, strVal)
		}
	}

	fieldValue.Set(reflect.ValueOf(serviceScripts))

	return nil
}
