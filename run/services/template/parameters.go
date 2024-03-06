package template

import (
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/parameters/parameter"
)

type Parameters map[string]Parameter

func NewParameters(parameters parameters.Parameters) Parameters {
	params := map[string]Parameter{}
	for key, param := range parameters {
		params[key] = *NewParameter(param)
	}
	return params
}

// GetString retrieves a string value directly from the parameters.
func (p Parameters) GetString(key string) string {
	if val, ok := p[key]; ok {
		return val.ToString()
	}
	return ""
}

// GetObject retrieves a map as a dynamic object with `ToString` method for each value.
func (p Parameters) GetObject(key string) Parameters {
	if val, ok := p[key]; ok {
		return val.ToObject()
	}
	return Parameters{}
}

// GetObjectString retrieves a string value by the secondKey of the object retrieved by the firstKey
func (p Parameters) GetObjectString(firstKey, secondKey string) string {
	if val, ok := p[firstKey]; ok {
		params := val.ToObject()
		if val, ok = params[secondKey]; ok {
			return val.ToString()
		}
	}
	return ""
}

// Parameter is used to encapsulate values and provide a ToString method.
type Parameter struct {
	parameter parameter.Parameter
}

func NewParameter(parameter parameter.Parameter) *Parameter {
	return &Parameter{parameter: parameter}
}

func (p *Parameter) IsNumber() bool {
	paramType := p.parameter.Type()
	return paramType == parameter.IntType || paramType == parameter.FloatType
}

func (p *Parameter) ToNumber() float64 {
	if p.parameter == nil {
		return 0.0
	}
	return p.parameter.FloatValue()
}

func (p *Parameter) ToString() string {
	if p.parameter == nil {
		return ""
	}
	return p.parameter.StringValue()
}

func (p *Parameter) ToObject() Parameters {
	if p.parameter == nil {
		return Parameters{}
	}
	return NewParameters(p.parameter.MapValue())
}
