package conf

import (
	"fmt"
	"github.com/bukka/wst/app"
	"reflect"
	"strings"
)

const (
	paramName     = "name"
	paramLoadable = "loadable"
	paramDefault  = "default"
	paramFactory  = "factory"
	paramEnum     = "enum"
	paramKeys     = "keys"
	paramKeymatch = "keymatch"
	paramString   = "string"
	paramBoolean  = "boolean"
)

// check if param is a valid param (one of param* constants)
func isValidParam(param string) bool {
	switch param {
	case paramName:
		fallthrough
	case paramLoadable:
		fallthrough
	case paramDefault:
		fallthrough
	case paramFactory:
		fallthrough
	case paramEnum:
		fallthrough
	case paramKeys:
		fallthrough
	case paramKeymatch:
		fallthrough
	case paramString:
		fallthrough
	case paramBoolean:
		return true
	default:
		return false
	}
}

// parseTag parses the 'wst' struct tag into a Field
func parseTag(tag string, env app.Env) map[string]string {
	// split the tag into parts
	parts := strings.Split(tag, ",")

	// create a map to store parameters
	params := make(map[string]string)

	// the name is the first part
	firstPart := parts[0]
	// loop start index
	startIndex := 1
	// check if = present as it is optional for name
	if strings.Contains(firstPart, "=") {
		startIndex = 0
	} else {
		params[paramName] = firstPart
	}

	// parse the rest of the parts as key=value parameters
	for _, part := range parts[startIndex:] {
		// split the part into a key and a value
		keyValue := strings.Split(part, "=")

		key := keyValue[0]
		if isValidParam(key) {
			var value string
			if len(keyValue) == 2 {
				value = keyValue[1]
			} else {
				// this is a boolean parameter
				value = "true"
			}
			params[key] = value
		} else {
			env.Logger().Errorf("Invalid parameter key: %s", key)
		}
	}

	return params
}

// parseField parses a struct field based on data and params
func parseField(data interface{}, fieldValue reflect.Value, params map[string]string, env app.Env) error {
	// TODO: field parsing
	return nil
}

// parseStruct parses a struct into a map of Fields
func parseStruct(data map[string]interface{}, s interface{}, env app.Env) error {
	t := reflect.TypeOf(s)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("wst")
		if tag == "" {
			continue
		}
		params := parseTag(tag, env)
		name, ok := params[paramName]
		if !ok {
			name = field.Name
		}
		fieldValue := reflect.ValueOf(s).Elem().FieldByName(field.Name)
		if fieldData, ok := data[name]; ok {
			if err := parseField(fieldData, fieldValue, params, env); err != nil {
				return err
			}
		} else if defaultValue, found := params[paramDefault]; found {
			fieldType := fieldValue.Type()
			defaultValueValue := reflect.ValueOf(defaultValue)
			if !defaultValueValue.CanConvert(fieldType) {
				return fmt.Errorf("default value %v for field %s cannot be converted to type %v",
					defaultValueValue, name, fieldType)
			}
			defaultValueConverted := defaultValueValue.Convert(fieldType)
			fieldValue.Set(defaultValueConverted)
		}
	}

	return nil
}

func ParseConfig(data interface{}, config *Config, env app.Env) error {
	if m, ok := data.(map[string]interface{}); ok {
		return parseStruct(m, config, env)
	} else {
		return fmt.Errorf("config data is not an object")
	}
}
