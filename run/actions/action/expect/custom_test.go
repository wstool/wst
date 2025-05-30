package expect

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	runtimeMocks "github.com/wstool/wst/mocks/generated/run/instances/runtime"
	parametersMocks "github.com/wstool/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters/parameter"
	serversMocks "github.com/wstool/wst/mocks/generated/run/servers"
	actionsMocks "github.com/wstool/wst/mocks/generated/run/servers/actions"
	servicesMocks "github.com/wstool/wst/mocks/generated/run/services"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/actions/action/request"
	"github.com/wstool/wst/run/environments/environment/output"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/parameters"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestExpectationActionMaker_MakeCustomAction(t *testing.T) {
	outputExpectation := &expectations.OutputExpectation{
		OrderType:      "exact",
		MatchType:      "regexp",
		OutputType:     "any",
		Messages:       nil,
		RenderTemplate: true,
	}
	responseExpectation := &expectations.ResponseExpectation{
		Request:            "last",
		Headers:            nil,
		BodyContent:        "test",
		BodyMatch:          "exact",
		BodyRenderTemplate: true,
	}

	tests := []struct {
		name              string
		config            *types.CustomExpectationAction
		defaultTimeout    int
		setupMocks        func(*testing.T, *servicesMocks.MockServiceLocator, *servicesMocks.MockService, *parametersMocks.MockMaker) parameters.Parameters
		getExpectedAction func(fndMock *appMocks.MockFoundation, svc *servicesMocks.MockService, finalParams parameters.Parameters) *customAction
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "successful custom action creation with parameter inheritance",
			config: &types.CustomExpectationAction{
				Service:   "validService",
				When:      "on_success",
				OnFailure: "fail",
				Custom: types.CustomExpectation{
					Name:       "validAction",
					Parameters: types.Parameters{"config_param": "config_value", "shared_key": "config_override"},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, svc *servicesMocks.MockService, paramsMaker *parametersMocks.MockMaker) parameters.Parameters {
				// Create different parameter sets to test inheritance chain
				configParams := parameters.Parameters{
					"config_param": parameterMocks.NewMockParameter(t),
					"shared_key":   parameterMocks.NewMockParameter(t),
				}

				expectParams := parameters.Parameters{
					"expect_param": parameterMocks.NewMockParameter(t),
					"shared_key":   parameterMocks.NewMockParameter(t), // Should be overridden by configParams
				}

				serverParams := parameters.Parameters{
					"server_param": parameterMocks.NewMockParameter(t),
					"shared_key":   parameterMocks.NewMockParameter(t), // Should be overridden by configParams or expectParams
					"expect_param": parameterMocks.NewMockParameter(t), // Should be overridden by expectParams
				}

				// Expected final parameters after inheritance chain (config -> expect -> server)
				finalParams := parameters.Parameters{
					"config_param": configParams["config_param"],
					"expect_param": expectParams["expect_param"],
					"server_param": serverParams["server_param"],
					"shared_key":   configParams["shared_key"], // Should come from config (highest priority)
				}

				sl.On("Find", "validService").Return(svc, nil)
				server := serversMocks.NewMockServer(t)
				svc.On("Server").Return(server)
				expectation := actionsMocks.NewMockExpectAction(t)

				paramsMaker.On("Make", types.Parameters{"config_param": "config_value", "shared_key": "config_override"}).Return(configParams, nil)

				server.On("ExpectAction", "validAction").Return(expectation, true)
				expectation.On("OutputExpectation").Return(outputExpectation)
				expectation.On("ResponseExpectation").Return(responseExpectation)
				expectation.On("Parameters").Return(expectParams)

				svc.On("ServerParameters").Return(serverParams)

				return finalParams
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svc *servicesMocks.MockService, finalParams parameters.Parameters) *customAction {
				return &customAction{
					CommonExpectation: &CommonExpectation{
						fnd:       fndMock,
						service:   svc,
						timeout:   5000 * 1e6,
						when:      action.OnSuccess,
						onFailure: action.Fail,
					},
					OutputExpectation:   outputExpectation,
					ResponseExpectation: responseExpectation,
					parameters:          finalParams,
				}
			},
		},
		{
			name: "parameters maker error",
			config: &types.CustomExpectationAction{
				Service: "validService",
				When:    "on_success",
				Custom: types.CustomExpectation{
					Name:       "invalidAction",
					Parameters: types.Parameters{"key": "value"},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, svc *servicesMocks.MockService, paramsMaker *parametersMocks.MockMaker) parameters.Parameters {
				sl.On("Find", "validService").Return(svc, nil)
				server := serversMocks.NewMockServer(t)
				svc.On("Server").Return(server)
				expectation := actionsMocks.NewMockExpectAction(t)
				server.On("ExpectAction", "invalidAction").Return(expectation, true)
				paramsMaker.On("Make", types.Parameters{"key": "value"}).Return(nil, errors.New("no params"))

				return nil
			},
			expectError:      true,
			expectedErrorMsg: "no params",
		},
		{
			name: "expectation action not found",
			config: &types.CustomExpectationAction{
				Service: "validService",
				When:    "on_success",
				Custom: types.CustomExpectation{
					Name: "invalidAction",
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, svc *servicesMocks.MockService, paramsMaker *parametersMocks.MockMaker) parameters.Parameters {
				sl.On("Find", "validService").Return(svc, nil)
				server := serversMocks.NewMockServer(t)
				svc.On("Server").Return(server)
				server.On("ExpectAction", "invalidAction").Return(nil, false)

				return nil
			},
			expectError:      true,
			expectedErrorMsg: "expectation action invalidAction not found",
		},
		{
			name: "service locator error",
			config: &types.CustomExpectationAction{
				Service: "invalidService",
				When:    "on_success",
				Custom: types.CustomExpectation{
					Name: "validAction",
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, svc *servicesMocks.MockService, paramsMaker *parametersMocks.MockMaker) parameters.Parameters {
				sl.On("Find", "invalidService").Return(nil, fmt.Errorf("service not found"))

				return nil
			},
			expectError:      true,
			expectedErrorMsg: "service not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			slMock := servicesMocks.NewMockServiceLocator(t)
			svcMock := servicesMocks.NewMockService(t)
			parametersMakerMock := parametersMocks.NewMockMaker(t)
			m := &ExpectationActionMaker{
				fnd:             fndMock,
				parametersMaker: parametersMakerMock,
			}

			expectedFinalParams := tt.setupMocks(t, slMock, svcMock, parametersMakerMock)

			got, err := m.MakeCustomAction(tt.config, slMock, tt.defaultTimeout)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				actualAction, ok := got.(*customAction)
				assert.True(t, ok)

				// Create the expected action using the final parameters
				expectedAction := tt.getExpectedAction(fndMock, svcMock, expectedFinalParams)

				// Compare the actual action with expected action
				assert.Equal(t, expectedAction, actualAction, "The customAction did not match expected structure")
			}
		})
	}
}

