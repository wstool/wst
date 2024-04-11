package overwrites

import (
	"errors"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/parser"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	loaderMocks "github.com/bukka/wst/mocks/conf/loader"
	parserMocks "github.com/bukka/wst/mocks/conf/parser"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestCreateOverwriter(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	parserMock := parserMocks.NewMockParser(t)
	tests := []struct {
		name   string
		fnd    app.Foundation
		parser parser.Parser
	}{
		{
			name:   "create merger",
			fnd:    fndMock,
			parser: parserMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := CreateOverwriter(tt.fnd, tt.parser)
			actualOverwriter, ok := merger.(*nativeOverwriter)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, actualOverwriter.fnd)
			assert.Equal(t, tt.parser, actualOverwriter.parser)
		})
	}
}

func Test_nativeOverwriter_Overwrite(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	loaderMock := loaderMocks.NewMockLoader(t)
	parser := parser.CreateParser(fndMock, loaderMock)
	tests := []struct {
		name       string
		overwrites map[string]string
		config     *types.Config
		want       *types.Config
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "overwrite config name only",
			overwrites: map[string]string{"name": "test"},
			config:     &types.Config{},
			want:       &types.Config{Name: "test"},
		},
		{
			name:       "overwrite config with spec",
			overwrites: map[string]string{"spec.workspace": "/var/www"},
			config:     &types.Config{},
			want:       &types.Config{Spec: types.Spec{Workspace: "/var/www"}},
		},
		{
			name:       "overwrite config with spec",
			overwrites: map[string]string{"spec.environments.docker.name_prefix": "prefix"},
			config: &types.Config{
				Spec: types.Spec{
					Environments: map[string]types.Environment{
						"docker": &types.DockerEnvironment{
							NamePrefix: "test",
						},
					},
				},
			},
			want: &types.Config{
				Spec: types.Spec{
					Environments: map[string]types.Environment{
						"docker": &types.DockerEnvironment{
							NamePrefix: "prefix",
						},
					},
				},
			},
		},
		{
			name: "overwrite config action",
			overwrites: map[string]string{
				"spec.instances[0].name":               "cool stuff",
				"spec.instances[0].actions[0].timeout": "20000",
			},
			config: &types.Config{
				Spec: types.Spec{
					Instances: []types.Instance{
						{
							Name: "test",
							Actions: []types.Action{
								&types.StartAction{
									Timeout: 50000,
								},
							},
						},
					},
				},
			},
			want: &types.Config{
				Spec: types.Spec{
					Instances: []types.Instance{
						{
							Name: "cool stuff",
							Actions: []types.Action{
								&types.StartAction{
									Timeout: 20000,
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overwriter := CreateOverwriter(fndMock, parser)
			err := overwriter.Overwrite(tt.config, tt.overwrites)

			if tt.errMsg != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, tt.want, tt.config)
				}
			}
		})
	}
}

func Test_nativeOverwriter_parseFieldNameAndIndex(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	parserMock := parserMocks.NewMockParser(t)
	tests := []struct {
		name          string
		ptr           string
		wantFieldName string
		wantIndex     int
		wantIsSlice   bool
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "normal field",
			ptr:           "instances",
			wantFieldName: "instances",
			wantIndex:     0,
			wantIsSlice:   false,
			wantErr:       false,
		},
		{
			name:          "field with index",
			ptr:           "instances[0]",
			wantFieldName: "instances",
			wantIndex:     0,
			wantIsSlice:   true,
			wantErr:       false,
		},
		{
			name:          "field with negative index - invalid",
			ptr:           "instances[-1]",
			wantFieldName: "instances",
			wantIndex:     -1,
			wantIsSlice:   true,
		},
		{
			name:          "field with non-integer index - invalid",
			ptr:           "instances[abc]",
			wantFieldName: "instances[abc]",
			wantIndex:     0,
			wantIsSlice:   false,
		},
		{
			name:          "nested field with index",
			ptr:           "config[3].setting",
			wantFieldName: "config[3].setting",
			wantIndex:     0,
			wantIsSlice:   false,
			wantErr:       false,
		},
		{
			name:          "nested field with index",
			ptr:           "config[-]",
			wantFieldName: "",
			wantIndex:     0,
			wantIsSlice:   false,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &nativeOverwriter{
				fnd:    fndMock,
				parser: parserMock,
			}
			gotFieldName, gotIndex, gotIsSlice, err := o.parseFieldNameAndIndex(tt.ptr)
			if tt.wantErr {
				if assert.Error(t, err) && tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, tt.wantFieldName, gotFieldName)
					assert.Equal(t, tt.wantIndex, gotIndex)
					assert.Equal(t, tt.wantIsSlice, gotIsSlice)
				}
			}
		})
	}
}

