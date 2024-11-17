package expect

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	expectationsMocks "github.com/wstool/wst/mocks/generated/run/expectations"
	runtimeMocks "github.com/wstool/wst/mocks/generated/run/instances/runtime"
	parametersMocks "github.com/wstool/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/wstool/wst/mocks/generated/run/parameters/parameter"
	servicesMocks "github.com/wstool/wst/mocks/generated/run/services"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/actions/action/request"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/parameters"
	"net/http"
	"testing"
	"time"
)

func TestExpectationActionMaker_MakeResponseAction(t *testing.T) {
	tests := []struct {
		name           string
		config         *types.ResponseExpectationAction
		defaultTimeout int
		setupMocks     func(
			*testing.T,
			*servicesMocks.MockServiceLocator,
			*servicesMocks.MockService,
			*expectationsMocks.MockMaker,
			*types.ResponseExpectationAction,
		) *expectations.ResponseExpectation
		getExpectedAction func(
			*appMocks.MockFoundation,
			*servicesMocks.MockService,
			*expectations.ResponseExpectation,
		) *responseAction
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful output action creation",
			config: &types.ResponseExpectationAction{
				Service: "validService",
				When:    "on_success",
				Response: types.ResponseExpectation{
					Request: "last",
					Headers: types.Headers{"h1": "test"},
					Body: types.ResponseBody{
						Content:        "data",
						Match:          "exact",
						RenderTemplate: true,
					},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(
				t *testing.T,
				sl *servicesMocks.MockServiceLocator,
				svc *servicesMocks.MockService,
				expectationMaker *expectationsMocks.MockMaker,
				config *types.ResponseExpectationAction,
			) *expectations.ResponseExpectation {
				sl.On("Find", "validService").Return(svc, nil)
				responseExpectation := &expectations.ResponseExpectation{
					Request:            "last",
					Headers:            types.Headers{"h1": "test"},
					BodyContent:        "data",
					BodyMatch:          expectations.MatchTypeExact,
					BodyRenderTemplate: true,
				}
				expectationMaker.On("MakeResponseExpectation", &config.Response).Return(responseExpectation, nil)
				return responseExpectation
			},
			getExpectedAction: func(
				fndMock *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				expectation *expectations.ResponseExpectation,
			) *responseAction {
				return &responseAction{
					CommonExpectation: &CommonExpectation{
						fnd:     fndMock,
						service: svc,
						timeout: 5000 * 1e6,
						when:    action.OnSuccess,
					},
					ResponseExpectation: expectation,
					parameters:          parameters.Parameters{},
				}
			},
		},
		{
			name: "failed output action creation because no service found",
			config: &types.ResponseExpectationAction{
				Service: "invalidService",
				When:    "on_success",
				Response: types.ResponseExpectation{
					Request: "last",
					Headers: types.Headers{"h1": "test"},
					Body: types.ResponseBody{
						Content:        "data",
						Match:          "exact",
						RenderTemplate: true,
					},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(
				t *testing.T,
				sl *servicesMocks.MockServiceLocator,
				svc *servicesMocks.MockService,
				expectationMaker *expectationsMocks.MockMaker,
				config *types.ResponseExpectationAction,
			) *expectations.ResponseExpectation {
				sl.On("Find", "invalidService").Return(nil, errors.New("svc not found"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "svc not found",
		},
		{
			name: "failed output action creation because output expectation creation failed",
			config: &types.ResponseExpectationAction{
				Service: "validService",
				When:    "on_success",
				Response: types.ResponseExpectation{
					Request: "last",
					Headers: types.Headers{"h1": "test"},
					Body: types.ResponseBody{
						Content:        "data",
						Match:          "exact",
						RenderTemplate: true,
					},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(
				t *testing.T,
				sl *servicesMocks.MockServiceLocator,
				svc *servicesMocks.MockService,
				expectationMaker *expectationsMocks.MockMaker,
				config *types.ResponseExpectationAction,
			) *expectations.ResponseExpectation {
				sl.On("Find", "validService").Return(svc, nil)
				expectationMaker.On("MakeResponseExpectation", &config.Response).Return(nil, errors.New("response failed"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "response failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			slMock := servicesMocks.NewMockServiceLocator(t)
			svcMock := servicesMocks.NewMockService(t)
			parametersMakerMock := parametersMocks.NewMockMaker(t)
			expectationsMakerMock := expectationsMocks.NewMockMaker(t)
			m := &ExpectationActionMaker{
				fnd:               fndMock,
				parametersMaker:   parametersMakerMock,
				expectationsMaker: expectationsMakerMock,
			}

			outputExpectation := tt.setupMocks(t, slMock, svcMock, expectationsMakerMock, tt.config)

			got, err := m.MakeResponseAction(tt.config, slMock, tt.defaultTimeout)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				actualAction, ok := got.(*responseAction)
				assert.True(t, ok)
				expectedAction := tt.getExpectedAction(fndMock, svcMock, outputExpectation)
				assert.Equal(t, expectedAction, actualAction)
			}
		})
	}
}

func Test_responseAction_Execute(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(
			t *testing.T,
			fnd *appMocks.MockFoundation,
			ctx context.Context,
			rd *runtimeMocks.MockData,
			svc *servicesMocks.MockService,
			params parameters.Parameters,
		)
		expectation      *expectations.ResponseExpectation
		want             bool
		expectErr        bool
		expectedErrorMsg string
	}{
		{
			name: "successful response with exact body match and default status",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "test", params).Return("test tmp", nil)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "test",
				BodyMatch:          expectations.MatchTypeExact,
				BodyRenderTemplate: true,
			},
			want: true,
		},
		{
			name: "successful response with exact body match specific status",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "test", params).Return("test tmp", nil)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "test",
				BodyMatch:          expectations.MatchTypeExact,
				BodyRenderTemplate: true,
				StatusCode:         200,
			},
			want: true,
		},
		{
			name: "successful response with exact no body match",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "test x", params).Return("test x", nil)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "test x",
				BodyMatch:          expectations.MatchTypeExact,
				BodyRenderTemplate: true,
			},
			want: false,
		},
		{
			name: "successful response with exact no body match and dry run",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(true)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "test x", params).Return("test x", nil)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "test x",
				BodyMatch:          expectations.MatchTypeExact,
				BodyRenderTemplate: true,
			},
			want: true,
		},
		{
			name: "successful response with regexp body match",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "te.t\\st[mn]p",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: false,
			},
			want: true,
		},
		{
			name: "successful response with regexp no body match",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "te.t\\stp",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: false,
			},
			want: false,
		},
		{
			name: "successful response with regexp no body match and dry run",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(true)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "te.t\\stp",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: false,
			},
			want: true,
		},
		{
			name: "failed response with regexp body match because invalid pattern",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "te.a(a",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: false,
			},
			want:      false,
			expectErr: true,
		},
		{
			name: "failed response with body match because rendering failed",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "tex", params).Return("", errors.New("failed render"))
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "tex",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: true,
			},
			want:             false,
			expectErr:        true,
			expectedErrorMsg: "failed render",
		},
		{
			name: "successful response with no headers match",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"accept": "application/json"},
				BodyContent:        "tex",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: true,
			},
			want: false,
		},
		{
			name: "successful response with no headers match and dry run",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(true)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"accept": "application/json"},
				BodyContent:        "tex",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: true,
			},
			want: true,
		},
		{
			name: "successful response with no status code match",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 201,
				}
				rd.On("Load", "response/last").Return(response, true)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"accept": "application/json"},
				BodyContent:        "tex",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: true,
				StatusCode:         200,
			},
			want: false,
		},
		{
			name: "successful response with no status code match and dry run",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(true)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				response := request.ResponseData{
					Body:       "test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 201,
				}
				rd.On("Load", "response/last").Return(response, true)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"accept": "application/json"},
				BodyContent:        "tex",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: true,
				StatusCode:         200,
			},
			want: true,
		},
		{
			name: "failed response match because invalid loaded data type",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				rd.On("Load", "response/last").Return("data", true)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "tex",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: true,
			},
			want:             false,
			expectErr:        true,
			expectedErrorMsg: "invalid response data type",
		},
		{
			name: "failed response match because failed data loading",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				rd.On("Load", "response/last").Return(nil, false)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "tex",
				BodyMatch:          expectations.MatchTypeRegexp,
				BodyRenderTemplate: true,
			},
			want:             false,
			expectErr:        true,
			expectedErrorMsg: "response data not found",
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

			tt.setupMocks(t, fndMock, ctx, dataMock, svcMock, params)

			a := &responseAction{
				CommonExpectation: &CommonExpectation{
					fnd:     fndMock,
					service: svcMock,
					timeout: 20 * 1e6,
				},
				ResponseExpectation: tt.expectation,
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

func Test_responseAction_Timeout(t *testing.T) {
	timeout := time.Duration(50 * 1e6)
	a := &responseAction{
		CommonExpectation: &CommonExpectation{
			fnd:     nil,
			service: nil,
			timeout: timeout,
		},
		ResponseExpectation: &expectations.ResponseExpectation{
			Request:            "last",
			Headers:            types.Headers{"h1": "test"},
			BodyContent:        "data",
			BodyMatch:          expectations.MatchTypeExact,
			BodyRenderTemplate: true,
		},
	}
	assert.Equal(t, timeout, a.Timeout())
}

func Test_responseAction_When(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &responseAction{
		CommonExpectation: &CommonExpectation{
			fnd:     fndMock,
			service: nil,
			when:    action.OnSuccess,
		},
	}
	assert.Equal(t, action.OnSuccess, a.When())
}
