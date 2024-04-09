package merger

import (
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_nativeMerger_MergeConfigs(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}

	tests := []struct {
		name           string
		configs        []*types.Config
		overwrites     map[string]string
		expectedConfig *types.Config
		wantErr        bool
		errMsg         string
	}{
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
			config, err := merger.MergeConfigs(tt.configs, tt.overwrites)

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
		StrField    string
		IntField    int
		StructField nestedStruct
		MapField    map[string]int
		SliceField  []string
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
			want: &testStruct{
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
			want: &testStruct{
				StructField: nestedStruct{
					Field1: "nested",
					Field2: 100,
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
			want: &testStruct{
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
			want: &testStruct{
				SliceField: []string{"three", "two"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := &nativeMerger{
				fnd: fndMock,
			}
			merger.mergeStructs(reflect.ValueOf(tt.dst).Elem(), reflect.ValueOf(tt.src).Elem())
			assert.Equal(t, tt.want, tt.dst)
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
		dst  interface{}
		src  interface{}
		want interface{}
	}{
		{
			name: "simple struct merge",
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
				"arrayKey": []int{1, 2, 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := nativeMerger{}
			dst := reflect.ValueOf(tt.dst)
			src := reflect.ValueOf(tt.src)
			if dst.Kind() == reflect.Ptr {
				dst = dst.Elem()
			}
			if src.Kind() == reflect.Ptr {
				src = src.Elem()
			}
			merger.mergeMaps(dst, src)
			assert.Equal(t, tt.want, tt.dst)
		})
	}
}
