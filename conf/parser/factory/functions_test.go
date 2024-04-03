package factory

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	parserMocks "github.com/bukka/wst/mocks/conf/parser"
	"github.com/stretchr/testify/assert"
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
	parserMock := parserMocks.NewMockParser(t)
	f := &FuncProvider{
		fnd:          fndMock,
		structParser: parserMock.ParseStruct,
	}

	tests := []struct {
		name          string
		funcName      string
		data          interface{}
		expectedValue interface{} // Expected value to be set by the factory function.
		wantErr       bool
	}{
		{
			name:          "createActions valid data",
			funcName:      "createActions",
			data:          []interface{}{},  // Adjust based on what createActions expects.
			expectedValue: []types.Action{}, // Expected result after createActions.
			wantErr:       false,
		},
		// Add other cases for each function.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			factoryFunc := f.GetFactoryFunc(tt.funcName)
			if factoryFunc == nil {
				t.Fatalf("GetFactoryFunc(%s) returned nil", tt.funcName)
			}

			// Prepare a reflect.Value that the factory function will operate on.
			fieldValue := reflect.New(reflect.TypeOf(tt.expectedValue)).Elem()
			err := factoryFunc(tt.data, fieldValue, "")

			if tt.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				// Compare the fieldValue after invocation to the expectedValue.
				assert.Equal(tt.expectedValue, fieldValue.Interface())
			}
		})
	}
}
