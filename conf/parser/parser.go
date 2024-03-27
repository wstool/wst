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
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/loader"
	"github.com/bukka/wst/conf/parser/factory"
	"github.com/bukka/wst/conf/types"
	"github.com/spf13/afero"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type ConfigParam string

const (
	paramName     ConfigParam = "name"
	paramLoadable             = "loadable"
	paramDefault              = "default"
	paramFactory              = "factory"
	paramEnum                 = "enum"
	paramKeys                 = "keys"
	paramPath                 = "path"
	paramString               = "string"
)

const pathKey = "wst/path"

type Parser interface {
	ParseConfig(data map[string]interface{}, config *types.Config, configPath string) error
}

type ConfigParser struct {
	fnd       app.Foundation
	loader    loader.Loader
	factories factory.Functions
}

// check if param is a valid param (one of param* constants)
func isValidParam(param string) bool {
	switch ConfigParam(param) {
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
	case paramPath:
		fallthrough
	case paramString:
		return true
	default:
		return false
	}
}

// parseTag parses the 'wst' struct tag into a Field
func (p ConfigParser) parseTag(tag string) (map[ConfigParam]string, error) {
	// split the tag into parts
	parts := strings.Split(tag, ",")

	// create a map to store parameters
	params := make(map[ConfigParam]string)

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
			params[ConfigParam(key)] = value
		} else {
			return nil, fmt.Errorf("invalid parameter key: %s", key)
		}
	}

	return params, nil
}

func (p ConfigParser) processDefaultParam(fieldName string, defaultValue string, fieldValue reflect.Value) error {
	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		intValue, err := strconv.Atoi(defaultValue)
		if err != nil {
			return fmt.Errorf("default value %q for field %s can't be converted to int: %w",
				defaultValue, fieldName, err)
		}
		fieldValue.SetInt(int64(intValue))
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(defaultValue)
		if err != nil {
			return fmt.Errorf("default value %q for field %s can't be converted to bool: %w",
				defaultValue, fieldName, err)
		}
		fieldValue.SetBool(boolValue)
	case reflect.String:
		fieldValue.SetString(defaultValue)
	default:
		return fmt.Errorf("default value %v for field %s cannot be converted to type %v",
			defaultValue, fieldName, fieldValue.Type())
	}
	return nil
}

func (p ConfigParser) processFactoryParam(
	factory string,
	data interface{},
	fieldValue reflect.Value,
	path string,
) error {
	factoryFunc := p.factories.GetFactoryFunc(factory)
	if factoryFunc == nil {
		return fmt.Errorf("factory function %s not found", factory)
	}
	return factoryFunc(data, fieldValue, path)
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

func (p ConfigParser) processPathParam(data interface{}, fieldValue reflect.Value, fieldName string, configPath string) error {
	// Assert data is a string
	path, ok := data.(string)
	if !ok {
		return fmt.Errorf("unexpected type %T for data, expected string", data)
	}

	fs := p.fnd.Fs()

	// If it's not an absolute path, prepend the path variable to it
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(configPath), path)
	}

	// Check if the constructed path exists using Afero's fs.Exists
	exists, err := afero.Exists(fs, path)
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("file path %s does not exist", path)
	}

	// Check that fieldValue is settable (it is addressable and was not obtained by
	// the use of unexported struct fields)
	if !fieldValue.CanSet() {
		return fmt.Errorf("field %s is not settable", fieldName)
	}

	// Check if fieldValue is of type string
	if fieldValue.Kind() != reflect.String {
		return fmt.Errorf("field %s is not of type string", fieldName)
	}

	// If no errors so far, set the fieldValue to the resolved file path
	fieldValue.SetString(path)

	return nil
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
			loadedData := make(map[string]interface{})
			for _, config := range configs {
				data := config.Data()
				data[pathKey] = config.Path()
				loadedData[config.Path()] = data
			}
			return loadedData, nil

		case reflect.Slice:
			loadedData := make([]map[string]interface{}, len(configs))
			for i, config := range configs {
				data := config.Data()
				data[pathKey] = config.Path()
				loadedData[i] = data
			}
			return loadedData, nil

		default:
			return nil, fmt.Errorf("type of field is neither map nor slice (kind=%s)", fieldValue.Kind())
		}
	}
	return data, nil
}

