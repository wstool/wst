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

package parameter

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/pkg/errors"
	"strconv"
)

type Parameter interface {
	BoolValue() bool
	IntValue() int
	FloatValue() float64
	StringValue() string
	ArrayValue() []Parameter
	MapValue() map[string]Parameter
	Type() Type
}

type Type int

const (
	NilType Type = iota
	BoolType
	IntType
	FloatType
	StringType
	ArrayType
	MapType
)

type Maker interface {
	Make(config interface{}) (Parameter, error)
}

type nativeMaker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{
		fnd: fnd,
	}
}

func (m *nativeMaker) Make(config interface{}) (Parameter, error) {
	p := &parameter{}

	// Use a type switch to handle different types.
	switch v := config.(type) {
	case bool:
		p.parameterType = BoolType
		p.boolValue = v
	case int:
		p.parameterType = IntType
		p.intValue = v
	case float64:
		p.parameterType = FloatType
		p.floatValue = v
	case string:
		p.parameterType = StringType
		p.stringValue = v
	case []interface{}:
		p.parameterType = ArrayType
		for _, elem := range v {
			if paramElem, err := m.Make(elem); err == nil {
				p.arrayValue = append(p.arrayValue, paramElem)
			} else {
				return nil, err
			}
		}
	case map[string]interface{}:
		if err := m.processMap(p, v); err != nil {
			return nil, err
		}
	case types.Parameters:
		if err := m.processMap(p, v); err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("unsupported type: %T", v)
	}

	return p, nil
}

func (m *nativeMaker) processMap(p *parameter, v map[string]interface{}) error {
	p.parameterType = MapType
	p.mapValue = make(map[string]Parameter)
	for key, elem := range v {
		if paramElem, err := m.Make(elem); err == nil {
			p.mapValue[key] = paramElem
		} else {
			return err
		}
	}
	return nil
}

type parameter struct {
	parameterType Type
	boolValue     bool
	intValue      int
	floatValue    float64
	stringValue   string
	arrayValue    []Parameter
	mapValue      map[string]Parameter
}

func (p *parameter) Type() Type {
	return p.parameterType
}

func (p *parameter) BoolValue() bool {
	switch p.parameterType {
	case BoolType:
		return p.boolValue
	case IntType, FloatType:
		return p.IntValue() != 0 // Reuse IntValue for simplicity
	case StringType:
		return p.stringValue != ""
	case ArrayType:
		return len(p.arrayValue) > 0
	case MapType:
		return len(p.mapValue) > 0
	default:
		return false
	}
}

func (p *parameter) IntValue() int {
	switch p.parameterType {
	case IntType:
		return p.intValue
	case BoolType:
		if p.boolValue {
			return 1
		}
		return 0
	case FloatType:
		return int(p.floatValue) // Note: Truncation
	case StringType:
		val, err := strconv.Atoi(p.stringValue)
		if err == nil {
			return val
		}
		return 0
	case ArrayType:
		return len(p.arrayValue)
	case MapType:
		return len(p.mapValue)
	default:
		return 0
	}
}

func (p *parameter) FloatValue() float64 {
	switch p.parameterType {
	case FloatType:
		return p.floatValue
	case IntType:
		return float64(p.intValue)
	case BoolType:
		if p.boolValue {
			return 1.0
		}
		return 0.0
	case ArrayType:
		return float64(len(p.arrayValue))
	case MapType:
		return float64(len(p.mapValue))
	default:
		return 0.0
	}
}

func (p *parameter) StringValue() string {
	switch p.parameterType {
	case StringType:
		return p.stringValue
	case BoolType:
		return strconv.FormatBool(p.boolValue)
	case IntType:
		return strconv.Itoa(p.intValue)
	case FloatType:
		return fmt.Sprintf("%v", p.floatValue)
	default:
		return fmt.Sprintf("%v", p)
	}
}

func (p *parameter) ArrayValue() []Parameter {
	if p.parameterType == ArrayType {
		return p.arrayValue
	} else if p.parameterType == MapType {
		// Convert map to array, focusing only on the map's values
		var convertedArray []Parameter
		for _, val := range p.mapValue {
			convertedArray = append(convertedArray, val)
		}
		return convertedArray
	} else {
		// For all other types, convert to an array with a single element
		return []Parameter{p}
	}
}

func (p *parameter) MapValue() map[string]Parameter {
	if p.parameterType == MapType {
		return p.mapValue
	} else if p.parameterType == ArrayType {
		// Convert array to map, using string indexes as keys
		convertedMap := make(map[string]Parameter, len(p.arrayValue))
		for i, val := range p.arrayValue {
			convertedMap[strconv.Itoa(i)] = val
		}
		return convertedMap
	} else {
		// For all other types, convert to a map with a single key "0"
		return map[string]Parameter{"0": p}
	}
}
