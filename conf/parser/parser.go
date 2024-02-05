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

package parser

import (
	"errors"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/loader"
	"github.com/bukka/wst/conf/types"
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
)

type Parser interface {
	ParseConfig(data map[string]interface{}, config *types.Config) error
}

type ConfigParser struct {
	env       app.Env
	loader    loader.Loader
	factories map[string]factoryFunc
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

func (p ConfigParser) processFactoryParam(factory string, data interface{}, fieldValue reflect.Value) error {
	factoryFunc, found := p.factories[factory]
	if !found {
		return fmt.Errorf("factory function %s not found", factory)
	}
	return factoryFunc(data, fieldValue)
}

func (p ConfigParser) processEnumParam(enums string, data interface{}, fieldName string) error {
	enumList := strings.Split(enums, "|")
	for _, enum := range enumList {
		if enum == data {
			return nil
		}
	}

	return fmt.Errorf("values %v are not valid for field %s", enums, fieldName)
}

func (p ConfigParser) processKeysParam(keys string, data interface{}, fieldName string) error {
	keysList := strings.Split(keys, "|")
	for _, key := range keysList {
		if _, ok := data.(map[string]interface{})[key]; ok {
			return nil
		}
	}

	return fmt.Errorf("keys %v are not valid for field %s", keys, fieldName)
}

func (p ConfigParser) processLoadableParam(data interface{}, fieldValue reflect.Value) (interface{}, error) {
	loadableData, isString := data.(string)
	if isString {
		configs, err := p.loader.GlobConfigs(loadableData)
		if err != nil {
			return nil, fmt.Errorf("loading configs: %v", err)
		}

		switch fieldValue.Kind() {
		case reflect.Map:
			loadedData := make(map[string]map[string]interface{})
			for _, config := range configs {
				loadedData[config.Path()] = config.Data()
			}
			return loadedData, nil

		case reflect.Slice:
			loadedData := make([]map[string]interface{}, len(configs))
			for i, config := range configs {
				loadedData[i] = config.Data()
			}
			return loadedData, nil

		default:
			return nil, fmt.Errorf("type of field is neither map nor slice (kind=%s)", fieldValue.Kind())
		}
	}
	return data, nil
}

func (p ConfigParser) processStringParam(fieldName string, data interface{}, fieldValuePtr reflect.Value) (bool, error) {
	fieldValue := fieldValuePtr.Elem()
	var isMap = fieldValue.Kind() == reflect.Map
	var isStruct = fieldValue.Kind() == reflect.Struct

	if isMap {
		mapData, ok := data.(map[string]string)
		if !ok {
			// Data is not a map
			return false, nil
		}
		// Process map values
		err := p.processMapValue(mapData, fieldValue, fieldName)
		if err != nil {
			return false, err
		}
	} else if isStruct {
		strData, isString := data.(string)
		if !isString {
			return false, nil // If not a string, mark it as not done and just ignore.
		}

		elem := reflect.New(fieldValuePtr.Type().Elem())

		// Use an empty map as temporary data to populate the struct
		err := p.parseStruct(make(map[string]interface{}), elem.Interface())
		if err != nil {
			return false, fmt.Errorf("error parsing struct for string param: %v", err)
		}

		// Set the string value to the appropriate sub-field
		err = p.setFieldByName(fieldValuePtr.Interface(), fieldName, strData)
		if err != nil {
			return false, fmt.Errorf("failed to set field %s : %v", fieldName, err)
		}
	} else {
		return false, errors.New("v must be a pointer to a struct or a map")
	}
	return true, nil
}

// setFieldByName sets the field of the struct v with the given name to the specified value.
// The value v must be a pointer to a struct, and the field should be exported and settable.
func (p ConfigParser) setFieldByName(v interface{}, name string, value string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct {
		// For struct:
		fv := rv.Elem().FieldByName(name)
		if !fv.IsValid() {
			return fmt.Errorf("not a valid field name: %s", name)
		}
		if !fv.CanSet() {
			return fmt.Errorf("cannot set the field: %s", name)
		}
		fv.SetString(value)
	}

	return nil
}

// processMapValue processes a map[string]interface{} where value is a string
// It goes through each value in map and if it's a string, sets it to a specified field.
func (p ConfigParser) processMapValue(mapData map[string]string, rv reflect.Value, fieldName string) error {
	keyType := rv.Type().Key()
	elemType := rv.Type().Elem()

	// Create a new map of the appropriate type
	resultMap := reflect.MakeMap(reflect.MapOf(keyType, elemType))

	for key, val := range mapData {
		// Create a new element of the appropriate type
		newElem := reflect.New(elemType)

		// Use an empty map as temporary data to populate the struct
		err := p.parseStruct(make(map[string]interface{}), newElem.Interface())
		if err != nil {
			return fmt.Errorf("error parsing struct for string param: %v", err)
		}

		// Set the string value to the appropriate sub-field
		err = p.setFieldByName(newElem.Interface(), fieldName, val)
		if err != nil {
			return fmt.Errorf("failed to set field %s : %v", fieldName, err)
		}

		// Insert the new element into the result map
		resultMap.SetMapIndex(reflect.ValueOf(key), newElem.Elem())
	}

	rv.Set(resultMap)
	return nil
}

// assignField assigns the provided data to the fieldValue.
func (p ConfigParser) assignField(data interface{}, fieldValue reflect.Value, fieldName string) error {
	switch fieldValue.Kind() {
	case reflect.Struct:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to convert data for field %s to map[string]interface{}", fieldName)
		}
		return p.parseStruct(dataMap, fieldValue.Addr().Interface())
	case reflect.Map:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to convert data for field %s to map[string]interface{}", fieldName)
		}
		for key, val := range dataMap {
			tempFieldValue := fieldValue.FieldByName(key)
			if tempFieldValue.IsValid() {
				return p.assignField(val, tempFieldValue, key)
			}
		}
	case reflect.Slice:
		dataSlice, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("unable to convert data for field %s to []interface{}", fieldName)
		}
		for i, val := range dataSlice {
			return p.assignField(val, fieldValue.Index(i), fmt.Sprintf("%s[%d]", fieldName, i))
		}
	default:
		v := reflect.ValueOf(data)
		if v.Type().ConvertibleTo(fieldValue.Type()) {
			fieldValue.Set(v.Convert(fieldValue.Type()))
			return nil
		}
		return fmt.Errorf("field %s could not be set", fieldName)
	}
	return nil
}

