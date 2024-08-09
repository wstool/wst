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
	"github.com/bukka/wst/conf/parser/location"
	"github.com/bukka/wst/conf/types"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"math"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type ConfigParam string

const (
	ConfigParamName     ConfigParam = "name"
	ConfigParamLoadable             = "loadable"
	ConfigParamDefault              = "default"
	ConfigParamFactory              = "factory"
	ConfigParamEnum                 = "enum"
	ConfigParamKeys                 = "keys"
	ConfigParamPath                 = "path"
	ConfigParamString               = "string"
)

const pathKey = "wst/path"

type Parser interface {
	ParseConfig(data map[string]interface{}, config *types.Config, configPath string) error
	ParseStruct(data map[string]interface{}, structure interface{}, configPath string) error
	ParseTag(tag string) (map[ConfigParam]string, error)
}

type ConfigParser struct {
	fnd       app.Foundation
	loader    loader.Loader
	factories factory.Functions
	loc       *location.Location
}

// check if param is a valid param (one of param* constants)
func isValidParam(param string) bool {
	switch ConfigParam(param) {
	case ConfigParamName:
		fallthrough
	case ConfigParamLoadable:
		fallthrough
	case ConfigParamDefault:
		fallthrough
	case ConfigParamFactory:
		fallthrough
	case ConfigParamEnum:
		fallthrough
	case ConfigParamKeys:
		fallthrough
	case ConfigParamPath:
		fallthrough
	case ConfigParamString:
		return true
	default:
		return false
	}
}

func (p *ConfigParser) Pos() string {
	pos := p.loc.String()
	if pos == "" {
		return "unknown"
	}
	return pos
}

// ParseTag parses the 'wst' struct tag into a Field
func (p *ConfigParser) ParseTag(tag string) (map[ConfigParam]string, error) {
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
		params[ConfigParamName] = firstPart
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
			return nil, errors.Errorf("field %s invalid parameter key: %s", p.Pos(), key)
		}
	}

	return params, nil
}

func (p *ConfigParser) processDefaultParam(fieldName string, defaultValue string, fieldValue reflect.Value) error {
	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		intValue, err := strconv.Atoi(defaultValue)
		if err != nil {
			return errors.Errorf("default value %q for field %s can't be converted to int: %v",
				defaultValue, p.Pos(), err)
		}
		fieldValue.SetInt(int64(intValue))
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(defaultValue)
		if err != nil {
			return errors.Errorf("default value %q for field %s can't be converted to bool: %v",
				defaultValue, p.Pos(), err)
		}
		fieldValue.SetBool(boolValue)
	case reflect.String:
		fieldValue.SetString(defaultValue)
	default:
		return errors.Errorf("default value %v for field %s cannot be converted to type %v",
			defaultValue, p.Pos(), fieldValue.Type())
	}
	return nil
}

func (p *ConfigParser) processFactoryParam(
	factory string,
	data interface{},
	fieldValue reflect.Value,
	path string,
) error {
	factoryFunc, err := p.factories.GetFactoryFunc(factory)
	if err != nil {
		return err
	}
	return factoryFunc(data, fieldValue, path)
}

func (p *ConfigParser) processEnumParam(enums string, data interface{}, fieldName string) error {
	enumList := strings.Split(enums, "|")
	for _, enum := range enumList {
		if enum == data {
			return nil
		}
	}

	return errors.Errorf("values %v are not valid for field %s", enums, p.Pos())
}

func (p *ConfigParser) processKeysParam(keys string, data interface{}, fieldName string) error {
	keysList := strings.Split(keys, "|")
	for _, key := range keysList {
		if _, ok := data.(map[string]interface{})[key]; ok {
			return nil
		}
	}

	return errors.Errorf("keys %v are not valid for field %s", keys, p.Pos())
}

