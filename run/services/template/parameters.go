package template

import "fmt"

type Parameters struct {
	data map[string]interface{}
}

func NewParameters(data map[string]interface{}) *Parameters {
	return &Parameters{data: data}
}

// GetString retrieves a string value directly from the parameters.
func (p *Parameters) GetString(key string) string {
	if val, ok := p.data[key].(string); ok {
		return val
	}
	return ""
}

// GetObject retrieves a map as a dynamic object with `ToString` method for each value.
func (p *Parameters) GetObject(key string) map[string]*DynamicValue {
	object := make(map[string]*DynamicValue)
	if val, ok := p.data[key].(map[string]interface{}); ok {
		for k, v := range val {
			object[k] = &DynamicValue{data: v}
		}
	}
	return object
}

// DynamicValue is used to encapsulate values and provide a ToString method.
type DynamicValue struct {
	data interface{}
}

func (dv *DynamicValue) ToString() string {
	switch v := dv.data.(type) {
	case string:
		return v
	case int, float64:
		return fmt.Sprintf("%v", v)
	// Add more cases as necessary.
	default:
		return ""
	}
}