type TestStruct struct {
	Name        string                  `wst:"name"`
	Age         int                     `wst:"age"`
	Nested      NestedStruct            `wst:"nested"`
	NestedSlice []NestedStruct          `wst:"nestedSlice"`
	NestedMap   map[string]NestedStruct `wst:"nestedMap"`
}

type NestedStruct struct {
	Field1 string `wst:"field1"`
	Field2 int    `wst:"field2"`
}

func Test_nativeOverwriter_overwriteStruct(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}

	tests := []struct {
		name       string
		ptrs       []string
		tags       map[parser.ConfigParam]string
		tagsErr    map[parser.ConfigParam]string
		val        string
		dst        interface{}
		wantErr    bool
		errMsg     string
		verifyFunc func(*testing.T, interface{})
	}{
		{
			name: "overwrite scalar field",
			ptrs: []string{"name"},
			val:  "John Doe",
			tags: map[parser.ConfigParam]string{
				"name": "name",
			},
			dst: &TestStruct{},
			verifyFunc: func(t *testing.T, dst interface{}) {
				assert.Equal(t, "John Doe", dst.(*TestStruct).Name)
			},
		},
		{
			name: "overwrite nested struct field",
			ptrs: []string{"nested", "field1"},
			val:  "New Value",
			tags: map[parser.ConfigParam]string{
				"name":   "name",
				"age":    "age",
				"nested": "nested",
				"field1": "field1",
			},
			dst: &TestStruct{Nested: NestedStruct{Field1: "Old Value"}},
			verifyFunc: func(t *testing.T, dst interface{}) {
				assert.Equal(t, "New Value", dst.(*TestStruct).Nested.Field1)
			},
		},
		{
			name: "overwrite nested slice element",
			ptrs: []string{"nestedSlice[0]", "field2"},
			tags: map[parser.ConfigParam]string{
				"name":        "name",
				"age":         "age",
				"nested":      "nested",
				"nestedSlice": "nestedSlice",
				"field1":      "field1",
				"field2":      "field2",
			},
			val: "42",
			dst: &TestStruct{NestedSlice: []NestedStruct{{Field2: 0}}},
			verifyFunc: func(t *testing.T, dst interface{}) {
				assert.Equal(t, 42, dst.(*TestStruct).NestedSlice[0].Field2)
			},
		},
		{
			name: "overwrite nested map element",
			ptrs: []string{"nestedMap", "key1", "field1"},
			tags: map[parser.ConfigParam]string{
				"name":        "name",
				"age":         "age",
				"nested":      "nested",
				"nestedSlice": "nestedSlice",
				"nestedMap":   "nestedMap",
				"field1":      "field1",
			},
			val:     "Updated",
			dst:     &TestStruct{NestedMap: map[string]NestedStruct{"key1": {Field1: "Original"}}},
			wantErr: true,
			errMsg:  "The value for field field1 cannot be set",
		},
		{
			name: "overwrite nonexistent field",
			ptrs: []string{"nonexistent"},
			tags: map[parser.ConfigParam]string{
				"name":        "name",
				"age":         "age",
				"nested":      "nested",
				"nestedSlice": "nestedSlice",
				"nestedMap":   "nestedMap",
			},
			val:     "Value",
			dst:     &TestStruct{},
			wantErr: true,
			errMsg:  "overwrite for field nonexistent not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parserMock := parserMocks.NewMockParser(t)
			for name, value := range tt.tags {
				parserMock.On("ParseTag", string(name)).Return(map[parser.ConfigParam]string{"name": value}, nil)
			}
			for name, msg := range tt.tagsErr {
				parserMock.On("ParseTag", name).Return(nil, errors.New(msg))
			}

			o := &nativeOverwriter{parser: parserMock, fnd: fndMock}
			err := o.overwriteStruct(reflect.ValueOf(tt.dst), tt.ptrs, tt.val)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.verifyFunc != nil {
					tt.verifyFunc(t, tt.dst)
				}
			}
		})
	}
}

