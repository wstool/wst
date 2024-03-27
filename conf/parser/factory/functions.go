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

func (f *FuncProvider) createEnvironments(data interface{}, fieldValue reflect.Value, path string) error {
	// Check if data is a map
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("data must be a map, got %T", data)
	}
	var environments map[string]types.Environment
	if _, ok = dataMap["common"]; ok {
		commonEnvironment := types.CommonEnvironment{}
		if err := f.structParser(dataMap, &commonEnvironment, path); err != nil {
			return err
		}
		environments["common"] = &commonEnvironment
	}
	if _, ok = dataMap["local"]; ok {
		localEnvironment := types.LocalEnvironment{}
		if err := f.structParser(dataMap, &localEnvironment, path); err != nil {
			return err
		}
		environments["local"] = &localEnvironment
	}
	if _, ok = dataMap["container"]; ok {
		containerEnvironment := types.ContainerEnvironment{}
		if err := f.structParser(dataMap, &containerEnvironment, path); err != nil {
			return err
		}
		environments["container"] = &containerEnvironment
	}
	if _, ok = dataMap["docker"]; ok {
		dockerEnvironment := types.DockerEnvironment{}
		if err := f.structParser(dataMap, &dockerEnvironment, path); err != nil {
			return err
		}
		environments["docker"] = &dockerEnvironment
	}
	if _, ok = dataMap["kubernetes"]; ok {
		kubernetesEnvironment := types.KubernetesEnvironment{}
		if err := f.structParser(dataMap, &kubernetesEnvironment, path); err != nil {
			return err
		}
		environments["kubernetes"] = &kubernetesEnvironment
	}

	fieldValue.Set(reflect.ValueOf(environments))

	return nil
}

func (f *FuncProvider) createHooks(data interface{}, fieldValue reflect.Value, path string) error {
	return nil
}

func (f *FuncProvider) createParameters(data interface{}, fieldValue reflect.Value, path string) error {
	return nil
}

func (f *FuncProvider) createSandboxes(data interface{}, fieldValue reflect.Value, path string) error {
	return nil
}

func (f *FuncProvider) createServerExpectations(data interface{}, fieldValue reflect.Value, path string) error {
	return nil
}

func (f *FuncProvider) createServiceScripts(data interface{}, fieldValue reflect.Value, path string) error {
	return nil
}
