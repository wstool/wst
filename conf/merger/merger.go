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

func (n *nativeMerger) MergeConfigs(configs []*types.Config, overwrites map[string]string) (*types.Config, error) {
	if len(configs) == 0 {
		return nil, nil // or return an empty config or an error
	}

	merged := &types.Config{}
	for _, config := range configs {
		mergeStructs(reflect.ValueOf(merged).Elem(), reflect.ValueOf(config).Elem())
	}

	// TODO: Apply overwrites to merged here, if needed

	return merged, nil
}

func mergeStructs(dst, src reflect.Value) {
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		dstField := dst.Field(i)

		switch srcField.Kind() {
		case reflect.Struct:
			mergeStructs(dstField, srcField)
		case reflect.Map:
			mergeMaps(dstField, srcField)
		case reflect.Slice:
			mergeSlices(dstField, srcField)
		default:
			if !isEmptyValue(srcField) {
				dstField.Set(srcField)
			}
		}
	}
}

func mergeMaps(dst, src reflect.Value) {
	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}
	for _, key := range src.MapKeys() {
		srcValue := src.MapIndex(key)
		dstValue := dst.MapIndex(key)

		// If key exists in both source and destination.
		if dstValue.IsValid() {
			// Determine the kind of srcValue and proceed accordingly.
			switch srcValue.Kind() {
			case reflect.Map:
				if dstValue.Kind() == reflect.Map {
					mergeMaps(dstValue, srcValue)
				} else {
					// Inconsistent types - overwrite.
					dst.SetMapIndex(key, srcValue)
				}
			case reflect.Struct:
				if dstValue.Kind() == reflect.Struct {
					mergeStructs(dstValue, srcValue)
				} else {
					// Inconsistent types - overwrite.
					dst.SetMapIndex(key, srcValue)
				}
			case reflect.Slice, reflect.Array:
				if dstValue.Kind() == reflect.Slice || dstValue.Kind() == reflect.Array {
					mergeSlices(dstValue, srcValue)
				} else {
					// Inconsistent types - overwrite.
					dst.SetMapIndex(key, srcValue)
				}
			default:
				// For non-composite types, or if types do not match, overwrite.
				dst.SetMapIndex(key, srcValue)
			}
		} else {
			// Key doesn't exist in dst, set the src value directly.
			dst.SetMapIndex(key, srcValue)
		}
	}
}

func mergeSlices(dst, src reflect.Value) {
	// First, check if the destination is a slice.
	// This check ensures we're only attempting to merge slices into slices.
	if dst.Kind() != reflect.Slice {
		fmt.Errorf("destination is not a slice, got %s", dst.Kind().String())
		return // Optionally return an error instead of just exiting.
	}

	// Ensure dst slice is large enough to contain all src elements.
	maxLength := src.Len()
	if dst.Len() < maxLength {
		// Extend the destination slice to match the source slice length, if necessary.
		newDst := reflect.MakeSlice(dst.Type(), maxLength, maxLength)
		reflect.Copy(newDst, dst)
		dst.Set(newDst)
	}

	// Iterate over src slice and merge elements into dst slice based on their kind.
	for i := 0; i < src.Len(); i++ {
		srcElem := src.Index(i)
		dstElem := dst.Index(i)
		// Check the kinds of srcElem and dstElem to decide on merging strategy.
		switch srcElem.Kind() {
		case reflect.Struct:
			if dstElem.Kind() == reflect.Struct {
				mergeStructs(dstElem, srcElem)
			} else {
				dstElem.Set(srcElem) // Potentially handle type mismatch differently.
			}
		case reflect.Map:
			if dstElem.Kind() == reflect.Map {
				mergeMaps(dstElem, srcElem)
			} else {
				dstElem.Set(srcElem) // Potentially handle type mismatch differently.
			}
		case reflect.Slice, reflect.Array:
			if dstElem.Kind() == reflect.Slice || dstElem.Kind() == reflect.Array {
				mergeSlices(dstElem, srcElem)
			} else {
				dstElem.Set(srcElem) // Potentially handle type mismatch differently.
			}
		default:
			// For non-composite types, simply overwrite the destination element.
			dstElem.Set(srcElem)
		}
	}
}
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