func Test_nativeOverwriter_overwriteMap(t *testing.T) {
	tests := []struct {
		name        string
		dst         interface{}
		ptrs        []string
		val         string
		wantErr     bool
		errMsg      string
		expectedMap interface{}
	}{
		{
			name: "overwrite direct map value",
			dst: map[string]string{
				"key1": "oldValue1",
				"key2": "oldValue2",
			},
			ptrs:        []string{"key1"},
			val:         "newValue1",
			expectedMap: map[string]string{"key1": "newValue1", "key2": "oldValue2"},
		},
		{
			name: "overwrite nested map value",
			dst: map[string]interface{}{
				"nestedMap": map[string]string{"nestedKey": "oldNestedValue"},
			},
			ptrs:        []string{"nestedMap", "nestedKey"},
			val:         "newNestedValue",
			expectedMap: map[string]interface{}{"nestedMap": map[string]string{"nestedKey": "newNestedValue"}},
		},
		{
			name: "overwrite slice within map",
			dst: map[string]interface{}{
				"sliceKey": []string{"sliceOldValue1", "sliceOldValue2"},
			},
			ptrs:        []string{"sliceKey[1]"},
			val:         "sliceNewValue2",
			expectedMap: map[string]interface{}{"sliceKey": []string{"sliceOldValue1", "sliceNewValue2"}},
		},
		{
			name: "overwrite struct field within map",
			dst: map[string]interface{}{
				"structKey": TestStruct{Name: "oldName"},
			},
			ptrs:        []string{"structKey", "Name"},
			val:         "newName",
			expectedMap: map[string]interface{}{"structKey": TestStruct{Name: "newName"}},
		},
		{
			name:    "key not found in map",
			dst:     map[string]string{"key": "value"},
			ptrs:    []string{"nonexistentKey"},
			val:     "newValue",
			wantErr: true,
			errMsg:  "key nonexistentKey not found in map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dstValue := reflect.ValueOf(tt.dst)
			o := &nativeOverwriter{}
			err := o.overwriteMap(dstValue, tt.ptrs, tt.val)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMap, tt.dst)
			}
		})
	}
}

func Test_nativeOverwriter_overwriteSlice(t *testing.T) {
	tests := []struct {
		name          string
		dst           interface{}
		ptrs          []string
		index         int
		val           string
		wantErr       bool
		errMsg        string
		expectedSlice interface{}
	}{
		{
			name:          "overwrite scalar value in slice",
			dst:           []string{"oldValue1", "oldValue2"},
			ptrs:          []string{},
			index:         1,
			val:           "newValue2",
			expectedSlice: []string{"oldValue1", "newValue2"},
		},
		{
			name: "overwrite struct field in slice",
			dst: []TestStruct{
				{Name: "oldName1"},
				{Name: "oldName2"},
			},
			ptrs:  []string{"Name"},
			index: 1,
			val:   "newName2",
			expectedSlice: []TestStruct{
				{Name: "oldName1"},
				{Name: "newName2"},
			},
		},
		{
			name:    "index out of range",
			dst:     []string{"value1", "value2"},
			ptrs:    []string{},
			index:   3,
			val:     "newValue",
			wantErr: true,
			errMsg:  "index out of range: 3",
		},
		{
			name:    "attempt to overwrite slice in slice (unsupported)",
			dst:     [][]string{{"oldValue1"}, {"oldValue2"}},
			ptrs:    []string{"0"},
			index:   1,
			val:     "newValue",
			wantErr: true,
			errMsg:  "array of arrays are not supported for overwrites",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dstValue := reflect.ValueOf(tt.dst)
			o := &nativeOverwriter{}
			err := o.overwriteSlice(dstValue, tt.ptrs, tt.index, tt.val)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSlice, tt.dst)
			}
		})
	}
}

func Test_nativeOverwriter_overwriteScalar(t *testing.T) {
	tests := []struct {
		name    string
		dst     interface{}
		ptrs    []string
		val     string
		want    interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "overwrite integer",
			dst:  new(int),
			val:  "42",
			want: 42,
		},
		{
			name:    "integer overflow error",
			dst:     new(int8),
			val:     "128",
			wantErr: true,
			errMsg:  "value 128 overflows",
		},
		{
			name: "overwrite float",
			dst:  new(float64),
			val:  "3.14",
			want: 3.14,
		},
		{
			name:    "float overflow error",
			dst:     new(float32),
			val:     "3.4028235e+38", // slightly above max float32
			wantErr: true,
			errMsg:  "value 3.4028235e+38 overflows",
		},
		{
			name: "overwrite string",
			dst:  new(string),
			val:  "hello",
			want: "hello",
		},
		{
			name:    "unsupported value kind",
			dst:     new(bool), // Assuming bool is not supported by overwriteScalar
			val:     "true",
			wantErr: true,
			errMsg:  "unsupported value kind bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dstValue := reflect.ValueOf(tt.dst).Elem()
			o := &nativeOverwriter{}
			err := o.overwriteScalar(dstValue, tt.ptrs, tt.val)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Use reflect.DeepEqual for comparison to handle different types properly
				if !reflect.DeepEqual(tt.dst, tt.want) {
					t.Errorf("overwriteScalar() got = %v, want %v", tt.dst, tt.want)
				}
			}
		})
	}
}
