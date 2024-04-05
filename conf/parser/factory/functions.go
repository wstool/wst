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
	GetFactoryFunc(funcName string) Func
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

func (f *FuncProvider) GetFactoryFunc(funcName string) Func {
	switch funcName {
	case "createActions":
		return f.createActions
	case "createContainerImage":
		return f.createContainerImage
	case "createEnvironments":
		return f.createEnvironments
	case "createHooks":
		return f.createHooks
	case "createParameters":
		return f.createParameters
	case "createSandboxes":
		return f.createSandboxes
	case "createServerExpectations":
		return f.createServerExpectations
	case "createServiceScripts":
		return f.createServiceScripts
	default:
		return nil
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

type hooksMapFactory func(key string, hook interface{}) (types.SandboxHook, error)

func (f *FuncProvider) createHooks(data interface{}, fieldValue reflect.Value, path string) error {
	// Check if data is a map
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors.Errorf("data for hooks must be a map, got %T", data)
	}

	hooksFactories := map[string]hooksMapFactory{
		"command": func(key string, hook interface{}) (types.SandboxHook, error) {
			hookMap, ok := hook.(map[string]interface{})
			if !ok {
				return nil, errors.Errorf("command hooks must be a map, got %T", hook)
			}

			if cmd, ok := hookMap["command"]; ok {
				shell, shellOk := hookMap["shell"].(string)
				command, commandOk := cmd.(string)
				if !commandOk {
					return nil, errors.Errorf("command must be a string")
				}
				if !shellOk {
					shell = "/bin/sh" // Default to /bin/sh if not specified
				}
				return &types.SandboxHookShellCommand{Command: command, Shell: shell}, nil
			} else if exe, ok := hookMap["executable"]; ok {
				executable, exeOk := exe.(string)
				if !exeOk {
					return nil, errors.Errorf("executable must be a string")
				}
				argsInterface, argsOk := hookMap["args"]
				var args []string
				if argsOk {
					switch v := argsInterface.(type) {
					case []interface{}:
						for _, arg := range v {
							strArg, ok := arg.(string)
							if !ok {
								return nil, errors.Errorf("args must be an array of strings but its item is of type %T", arg)
							}
							args = append(args, strArg)
						}
					default:
						return nil, errors.Errorf("args must be an array of strings but it is not an array")
					}
				}
				return &types.SandboxHookArgsCommand{Executable: executable, Args: args}, nil
			} else {
				return nil, errors.Errorf("command hooks data is invalid")
			}
		},
		"signal": func(key string, hook interface{}) (types.SandboxHook, error) {
			if strHook, ok := hook.(string); ok {
				return &types.SandboxHookSignal{
					IsString:    true,
					StringValue: strHook,
				}, nil
			}
			if intHook, ok := hook.(int); ok {
				return &types.SandboxHookSignal{
					IsString: false,
					IntValue: intHook,
				}, nil
			}
			return nil, errors.Errorf("invalid signal hook type %t, only string and int is allowed", hook)
		},
	}

	hooks := make(map[string]types.SandboxHook, len(dataMap))
	for key, hookData := range dataMap {
		if factory, ok := hooksFactories[key]; ok {
			hook, err := factory(key, hookData)
			if err != nil {
				return err
			}
			hooks[key] = hook
		} else {
			return errors.Errorf("unknown environment type: %s", key)
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
