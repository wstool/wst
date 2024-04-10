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

package merger

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"reflect"
)

type Merger interface {
	MergeConfigs(configs []*types.Config, overwrites map[string]string) (*types.Config, error)
}

type nativeMerger struct {
	fnd app.Foundation
}

func CreateMerger(fnd app.Foundation) Merger {
	return &nativeMerger{
		fnd: fnd,
	}
}

func (m *nativeMerger) MergeConfigs(configs []*types.Config, overwrites map[string]string) (*types.Config, error) {
	if len(configs) == 0 {
		return nil, nil // or return an empty config or an error
	}

	mergedConfig := &types.Config{}
	for _, config := range configs {
		mergedStructValue := m.mergeStructs(reflect.ValueOf(mergedConfig).Elem(), reflect.ValueOf(config).Elem())
		if newMergedConfig, ok := mergedStructValue.Interface().(types.Config); ok {
			mergedConfig = &newMergedConfig
		} else {
			return nil, fmt.Errorf("failed to merge configs")
		}
	}

	// TODO: Apply overwrites to merged here, if needed

	return mergedConfig, nil
}

func (m *nativeMerger) mergeStructs(dst, src reflect.Value) reflect.Value {
	// Create a new instance of the struct based on dst's type.
	mergedStruct := reflect.New(dst.Type()).Elem()

	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		dstField := dst.Field(i)

		if srcField.IsZero() && dstField.IsZero() {
			continue
		}

		// Dereference pointers if necessary.
		if dstField.Kind() == reflect.Ptr && !dstField.IsNil() {
			dstField = dstField.Elem()
		}
		if srcField.Kind() == reflect.Ptr && !srcField.IsNil() {
			srcField = srcField.Elem()
		}

		// Perform merging based on the field's kind.
		switch srcField.Kind() {
		case reflect.Struct:
			mergedField := m.mergeStructs(dstField, srcField) // Recursive call for nested structs.
			mergedStruct.Field(i).Set(mergedField)            // Set the merged result to the corresponding field in the new struct.
		case reflect.Map:
			mergedMap := m.mergeMaps(dstField, srcField) // Assume mergeMaps also returns a reflect.Value now.
			mergedStruct.Field(i).Set(mergedMap)
		case reflect.Slice:
			mergedSlice := m.mergeSlices(dstField, srcField) // Assume mergeSlices returns a reflect.Value.
			mergedStruct.Field(i).Set(mergedSlice)
		default:
			// For simple types, use the source field if it's not zero, otherwise retain the destination field.
			if !srcField.IsZero() {
				mergedStruct.Field(i).Set(srcField)
			} else {
				mergedStruct.Field(i).Set(dstField)
			}
		}
	}

	return mergedStruct
}

func (m *nativeMerger) mergeMaps(dst, src reflect.Value) reflect.Value {
	// Initialize a new map of the same type as src to hold the merged results.
	mergedMap := reflect.MakeMap(src.Type())

	for _, key := range src.MapKeys() {
		srcValue := src.MapIndex(key)

		// Attempt to get the corresponding value from dst, if it exists.
		var dstValue reflect.Value
		if !dst.IsNil() {
			dstValue = dst.MapIndex(key)
		}

		// Handle if the source value is a pointer or interface and dereference if necessary.
		if (srcValue.Kind() == reflect.Interface || srcValue.Kind() == reflect.Ptr) && !srcValue.IsNil() {
			srcValue = srcValue.Elem()
		}

		// Handle if the destination value is a pointer or interface and dereference if necessary.
		if (dstValue.Kind() == reflect.Interface || dstValue.Kind() == reflect.Ptr) && !dstValue.IsNil() {
			dstValue = dstValue.Elem()
		}

		// If there is no corresponding destination value or it's a different kind, simply copy the source value.
		if !dstValue.IsValid() || dstValue.Kind() != srcValue.Kind() {
			mergedMap.SetMapIndex(key, srcValue)
			continue
		}

		// If both source and destination values exist and are of the same kind, merge them based on their type.
		switch srcValue.Kind() {
		case reflect.Map:
			mergedValue := m.mergeMaps(dstValue, srcValue) // Recursively merge maps.
			mergedMap.SetMapIndex(key, mergedValue)
		case reflect.Struct:
			mergedValue := m.mergeStructs(dstValue, srcValue) // Recursively merge structs.
			mergedMap.SetMapIndex(key, mergedValue)
		case reflect.Slice, reflect.Array:
			mergedValue := m.mergeSlices(dstValue, srcValue) // Merge slices.
			mergedMap.SetMapIndex(key, mergedValue)
		default:
			// For non-composite types or if types do not match, overwrite with source value.
			mergedMap.SetMapIndex(key, srcValue)
		}
	}

	for _, key := range dst.MapKeys() {
		srcValue := src.MapIndex(key)
		if !srcValue.IsValid() {
			dstValue := dst.MapIndex(key)
			mergedMap.SetMapIndex(key, dstValue)
		}
	}

	return mergedMap
}

func (m *nativeMerger) mergeSlices(dst, src reflect.Value) reflect.Value {
	maxLength := max(dst.Len(), src.Len())
	newSlice := reflect.MakeSlice(dst.Type(), maxLength, maxLength)
	reflect.Copy(newSlice, dst)

	for i := 0; i < src.Len(); i++ {
		srcElem := src.Index(i)
		if i < dst.Len() {
			dstElem := newSlice.Index(i)
			// Merge elements based on their kind.
			var mergedVal reflect.Value
			switch srcElem.Kind() {
			case reflect.Struct:
				mergedVal = m.mergeStructs(dstElem, srcElem)
			case reflect.Map:
				mergedVal = m.mergeMaps(dstElem, srcElem)
			case reflect.Slice, reflect.Array:
				mergedVal = m.mergeSlices(dstElem, srcElem)
			default:
				mergedVal = srcElem
			}
			dstElem.Set(mergedVal)
		} else {
			// Copy srcElem to newSlice if it's beyond the original dst length.
			newSlice.Index(i).Set(srcElem)
		}
	}

	return newSlice // Return the merged slice.
}
