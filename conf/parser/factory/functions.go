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
	"reflect"
)

type Func func(data interface{}, fieldValue reflect.Value) error

// Functions define an interface for factory functions
type Functions interface {
	GetFactoryFunc(funcName string) Func
}

// FuncProvider contains the implementation of the FactoryFunctions
type FuncProvider struct {
	fnd            app.Foundation
	actionsFactory ActionsFactory
}

func CreateFactories(fnd app.Foundation) Functions {
	return &FuncProvider{
		fnd:            fnd,
		actionsFactory: CreateActionsFactory(fnd),
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
	case "createEnvironments":
		return f.createEnvironments
	case "createHooks":
		return f.createHooks
	case "createParameters":
		return f.createParameters
	case "createSandboxes":
		return f.createSandboxes
	case "createServerExpectation":
		return f.createServiceScripts
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

func (f *FuncProvider) createEnvironments(data interface{}, fieldValue reflect.Value) error {
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
