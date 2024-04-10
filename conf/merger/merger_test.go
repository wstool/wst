package merger

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestCreateMerger(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	tests := []struct {
		name string
		fnd  app.Foundation
	}{
		{
			name: "create merger",
			fnd:  fndMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := CreateMerger(tt.fnd)
			actualMerger, ok := merger.(*nativeMerger)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, actualMerger.fnd)
		})
	}
}

func Test_nativeMerger_MergeConfigs(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}

	tests := []struct {
		name           string
		configs        []*types.Config
		expectedConfig *types.Config
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "Merge basic fields and complex structures",
			configs:        []*types.Config{},
			expectedConfig: nil,
			wantErr:        true,
			errMsg:         "no config has been provided for merging",
		},
		{
			name: "Merge basic fields and complex structures",
			configs: []*types.Config{
				{
					Version:     "1.0",
					Name:        "Config 1",
					Description: "Description 1",
					Spec: types.Spec{
						Environments: map[string]types.Environment{
							"common": &types.CommonEnvironment{
								Ports: types.EnvironmentPorts{Start: 8000, End: 8080},
							},
							"docker": &types.DockerEnvironment{
								NamePrefix: "prefix-1",
							},
						},
						Servers: []types.Server{
							{Name: "Server 1", User: "user1"},
						},
					},
				},
				{
					Version:     "1.0",
					Name:        "Config 2",
					Description: "Description 2",
					Spec: types.Spec{
						Environments: map[string]types.Environment{
							"docker": &types.DockerEnvironment{
								NamePrefix: "prefix-2",
							},
						},
						Servers: []types.Server{
							{Name: "Server 2", User: "user2"},
						},
					},
				},
			},
			expectedConfig: &types.Config{
				Version:     "1.0",
				Name:        "Config 2",
				Description: "Description 2",
				Spec: types.Spec{
					Environments: map[string]types.Environment{
						"common": &types.CommonEnvironment{
							Ports: types.EnvironmentPorts{Start: 8000, End: 8080},
						},
						"docker": &types.DockerEnvironment{
							NamePrefix: "prefix-2",
						},
					},
					Servers: []types.Server{
						{Name: "Server 2", User: "user2"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := &nativeMerger{
				fnd: fndMock,
			}
			config, err := merger.MergeConfigs(tt.configs)

			if tt.errMsg != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, tt.expectedConfig, config)
				}
			}
		})
	}
}

func Test_nativeMerger_mergeStructs(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}

	type nestedStruct struct {
		Field1 string
		Field2 int
	}
	type testStruct struct {
		StrField       string
		IntField       int
		StructField    nestedStruct
		PtrStructField *nestedStruct
		MapField       map[string]int
		SliceField     []string
	}

	tests := []struct {
		name string
		dst  interface{}
		src  interface{}
		want interface{}
	}{
		{
			name: "simple fields",
			dst:  &testStruct{},
			src: &testStruct{
				StrField: "test",
				IntField: 42,
			},
			want: testStruct{
				StrField: "test",
				IntField: 42,
			},
		},
		{
			name: "nested struct",
			dst:  &testStruct{},
			src: &testStruct{
				StructField: nestedStruct{
					Field1: "nested",
					Field2: 100,
				},
			},
			want: testStruct{
				StructField: nestedStruct{
					Field1: "nested",
					Field2: 100,
				},
			},
		},

		{
			name: "nested struct references",
			dst: &testStruct{
				PtrStructField: &nestedStruct{
					Field1: "nested 1",
					Field2: 100,
				},
			},
			src: &testStruct{
				PtrStructField: &nestedStruct{
					Field2: 200,
				},
			},
			want: testStruct{
				PtrStructField: &nestedStruct{
					Field1: "nested 1",
					Field2: 200,
				},
			},
		},
		{
			name: "map field",
			dst: &testStruct{
				MapField: map[string]int{
					"existing": 1,
				},
			},
			src: &testStruct{
				MapField: map[string]int{
					"new": 2,
				},
			},
			want: testStruct{
				MapField: map[string]int{
					"existing": 1,
					"new":      2,
				},
			},
		},
		{
			name: "slice field",
			dst: &testStruct{
				SliceField: []string{"one", "two"},
			},
			src: &testStruct{
				SliceField: []string{"three"},
			},
			want: testStruct{
				SliceField: []string{"three", "two"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := &nativeMerger{
				fnd: fndMock,
			}
			mergedStruct := merger.mergeStructs(reflect.ValueOf(tt.dst).Elem(), reflect.ValueOf(tt.src).Elem())
			assert.Equal(t, tt.want, mergedStruct.Interface())
		})
	}
}

