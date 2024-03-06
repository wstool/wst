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

// Parameter is used to encapsulate values and provide a ToString method.
type Parameter struct {
	parameter parameter.Parameter
}

func NewParameter(parameter parameter.Parameter) *Parameter {
	return &Parameter{parameter: parameter}
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
