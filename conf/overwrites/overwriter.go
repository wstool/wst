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

package overwrites

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/parser"
	"github.com/bukka/wst/conf/types"
	"github.com/pkg/errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Overwriter interface {
	Overwrite(config *types.Config, overwrites map[string]string) error
}

type nativeOverwriter struct {
	fnd    app.Foundation
	parser parser.Parser
}

func CreateOverwriter(fnd app.Foundation, parser parser.Parser) Overwriter {
	return &nativeOverwriter{
		fnd:    fnd,
		parser: parser,
	}
}

func (t *nativeOverwriter) Overwrite(config *types.Config, overwrites map[string]string) error {
	dst := reflect.ValueOf(config)
	for key, val := range overwrites {
		ptr := strings.Split(key, ".")
		err := t.overwriteStruct(dst, ptr, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *nativeOverwriter) overwriteStruct(dst reflect.Value, ptrs []string, val string) error {
	if (dst.Kind() == reflect.Ptr || dst.Kind() == reflect.Interface) && !dst.IsNil() {
		dst = dst.Elem()
	}

	if len(ptrs) == 0 {
		return errors.Errorf("overwrite canot be for object")
	}

	ptr := ptrs[0]
	searchedFieldName, index, isSlice, err := t.parseFieldNameAndIndex(ptr)
	if err != nil {
		return err
	}
	ptrs = ptrs[1:]
	for i := 0; i < dst.NumField(); i++ {
		field := dst.Type().Field(i)
		tag := field.Tag.Get("wst")
		if tag == "" {
			continue
		}
		params, err := t.parser.ParseTag(tag)
		if err != nil {
			return err
		}

		fieldName, ok := params[parser.ConfigParamName]
		if !ok {
			fieldName = field.Name
		}
		if searchedFieldName == fieldName {
			fieldValue := dst.Field(i)
			fieldValueKind := fieldValue.Kind()
			if (fieldValueKind == reflect.Ptr || fieldValueKind == reflect.Interface) && !fieldValue.IsNil() {
				fieldValue = fieldValue.Elem()
				fieldValueKind = fieldValue.Kind()
			}

			if fieldValueKind == reflect.Slice {
				if !isSlice {
					return errors.Errorf("array index is missing for the array %s", fieldName)
				}
				return t.overwriteSlice(fieldValue, ptrs, index, val)
			} else {
				if isSlice {
					return errors.Errorf("array index can be used for arrays only")
				}
				switch fieldValue.Kind() {
				case reflect.Struct:
					return t.overwriteStruct(fieldValue, ptrs, val)
				case reflect.Map:
					return t.overwriteMap(fieldValue, ptrs, val)
				default:
					if dst.CanSet() {
						return t.overwriteScalar(fieldValue, ptrs, val)
					} else {
						return errors.Errorf("The value for field %s cannot be set", fieldName)
					}
				}
			}
		}
	}
	return errors.Errorf("overwrite for field %s not found", ptr)
}

func (t *nativeOverwriter) overwriteMap(dst reflect.Value, ptrs []string, val string) error {
	if len(ptrs) == 0 {
		return errors.Errorf("pointer path is empty, cannot proceed with map overwriteation")
	}

	// Ensure the dst is a map and is addressable
	dstKind := dst.Kind()
	if (dstKind == reflect.Ptr || dstKind == reflect.Interface) && !dst.IsNil() {
		dst = dst.Elem()
		dstKind = dst.Kind()
	}

	if dst.Kind() != reflect.Map {
		return errors.Errorf("destination is not a map or cannot be set")
	}

	ptr := ptrs[0]
	searchedKey, index, isSlice, err := t.parseFieldNameAndIndex(ptr)
	if err != nil {
		return err
	}

	if !isSlice && len(ptrs) == 1 {
		// Directly set the value in the map
		keyValue := reflect.ValueOf(ptr)
		valValue := reflect.ValueOf(val)
		dst.SetMapIndex(keyValue, valValue)
		return nil
	}

	// If the map contains the key, get its value; otherwise, return an error
	keyValue := reflect.ValueOf(searchedKey)
	existingValue := dst.MapIndex(keyValue)
	if !existingValue.IsValid() {
		return errors.Errorf("key %s not found in map", ptr)
	}

	existingValueKind := existingValue.Kind()
	if existingValueKind == reflect.Interface && !existingValue.IsNil() {
		existingValue = existingValue.Elem()
		existingValueKind = existingValue.Kind()
	}
	if existingValueKind == reflect.Ptr && !existingValue.IsNil() {
		existingValue = existingValue.Elem()
		existingValueKind = existingValue.Kind()
	}

	// Recursively overwrite the existing value
	ptrs = ptrs[1:] // Move to the next part of the path
	if existingValueKind == reflect.Slice {
		if !isSlice {
			return errors.Errorf("array index is missing for the array %s", searchedKey)
		}
		return t.overwriteSlice(existingValue, ptrs, index, val)
	} else {
		if isSlice {
			return errors.Errorf("array index can be used for arrays only")
		}
		switch existingValueKind {
		case reflect.Struct:
			return t.overwriteStruct(existingValue, ptrs, val)
		case reflect.Map:
			return t.overwriteMap(existingValue, ptrs, val)
		default:
			return t.overwriteScalar(existingValue, ptrs, val)
		}
	}
}

func (t *nativeOverwriter) overwriteSlice(dst reflect.Value, ptrs []string, index int, val string) error {
	// Ensure the dst is a map and is addressable
	dstKind := dst.Kind()
	if (dstKind == reflect.Ptr || dstKind == reflect.Interface) && !dst.IsNil() {
		dst = dst.Elem()
		dstKind = dst.Kind()
	}
	// Ensure the dst is a slice and is addressable
	if dstKind != reflect.Slice {
		return errors.Errorf("destination is not a slice")
	}

	if index < 0 || index >= dst.Len() {
		return errors.Errorf("index out of range: %d", index)
	}

	existingValue := dst.Index(index)
	if len(ptrs) == 0 {
		return t.overwriteScalar(existingValue, ptrs, val)
	}

	existingValueKind := existingValue.Kind()
	if existingValueKind == reflect.Interface && !existingValue.IsNil() {
		existingValue = existingValue.Elem()
		existingValueKind = existingValue.Kind()
	}
	if existingValueKind == reflect.Ptr && !existingValue.IsNil() {
		existingValue = existingValue.Elem()
		existingValueKind = existingValue.Kind()
	}

	// Recursively overwrite the existing value
	if existingValueKind == reflect.Map {
		return t.overwriteMap(existingValue, ptrs, val)
	} else if existingValueKind == reflect.Slice {
		return errors.Errorf("array of arrays are not supported for overwrites")
	} else if existingValueKind == reflect.Struct {
		return t.overwriteStruct(existingValue, ptrs, val)
	} else {
		return t.overwriteScalar(existingValue, ptrs, val)
	}
}

func (t *nativeOverwriter) overwriteScalar(dst reflect.Value, ptrs []string, val string) error {
	switch dst.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		intVal, err := strconv.ParseInt(val, 10, 64) // Convert string to int64
		if err != nil {
			return errors.Errorf("failed to convert %s to integer: %v", val, err)
		}
		if dst.OverflowInt(intVal) {
			return errors.Errorf("value %s overflows", val)
		}
		dst.SetInt(intVal) // Set the converted integer value

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(val, 64) // Convert string to float64
		if err != nil {
			return errors.Errorf("failed to convert %s to float: %v", val, err)
		}
		if dst.OverflowFloat(floatVal) {
			return errors.Errorf("value %s overflows", val)
		}
		dst.SetFloat(floatVal) // Set the converted float value

	default:
		if len(ptrs) > 0 {
			return errors.Errorf("overwrite field %s is not nestable", dst)
		}
		// For string fields, directly set the value
		if dst.Kind() == reflect.String {
			dst.SetString(val)
		} else {
			return errors.Errorf("unsupported value kind %s", dst.Kind())
		}
	}

	return nil
}

func (t *nativeOverwriter) parseFieldNameAndIndex(ptr string) (fieldName string, index int, isSlice bool, err error) {
	// Example: "instances[0]" -> fieldName: "instances", index: 0
	re := regexp.MustCompile(`^([a-zA-Z_]\w*)\[(-?\d*)\]$`)
	matches := re.FindStringSubmatch(ptr)
	if matches != nil {
		fieldName = matches[1]
		index, err = strconv.Atoi(matches[2])
		if err != nil {
			return "", 0, false, errors.Errorf("invalid index value in %s", ptr)
		}
		return fieldName, index, true, nil
	}
	return ptr, 0, false, nil // No index notation, return original ptr as fieldName
}