func (p *ConfigParser) processPathParam(data interface{}, fieldValue reflect.Value, fieldName string, configPath string) error {
	// Assert data is a string
	path, ok := data.(string)
	if !ok {
		return errors.Errorf("unexpected type %T for data in field %s, expected string", data, p.Pos())
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
		return errors.Errorf("file path %s for field %s does not exist", path, p.Pos())
	}

	// Check that fieldValue is settable (it is addressable and was not obtained by
	// the use of unexported struct fields)
	if !fieldValue.CanSet() {
		return errors.Errorf("field %s is not settable", p.Pos())
	}

	// Check if fieldValue is of type string
	if fieldValue.Kind() != reflect.String {
		return errors.Errorf("field %s is not of type string", p.Pos())
	}

	// If no errors so far, set the fieldValue to the resolved file path
	fieldValue.SetString(path)

	return nil
}

func (p *ConfigParser) processLoadableParam(data interface{}, fieldValue reflect.Value, path string) (interface{}, error) {
	loadableData, isString := data.(string)
	if isString {
		configs, err := p.loader.GlobConfigs(loadableData, filepath.Dir(path))
		if err != nil {
			return nil, errors.Errorf("loading configs for field %s: %v", p.Pos(), err)
		}

		switch fieldValue.Kind() {
		case reflect.Map:
			loadedData := make(map[string]interface{})
			for _, config := range configs {
				data := config.Data()
				data[pathKey] = config.Path()
				loadedData = data
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
			return nil, errors.Errorf(
				"type of field %s is neither map nor slice (kind=%s)",
				p.Pos(),
				fieldValue.Kind().String(),
			)
		}
	}
	return data, nil
}

func (p *ConfigParser) processStringParam(
	fieldName string,
	data interface{},
	fieldValue reflect.Value,
	path string,
) (bool, error) {
	kind := fieldValue.Kind()

	if kind != reflect.Struct && kind != reflect.Map && kind != reflect.Interface && kind != reflect.Ptr {
		return false, errors.Errorf("field %s must be a struct or interface type or a pointer to such", p.Pos())
	}

	if kind == reflect.Ptr || kind == reflect.Interface {
		fieldValue = fieldValue.Elem()
	}
	fieldValueKind := fieldValue.Kind()

	if fieldValueKind == reflect.Map {
		mapData, ok := data.(map[string]interface{})
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
		err := p.ParseStruct(make(map[string]interface{}), fieldValuePtrInterface, path)
		if err != nil {
			return false, errors.Errorf("error parsing struct in field %s for string param: %v", p.Pos(), err)
		}

		// Set the string value to the appropriate sub-field
		err = p.setFieldByName(fieldValuePtrInterface, fieldName, strData)
		if err != nil {
			return false, errors.Errorf("failed to set field %s : %v", p.Pos(), err)
		}
	} else {
		return false, errors.Errorf("field %s value must be a pointer to a struct or a map", p.Pos())
	}
	return true, nil
}

// setFieldByName sets the field of the struct v with the given name to the specified value.
// The value v must be a pointer to a struct, and the field should be exported and settable.
func (p *ConfigParser) setFieldByName(v interface{}, name string, value string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct {
		// For struct:
		fv := rv.Elem().FieldByName(name)
		if !fv.IsValid() {
			return errors.Errorf("field %s does not have a valid field name: %s", p.Pos(), name)
		}
		if !fv.CanSet() {
			return errors.Errorf("field %s cannot set the field: %s", p.Pos(), name)
		}
		fv.SetString(value)
	}

	return nil
}

// processMapValue processes a map[string]interface{} where value is a string
// It goes through each value in map and if it's a string, sets it to a specified field.
func (p *ConfigParser) processMapValue(
	mapData map[string]interface{},
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

		var elemData map[string]interface{}
		mapVal, valIsMap := val.(map[string]interface{})
		if valIsMap {
			elemData = mapVal
		} else {
			elemData = make(map[string]interface{})
		}

		// Use an empty map as temporary data to populate the struct
		err := p.ParseStruct(elemData, newElem.Interface(), path)
		if err != nil {
			return errors.Errorf("field %s error parsing struct for string param: %v", p.Pos(), err)
		}

		if !valIsMap {
			// Set the string value to the appropriate sub-field
			strVal, ok := val.(string)
			if !ok {
				return errors.Errorf(
					"field %s invalid map value type for string param - expected string, got %T",
					p.Pos(),
					val,
				)
			}
			err = p.setFieldByName(newElem.Interface(), fieldName, strVal)
			if err != nil {
				return errors.Errorf("failed to set field %s: %v", p.Pos(), err)
			}
		}

		// Insert the new element into the result map
		resultMap.SetMapIndex(reflect.ValueOf(key), newElem.Elem())
	}

	rv.Set(resultMap)
	return nil
}

