package expect

import (
	"context"
	"errors"
	"fmt"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	parametersMocks "github.com/bukka/wst/mocks/run/parameters"
	parameterMocks "github.com/bukka/wst/mocks/run/parameters/parameter"
	serversMocks "github.com/bukka/wst/mocks/run/servers"
	actionsMocks "github.com/bukka/wst/mocks/run/servers/actions"
	servicesMocks "github.com/bukka/wst/mocks/run/services"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/parameters"
	"github.com/stretchr/testify/assert"
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
			name:           "successful custom action creation",
			config:         &types.CustomExpectationAction{Service: "validService", Name: "validAction", Parameters: types.Parameters{"key": "value"}},
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
					},
					OutputExpectation:   outputExpectation,
					ResponseExpectation: responseExpectation,
					parameters:          params,
				}
			},
		},
		{
			name:           "parameters maker error",
			config:         &types.CustomExpectationAction{Service: "validService", Name: "invalidAction", Parameters: types.Parameters{"key": "value"}},
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
			name:           "expectation action not found",
			config:         &types.CustomExpectationAction{Service: "validService", Name: "invalidAction"},
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
			name:           "service locator error",
			config:         &types.CustomExpectationAction{Service: "invalidService", Name: "validAction"},
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
	type fields struct {
		CommonExpectation   *CommonExpectation
		OutputExpectation   *expectations.OutputExpectation
		ResponseExpectation *expectations.ResponseExpectation
		parameters          parameters.Parameters
	}
	type args struct {
		ctx     context.Context
		runData runtime.Data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &customAction{
				CommonExpectation:   tt.fields.CommonExpectation,
				OutputExpectation:   tt.fields.OutputExpectation,
				ResponseExpectation: tt.fields.ResponseExpectation,
				parameters:          tt.fields.parameters,
			}
			got, err := a.Execute(tt.args.ctx, tt.args.runData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Execute() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_customAction_Timeout(t *testing.T) {
	type fields struct {
		CommonExpectation   *CommonExpectation
		OutputExpectation   *expectations.OutputExpectation
		ResponseExpectation *expectations.ResponseExpectation
		parameters          parameters.Parameters
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &customAction{
				CommonExpectation:   tt.fields.CommonExpectation,
				OutputExpectation:   tt.fields.OutputExpectation,
				ResponseExpectation: tt.fields.ResponseExpectation,
				parameters:          tt.fields.parameters,
			}
			if got := a.Timeout(); got != tt.want {
				t.Errorf("Timeout() = %v, want %v", got, tt.want)
			}
		})
	}
}