func Test_customAction_Execute(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(
			*testing.T,
			*appMocks.MockFoundation,
			context.Context,
			*runtimeMocks.MockData,
			*servicesMocks.MockService,
			parameters.Parameters,
			output.Type,
		)
		outputExpectation   *expectations.OutputExpectation
		responseExpectation *expectations.ResponseExpectation
		parameters          parameters.Parameters
		want                bool
		expectedOutputType  output.Type
		expectErr           bool
		expectedErrorMsg    string
	}{
		{
			name: "output expectation set",
			outputExpectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeFixed,
				MatchType:      expectations.MatchTypeExact,
				OutputType:     expectations.OutputTypeAny,
				Messages:       []string{"test"},
				RenderTemplate: true,
			},
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				svc.On("RenderTemplate", "test", params).Return("test tmp", nil)
				r := strings.NewReader("test tmp")
				svc.On("OutputReader", ctx, outputType).Return(r, nil)
			},
			expectedOutputType: output.Any,
			want:               true,
		},
		{
			name: "response expectation set",
			responseExpectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "test",
				BodyMatch:          expectations.MatchTypeExact,
				BodyRenderTemplate: true,
			},
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:    "test tmp",
					Headers: http.Header{"content-type": []string{"application/json"}},
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "test", params).Return("test tmp", nil)
			},
			expectedOutputType: output.Any,
			want:               true,
		},
		{
			name: "unknown expectation set",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
			},
			expectedOutputType: output.Any,
			want:               false,
			expectErr:          true,
			expectedErrorMsg:   "no expectation set",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := parameters.Parameters{
				"test": parameterMocks.NewMockParameter(t),
			}
			fndMock := appMocks.NewMockFoundation(t)
			dataMock := runtimeMocks.NewMockData(t)
			svcMock := servicesMocks.NewMockService(t)
			ctx := context.Background()

			tt.setupMocks(t, fndMock, ctx, dataMock, svcMock, params, tt.expectedOutputType)

			a := &customAction{
				CommonExpectation: &CommonExpectation{
					fnd:     fndMock,
					service: svcMock,
					timeout: 20 * 1e6,
				},
				OutputExpectation:   tt.outputExpectation,
				ResponseExpectation: tt.responseExpectation,
				parameters:          params,
			}

			got, err := a.Execute(ctx, dataMock)

			if tt.expectErr {
				assert.Error(t, err)
				assert.False(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_customAction_Timeout(t *testing.T) {
	timeout := time.Duration(50 * 1e6)
	a := &customAction{
		CommonExpectation: &CommonExpectation{
			fnd:     nil,
			service: nil,
			timeout: timeout,
		},
	}
	assert.Equal(t, timeout, a.Timeout())
}

func Test_customAction_When(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &customAction{
		CommonExpectation: &CommonExpectation{
			fnd:     fndMock,
			service: nil,
			when:    action.OnSuccess,
		},
	}
	assert.Equal(t, action.OnSuccess, a.When())
}
