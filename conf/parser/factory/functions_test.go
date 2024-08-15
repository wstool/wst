package factory

import (
	"errors"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/parser/location"
	"github.com/bukka/wst/conf/types"
	localMocks "github.com/bukka/wst/mocks/authored/local"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

func TestCreateFactories(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	parserMock := localMocks.NewMockMiniParser(t)
	testData := map[string]interface{}{"exampleKey": "exampleValue"}
	testStructure := make(map[string]interface{})
	testPath := "testPath"
	parserMock.On("ParseStruct", testData, &testStructure, testPath).
		Return(nil).Once()
	actionFactory := CreateActionsFactory(fndMock, parserMock.ParseStruct, location.CreateLocation())

	tests := []struct {
		name         string
		fnd          app.Foundation
		structParser StructParser
		want         Functions
	}{
		{
			name:         "Testing CreateLoader",
			fnd:          fndMock,
			structParser: parserMock.ParseStruct,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateFactories(tt.fnd, tt.structParser, "wst/path", location.CreateLocation())
			funcProvider, ok := got.(*FuncProvider)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, funcProvider.fnd)
			assert.IsType(t, actionFactory, funcProvider.actionsFactory)
			// assert struct parser call
			err := funcProvider.structParser(testData, &testStructure, testPath)
			assert.NoError(t, err)
			parserMock.AssertExpectations(t)
		})
	}
}

