package factory

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/parser/location"
	"github.com/wstool/wst/conf/types"
	localMocks "github.com/wstool/wst/mocks/authored/local"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	"testing"
)

func TestCreateActionsFactory(t *testing.T) {
	fndMock := &appMocks.MockFoundation{}
	parserMock := localMocks.NewMockMiniParser(t)
	testData := map[string]interface{}{"exampleKey": "exampleValue"}
	testStructure := make(map[string]interface{})
	testPath := "testPath"
	parserMock.On("ParseStruct", testData, &testStructure, testPath).
		Return(nil).Once()

	tests := []struct {
		name         string
		fnd          app.Foundation
		structParser StructParser
		want         ActionsFactory
	}{
		{
			name:         "Testing CreateLoader",
			fnd:          fndMock,
			structParser: parserMock.ParseStruct,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateActionsFactory(tt.fnd, tt.structParser, location.CreateLocation())
			factory, ok := got.(*NativeActionsFactory)
			assert.True(t, ok)
			assert.Equal(t, tt.fnd, factory.fnd)
			// assert struct parser call
			err := factory.structParser(testData, &testStructure, testPath)
			assert.NoError(t, err)
			parserMock.AssertExpectations(t)
		})
	}
}

func TestNativeActionsFactory_ParseActions(t *testing.T) {
	const staticPath = "testPath"

	tests := []struct {
		name           string
		actions        []interface{}
		mockParseCalls []struct {
			data map[string]interface{}
			path string
			err  error
		}
		want    []types.Action
		wantErr bool
		errMsg  string
	}{
		{
			name: "Empty action string",
			actions: []interface{}{
				"",
			},
			wantErr: true,
			errMsg:  "action [0] string cannot be empty",
		},
		{
			name: "Valid bench action",
			actions: []interface{}{
				"bench/serviceName",
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.BenchAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Valid execute action",
			actions: []interface{}{
				"execute/serviceName",
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.ExecuteAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Invalid action format - too many elements",
			actions: []interface{}{
				"action/service/custom/extra",
			},
			wantErr: true,
			errMsg:  "action [0].action/service/custom/extra string cannot be composed of more than three elements",
		},
		{
			name: "Valid custom expectation action string",
			actions: []interface{}{
				map[string]interface{}{
					"expect/serviceName/customName": map[string]interface{}{"response": "expectedResponse"},
				},
			},
			mockParseCalls: nil,
			want: []types.Action{
				&types.CustomExpectationAction{
					Service:   "serviceName",
					When:      "on_success",
					OnFailure: "fail",
					Custom: types.CustomExpectation{
						Name:       "customName",
						Parameters: map[string]interface{}{"response": "expectedResponse"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid custom expectation action map",
			actions: []interface{}{
				map[string]interface{}{
					"expect": map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"when":    "always",
						"name":    "cname",
						"custom": map[string]interface{}{
							"id": "data",
						},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"name":    "cname",
						"when":    "always",
						"custom": map[string]interface{}{
							"id": "data",
						},
					},
					path: "testPath",
					err:  nil,
				},
			},
			want: []types.Action{
				&types.CustomExpectationAction{},
			},
			wantErr: false,
		},
		{
			name: "Valid custom expectation with inline service",
			actions: []interface{}{
				map[string]interface{}{
					"expect/serviceName": map[string]interface{}{
						"timeout": 1000,
						"name":    "cname",
						"custom": map[string]interface{}{
							"id": "data",
						},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{
						"timeout": 1000,
						"name":    "cname",
						"custom": map[string]interface{}{
							"id": "data",
						},
					},
					path: "testPath",
					err:  nil,
				},
			},
			want: []types.Action{
				&types.CustomExpectationAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Valid metrics expectation action",
			actions: []interface{}{
				map[string]interface{}{
					"expect": map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"metrics": map[string]interface{}{
							"id": "data",
							"rules": map[string]interface{}{
								"metric": "name",
							},
						},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"metrics": map[string]interface{}{
							"id": "data",
							"rules": map[string]interface{}{
								"metric": "name",
							},
						},
					},
					path: "testPath",
					err:  nil,
				},
			},
			want: []types.Action{
				&types.MetricsExpectationAction{},
			},
			wantErr: false,
		},
		{
			name: "Valid metrics expectation action with inline service",
			actions: []interface{}{
				map[string]interface{}{
					"expect/serviceName": map[string]interface{}{
						"timeout": 1000,
						"metrics": map[string]interface{}{
							"id": "data",
							"rules": map[string]interface{}{
								"metric": "name",
							},
						},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{
						"timeout": 1000,
						"metrics": map[string]interface{}{
							"id": "data",
							"rules": map[string]interface{}{
								"metric": "name",
							},
						},
					},
					path: "testPath",
					err:  nil,
				},
			},
			want: []types.Action{
				&types.MetricsExpectationAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Valid output expectation action",
			actions: []interface{}{
				map[string]interface{}{
					"expect": map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"output": map[string]interface{}{
							"message": "data",
						},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"output": map[string]interface{}{
							"message": "data",
						},
					},
					path: "testPath",
					err:  nil,
				},
			},
			want: []types.Action{
				&types.OutputExpectationAction{},
			},
			wantErr: false,
		},
		{
			name: "Valid output expectation action with inline service",
			actions: []interface{}{
				map[string]interface{}{
					"expect/serviceName": map[string]interface{}{
						"timeout": 1000,
						"output": map[string]interface{}{
							"message": "data",
						},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{
						"timeout": 1000,
						"output": map[string]interface{}{
							"message": "data",
						},
					},
					path: "testPath",
					err:  nil,
				},
			},
			want: []types.Action{
				&types.OutputExpectationAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Failed parsing of the output expectation action",
			actions: []interface{}{
				map[string]interface{}{
					"expect": map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"output": map[string]interface{}{
							"message": "data",
						},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"output": map[string]interface{}{
							"message": "data",
						},
					},
					path: "testPath",
					err:  errors.New("parsing failed"),
				},
			},
			want:    nil,
			wantErr: true,
			errMsg:  "parsing failed",
		},
		{
			name: "Invalid expectation key",
			actions: []interface{}{
				map[string]interface{}{
					"expect": map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"unknown": map[string]interface{}{
							"message": "data",
						},
					},
				},
			},
			mockParseCalls: nil,
			want:           nil,
			wantErr:        true,
			errMsg:         "invalid expectation key unknown",
		},
		{
			name: "Valid response expectation action",
			actions: []interface{}{
				map[string]interface{}{
					"expect": map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"response": map[string]interface{}{
							"request": "data",
						},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{
						"service": "serviceName",
						"timeout": 1000,
						"response": map[string]interface{}{
							"request": "data",
						},
					},
					path: "testPath",
					err:  nil,
				},
			},
			want: []types.Action{
				&types.ResponseExpectationAction{},
			},
			wantErr: false,
		},
		{
			name: "Invalid expectation key",
			actions: []interface{}{
				map[string]interface{}{
					"expect/serviceName": map[string]interface{}{"invalidKey": "someValue"},
				},
			},
			wantErr: true,
			errMsg:  "invalid expectation key invalidKey",
		},
		{
			name: "Multiple expectation types error",
			actions: []interface{}{
				map[string]interface{}{
					"expect/serviceName": map[string]interface{}{"metrics": map[string]interface{}{}, "output": map[string]interface{}{}},
				},
			},
			wantErr: true,
			errMsg:  "expression cannot have multiple types - additional key",
		},
		{
			name: "Valid not action",
			actions: []interface{}{
				map[string]interface{}{
					"not": map[string]interface{}{"action": map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"action": map[string]interface{}{}},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.NotAction{},
			},
			wantErr: false,
		},
		{
			name: "Invalid not action - service name present",
			actions: []interface{}{
				map[string]interface{}{
					"not/serviceName": map[string]interface{}{
						"action": map[string]interface{}{},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{
						"action": map[string]interface{}{},
					},
					path: staticPath,
					err:  nil,
				},
			},
			want:    nil,
			wantErr: true,
			errMsg:  "service name not allowed for action not",
		},
		{
			name: "Valid parallel action",
			actions: []interface{}{
				map[string]interface{}{
					"parallel": map[string]interface{}{"actions": []map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"actions": []map[string]interface{}{}},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.ParallelAction{},
			},
			wantErr: false,
		},
		{
			name: "Valid reload action",
			actions: []interface{}{
				map[string]interface{}{
					"reload/serviceName": map[string]interface{}{"services": []map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"services": []map[string]interface{}{}},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.ReloadAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Valid request action",
			actions: []interface{}{
				map[string]interface{}{
					"request/serviceName": map[string]interface{}{"id": "test"},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"id": "test"},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.RequestAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Valid restart action",
			actions: []interface{}{
				map[string]interface{}{
					"restart/serviceName": map[string]interface{}{"services": []map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"services": []map[string]interface{}{}},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.RestartAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Valid sequential action",
			actions: []interface{}{
				map[string]interface{}{
					"sequential": map[string]interface{}{"actions": []map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"actions": []map[string]interface{}{}},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.SequentialAction{},
			},
			wantErr: false,
		},
		{
			name: "Valid sequential action with service name and custom name",
			actions: []interface{}{
				map[string]interface{}{
					"sequential/serviceName/customName": map[string]interface{}{
						"actions": []map[string]interface{}{},
					},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"actions": []map[string]interface{}{}},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.SequentialAction{
					Service: "serviceName",
					Name:    "customName",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid start action",
			actions: []interface{}{
				map[string]interface{}{
					"start/serviceName": map[string]interface{}{"services": []map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"services": []map[string]interface{}{}},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.StartAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Valid stop action",
			actions: []interface{}{
				map[string]interface{}{
					"stop/serviceName": map[string]interface{}{"services": []map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"services": []map[string]interface{}{}},
					path: staticPath,
					err:  nil,
				},
			},
			want: []types.Action{
				&types.StopAction{Service: "serviceName"},
			},
			wantErr: false,
		},
		{
			name: "Invalid stop action because of custom name",
			actions: []interface{}{
				map[string]interface{}{
					"stop/serviceName/customName": map[string]interface{}{"services": []map[string]interface{}{}},
				},
			},
			mockParseCalls: []struct {
				data map[string]interface{}
				path string
				err  error
			}{
				{
					data: map[string]interface{}{"services": []map[string]interface{}{}},
					path: staticPath,
					err:  nil,
				},
			},
			want:    nil,
			wantErr: true,
			errMsg:  "custom name not allowed for action stop",
		},
		{
			name: "Unknown action",
			actions: []interface{}{
				"unsupported",
			},
			wantErr: true,
			errMsg:  "unknown action unsupported",
		},
		{
			name: "Unsupported action type",
			actions: []interface{}{
				1,
			},
			wantErr: true,
			errMsg:  "unsupported action [0] type int",
		},
		{
			name: "Unsupported action type",
			actions: []interface{}{
				"unsupported",
			},
			wantErr: true,
			errMsg:  "unknown action unsupported",
		},
		{
			name: "Invalid action format - exactly one elelemnt",
			actions: []interface{}{
				map[string]interface{}{
					"extra":            "test",
					"stop/serviceName": map[string]interface{}{"services": []map[string]interface{}{}},
				},
			},
			wantErr: true,
			errMsg:  "invalid action [0] format - exactly one item in map is required",
		},
		{
			name: "Invalid action format - not an object",
			actions: []interface{}{
				map[string]interface{}{"action": "value"},
			},
			wantErr: true,
			errMsg:  "invalid action [0].action format - action value must be an object",
		},
		{
			name: "Invalid action format - empty",
			actions: []interface{}{
				map[string]interface{}{},
			},
			wantErr: true,
			errMsg:  "invalid action [0] format - empty object is not valid action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parserMock := localMocks.NewMockMiniParser(t)
			assert := assert.New(t)

			// Setup mock expectations
			totalCalls := 0 // Initialize a counter for the total expected calls to ParseStruct
			for _, call := range tt.mockParseCalls {
				parserMock.On("ParseStruct", call.data, mock.Anything, call.path).Return(call.err).Once()
				totalCalls++ // Increment for each mock call setup
			}

			f := &NativeActionsFactory{
				structParser: parserMock.ParseStruct,
				loc:          location.CreateLocation(),
			}

			got, err := f.ParseActions(tt.actions, staticPath)

			if tt.wantErr {
				assert.Error(err)
				if tt.errMsg != "" {
					assert.ErrorContains(err, tt.errMsg)
				}
			} else {
				assert.NoError(err)
				// Compare the received actions with the expected ones if no error is expected
				assert.Equal(tt.want, got)
			}

			// Ensure all expectations on the mock are met
			parserMock.AssertExpectations(t)
			// Additionally, assert that ParseStruct was called the expected number of times
			parserMock.AssertNumberOfCalls(t, "ParseStruct", totalCalls)
		})
	}
}
