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
	paramString   = "string"
	paramBoolean  = "boolean"
)

type Parser interface {
	ParseConfig(data map[string]interface{}, config *Config) error
}

type ConfigParser struct {
	env app.Env
}

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
	case paramString:
		fallthrough
	case paramBoolean:
		return true
	default:
		return false
	}
}

// parseTag parses the 'wst' struct tag into a Field
func (p ConfigParser) parseTag(tag string) map[string]string {
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
			p.env.Logger().Errorf("Invalid parameter key: %s", key)
		}
	}

	return params
}

// parseField parses a struct field based on data and params
func (p ConfigParser) parseField(data interface{}, fieldValue reflect.Value, params map[string]string) error {

	return nil
}

// parseStruct parses a struct into a map of Fields
func (p ConfigParser) parseStruct(data map[string]interface{}, s interface{}) error {
	t := reflect.TypeOf(s)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("wst")
		if tag == "" {
			continue
		}
		params := p.parseTag(tag)
		name, ok := params[paramName]
		if !ok {
			name = field.Name
		}
		fieldValue := reflect.ValueOf(s).Elem().FieldByName(field.Name)
		if fieldData, ok := data[name]; ok {
			if err := p.parseField(fieldData, fieldValue, params); err != nil {
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
		} else {
			fieldValue.SetZero()
		}
	}

	return nil
}

func (p ConfigParser) ParseConfig(data map[string]interface{}, config *Config) error {
	return p.parseStruct(data, config)
}

func CreateParser(env app.Env) Parser {
	return &ConfigParser{
		env: env,
	}
}