// parseField parses a struct field based on data and params
func (p ConfigParser) parseField(data interface{}, s interface{}, fieldName string, params map[string]string) error {
	var err error
	valueOfS := reflect.Indirect(reflect.ValueOf(s))
	if !valueOfS.IsValid() {
		return fmt.Errorf("invalid parameter s: cannot be accessed")
	}
	if !valueOfS.Type().ConvertibleTo(reflect.TypeOf(s)) {
		return fmt.Errorf("type %T cannot be converted to the same type as s (%T)", valueOfS, s)
	}
	fieldValue := valueOfS.FieldByName(fieldName)

	if factory, hasFactory := params[paramFactory]; hasFactory {
		if err = p.processFactoryParam(factory, data, fieldValue); err != nil {
			return err
		}
		// factory should set everything so there is no need to continue
		return nil
	}

	if stringValue, hasString := params[paramString]; hasString {
		var done bool
		if done, err = p.processStringParam(stringValue, data, fieldValue); err != nil {
			return err
		}
		if done {
			return nil
		}
	}

	if _, isLoadable := params[paramLoadable]; isLoadable {
		if data, err = p.processLoadableParam(data, fieldValue); err != nil {
			return err
		}
	}

	if enums, hasEnum := params[paramEnum]; hasEnum {
		if err = p.processEnumParam(enums, data, fieldName); err != nil {
			return err
		}
	}

	if keys, hasKeys := params[paramKeys]; hasKeys {
		if err = p.processKeysParam(keys, data, fieldName); err != nil {
			return err
		}
	}

	// assign the processed data to the field
	if err = p.assignField(data, fieldValue, fieldName); err != nil {
		return err
	}

	return nil
}

// parseStruct parses a struct into a map of Fields
func (p ConfigParser) parseStruct(data map[string]interface{}, s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %T", s)
	}
	t := v.Elem().Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("wst")
		if tag == "" {
			continue
		}
		params := p.parseTag(tag)
		fieldName, ok := params[paramName]
		if !ok {
			fieldName = field.Name
		}
		if fieldData, ok := data[fieldName]; ok {
			if err := p.parseField(fieldData, s, fieldName, params); err != nil {
				return err
			}
		} else {
			fieldValue := reflect.ValueOf(s).Elem().FieldByName(field.Name)
			if defaultValue, found := params[paramDefault]; found {
				fieldType := fieldValue.Type()
				defaultValueValue := reflect.ValueOf(defaultValue)
				if !defaultValueValue.CanConvert(fieldType) {
					return fmt.Errorf("default value %v for field %s cannot be converted to type %v",
						defaultValueValue, fieldName, fieldType)
				}
				defaultValueConverted := defaultValueValue.Convert(fieldType)
				fieldValue.Set(defaultValueConverted)
			} else {
				fieldValue.SetZero()
			}
		}
	}

	return nil
}

func (p ConfigParser) ParseConfig(data map[string]interface{}, config *types.Config) error {
	return p.parseStruct(data, config)
}

func CreateParser(env app.Env, loader loader.Loader) Parser {
	return &ConfigParser{
		env:       env,
		loader:    loader,
		factories: getFactories(),
	}
}
