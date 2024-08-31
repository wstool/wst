package expect

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	runtimeMocks "github.com/bukka/wst/mocks/generated/run/instances/runtime"
	parametersMocks "github.com/bukka/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	serversMocks "github.com/bukka/wst/mocks/generated/run/servers"
	actionsMocks "github.com/bukka/wst/mocks/generated/run/servers/actions"
	servicesMocks "github.com/bukka/wst/mocks/generated/run/services"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/actions/action/request"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/parameters"
	"github.com/stretchr/testify/assert"
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
		setupMocks        func(*testing.T, *servicesMocks.MockServiceLocator, *servicesMocks.MockService, *parametersMocks.MockMaker, parameters.Parameters)
		getExpectedAction func(fndMock *appMocks.MockFoundation, svc *servicesMocks.MockService, params parameters.Parameters) *customAction
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "successful custom action creation",
			config: &types.CustomExpectationAction{
				Service: "validService",
				When:    "on_success",
				Custom: types.CustomExpectation{
					Name:       "validAction",
					Parameters: types.Parameters{"key": "value"},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, svc *servicesMocks.MockService, paramsMaker *parametersMocks.MockMaker, params parameters.Parameters) {
				sl.On("Find", "validService").Return(svc, nil)
				server := serversMocks.NewMockServer(t)
				svc.On("Server").Return(server)
				expectation := actionsMocks.NewMockExpectAction(t)
				paramsMaker.On("Make", types.Parameters{"key": "value"}).Return(params, nil)
				server.On("ExpectAction", "validAction").Return(expectation, true)
				server.On("Parameters").Return(params)
				expectation.On("OutputExpectation").Return(outputExpectation)
				expectation.On("ResponseExpectation").Return(responseExpectation)
				expectation.On("Parameters").Return(params)
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svc *servicesMocks.MockService, params parameters.Parameters) *customAction {
				return &customAction{
					CommonExpectation: &CommonExpectation{
						fnd:     fndMock,
						service: svc,
						timeout: 5000 * 1e6,
						when:    action.OnSuccess,
					},
					OutputExpectation:   outputExpectation,
					ResponseExpectation: responseExpectation,
					parameters:          params,
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
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, svc *servicesMocks.MockService, paramsMaker *parametersMocks.MockMaker, params parameters.Parameters) {
				sl.On("Find", "validService").Return(svc, nil)
				server := serversMocks.NewMockServer(t)
				svc.On("Server").Return(server)
				expectation := actionsMocks.NewMockExpectAction(t)
				server.On("ExpectAction", "invalidAction").Return(expectation, true)
				paramsMaker.On("Make", types.Parameters{"key": "value"}).Return(nil, errors.New("no params"))
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
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, svc *servicesMocks.MockService, paramsMaker *parametersMocks.MockMaker, params parameters.Parameters) {
				sl.On("Find", "validService").Return(svc, nil)
				server := serversMocks.NewMockServer(t)
				svc.On("Server").Return(server)
				server.On("ExpectAction", "invalidAction").Return(nil, false)
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
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, svc *servicesMocks.MockService, paramsMaker *parametersMocks.MockMaker, params parameters.Parameters) {
				sl.On("Find", "invalidService").Return(nil, fmt.Errorf("service not found"))
			},
			expectError:      true,
			expectedErrorMsg: "service not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := parameters.Parameters{
				"test": parameterMocks.NewMockParameter(t),
			}
			fndMock := appMocks.NewMockFoundation(t)
			slMock := servicesMocks.NewMockServiceLocator(t)
			svcMock := servicesMocks.NewMockService(t)
			parametersMakerMock := parametersMocks.NewMockMaker(t)
			m := &ExpectationActionMaker{
				fnd:             fndMock,
				parametersMaker: parametersMakerMock,
			}

			tt.setupMocks(t, slMock, svcMock, parametersMakerMock, params)

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
				expectedAction := tt.getExpectedAction(fndMock, svcMock, params)
				assert.Equal(t, expectedAction, actualAction)
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
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
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