// assignField assigns the provided data to the fieldValue.
func (p *ConfigParser) assignField(data interface{}, fieldValue reflect.Value, fieldName string, path string) error {
	switch fieldValue.Kind() {
	case reflect.Struct:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return errors.Errorf("unable to convert data for field %s to map[string]interface{}", p.Pos())
		}
		return p.ParseStruct(dataMap, fieldValue.Addr().Interface(), path)
	case reflect.Map:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return errors.Errorf("unable to convert data for field %s to map[string]interface{}", p.Pos())
		}
		if fieldValue.IsNil() {
			// Initialize a new map
			fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
		}
		p.loc.StartObject()
		for key, val := range dataMap {
			p.loc.SetField(key)
			newVal := reflect.New(fieldValue.Type().Elem())
			err := p.assignField(val, newVal.Elem(), fieldName, path)
			if err != nil {
				return err
			}
			fieldValue.SetMapIndex(reflect.ValueOf(key), newVal.Elem())
		}
		p.loc.EndObject()
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
			return errors.Errorf("unable to convert data for field %s to []interface{}", p.Pos())
		}
		if fieldValue.IsNil() || fieldValue.Len() < len(dataSlice) {
			// Make a new slice to accommodate all elements
			fieldValue.Set(reflect.MakeSlice(fieldValue.Type(), len(dataSlice), len(dataSlice)))
		}
		p.loc.StartArray()
		for i, val := range dataSlice {
			p.loc.SetIndex(i)
			err := p.assignField(val, fieldValue.Index(i), fmt.Sprintf("%s[%d]", fieldName, i), path)
			if err != nil {
				return err
			}
		}
		p.loc.EndArray()
	default:
		v := reflect.ValueOf(data)
		targetType := fieldValue.Type()
		sourceType := v.Type()

		// Handle the case where the types are directly assignable
		if sourceType.AssignableTo(targetType) {
			fieldValue.Set(v)
			return nil
		}
		sourceKind := sourceType.Kind()

		// Specific conversion from float / int64 to int32 or int16 with overflow check
		isFloat := sourceKind == reflect.Float64 || sourceKind == reflect.Float32
		if isFloat || sourceKind == reflect.Int || sourceKind == reflect.Int64 {
			var intVal int64
			if isFloat {
				floatVal := v.Float()

				// Check if the float64 has a fractional part
				if floatVal == math.Floor(floatVal) {
					intVal = int64(floatVal)
				} else {
					return errors.Errorf("field %s could not be set: "+
						"float64 value has a fractional part and cannot be converted to integer types", p.Pos())
				}
			} else {
				intVal = v.Int()
			}

			targetKind := targetType.Kind()
			switch targetKind {
			case reflect.Int64:
				fieldValue.SetInt(intVal)
				return nil
			case reflect.Int32:
				if intVal < -1<<31 || intVal > 1<<31-1 {
					return errors.Errorf(
						"field %s overflow error: int64 value %d is out of range for int32",
						fieldName,
						intVal,
					)
				}
				fieldValue.SetInt(intVal)
				return nil
			case reflect.Int16:
				if intVal < -1<<15 || intVal > 1<<15-1 {
					return errors.Errorf(
						"field %s overflow error: int64 value %d is out of range for int16",
						fieldName,
						intVal,
					)
				}
				fieldValue.SetInt(intVal)
				return nil
			case reflect.Int8:
				if intVal < -1<<7 || intVal > 1<<7-1 {
					return errors.Errorf(
						"field %s overflow error: int64 value %d is out of range for int8",
						fieldName,
						intVal,
					)
				}
				fieldValue.SetInt(intVal)
				return nil
			case reflect.String:
				return errors.Errorf("field %s is an integer and cannot be converted to string", p.Pos())
			}
		}

		// General case for type conversion if it's convertible
		if v.Type().ConvertibleTo(fieldValue.Type()) {
			convertedValue := v.Convert(fieldValue.Type())
			fieldValue.Set(convertedValue)
			return nil
		}

		return errors.Errorf("field %s could not be set due to type mismatch or non-convertible types", p.Pos())

	}
	return nil
}