func TestFuncProvider_GetFactoryFunc(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	path := "/var/www/ws"

	tests := []struct {
		name           string
		funcName       string
		data           interface{}
		mockParseCalls []struct {
			data map[string]interface{}
			err  error
		}
		expectedValue      interface{} // Expected value to be set by the factory function.
		expectedNoFunction bool
		wantErr            bool
		errMsg             string
	}{
		// ACTION
		{
			name:     "createAction valid data",
			funcName: "createAction",
			data: map[string]interface{}{
				"start/serviceName": map[string]interface{}{"services": []map[string]interface{}{}},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"services": []map[string]interface{}{}},
					err:  nil,
				},
			},
			expectedValue: &types.StartAction{
				Service: "serviceName",
			},
			wantErr: false,
		},
		{
			name:     "createAction fails on invalid type",
			funcName: "createAction",
			data: map[string]interface{}{
				"wrong": map[string]interface{}{"services": []map[string]interface{}{}},
			},
			expectedValue: &types.StartAction{},
			wantErr:       true,
			errMsg:        "unknown action wrong at wrong",
		},
		{
			name:          "createActions valid data empty",
			funcName:      "createActions",
			data:          []interface{}{},
			expectedValue: []types.Action{},
			wantErr:       false,
		},
		{
			name:     "createActions valid data with service",
			funcName: "createActions",
			data: []interface{}{
				map[string]interface{}{
					"start/serviceName": map[string]interface{}{"services": []map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"services": []map[string]interface{}{}},
					err:  nil,
				},
			},
			expectedValue: []types.Action{
				&types.StartAction{
					Service: "serviceName",
				},
			},
			wantErr: false,
		},
		{
			name:     "createActions fails on creating struct",
			funcName: "createActions",
			data: []interface{}{
				map[string]interface{}{
					"start/serviceName": map[string]interface{}{"services": []map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"services": []map[string]interface{}{}},
					err:  errors.New("invalid data"),
				},
			},
			expectedValue: []types.Action{},
			wantErr:       true,
			errMsg:        "invalid data",
		},
		{
			name:          "createActions fails on invalid data",
			funcName:      "createActions",
			data:          1234,
			expectedValue: []types.Action{},
			wantErr:       true,
			errMsg:        "data must be an array, got int",
		},
		// Container image
		{
			name:     "createContainerImage with name and tag",
			funcName: "createContainerImage",
			data:     "imageName:1.0",
			expectedValue: types.ContainerImage{
				Name: "imageName",
				Tag:  "1.0",
			},
			wantErr: false,
		},
		{
			name:     "createContainerImage with name only",
			funcName: "createContainerImage",
			data:     "imageName",
			expectedValue: types.ContainerImage{
				Name: "imageName",
				Tag:  "latest", // Expecting the default tag to be 'latest'
			},
			wantErr: false,
		},
		{
			name:     "createContainerImage with map input",
			funcName: "createContainerImage",
			data: map[string]interface{}{
				"name": "imageName",
				"tag":  "1.0",
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"name": "imageName", "tag": "1.0"},
					err:  nil,
				},
			},
			expectedValue: types.ContainerImage{},
			wantErr:       false,
		},
		{
			name:     "createContainerImage with failed parsing",
			funcName: "createContainerImage",
			data: map[string]interface{}{
				"name": "imageName",
				"tag":  "1.0",
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"name": "imageName", "tag": "1.0"},
					err:  errors.New("container parsing failed"),
				},
			},
			expectedValue: types.ContainerImage{},
			wantErr:       true,
			errMsg:        "container parsing failed",
		},
		{
			name:          "createContainerImage fails on unsupported type",
			funcName:      "createContainerImage",
			data:          1234,                   // Invalid type, expecting string or map
			expectedValue: types.ContainerImage{}, // Default empty value since it should fail
			wantErr:       true,
			errMsg:        "unsupported type for image data",
		},
		// Environment
		{
			name:     "createEnvironments with valid data for single type",
			funcName: "createEnvironments",
			data: map[string]interface{}{
				"common": map[string]interface{}{"config": "commonConfig"},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"config": "commonConfig"},
					err:  nil,
				},
			},
			expectedValue: map[string]types.Environment{
				"common": &types.CommonEnvironment{},
			},
			wantErr: false,
		},
		{
			name:     "createEnvironments with valid data for multiple types",
			funcName: "createEnvironments",
			data: map[string]interface{}{
				"common":     map[string]interface{}{"config": "commonConfig"},
				"local":      map[string]interface{}{"path": "/local/path"},
				"container":  map[string]interface{}{"image": "test:1.0"},
				"docker":     map[string]interface{}{"image": "test:1.1"},
				"kubernetes": map[string]interface{}{"image": "test:1.2"},
				"wst/path":   "/var/wst",
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"config": "commonConfig"},
					err:  nil,
				},
				{
					data: map[string]interface{}{"path": "/local/path"},
					err:  nil,
				},
				{
					data: map[string]interface{}{"image": "test:1.0"},
					err:  nil,
				},
				{
					data: map[string]interface{}{"image": "test:1.1"},
					err:  nil,
				},
				{
					data: map[string]interface{}{"image": "test:1.2"},
					err:  nil,
				},
			},
			expectedValue: map[string]types.Environment{
				"common":     &types.CommonEnvironment{},
				"local":      &types.LocalEnvironment{},
				"container":  &types.ContainerEnvironment{},
				"docker":     &types.DockerEnvironment{},
				"kubernetes": &types.KubernetesEnvironment{},
			},
			wantErr: false,
		},
		{
			name:     "createEnvironments with failed struct parsing",
			funcName: "createEnvironments",
			data: map[string]interface{}{
				"common": map[string]interface{}{"config": "commonConfig"},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"config": "commonConfig"},
					err:  errors.New("invalid env"),
				},
			},
			expectedValue: map[string]interface{}{},
			wantErr:       true,
			errMsg:        "invalid env",
		},
		{
			name:     "createEnvironments with invalid env data",
			funcName: "createEnvironments",
			data: map[string]interface{}{
				"common": "test",
			},
			expectedValue: map[string]interface{}{},
			wantErr:       true,
			errMsg:        "data for value in environments must be a map, got string",
		},
		{
			name:     "createEnvironments with unsupported environment type",
			funcName: "createEnvironments",
			data: map[string]interface{}{
				"unsupported": map[string]interface{}{"config": "someConfig"},
			},
			expectedValue: map[string]interface{}{}, // Expected no environments to be created
			wantErr:       true,
			errMsg:        "unknown environments type: unsupported",
		},
		{
			name:          "createEnvironments with invalid data structure",
			funcName:      "createEnvironments",
			data:          "invalidDataStructure",   // Not a map
			expectedValue: map[string]interface{}{}, // Expected no environments to be created
			wantErr:       true,
			errMsg:        "data for environments must be a map, got string",
		},
		// Hooks
		{
			name:     "createHooks with valid native hook",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"native": map[string]interface{}{
						"enabled": true,
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{
						"enabled": true,
					},
					err: nil,
				},
			},
			expectedValue: map[string]types.SandboxHook{
				"start": &types.SandboxHookNative{},
			},
			wantErr: false,
		},
		{
			name:     "createHooks with valid command hook (shell command)",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"command": map[string]interface{}{
						"command": "echo 'Hello World'",
						"shell":   "/bin/bash",
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{
						"command": "echo 'Hello World'",
						"shell":   "/bin/bash",
					},
					err: nil,
				},
			},
			expectedValue: map[string]types.SandboxHook{
				"start": &types.SandboxHookShellCommand{},
			},
			wantErr: false,
		},
		{
			name:     "createHooks with valid command hook (executable)",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"command": map[string]interface{}{
						"executable": "myApp",
						"args":       []interface{}{"arg1", "arg2"},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{
						"executable": "myApp",
						"args":       []interface{}{"arg1", "arg2"},
					},
					err: nil,
				},
			},
			expectedValue: map[string]types.SandboxHook{
				"start": &types.SandboxHookArgsCommand{},
			},
			wantErr: false,
		},
		{
			name:     "createHooks with valid signal hook (string)",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"signal": "SIGHUP",
				},
			},
			expectedValue: map[string]types.SandboxHook{
				"start": &types.SandboxHookSignal{
					IsString:    true,
					StringValue: "SIGHUP",
				},
			},
			wantErr: false,
		},
		{
			name:     "createHooks with valid signal hook (int)",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"signal": 9, // Typically SIGKILL
				},
			},
			expectedValue: map[string]types.SandboxHook{
				"start": &types.SandboxHookSignal{
					IsString: false,
					IntValue: 9,
				},
			},
			wantErr: false,
		},
		{
			name:     "createHooks with signal hook invalid type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"signal": map[string]interface{}{}, // Invalid type, expecting a string or int
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "invalid signal hook type map[], only string and int is allowed",
		},
		{
			name:     "createHooks with unsupported hook type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"command": map[string]interface{}{
						"command": "invalid",
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{
						"command": "invalid",
					},
					err: errors.New("invalid command"),
				},
			},
			expectedValue: map[string]types.SandboxHook{}, // No hooks should be created
			wantErr:       true,
			errMsg:        "invalid command",
		},
		{
			name:          "createHooks fails on invalid hook data structure",
			funcName:      "createHooks",
			data:          "invalidHookData",        // Not a map
			expectedValue: map[string]interface{}{}, // No hooks should be created
			wantErr:       true,
			errMsg:        "data for hooks must be a map, got string",
		},
		{
			name:     "createHooks with command hook invalid command hooks type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"command": 123,
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "command hooks must be a map, got int",
		},
		{
			name:     "createHooks with command hook invalid command items",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"command": map[string]interface{}{
						"field": map[string]interface{}{},
					},
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "command hooks data is invalid",
		},
		{
			name:     "createHooks with command hook invalid native hooks type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"native": 123,
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "native hook must be a map, got int",
		},
		{
			name:     "createHooks with unknown type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"field": map[string]interface{}{},
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "unknown hook type: field",
		},
		{
			name:     "createHooks with more than one types",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{
					"command": map[string]interface{}{},
					"signal":  map[string]interface{}{},
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "hook data must have only one element",
		},
		{
			name:     "createHooks with no type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": map[string]interface{}{},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "hook data cannot be an empty map",
		},
		{
			name:     "createHooks with invalid type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"start": 1,
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "hook data must be a map, got int",
		},
		// Parameters
		{
			name:     "createParameters with valid map data",
			funcName: "createParameters",
			data: map[string]interface{}{
				"param1": "value1",
				"param2": 123,
				"param3": true,
				"param4": map[string]interface{}{
					"param41": []interface{}{
						map[string]interface{}{
							"param5": 2,
						},
						5,
					},
					"param42": 3,
				},
			},
			expectedValue: types.Parameters{
				"param1": "value1",
				"param2": 123,
				"param3": true,
				"param4": types.Parameters{
					"param41": []interface{}{
						types.Parameters{
							"param5": 2,
						},
						5,
					},
					"param42": 3,
				},
			},
			wantErr: false,
		},
		{
			name:          "createParameters fails on invalid data type (int)",
			funcName:      "createParameters",
			data:          123,                // Invalid data type, expecting a map.
			expectedValue: types.Parameters{}, // No parameters should be set due to error.
			wantErr:       true,
			errMsg:        "data for parameters must be a map, got int",
		},
		// Sandboxes
		{
			name:     "createSandboxes with multiple valid sandbox types",
			funcName: "createSandboxes",
			data: map[string]interface{}{
				"common":     map[string]interface{}{"config": "commonConfig"},
				"local":      map[string]interface{}{"path": "/local/path"},
				"container":  map[string]interface{}{"image": "test:1.0"},
				"docker":     map[string]interface{}{"image": "test:1.1"},
				"kubernetes": map[string]interface{}{"image": "test:1.2"},
				"wst/path":   "/var/wst",
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"config": "commonConfig"},
					err:  nil,
				},
				{
					data: map[string]interface{}{"path": "/local/path"},
					err:  nil,
				},
				{
					data: map[string]interface{}{"image": "test:1.0"},
					err:  nil,
				},
				{
					data: map[string]interface{}{"image": "test:1.1"},
					err:  nil,
				},
				{
					data: map[string]interface{}{"image": "test:1.2"},
					err:  nil,
				},
			},
			expectedValue: map[string]types.Sandbox{
				"common":     &types.CommonSandbox{},
				"local":      &types.LocalSandbox{},
				"container":  &types.ContainerSandbox{},
				"docker":     &types.DockerSandbox{},
				"kubernetes": &types.KubernetesSandbox{},
			},
			wantErr: false,
		},
		{
			name:     "createSandboxes with unsupported sandbox type",
			funcName: "createSandboxes",
			data: map[string]interface{}{
				"unsupported": map[string]interface{}{}, // An unsupported sandbox type
			},
			expectedValue: map[string]interface{}{}, // Expecting no sandboxes to be created due to error
			wantErr:       true,
			errMsg:        "unknown sandboxes type: unsupported",
		},
		// Server expectations
		{
			name:     "createServerExpectations with valid data for multiple types",
			funcName: "createServerExpectations",
			data: map[string]interface{}{
				"expectation1": map[string]interface{}{"metrics": map[string]interface{}{}},
				"expectation2": map[string]interface{}{
					"output":     map[string]interface{}{},
					"parameters": map[string]interface{}{},
				},
				"expectation3": map[string]interface{}{"response": map[string]interface{}{}},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{ // Corresponds to "metrics"
					data: map[string]interface{}{"metrics": map[string]interface{}{}},
					err:  nil,
				},
				{ // Corresponds to "output"
					data: map[string]interface{}{
						"output":     map[string]interface{}{},
						"parameters": map[string]interface{}{},
					},
					err: nil,
				},
				{ // Corresponds to "response"
					data: map[string]interface{}{"response": map[string]interface{}{}},
					err:  nil,
				},
			},
			expectedValue: map[string]types.ServerExpectationAction{
				"expectation1": &types.ServerMetricsExpectation{},
				"expectation2": &types.ServerOutputExpectation{},
				"expectation3": &types.ServerResponseExpectation{},
			},
			wantErr: false,
		},
		{
			name:     "createServerExpectations with multiple parsed data",
			funcName: "createServerExpectations",
			data: map[string]interface{}{
				"expectation2": map[string]interface{}{
					"metrics":    map[string]interface{}{},
					"output":     map[string]interface{}{},
					"parameters": map[string]interface{}{},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{
						"metrics":    map[string]interface{}{},
						"output":     map[string]interface{}{},
						"parameters": map[string]interface{}{},
					},
					err: nil,
				},
			},
			expectedValue: map[string]interface{}{},
			wantErr:       true,
			errMsg:        "expectation cannot have multiple types - additional key ",
		},
		{
			name:     "createServerExpectations with failed parsing",
			funcName: "createServerExpectations",
			data: map[string]interface{}{
				"expectation1": map[string]interface{}{"metrics": map[string]interface{}{}},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				err  error
			}{
				{
					data: map[string]interface{}{"metrics": map[string]interface{}{}},
					err:  errors.New("parsing failed"),
				},
			},
			expectedValue: map[string]interface{}{},
			wantErr:       true,
			errMsg:        "parsing failed",
		},
		{
			name:          "createServerExpectations with invalid data type",
			funcName:      "createServerExpectations",
			data:          1,
			expectedValue: map[string]interface{}{}, // Expecting no expectations to be created due to error
			wantErr:       true,
			errMsg:        "data for server action expectations must be a map, got int",
		},
		{
			name:     "createServerExpectations with invalid data value type",
			funcName: "createServerExpectations",
			data: map[string]interface{}{
				"test": "ss",
			},
			expectedValue: map[string]interface{}{}, // Expecting no expectations to be created due to error
			wantErr:       true,
			errMsg:        "data for value in server action expectations must be a map, got string",
		},
		{
			name:     "createServerExpectations with unsupported expectation key",
			funcName: "createServerExpectations",
			data: map[string]interface{}{
				"invalidExpectation": map[string]interface{}{"unsupportedKey": map[string]interface{}{}},
			},
			expectedValue: map[string]interface{}{}, // Expecting no expectations to be created due to error
			wantErr:       true,
			errMsg:        "invalid server expectation key unsupportedKey",
		},
		// Service scripts
		{
			name:     "createServiceScripts with all scripts included (bool)",
			funcName: "createServiceScripts",
			data:     true,
			expectedValue: types.ServiceScripts{
				IncludeAll: true,
			},
			wantErr: false,
		},
		{
			name:     "createServiceScripts with selected scripts included (string array)",
			funcName: "createServiceScripts",
			data:     []interface{}{"script1.sh", "script2.sh"},
			expectedValue: types.ServiceScripts{
				IncludeList: []string{"script1.sh", "script2.sh"},
			},
			wantErr: false,
		},
		{
			name:          "createServiceScripts with non string item type",
			funcName:      "createServiceScripts",
			data:          []interface{}{"script1.sh", 1},
			expectedValue: types.ServiceScripts{},
			wantErr:       true,
			errMsg:        "invalid services scripts item type at index 1, expected string but got int",
		},
		{
			name:          "createServiceScripts with invalid data type (int)",
			funcName:      "createServiceScripts",
			data:          123,                    // Invalid type, expecting bool or []string
			expectedValue: types.ServiceScripts{}, // No scripts should be included due to error
			wantErr:       true,
			errMsg:        "invalid services scripts type, expected bool or string array but got int",
		},
		{
			name:          "createServiceScripts with invalid data type (map)",
			funcName:      "createServiceScripts",
			data:          map[string]interface{}{"script1": "script1.sh"}, // Invalid type, expecting bool or []string
			expectedValue: types.ServiceScripts{},
			wantErr:       true,
			errMsg:        "invalid services scripts type, expected bool or string array but got map[string]interface {}",
		},
		// Unknown
		{
			name:               "create unknown function",
			funcName:           "createUnknown",
			data:               map[string]interface{}{"script1": "script1.sh"},
			expectedNoFunction: true,
			wantErr:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			parserMock := localMocks.NewMockMiniParser(t)
			f := CreateFactories(fndMock, parserMock.ParseStruct, "wst/path", location.CreateLocation())

			// Setup mock expectations
			totalCalls := 0
			for _, call := range tt.mockParseCalls {
				parserMock.On("ParseStruct", call.data, mock.Anything, path).Return(call.err).Once()
				totalCalls++
			}

			factoryFunc, err := f.GetFactoryFunc(tt.funcName)
			if err != nil {
				if tt.expectedNoFunction {
					return
				}
				t.Fatalf("GetFactoryFunc(%s) returned nil", tt.funcName)
			}

			// Prepare a reflect.Value that the factory function will operate on.
			fieldValue := reflect.New(reflect.TypeOf(tt.expectedValue)).Elem()
			err = factoryFunc(tt.data, fieldValue, path)

			if tt.wantErr {
				assert.Error(err)
				if tt.errMsg != "" {
					assert.ErrorContains(err, tt.errMsg)
				}
			} else {
				assert.NoError(err)
				// Compare the fieldValue after invocation to the expectedValue.
				actualValue := fieldValue.Interface()
				expectedValue := reflect.ValueOf(tt.expectedValue)

				if actual := reflect.ValueOf(actualValue); actual.Kind() == reflect.Slice {
					if actual.IsNil() {
						assert.Equal(0, expectedValue.Len(), "Expected slice length to be 0, but got nil slice")
					} else {
						assert.Equal(expectedValue.Len(), actual.Len(), "Slice lengths differ")
						if expectedValue.Len() > 0 {
							assert.Equal(tt.expectedValue, actualValue, "Expected and actual slices differ")
						}
					}
				} else {
					assert.Equal(tt.expectedValue, actualValue, "Expected and actual values are not equal")
				}
				// Ensure all expectations on the mock are met
				parserMock.AssertExpectations(t)
				// Additionally, assert that ParseStruct was called the expected number of times
				parserMock.AssertNumberOfCalls(t, "ParseStruct", totalCalls)
			}
		})
	}
}
