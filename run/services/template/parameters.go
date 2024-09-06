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

package template

import (
	"fmt"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/parameters/parameter"
	"reflect"
)

type Parameters map[string]*Parameter

func NewParameters(origParams parameters.Parameters, tmpl Template) Parameters {
	params := map[string]*Parameter{}
	for key, param := range origParams {
		params[key] = NewParameter(param, origParams, tmpl)
	}
	return params
}

// GetString retrieves a string value directly from the parameters.
func (p Parameters) GetString(key string) (string, error) {
	if val, ok := p[key]; ok {
		return val.ToString()
	}
	return "", fmt.Errorf("object string for key [%s] not found", key)
}

// GetObject retrieves a map as a dynamic object with `ToString` method for each value.
func (p Parameters) GetObject(key string) Parameters {
	if val, ok := p[key]; ok {
		return val.ToObject()
	}
	return Parameters{}
}

// GetObjectString retrieves a string value by the secondKey of the object retrieved by the firstKey
func (p Parameters) GetObjectString(firstKey, secondKey string) (string, error) {
	if val, ok := p[firstKey]; ok {
		params := val.ToObject()
		if val, ok = params[secondKey]; ok {
			return val.ToString()
		}
	}
	return "", fmt.Errorf("object string for key [%s][%s] not found", firstKey, secondKey)
}

// Parameter is used to encapsulate values and provide a ToString method.
type Parameter struct {
	param            parameter.Parameter
	params           parameters.Parameters
	renderedValue    string
	isRendered       bool
	startedRendering bool
	tmpl             Template
}

func NewParameter(param parameter.Parameter, params parameters.Parameters, tmpl Template) *Parameter {
	return &Parameter{
		param:            param,
		params:           params,
		tmpl:             tmpl,
		renderedValue:    "",
		isRendered:       false,
		startedRendering: false,
	}
}

func (p *Parameter) IsObject() bool {
	return p.param.Type() == parameter.MapType
}

func (p *Parameter) IsNumber() bool {
	paramType := p.param.Type()
	return paramType == parameter.IntType || paramType == parameter.FloatType
}

func (p *Parameter) ToNumber() float64 {
	if p.param == nil {
		return 0.0
	}
	return p.param.FloatValue()
}

func (p *Parameter) ToString() (string, error) {
	if reflect.ValueOf(p.param).IsNil() {
		return "", fmt.Errorf("trying to render not set parameter")
	}
	if p.isRendered {
		return p.renderedValue, nil
	}
	if p.startedRendering {
		return "", fmt.Errorf("recursive rendering of parameter %s detected", p.param.StringValue())
	}
	p.startedRendering = true
	renderedValue, err := p.tmpl.RenderToString(p.param.StringValue(), p.params)
	if err != nil {
		return "", err
	}
	p.isRendered = true
	p.renderedValue = renderedValue
	return p.renderedValue, nil
}

func (p *Parameter) ToObject() Parameters {
	if p.param == nil {
		return Parameters{}
	}
	return NewParameters(p.param.MapValue(), p.tmpl)
}
