package actions

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/conf/types"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	expectationsMocks "github.com/wstool/wst/mocks/generated/run/expectations"
	parametersMocks "github.com/wstool/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters/parameter"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/parameters"
	"testing"
)

func testParams(t *testing.T, len int) []*parameterMocks.MockParameter {
	params := make([]*parameterMocks.MockParameter, len)
	for i := 0; i < len; i++ {
		param := parameterMocks.NewMockParameter(t)
		// Differentiate params
		param.TestData().Set("id", i)
		params[i] = param
	}
	return params
}

func TestActions_Inherit(t *testing.T) {
	params := testParams(t, 2)
	tests := []struct {
		name            string
		childActions    *Actions
		parentActions   *Actions
		expectedActions *Actions
	}{
		{
			name: "inherit new expectations",
			childActions: &Actions{
				Expect: map[string]ExpectAction{
					"existing": &expectOutputAction{
						parameters:        parameters.Parameters{"param1": params[0]},
						outputExpectation: &expectations.OutputExpectation{Messages: []string{"output1"}},
					},
				},
				Sequential: map[string]SequentialAction{
					"existing-seq": &nativeSequentialAction{
						actions: []types.Action{
							&types.RequestAction{Service: "service1"},
						},
					},
				},
			},
			parentActions: &Actions{
				Expect: map[string]ExpectAction{
					"new": &expectResponseAction{
						parameters:          parameters.Parameters{"param2": params[1]},
						responseExpectation: &expectations.ResponseExpectation{BodyContent: "resp"},
					},
				},
				Sequential: map[string]SequentialAction{
					"new-seq": &nativeSequentialAction{
						actions: []types.Action{
							&types.RequestAction{Service: "service2"},
						},
					},
				},
			},
			expectedActions: &Actions{
				Expect: map[string]ExpectAction{
					"existing": &expectOutputAction{
						parameters:        parameters.Parameters{"param1": params[0]},
						outputExpectation: &expectations.OutputExpectation{Messages: []string{"output1"}},
					},
					"new": &expectResponseAction{
						parameters:          parameters.Parameters{"param2": params[1]},
						responseExpectation: &expectations.ResponseExpectation{BodyContent: "resp"},
					},
				},
				Sequential: map[string]SequentialAction{
					"existing-seq": &nativeSequentialAction{
						actions: []types.Action{
							&types.RequestAction{Service: "service1"},
						},
					},
					"new-seq": &nativeSequentialAction{
						actions: []types.Action{
							&types.RequestAction{Service: "service2"},
						},
					},
				},
			},
		},
		{
			name: "do not override existing expectations and sequential actions",
			childActions: &Actions{
				Expect: map[string]ExpectAction{
					"common": &expectResponseAction{
						parameters:          parameters.Parameters{"param1": params[0]},
						responseExpectation: &expectations.ResponseExpectation{BodyContent: "resp"},
					},
				},
				Sequential: map[string]SequentialAction{
					"common-seq": &nativeSequentialAction{
						actions: []types.Action{
							&types.RequestAction{Service: "service1"},
						},
					},
				},
			},
			parentActions: &Actions{
				Expect: map[string]ExpectAction{
					"common": &expectResponseAction{
						parameters:          parameters.Parameters{"param1": params[1]},
						responseExpectation: &expectations.ResponseExpectation{BodyContent: "resp2"},
					},
				},
				Sequential: map[string]SequentialAction{
					"common-seq": &nativeSequentialAction{
						actions: []types.Action{
							&types.RequestAction{Service: "service2"},
						},
					},
				},
			},
			expectedActions: &Actions{
				Expect: map[string]ExpectAction{
					"common": &expectResponseAction{
						parameters:          parameters.Parameters{"param1": params[0]},
						responseExpectation: &expectations.ResponseExpectation{BodyContent: "resp"},
					},
				},
				Sequential: map[string]SequentialAction{
					"common-seq": &nativeSequentialAction{
						actions: []types.Action{
							&types.RequestAction{Service: "service1"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.childActions.Inherit(tt.parentActions)
			assert.Equal(t, tt.expectedActions, tt.childActions)
		})
	}
}

func Test_nativeMaker_Make(t *testing.T) {
	params := testParams(t, 2)
	tests := []struct {
		name             string
		configActions    *types.ServerActions
		setupMocks       func(t *testing.T, em *expectationsMocks.MockMaker, pm *parametersMocks.MockMaker)
		expectedResult   *Actions
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful actions creation",
			configActions: &types.ServerActions{
				Expect: map[string]types.ServerExpectationAction{
					"output": &types.ServerOutputExpectation{
						Parameters: types.Parameters{"key": "value"},
						Output: types.OutputExpectation{
							Match: "output match",
						},
					},
					"response": &types.ServerResponseExpectation{
						Parameters: types.Parameters{"key2": "value2"},
						Response: types.ResponseExpectation{
							Body: types.ResponseBody{Content: "resp"},
						},
					},
				},
				Sequential: map[string]types.ServerSequentialAction{
					"start": {
						Actions: []types.Action{
							types.StartAction{
								Service: "svc",
							},
						},
					},
				},
			},
			setupMocks: func(t *testing.T, em *expectationsMocks.MockMaker, pm *parametersMocks.MockMaker) {
				pm.On("Make", types.Parameters{
					"key": "value",
				}).Return(parameters.Parameters{"key": params[0]}, nil)
				pm.On("Make", types.Parameters{
					"key2": "value2",
				}).Return(parameters.Parameters{"key2": params[1]}, nil)
				em.On("MakeOutputExpectation", &types.OutputExpectation{
					Match: "output match",
				}).Return(&expectations.OutputExpectation{Messages: []string{"output match"}}, nil)
				em.On("MakeResponseExpectation", &types.ResponseExpectation{
					Body: types.ResponseBody{Content: "resp"},
				}).Return(&expectations.ResponseExpectation{BodyContent: "resp"}, nil)
			},
			expectedResult: &Actions{
				Expect: map[string]ExpectAction{
					"output": &expectOutputAction{
						parameters:        parameters.Parameters{"key": params[0]},
						outputExpectation: &expectations.OutputExpectation{Messages: []string{"output match"}},
					},
					"response": &expectResponseAction{
						parameters:          parameters.Parameters{"key2": params[1]},
						responseExpectation: &expectations.ResponseExpectation{BodyContent: "resp"},
					},
				},
				Sequential: map[string]SequentialAction{
					"start": &nativeSequentialAction{
						actions: []types.Action{
							types.StartAction{
								Service: "svc",
							},
						},
					},
				},
			},
		},
		{
			name: "error in output expectation making",
			configActions: &types.ServerActions{
				Expect: map[string]types.ServerExpectationAction{
					"output": &types.ServerOutputExpectation{
						Parameters: map[string]interface{}{"key": "value"},
						Output: types.OutputExpectation{
							Match: "output match",
						},
					},
				},
			},
			setupMocks: func(t *testing.T, em *expectationsMocks.MockMaker, pm *parametersMocks.MockMaker) {
				pm.On("Make", types.Parameters{
					"key": "value",
				}).Return(parameters.Parameters{"key": params[0]}, nil)
				em.On("MakeOutputExpectation", &types.OutputExpectation{
					Match: "output match",
				}).Return(nil, errors.New("expectation error"))
			},
			expectError:      true,
			expectedErrorMsg: "expectation error",
		},
		{
			name: "error in output params making",
			configActions: &types.ServerActions{
				Expect: map[string]types.ServerExpectationAction{
					"output": &types.ServerOutputExpectation{
						Parameters: map[string]interface{}{"key": "value"},
						Output: types.OutputExpectation{
							Match: "output match",
						},
					},
				},
			},
			setupMocks: func(t *testing.T, em *expectationsMocks.MockMaker, pm *parametersMocks.MockMaker) {
				pm.On("Make", types.Parameters{
					"key": "value",
				}).Return(nil, errors.New("params error"))

			},
			expectError:      true,
			expectedErrorMsg: "params error",
		},
		{
			name: "error in response expectation making",
			configActions: &types.ServerActions{
				Expect: map[string]types.ServerExpectationAction{
					"response": &types.ServerResponseExpectation{
						Parameters: types.Parameters{"key2": "value2"},
						Response: types.ResponseExpectation{
							Body: types.ResponseBody{Content: "resp"},
						},
					},
				},
			},
			setupMocks: func(t *testing.T, em *expectationsMocks.MockMaker, pm *parametersMocks.MockMaker) {
				pm.On("Make", types.Parameters{
					"key2": "value2",
				}).Return(parameters.Parameters{"key": params[0]}, nil)
				em.On("MakeResponseExpectation", &types.ResponseExpectation{
					Body: types.ResponseBody{Content: "resp"},
				}).Return(nil, errors.New("expectation error"))
			},
			expectError:      true,
			expectedErrorMsg: "expectation error",
		},
		{
			name: "error in response params making",
			configActions: &types.ServerActions{
				Expect: map[string]types.ServerExpectationAction{
					"response": &types.ServerResponseExpectation{
						Parameters: types.Parameters{"key2": "value2"},
						Response: types.ResponseExpectation{
							Body: types.ResponseBody{Content: "resp"},
						},
					},
				},
			},
			setupMocks: func(t *testing.T, em *expectationsMocks.MockMaker, pm *parametersMocks.MockMaker) {
				pm.On("Make", types.Parameters{
					"key2": "value2",
				}).Return(nil, errors.New("params error"))
			},
			expectError:      true,
			expectedErrorMsg: "params error",
		},
		{
			name: "error in expectation type",
			configActions: &types.ServerActions{
				Expect: map[string]types.ServerExpectationAction{
					"response": "string",
				},
			},
			setupMocks:       func(t *testing.T, em *expectationsMocks.MockMaker, pm *parametersMocks.MockMaker) {},
			expectError:      true,
			expectedErrorMsg: "invalid server expectation type string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			expectationsMock := expectationsMocks.NewMockMaker(t)
			parametersMock := parametersMocks.NewMockMaker(t)
			maker := CreateMaker(fndMock, expectationsMock, parametersMock)

			tt.setupMocks(t, expectationsMock, parametersMock)

			result, err := maker.Make(tt.configActions)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			expectationsMock.AssertExpectations(t)
			parametersMock.AssertExpectations(t)
		})
	}
}
