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

	merged := &types.Config{}
	for _, config := range configs {
		m.mergeStructs(reflect.ValueOf(merged).Elem(), reflect.ValueOf(config).Elem())
	}

	// TODO: Apply overwrites to merged here, if needed

	return merged, nil
}

func (m *nativeMerger) mergeStructs(dst, src reflect.Value) {
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		dstField := dst.Field(i)

		if dstField.Kind() == reflect.Ptr && !dstField.IsNil() {
			dstField = dstField.Elem()
		}

		if srcField.Kind() == reflect.Ptr && !srcField.IsNil() {
			srcField = srcField.Elem()
		}

		switch srcField.Kind() {
		case reflect.Struct:
			if dstField.CanAddr() {
				m.mergeStructs(dstField.Addr().Elem(), srcField)
			} else {
				temp := reflect.New(dstField.Type()).Elem()
				m.mergeStructs(temp, srcField)
				dstField.Set(temp)
			}
		case reflect.Map:
			m.mergeMaps(dstField, srcField)
		case reflect.Slice:
			m.mergeSlices(dstField, srcField)
		default:
			if !srcField.IsZero() {
				dstField.Set(srcField)
			}
		}
	}
}

func (m *nativeMerger) mergeMaps(dst, src reflect.Value) {
	if dst.IsNil() {
		dst.Set(reflect.MakeMap(src.Type()))
	}

	for _, key := range src.MapKeys() {
		srcValue := src.MapIndex(key)
		dstValue := dst.MapIndex(key)

		// Handle if the source value is a pointer and dereference if necessary
		if srcValue.Kind() == reflect.Ptr && !srcValue.IsNil() {
			srcValue = srcValue.Elem()
		}

		// Create a new value if dstValue is not valid or both src and dst are maps
		if !dstValue.IsValid() || (srcValue.Kind() == reflect.Map && (!dstValue.IsValid() || dstValue.Kind() == reflect.Map)) {
			if srcValue.Kind() == reflect.Map {
				if !dstValue.IsValid() || dstValue.Kind() == reflect.Map {
					// Initialize new map or merge existing maps
					newMap := reflect.MakeMap(srcValue.Type())
					if dstValue.IsValid() {
						m.mergeMaps(newMap, srcValue)
						dst.SetMapIndex(key, newMap)
					} else {
						dst.SetMapIndex(key, srcValue)
					}
				}
			} else {
				// Directly set srcValue in dst for non-map types
				dst.SetMapIndex(key, srcValue)
			}
		} else if srcValue.Kind() == reflect.Struct || (srcValue.Kind() == reflect.Ptr && srcValue.Elem().Kind() == reflect.Struct) {
			// Handle merging of struct types, including handling pointers to structs
			if dstValue.Kind() == reflect.Ptr && !dstValue.IsNil() {
				// Merge into existing struct pointer
				m.mergeStructs(dstValue.Elem(), srcValue.Elem())
			} else if dstValue.Kind() == reflect.Struct {
				// Merge into existing struct
				m.mergeStructs(dstValue, srcValue)
			} else {
				// Replace destination with source value
				dst.SetMapIndex(key, srcValue)
			}
		} else {
			// For all other cases, replace destination with source value
			dst.SetMapIndex(key, srcValue)
		}
	}
}

func (m *nativeMerger) mergeSlices(dst, src reflect.Value) {
	maxLength := max(dst.Len(), src.Len())
	newSlice := reflect.MakeSlice(dst.Type(), maxLength, maxLength)
	reflect.Copy(newSlice, dst)

	for i := 0; i < src.Len(); i++ {
		srcElem := src.Index(i)
		if i < dst.Len() {
			dstElem := newSlice.Index(i)
			switch srcElem.Kind() {
			case reflect.Struct:
				m.mergeStructs(dstElem, srcElem)
			case reflect.Map:
				m.mergeMaps(dstElem, srcElem)
			case reflect.Slice, reflect.Array:
				m.mergeSlices(dstElem, srcElem)
			default:
				dstElem.Set(srcElem)
			}
		} else {
			newSlice.Index(i).Set(srcElem)
		}
	}

	dst.Set(newSlice)
}

// Helper function to find the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
