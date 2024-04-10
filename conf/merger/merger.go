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
	if (src.Kind() == reflect.Ptr || src.Kind() == reflect.Interface) && !src.IsNil() {
		src = src.Elem()
	}
	dstIsPtr := dst.Kind() == reflect.Ptr
	if (dstIsPtr || dst.Kind() == reflect.Interface) && !dst.IsNil() {
		dst = dst.Elem()
	}

	mergedStructPtr := reflect.New(dst.Type())
	mergedStruct := mergedStructPtr.Elem()

	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		dstField := dst.Field(i)

		if srcField.IsZero() && dstField.IsZero() {
			continue
		}

		srcFieldKind := srcField.Kind()
		if srcFieldKind == reflect.Ptr && !srcField.IsNil() {
			srcFieldKind = srcField.Elem().Kind()
		}

		// Perform merging based on the field's kind.
		switch srcFieldKind {
		case reflect.Struct:
			mergedField := m.mergeStructs(dstField, srcField)
			mergedStruct.Field(i).Set(mergedField)
		case reflect.Map:
			mergedMap := m.mergeMaps(dstField, srcField)
			mergedStruct.Field(i).Set(mergedMap)
		case reflect.Slice:
			mergedSlice := m.mergeSlices(dstField, srcField)
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
	if dstIsPtr {
		return mergedStructPtr
	}
	return mergedStruct
}

func (m *nativeMerger) mergeMaps(dst, src reflect.Value) reflect.Value {
	if dst.IsNil() {
		return src
	}
	if src.IsNil() {
		return dst
	}
	if src.Kind() == reflect.Interface || src.Kind() == reflect.Ptr {
		src = src.Elem()
	}
	if dst.Kind() == reflect.Interface || dst.Kind() == reflect.Ptr {
		dst = dst.Elem()
	}

	// Initialize a new map of the same type as src to hold the merged results.
	mergedMap := reflect.MakeMap(src.Type())

	for _, key := range src.MapKeys() {
		srcValue := src.MapIndex(key)
		dstValue := dst.MapIndex(key)

		// Get source value kind.
		srcValueKind := srcValue.Kind()
		if (srcValue.Kind() == reflect.Interface || srcValueKind == reflect.Ptr) && !srcValue.IsNil() {
			srcValueKind = srcValue.Elem().Kind()
		}

		// Get destination value kind.
		dstValueKind := dstValue.Kind()
		if (dstValueKind == reflect.Interface || dstValueKind == reflect.Ptr) && !dstValue.IsNil() {
			dstValueKind = dstValue.Elem().Kind()
		}

		// If there is no corresponding destination value or it's a different kind, simply copy the source value.
		if !dstValue.IsValid() || dstValueKind != srcValueKind {
			mergedMap.SetMapIndex(key, srcValue)
			continue
		}

		// If both source and destination values exist and are of the same kind, merge them based on their type.
		switch srcValueKind {
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
	if dst.IsNil() {
		return src
	}
	if src.IsNil() {
		return dst
	}
	if src.Kind() == reflect.Interface || src.Kind() == reflect.Ptr {
		src = src.Elem()
	}
	if dst.Kind() == reflect.Interface || dst.Kind() == reflect.Ptr {
		dst = dst.Elem()
	}

	maxLength := max(dst.Len(), src.Len())
	newSlice := reflect.MakeSlice(dst.Type(), maxLength, maxLength)
	reflect.Copy(newSlice, dst)

	for i := 0; i < src.Len(); i++ {
		srcElem := src.Index(i)
		if i < dst.Len() {
			dstElem := newSlice.Index(i)
			srcElemKind := srcElem.Kind()
			if (srcElemKind == reflect.Interface || srcElemKind == reflect.Ptr) && !srcElem.IsNil() {
				srcElemKind = srcElem.Elem().Kind()
			}
			// Merge elements based on their kind.
			var mergedVal reflect.Value
			switch srcElemKind {
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