// parseField parses a struct field based on data and params
func (p *ConfigParser) parseField(
	data interface{},
	fieldValue reflect.Value,
	fieldName string,
	params map[ConfigParam]string,
	path string,
) error {
	var err error

	if _, isLoadable := params[ConfigParamLoadable]; isLoadable {
		if data, err = p.processLoadableParam(data, fieldValue, path); err != nil {
			return err
		}
	}

	if factoryName, hasFactory := params[ConfigParamFactory]; hasFactory {
		if err = p.processFactoryParam(factoryName, data, fieldValue, path); err != nil {
			return err
		}
		// factory should set everything so there is no need to continue
		return nil
	}

	if stringValue, hasString := params[ConfigParamString]; hasString {
		var done bool
		if done, err = p.processStringParam(stringValue, data, fieldValue, path); err != nil {
			return err
		}
		if done {
			return nil
		}
	}

	if enums, hasEnum := params[ConfigParamEnum]; hasEnum {
		if err = p.processEnumParam(enums, data, fieldName); err != nil {
			return err
		}
	}

	if keys, hasKeys := params[ConfigParamKeys]; hasKeys {
		if err = p.processKeysParam(keys, data, fieldName); err != nil {
			return err
		}
	}

	if _, hasPath := params[ConfigParamPath]; hasPath {
		if err = p.processPathParam(data, fieldValue, fieldName, path); err != nil {
			return err
		}
	} else if err = p.assignField(data, fieldValue, fieldName, path); err != nil {
		return err
	}

	return nil
}

// ParseStruct parses a struct into a map of Fields
func (p *ConfigParser) ParseStruct(
	data map[string]interface{},
	structure interface{},
	configPath string,
) error {
	p.loc.StartObject()
	structValuePtr := reflect.ValueOf(structure)
	if structValuePtr.Kind() != reflect.Ptr || structValuePtr.Elem().Kind() != reflect.Struct {
		return errors.Errorf("field %s expected a pointer to a struct, got %T", p.Pos(), structure)
	}
	structValue := structValuePtr.Elem()
	structType := structValue.Type()

	if newPathInterface, newPathFound := data[pathKey]; newPathFound {
		newPath, ok := newPathInterface.(string)
		if !ok {
			return errors.Errorf("field %s unexpected type %T for path", p.Pos(), newPathInterface)
		}
		configPath = newPath
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("wst")
		if tag == "" {
			continue
		}
		p.loc.SetField(field.Name)
		params, err := p.ParseTag(tag)
		if err != nil {
			return err
		}

		fieldName, ok := params[ConfigParamName]
		if !ok {
			fieldName = field.Name
		} else {
			p.loc.SetField(fieldName)
		}
		fieldValue := structValue.FieldByName(field.Name)
		if fieldData, ok := data[fieldName]; ok {
			if err = p.parseField(fieldData, fieldValue, fieldName, params, configPath); err != nil {
				return err
			}
		} else if defaultValue, found := params[ConfigParamDefault]; found {
			if err = p.processDefaultParam(fieldName, defaultValue, fieldValue); err != nil {
				return err
			}
		}
	}
	p.loc.EndObject()

	return nil
}

func (p *ConfigParser) ParseConfig(data map[string]interface{}, config *types.Config, configPath string) error {
	p.loc.Reset()
	return p.ParseStruct(data, config, configPath)
}

func CreateParser(fnd app.Foundation, loader loader.Loader) Parser {
	configParser := &ConfigParser{
		fnd:    fnd,
		loader: loader,
		loc:    location.CreateLocation(),
	}

	factories := factory.CreateFactories(fnd, configParser.ParseStruct, pathKey, configParser.loc)
	configParser.factories = factories

	return configParser
}
