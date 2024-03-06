package parameter

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
)

type Parameter interface {
	IsNil() bool
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
	NilType = iota
	BoolType
	IntType
	FloatType
	StringType
	ArrayType
	MapType
)

type Maker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd: fnd,
	}
}

func (m *Maker) Make(config types.Parameter) (Parameter, error) {
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
		p.parameterType = MapType
		p.mapValue = make(map[string]Parameter)
		for key, elem := range v {
			if paramElem, err := m.Make(elem); err == nil {
				p.mapValue[key] = paramElem
			} else {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}

	return p, nil
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

func (p *parameter) IsNil() bool {
	return p.parameterType == NilType
}

func (p *parameter) BoolValue() bool {
	if p.parameterType != BoolType {
		panic("parameter is not a bool")
	}
	return p.boolValue
}

func (p *parameter) IntValue() int {
	if p.parameterType != IntType {
		panic("parameter is not an int")
	}
	return p.intValue
}

func (p *parameter) FloatValue() float64 {
	if p.parameterType != FloatType {
		panic("parameter is not a float")
	}
	return p.floatValue
}

func (p *parameter) StringValue() string {
	if p.parameterType != StringType {
		panic("parameter is not a string")
	}
	return p.stringValue
}

func (p *parameter) ArrayValue() []Parameter {
	if p.parameterType != ArrayType {
		panic("parameter is not an array")
	}
	return p.arrayValue
}

func (p *parameter) MapValue() map[string]Parameter {
	if p.parameterType != MapType {
		panic("parameter is not a map")
	}
	return p.mapValue
}

func (p *parameter) Type() Type {
	return p.parameterType
}
