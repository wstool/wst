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
					// Ensure all expectations on the mock are met
					parserMock.AssertExpectations(t)
					// Additionally, assert that ParseStruct was called the expected number of times
					parserMock.AssertNumberOfCalls(t, "ParseStruct", totalCalls)
				} else {
					t.Errorf("Actual value is not a slice; got type %T", actualValue)
				}
			}
		})
	}
}