func (p ConfigParser) processStringParam(
	fieldName string,
	data interface{},
	fieldValue reflect.Value,
	path string,
) (bool, error) {
	kind := fieldValue.Kind()

	if kind != reflect.Struct && kind != reflect.Interface && kind != reflect.Ptr {
		return false, fmt.Errorf("field %s must be a struct or interface type or a pointer to such", fieldName)
	}

	if kind == reflect.Ptr || kind == reflect.Interface {
		fieldValue = fieldValue.Elem()
	}
	fieldValueKind := fieldValue.Kind()

	if fieldValueKind == reflect.Map {
		mapData, ok := data.(map[string]string)
		if !ok {
			// Data is not a map
			return false, nil
		}
		// Process map values
		err := p.processMapValue(mapData, fieldValue, fieldName, path)
		if err != nil {
			return false, err
		}
	} else if fieldValueKind == reflect.Struct {
		strData, isString := data.(string)
		if !isString {
			return false, nil // If not a string, mark it as not done and just ignore.
		}

		fieldValuePtrInterface := fieldValue.Addr().Interface()
		// Use an empty map as temporary data to populate the struct
		err := p.parseStruct(make(map[string]interface{}), fieldValuePtrInterface, path)
		if err != nil {
			return false, fmt.Errorf("error parsing struct for string param: %v", err)
		}

		// Set the string value to the appropriate sub-field
		err = p.setFieldByName(fieldValuePtrInterface, fieldName, strData)
		if err != nil {
			return false, fmt.Errorf("failed to set field %s : %v", fieldName, err)
		}
	} else {
		return false, fmt.Errorf("field %s value must be a pointer to a struct or a map", fieldName)
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
func (p ConfigParser) processMapValue(
	mapData map[string]string,
	rv reflect.Value,
	fieldName string,
	path string,
) error {
	keyType := rv.Type().Key()
	elemType := rv.Type().Elem()

	// Create a new map of the appropriate type
	resultMap := reflect.MakeMap(reflect.MapOf(keyType, elemType))

	for key, val := range mapData {
		// Create a new element of the appropriate type
		newElem := reflect.New(elemType)

		// Use an empty map as temporary data to populate the struct
		err := p.parseStruct(make(map[string]interface{}), newElem.Interface(), path)
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
func (p ConfigParser) assignField(data interface{}, fieldValue reflect.Value, fieldName string, path string) error {
	switch fieldValue.Kind() {
	case reflect.Struct:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to convert data for field %s to map[string]interface{}", fieldName)
		}
		return p.parseStruct(dataMap, fieldValue.Addr().Interface(), path)
	case reflect.Map:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to convert data for field %s to map[string]interface{}", fieldName)
		}
		if fieldValue.IsNil() {
			// Initialize a new map
			fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
		}
		for key, val := range dataMap {
			newVal := reflect.New(fieldValue.Type().Elem())
			err := p.assignField(val, newVal.Elem(), fieldName, path)
			if err != nil {
				return err
			}
			fieldValue.SetMapIndex(reflect.ValueOf(key), newVal.Elem())
		}
	case reflect.Slice:
		var dataSlice []interface{}
		// check if data is a []interface{}
		if intermediate, ok := data.([]interface{}); ok {
			dataSlice = intermediate
		} else if intermediate, ok := data.([]map[string]interface{}); ok {
			// if data is a  []map[string]interface{} (e.g. returned by config loader), convert it to []interface{}
			dataSlice = make([]interface{}, len(intermediate))
			for i, v := range intermediate {
				dataSlice[i] = v
			}
		} else {
			return fmt.Errorf("unable to convert data for field %s to []interface{}", fieldName)
		}
		if fieldValue.IsNil() || fieldValue.Len() < len(dataSlice) {
			// Make a new slice to accommodate all elements
			fieldValue.Set(reflect.MakeSlice(fieldValue.Type(), len(dataSlice), len(dataSlice)))
		}
		for i, val := range dataSlice {
			err := p.assignField(val, fieldValue.Index(i), fmt.Sprintf("%s[%d]", fieldName, i), path)
			if err != nil {
				return err
			}
		}
	default:
		v := reflect.ValueOf(data)
		if fieldValue.Type().Kind() != v.Type().Kind() {
			return fmt.Errorf("field %s could not be set due to type mismatch", fieldName)
		}
		if !v.Type().ConvertibleTo(fieldValue.Type()) {
			return fmt.Errorf("field %s could not be set", fieldName)
		}
		fieldValue.Set(v.Convert(fieldValue.Type()))
	}
	return nil
}

// parseField parses a struct field based on data and params
func (p ConfigParser) parseField(
	data interface{},
	fieldValue reflect.Value,
	fieldName string,
	params map[ConfigParam]string,
	path string,
) error {
	var err error

	if factoryName, hasFactory := params[paramFactory]; hasFactory {
		if err = p.processFactoryParam(factoryName, data, fieldValue, path); err != nil {
			return err
		}
		// factory should set everything so there is no need to continue
		return nil
	}

	if stringValue, hasString := params[paramString]; hasString {
		var done bool
		if done, err = p.processStringParam(stringValue, data, fieldValue, path); err != nil {
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

	if _, hasPath := params[paramPath]; hasPath {
		if err = p.processPathParam(data, fieldValue, fieldName, path); err != nil {
			return err
		}
	} else if err = p.assignField(data, fieldValue, fieldName, path); err != nil {
		return err
	}

	return nil
}

// parseStruct parses a struct into a map of Fields
func (p ConfigParser) parseStruct(data map[string]interface{}, structure interface{}, path string) error {
	structValuePtr := reflect.ValueOf(structure)
	if structValuePtr.Kind() != reflect.Ptr || structValuePtr.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %T", structure)
	}
	structValue := structValuePtr.Elem()
	structType := structValue.Type()

	if newPathInterface, newPathFound := data[pathKey]; newPathFound {
		newPath, ok := newPathInterface.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for path", newPathInterface)
		}
		path = newPath
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("wst")
		if tag == "" {
			continue
		}
		params, err := p.parseTag(tag)
		if err != nil {
			return err
		}

		fieldName, ok := params[paramName]
		if !ok {
			fieldName = field.Name
		}
		fieldValue := structValue.FieldByName(field.Name)
		if fieldData, ok := data[fieldName]; ok {
			if err := p.parseField(fieldData, fieldValue, fieldName, params, path); err != nil {
				return err
			}
		} else {
			if defaultValue, found := params[paramDefault]; found {
				if err := p.processDefaultParam(fieldName, defaultValue, fieldValue); err != nil {
					return err
				}
			} else {
				fieldValue.SetZero()
			}
		}
	}

	return nil
}

func (p ConfigParser) ParseConfig(data map[string]interface{}, config *types.Config, configPath string) error {
	return p.parseStruct(data, config, configPath)
}

func CreateParser(fnd app.Foundation, loader loader.Loader) Parser {
	configParser := &ConfigParser{
		fnd:    fnd,
		loader: loader,
	}

	factories := factory.CreateFactories(fnd, configParser.parseStruct)
	configParser.factories = factories

	return configParser
}
