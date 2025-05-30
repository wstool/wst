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
		) (*expectations.ResponseExpectation, parameters.Parameters)
		getExpectedAction func(
			*appMocks.MockFoundation,
			*servicesMocks.MockService,
			*expectations.ResponseExpectation,
			parameters.Parameters,
		) *responseAction
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful response action creation",
			config: &types.ResponseExpectationAction{
				Service:   "validService",
				When:      "on_success",
				OnFailure: "ignore",
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
			) (*expectations.ResponseExpectation, parameters.Parameters) {
				// Create server parameters to be used in the action
				serverParams := parameters.Parameters{
					"response_param1": parameterMocks.NewMockParameter(t),
					"response_param2": parameterMocks.NewMockParameter(t),
				}

				sl.On("Find", "validService").Return(svc, nil)
				responseExpectation := &expectations.ResponseExpectation{
					Request:            "last",
					Headers:            types.Headers{"h1": "test"},
					BodyContent:        "data",
					BodyMatch:          expectations.MatchTypeExact,
					BodyRenderTemplate: true,
				}
				expectationMaker.On("MakeResponseExpectation", &config.Response).Return(responseExpectation, nil)

				// Mock the ServerParameters method to return our test parameters
				svc.On("ServerParameters").Return(serverParams)

				return responseExpectation, serverParams
			},
			getExpectedAction: func(
				fndMock *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				expectation *expectations.ResponseExpectation,
				serverParams parameters.Parameters,
			) *responseAction {
				return &responseAction{
					CommonExpectation: &CommonExpectation{
						fnd:       fndMock,
						service:   svc,
						timeout:   5000 * 1e6,
						when:      action.OnSuccess,
						onFailure: action.Ignore,
					},
					ResponseExpectation: expectation,
					parameters:          serverParams,
				}
			},
		},
		{
			name: "failed response action creation because no service found",
			config: &types.ResponseExpectationAction{
				Service:   "invalidService",
				When:      "on_success",
				OnFailure: "fail",
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
			) (*expectations.ResponseExpectation, parameters.Parameters) {
				sl.On("Find", "invalidService").Return(nil, errors.New("svc not found"))
				return nil, nil
			},
			expectError:      true,
			expectedErrorMsg: "svc not found",
		},
		{
			name: "failed response action creation because response expectation creation failed",
			config: &types.ResponseExpectationAction{
				Service:   "validService",
				When:      "on_success",
				OnFailure: "fail",
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
			) (*expectations.ResponseExpectation, parameters.Parameters) {
				sl.On("Find", "validService").Return(svc, nil)
				expectationMaker.On("MakeResponseExpectation", &config.Response).Return(nil, errors.New("response failed"))
				return nil, nil
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

			responseExpectation, serverParams := tt.setupMocks(t, slMock, svcMock, expectationsMakerMock, tt.config)

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
				expectedAction := tt.getExpectedAction(fndMock, svcMock, responseExpectation, serverParams)
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
			name: "successful response with prefix body match and default status",
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
					Body:       "test tmp message",
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
				BodyMatch:          expectations.MatchTypePrefix,
				BodyRenderTemplate: true,
			},
			want: true,
		},
		{
			name: "successful response with suffix body match and default status",
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
					Body:       "message test tmp",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "tmp", params).Return("tmp", nil)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "tmp",
				BodyMatch:          expectations.MatchTypeSuffix,
				BodyRenderTemplate: true,
			},
			want: true,
		},
		{
			name: "successful response with infix body match and default status",
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
					Body:       "start test tmp end",
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
				BodyMatch:          expectations.MatchTypeInfix,
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
			name: "failed response with prefix body mismatch",
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
					Body:       "different start message",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "start", params).Return("test", nil)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "start",
				BodyMatch:          expectations.MatchTypePrefix,
				BodyRenderTemplate: true,
			},
			want: false,
		},
		{
			name: "failed response with suffix body mismatch",
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
					Body:       "message ending differently",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "ending", params).Return("test", nil)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "ending",
				BodyMatch:          expectations.MatchTypeSuffix,
				BodyRenderTemplate: true,
			},
			want: false,
		},
		{
			name: "failed response with infix body mismatch",
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
					Body:       "message without expected content",
					Headers:    http.Header{"content-type": []string{"application/json"}},
					StatusCode: 200,
				}
				rd.On("Load", "response/last").Return(response, true)
				svc.On("RenderTemplate", "test", params).Return("test", nil)
			},
			expectation: &expectations.ResponseExpectation{
				Request:            "last",
				Headers:            types.Headers{"content-type": "application/json"},
				BodyContent:        "test",
				BodyMatch:          expectations.MatchTypeInfix,
				BodyRenderTemplate: true,
			},
			want: false,
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
			fnd:       fndMock,
			service:   nil,
			when:      action.OnSuccess,
			onFailure: action.Skip,
		},
	}
	assert.Equal(t, action.OnSuccess, a.When())
}

func Test_responseAction_OnFailure(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &responseAction{
		CommonExpectation: &CommonExpectation{
			fnd:       fndMock,
			service:   nil,
			when:      action.OnSuccess,
			onFailure: action.Skip,
		},
	}
	assert.Equal(t, action.Skip, a.OnFailure())
}