func Test_nativeMerger_mergeMaps(t *testing.T) {
	type complexStruct struct {
		NestedField1 int
		NestedField2 string
	}
	tests := []struct {
		name string
		dst  map[string]interface{}
		src  map[string]interface{}
		want map[string]interface{}
	}{
		{
			name: "simple struct merge with single item",
			dst: map[string]interface{}{
				"key": complexStruct{NestedField1: 1},
			},
			src: map[string]interface{}{
				"key": complexStruct{NestedField1: 2},
			},
			want: map[string]interface{}{
				"key": complexStruct{NestedField1: 2},
			},
		},
		{
			name: "simple struct merge with different number of items",
			dst: map[string]interface{}{
				"key1": complexStruct{NestedField1: 1},
				"key2": complexStruct{NestedField1: 2},
			},
			src: map[string]interface{}{
				"key1": complexStruct{NestedField1: 3},
			},
			want: map[string]interface{}{
				"key1": complexStruct{NestedField1: 3},
				"key2": complexStruct{NestedField1: 2},
			},
		},
		{
			name: "recursive merge with overlapping keys",
			dst: map[string]interface{}{
				"key1": map[string]interface{}{
					"subkey1": "value1",
					"subkey2": complexStruct{NestedField1: 1},
				},
				"key2": "original value2",
			},
			src: map[string]interface{}{
				"key1": map[string]interface{}{
					"subkey1": "overwritten value1",
					"subkey2": complexStruct{NestedField2: "nested value2"},
					"subkey3": "new value3",
				},
				"key3": "new value3",
			},
			want: map[string]interface{}{
				"key1": map[string]interface{}{
					"subkey1": "overwritten value1",
					"subkey2": complexStruct{NestedField1: 1, NestedField2: "nested value2"},
					"subkey3": "new value3",
				},
				"key2": "original value2",
				"key3": "new value3",
			},
		},
		{
			name: "recursive merge with nil sources",
			dst: map[string]interface{}{
				"key": "val",
			},
			src: nil,
			want: map[string]interface{}{
				"key": "val",
			},
		},
		{
			name: "recursive merge with nil destination",
			dst:  nil,
			src: map[string]interface{}{
				"key": "val",
			},
			want: map[string]interface{}{
				"key": "val",
			},
		},
		{
			name: "recursive merge with non-overlapping keys",
			dst: map[string]interface{}{
				"key1": "value1",
			},
			src: map[string]interface{}{
				"key2": "value2",
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "merge maps containing slices",
			dst: map[string]interface{}{
				"arrayKey": []int{1, 2},
			},
			src: map[string]interface{}{
				"arrayKey": []int{3},
			},
			want: map[string]interface{}{
				"arrayKey": []int{3, 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := nativeMerger{}
			dst := reflect.ValueOf(tt.dst)
			src := reflect.ValueOf(tt.src)
			mergedMap := merger.mergeMaps(dst, src)
			assert.Equal(t, tt.want, mergedMap.Interface())
		})
	}
}

func Test_nativeMerger_mergeSlices(t *testing.T) {
	type complexStruct struct {
		NestedField1 int
		NestedField2 string
	}
	tests := []struct {
		name string
		dst  []interface{}
		src  []interface{}
		want []interface{}
	}{
		{
			name: "simple struct merge with single item",
			dst: []interface{}{
				complexStruct{NestedField1: 1},
			},
			src: []interface{}{
				complexStruct{NestedField1: 2},
			},
			want: []interface{}{
				complexStruct{NestedField1: 2},
			},
		},
		{
			name: "simple struct merge with single item",
			dst: []interface{}{
				&complexStruct{NestedField1: 1},
				&complexStruct{NestedField1: 3},
			},
			src: []interface{}{
				&complexStruct{NestedField1: 2},
			},
			want: []interface{}{
				&complexStruct{NestedField1: 2},
				&complexStruct{NestedField1: 3},
			},
		},
		{
			name: "merge maps",
			dst: []interface{}{
				map[string]interface{}{
					"subkey1": "value1",
					"subkey2": complexStruct{NestedField1: 1},
				},
				"original value2",
			},
			src: []interface{}{
				map[string]interface{}{
					"subkey1": "overwritten value1",
					"subkey2": complexStruct{NestedField2: "nested value2"},
					"subkey3": "new value3",
				},
				"new value3",
			},
			want: []interface{}{
				map[string]interface{}{
					"subkey1": "overwritten value1",
					"subkey2": complexStruct{NestedField1: 1, NestedField2: "nested value2"},
					"subkey3": "new value3",
				},
				"new value3",
			},
		},
		{
			name: "merge with more sources",
			dst: []interface{}{
				"val",
			},
			src: []interface{}{
				"val1",
				"val2",
				"val3",
			},
			want: []interface{}{
				"val1",
				"val2",
				"val3",
			},
		},
		{
			name: "merge with nil sources",
			dst: []interface{}{
				"val",
			},
			src: nil,
			want: []interface{}{
				"val",
			},
		},
		{
			name: "merge with nil destination",
			dst:  nil,
			src: []interface{}{
				"val",
			},
			want: []interface{}{
				"val",
			},
		},
		{
			name: "merge slices containing slices",
			dst: []interface{}{
				[]int{1, 2},
			},
			src: []interface{}{
				[]int{3},
			},
			want: []interface{}{
				[]int{3, 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := nativeMerger{}
			dst := reflect.ValueOf(tt.dst)
			src := reflect.ValueOf(tt.src)
			mergedSlice := merger.mergeSlices(dst, src)
			assert.Equal(t, tt.want, mergedSlice.Interface())
		})
	}
}
