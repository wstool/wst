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
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
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
	case "createServerExpectation":
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
		return fmt.Errorf("data must be an array, got %T", data)
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
		return fmt.Errorf("unsupported type for image data")
	}

	fieldValue.Set(reflect.ValueOf(img))

	return nil
}

type typeMapFactory func(key string) interface{}

func (f *FuncProvider) processTypeMap(
	name string,
	data interface{},
	factories map[string]typeMapFactory,
	path string,
) (map[string]interface{}, error) {
	// Check if data is a map
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data for %s must be a map, got %T", name, data)
	}

	var result map[string]interface{}
	var factory typeMapFactory
	var valMap map[string]interface{}
	for key, val := range dataMap {
		if factory, ok = factories[key]; ok {
			valMap, ok = val.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("data for value in %s must be a map, got %T", name, data)
			}
			structure := factory(key)
			if err := f.structParser(valMap, structure, path); err != nil {
				return nil, err
			}
			result[key] = structure
		} else {
			return nil, fmt.Errorf("unknown environment type: %s", key)
		}
	}

	return result, nil
}

func (f *FuncProvider) createEnvironments(data interface{}, fieldValue reflect.Value, path string) error {
	environmentFactories := map[string]typeMapFactory{
		"common": func(key string) interface{} {
			return &types.CommonEnvironment{}
		},
		"local": func(key string) interface{} {
			return &types.LocalEnvironment{}
		},
		"container": func(key string) interface{} {
			return &types.ContainerEnvironment{}
		},
		"docker": func(key string) interface{} {
			return &types.DockerEnvironment{}
		},
		"kubernetes": func(key string) interface{} {
			return &types.KubernetesEnvironment{}
		},
	}
	environments, err := f.processTypeMap("environments", data, environmentFactories, path)
	if err != nil {
		return err
	}

	fieldValue.Set(reflect.ValueOf(environments))

	return nil
}

type hooksMapFactory func(key string, hook interface{}) (interface{}, error)

func (f *FuncProvider) createHooks(data interface{}, fieldValue reflect.Value, path string) error {
	// Check if data is a map
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("data for hooks must be a map, got %T", data)
	}

	hooksFactories := map[string]hooksMapFactory{
		"command": func(key string, hook interface{}) (interface{}, error) {
			hookMap, ok := hook.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("command hooks must be a map, got %T", hook)
			}

			if cmd, ok := hookMap["command"]; ok {
				shell, shellOk := hookMap["shell"].(string)
				command, commandOk := cmd.(string)
				if !commandOk {
					return nil, fmt.Errorf("command must be a string")
				}
				if !shellOk {
					shell = "/bin/sh" // Default to /bin/sh if not specified
				}
				return &types.SandboxHookShellCommand{Command: command, Shell: shell}, nil
			} else if exe, ok := hookMap["executable"]; ok {
				executable, exeOk := exe.(string)
				if !exeOk {
					return nil, fmt.Errorf("executable must be a string")
				}
				argsInterface, argsOk := hookMap["args"]
				var args []string
				if argsOk {
					switch v := argsInterface.(type) {
					case []interface{}:
						for _, arg := range v {
							strArg, ok := arg.(string)
							if !ok {
								return nil, fmt.Errorf("args must be an array of strings")
							}
							args = append(args, strArg)
						}
					default:
						return nil, fmt.Errorf("args must be an array of strings")
					}
				}
				return &types.SandboxHookArgsCommand{Executable: executable, Args: args}, nil
			} else {
				return nil, fmt.Errorf("command hooks data is invalid")
			}
		},
		"signal": func(key string, hook interface{}) (interface{}, error) {
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
			return nil, fmt.Errorf("invalid signal hook type %t, only string and int is allowed", hook)
		},
	}

	var hooks map[string]interface{}
	for key, hook := range dataMap {
		if factory, ok := hooksFactories[key]; ok {
			env, err := factory(key, hook)
			if err != nil {
				return err
			}
			hooks[key] = env
		} else if key == "signal" {
			return fmt.Errorf("unknown environment type: %s", key)
		}
	}

	fieldValue.Set(reflect.ValueOf(hooks))

	return nil

}

func (f *FuncProvider) createParameters(data interface{}, fieldValue reflect.Value, path string) error {
	_, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("data for parameters must be a map, got %T", data)
	}

	fieldValue.Set(reflect.ValueOf(data))

	return nil
}

func (f *FuncProvider) createSandboxes(data interface{}, fieldValue reflect.Value, path string) error {
	sandboxFactories := map[string]typeMapFactory{
		"common": func(key string) interface{} {
			return &types.CommonSandbox{}
		},
		"local": func(key string) interface{} {
			return &types.LocalSandbox{}
		},
		"container": func(key string) interface{} {
			return &types.ContainerSandbox{}
		},
		"docker": func(key string) interface{} {
			return &types.DockerSandbox{}
		},
		"kubernetes": func(key string) interface{} {
			return &types.KubernetesSandbox{}
		},
	}
	sandboxes, err := f.processTypeMap("sandboxes", data, sandboxFactories, path)
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
		return fmt.Errorf("data for server action expectations must be a map, got %T", data)
	}
	expectations := make(map[string]types.Action)
	var structure interface{}
	for key, val := range dataMap {
		expData, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("data for value in server action expectations must be a map, got %T", data)
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
				return fmt.Errorf("invalid server expression key %s", expKey)
			}
			if parsed {
				return fmt.Errorf("expression cannot have multiple types - additional key %s", expKey)
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
	return nil
}
