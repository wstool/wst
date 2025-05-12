package request

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	runtimeMocks "github.com/wstool/wst/mocks/generated/run/instances/runtime"
	servicesMocks "github.com/wstool/wst/mocks/generated/run/services"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/services"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestCreateActionMaker(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	tests := []struct {
		name string
		fnd  app.Foundation
	}{
		{
			name: "create maker",
			fnd:  fndMock,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateActionMaker(tt.fnd)
			assert.Equal(t, tt.fnd, got.fnd)
		})
	}
}

func TestActionMaker_Make(t *testing.T) {
	tests := []struct {
		name              string
		config            *types.RequestAction
		defaultTimeout    int
		setupMocks        func(*testing.T, *servicesMocks.MockServiceLocator) services.Service
		getExpectedAction func(*appMocks.MockFoundation, services.Service) *Action
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "successful request action creation with default timeout",
			config: &types.RequestAction{
				Service:    "validService",
				Timeout:    0,
				When:       "on_success",
				OnFailure:  "fail",
				Id:         "last",
				Path:       "/test",
				EncodePath: true,
				Method:     "GET",
				Headers: types.Headers{
					"content-type": "application/json",
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Service {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", "validService").Return(svc, nil)
				return svc
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svc services.Service) *Action {
				return &Action{
					fnd:        fndMock,
					service:    svc,
					timeout:    5000 * time.Millisecond,
					when:       action.OnSuccess,
					onFailure:  action.Fail,
					id:         "last",
					path:       "/test",
					encodePath: true,
					method:     "GET",
					headers: types.Headers{
						"content-type": "application/json",
					},
				}
			},
		},
		{
			name: "successful request action creation with config timeout",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Path:      "/t1",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Service {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", "validService").Return(svc, nil)
				return svc
			},
			getExpectedAction: func(fndMock *appMocks.MockFoundation, svc services.Service) *Action {
				return &Action{
					fnd:       fndMock,
					service:   svc,
					timeout:   3000 * time.Millisecond,
					when:      action.OnSuccess,
					onFailure: action.Fail,
					id:        "new",
					path:      "/t1",
					method:    "POST",
					headers: types.Headers{
						"content-type": "application/json",
					},
				}
			},
		},
		{
			name: "failure request action creation because service not found",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Path:      "/t1",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator) services.Service {
				sl.On("Find", "validService").Return(nil, errors.New("nf"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "nf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			slMock := servicesMocks.NewMockServiceLocator(t)
			m := &ActionMaker{
				fnd: fndMock,
			}

			svcs := tt.setupMocks(t, slMock)

			got, err := m.Make(tt.config, slMock, tt.defaultTimeout)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				actualAction, ok := got.(*Action)
				assert.True(t, ok)
				expectedAction := tt.getExpectedAction(fndMock, svcs)
				assert.Equal(t, expectedAction, actualAction)
			}
		})
	}
}

type bodyReader struct {
	msg string
	err string
}

func (b *bodyReader) Read(p []byte) (n int, err error) {
	if len(b.err) > 0 {
		return 0, errors.New(b.err)
	}
	if len(b.msg) > 0 {
		n = copy(p, b.msg)
		b.msg = b.msg[n:]
		return n, nil
	}
	return 0, io.EOF
}

func (b *bodyReader) Close() error {
	return nil
}

func TestAction_Execute(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		path       string
		encodePath bool
		method     string
		headers    types.Headers
		setupMocks func(
			t *testing.T,
			ctx context.Context,
			rd *runtimeMocks.MockData,
			fnd *appMocks.MockFoundation,
			svc *servicesMocks.MockService,
		)
		contextSetup     func() context.Context
		want             bool
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:       "successful execution with path encoding",
			id:         "r1",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				url := "https://example.com/test"
				svc.On("PublicUrl", "/test").Return(url, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				assert.Nil(t, err)
				expectedRequest.Header.Add("content-type", "application/json")
				expectedRequest.Header.Add("user-agent", "wst")
				body := &bodyReader{msg: "test"}
				header := http.Header{
					"content-type": []string{"application/json"},
				}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient").Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "test",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution without path encoding",
			id:         "r1",
			path:       "/test",
			encodePath: false,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				publicUrl := "https://example.com/test"
				svc.On("PublicUrl", "/test").Return(publicUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", publicUrl, nil)
				assert.Nil(t, err)
				expectedRequest.Header.Add("content-type", "application/json")
				expectedRequest.Header.Add("user-agent", "wst")
				expectedRequest.URL = &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Opaque: "//example.com/test",
				}
				body := &bodyReader{msg: "test"}
				header := http.Header{
					"content-type": []string{"application/json"},
				}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient").Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "test",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "failed execution due to failed storing",
			id:         "r1",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				url := "https://example.com/test"
				svc.On("PublicUrl", "/test").Return(url, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				assert.Nil(t, err)
				expectedRequest.Header.Add("content-type", "application/json")
				expectedRequest.Header.Add("user-agent", "wst")
				body := &bodyReader{msg: "test"}
				header := http.Header{
					"content-type": []string{"application/json"},
				}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient").Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "test",
					Headers: header,
				}).Return(errors.New("store failed"))
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "store failed",
		},
		{
			name:       "failed execution due to failed response reading",
			id:         "r1",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				url := "https://example.com/test"
				svc.On("PublicUrl", "/test").Return(url, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				assert.Nil(t, err)
				expectedRequest.Header.Add("content-type", "application/json")
				expectedRequest.Header.Add("user-agent", "wst")
				body := &bodyReader{err: "failed read"}
				header := http.Header{
					"content-type": []string{"application/json"},
				}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient").Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "failed read",
		},
		{
			name:       "failed execution due to context being done",
			id:         "r1",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				url := "https://example.com/test"
				svc.On("PublicUrl", "/test").Return(url, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				assert.Nil(t, err)
				expectedRequest.Header.Add("content-type", "application/json")
				expectedRequest.Header.Add("user-agent", "wst")
				body := &bodyReader{err: "failed read"}
				header := http.Header{
					"content-type": []string{"application/json"},
				}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient").Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
			},
			contextSetup: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "context canceled",
		},
		{
			name:       "failed execution due to client do failure",
			id:         "r1",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				url := "https://example.com/test"
				svc.On("PublicUrl", "/test").Return(url, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				assert.Nil(t, err)
				expectedRequest.Header.Add("content-type", "application/json")
				expectedRequest.Header.Add("user-agent", "wst")
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient").Return(client)
				client.On("Do", expectedRequest).Return(nil, errors.New("client fail"))
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "client fail",
		},
		{
			name:       "failed execution due to invalid request",
			id:         "r1",
			path:       "/test",
			encodePath: true,
			method:     "=",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				url := "https://example.com/test"
				svc.On("PublicUrl", "/test").Return(url, nil)
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "invalid method",
		},
		{
			name:       "failed execution due to public url failing",
			id:         "r1",
			path:       "/test",
			encodePath: true,
			method:     "=",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				svc.On("PublicUrl", "/test").Return("", errors.New("pub url"))
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "pub url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			runDataMock := runtimeMocks.NewMockData(t)
			svcMock := servicesMocks.NewMockService(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Return(mockLogger.SugaredLogger)
			var ctx context.Context
			if tt.contextSetup == nil {
				ctx = context.Background()
			} else {
				ctx = tt.contextSetup()
			}

			tt.setupMocks(t, ctx, runDataMock, fndMock, svcMock)

			a := &Action{
				fnd:        fndMock,
				service:    svcMock,
				timeout:    3000 * time.Millisecond,
				id:         tt.id,
				path:       tt.path,
				encodePath: tt.encodePath,
				method:     tt.method,
				headers:    tt.headers,
			}

			got, err := a.Execute(ctx, runDataMock)

			if tt.expectError {
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

func TestAction_Timeout(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &Action{
		fnd:     fndMock,
		timeout: 2000 * time.Millisecond,
	}
	assert.Equal(t, 2000*time.Millisecond, a.Timeout())
}

func TestAction_OnFailure(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &Action{
		fnd:       fndMock,
		when:      action.OnSuccess,
		onFailure: action.Skip,
	}
	assert.Equal(t, action.Skip, a.OnFailure())
}

func TestAction_When(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &Action{
		fnd:  fndMock,
		when: action.OnSuccess,
	}
	assert.Equal(t, action.OnSuccess, a.When())
}
