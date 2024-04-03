package factory

import (
	"errors"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	parserMocks "github.com/bukka/wst/mocks/conf/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

func TestCreateFactories(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	parserMock := parserMocks.NewMockParser(t)
	testData := map[string]interface{}{"exampleKey": "exampleValue"}
	testStructure := make(map[string]interface{})
	testPath := "testPath"
	parserMock.On("ParseStruct", testData, &testStructure, testPath).
		Return(nil).Once()
	actionFactory := CreateActionsFactory(fndMock, parserMock.ParseStruct)

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
			got := CreateFactories(tt.fnd, tt.structParser)
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
		expectedValue interface{} // Expected value to be set by the factory function.
		wantErr       bool
		errMsg        string
	}{
		// ACTION
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
			expectedValue: map[string]interface{}{
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
			expectedValue: map[string]interface{}{
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
			errMsg:        "unknown environment type: unsupported",
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
			name:     "createHooks with valid command hook (shell command)",
			funcName: "createHooks",
			data: map[string]interface{}{
				"command": map[string]interface{}{
					"command": "echo 'Hello World'",
					"shell":   "/bin/bash",
				},
			},
			expectedValue: map[string]interface{}{
				"command": &types.SandboxHookShellCommand{
					Command: "echo 'Hello World'",
					Shell:   "/bin/bash",
				},
			},
			wantErr: false,
		},

		{
			name:     "createHooks with valid command hook (default shell command)",
			funcName: "createHooks",
			data: map[string]interface{}{
				"command": map[string]interface{}{
					"command": "echo 'Hello World'",
				},
			},
			expectedValue: map[string]interface{}{
				"command": &types.SandboxHookShellCommand{
					Command: "echo 'Hello World'",
					Shell:   "/bin/sh",
				},
			},
			wantErr: false,
		},
		{
			name:     "createHooks with valid command hook (executable)",
			funcName: "createHooks",
			data: map[string]interface{}{
				"command": map[string]interface{}{
					"executable": "myApp",
					"args":       []interface{}{"arg1", "arg2"},
				},
			},
			expectedValue: map[string]interface{}{
				"command": &types.SandboxHookArgsCommand{
					Executable: "myApp",
					Args:       []string{"arg1", "arg2"},
				},
			},
			wantErr: false,
		},
		{
			name:     "createHooks with valid signal hook (string)",
			funcName: "createHooks",
			data: map[string]interface{}{
				"signal": "SIGHUP",
			},
			expectedValue: map[string]interface{}{
				"signal": &types.SandboxHookSignal{
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
				"signal": 9, // Typically SIGKILL
			},
			expectedValue: map[string]interface{}{
				"signal": &types.SandboxHookSignal{
					IsString: false,
					IntValue: 9,
				},
			},
			wantErr: false,
		},
		{
			name:     "createHooks with unsupported hook type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"unsupported": map[string]interface{}{},
			},
			expectedValue: map[string]interface{}{}, // No hooks should be created
			wantErr:       true,
			errMsg:        "unknown environment type: unsupported",
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
			name:     "createHooks with command hook missing command",
			funcName: "createHooks",
			data: map[string]interface{}{
				"command": map[string]interface{}{ // Missing "command" or "executable" key
					"shell": "/bin/bash",
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "command hooks data is invalid",
		},
		{
			name:     "createHooks with command hook invalid command hooks type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"command": 123,
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "command hooks must be a map, got int",
		},
		{
			name:     "createHooks with command hook invalid inner command type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"command": map[string]interface{}{
					"command": 123, // Invalid type, expecting a string
					"shell":   "/bin/bash",
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "command must be a string",
		},

		{
			name:     "createHooks with command hook invalid args item type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"command": map[string]interface{}{
					"executable": 1,
					"args":       []interface{}{"a"},
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "executable must be a string",
		},
		{
			name:     "createHooks with command hook invalid args item type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"command": map[string]interface{}{
					"executable": "myApp",
					"args":       []interface{}{1}, // Invalid type, expecting an array of strings
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "args must be an array of strings but its item is of type int",
		},
		{
			name:     "createHooks with command hook invalid args type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"command": map[string]interface{}{
					"executable": "myApp",
					"args":       "notAnArray", // Invalid type, expecting an array of strings
				},
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "args must be an array of strings but it is not an array",
		},
		{
			name:     "createHooks with signal hook invalid type",
			funcName: "createHooks",
			data: map[string]interface{}{
				"signal": map[string]interface{}{}, // Invalid type, expecting a string or int
			},
			expectedValue: map[string]interface{}{}, // Expecting no valid hooks due to error
			wantErr:       true,
			errMsg:        "invalid signal hook type map[], only string and int is allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			parserMock := parserMocks.NewMockParser(t)
			f := CreateFactories(fndMock, parserMock.ParseStruct)

			// Setup mock expectations
			totalCalls := 0
			for _, call := range tt.mockParseCalls {
				parserMock.On("ParseStruct", call.data, mock.Anything, path).Return(call.err).Once()
				totalCalls++
			}

			factoryFunc := f.GetFactoryFunc(tt.funcName)
			if factoryFunc == nil {
				t.Fatalf("GetFactoryFunc(%s) returned nil", tt.funcName)
			}

			// Prepare a reflect.Value that the factory function will operate on.
			fieldValue := reflect.New(reflect.TypeOf(tt.expectedValue)).Elem()
			err := factoryFunc(tt.data, fieldValue, path)

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
