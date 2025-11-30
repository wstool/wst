package request

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	runtimeMocks "github.com/wstool/wst/mocks/generated/run/instances/runtime"
	certificatesMocks "github.com/wstool/wst/mocks/generated/run/resources/certificates"
	servicesMocks "github.com/wstool/wst/mocks/generated/run/services"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/resources/certificates"
	"github.com/wstool/wst/run/services"
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
		setupMocks        func(*testing.T, *servicesMocks.MockServiceLocator, *appMocks.MockFoundation) services.Service
		getExpectedAction func(*appMocks.MockFoundation, services.Service) *Action
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name: "successful request action creation with default timeout and default protocols (HTTP)",
			config: &types.RequestAction{
				Service:    "validService",
				Timeout:    0,
				When:       "on_success",
				OnFailure:  "fail",
				Id:         "last",
				Path:       "/test",
				Scheme:     "http",
				EncodePath: true,
				Method:     "GET",
				Headers: types.Headers{
					"content-type": "application/json",
				},
				TLS: types.TLSClientConfig{
					SkipVerify: false,
					CACert:     "",
				},
				Protocols: []string{},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, fnd *appMocks.MockFoundation) services.Service {
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
					scheme:     "http",
					path:       "/test",
					encodePath: true,
					method:     "GET",
					headers: types.Headers{
						"content-type": "application/json",
					},
					body: &types.RequestBody{},
					tls: &types.TLSClientConfig{
						SkipVerify: false,
						CACert:     "",
					},
					protocols: []Protocol{ProtocolHTTP11},
				}
			},
		},
		{
			name: "successful request action creation with default protocols (HTTPS)",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Scheme:    "https",
				Path:      "/t1",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
				TLS: types.TLSClientConfig{
					SkipVerify: false,
					CACert:     "",
				},
				Protocols: []string{},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, fnd *appMocks.MockFoundation) services.Service {
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
					scheme:    "https",
					path:      "/t1",
					method:    "POST",
					headers: types.Headers{
						"content-type": "application/json",
					},
					body: &types.RequestBody{},
					tls: &types.TLSClientConfig{
						SkipVerify: false,
						CACert:     "",
					},
					protocols: []Protocol{ProtocolHTTP11, ProtocolHTTP2},
				}
			},
		},
		{
			name: "successful request with explicit HTTP/1.1 only",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Scheme:    "https",
				Path:      "/t1",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
				TLS: types.TLSClientConfig{
					SkipVerify: true,
					CACert:     "",
				},
				Protocols: []string{"http1.1"},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, fnd *appMocks.MockFoundation) services.Service {
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
					scheme:    "https",
					path:      "/t1",
					method:    "POST",
					headers: types.Headers{
						"content-type": "application/json",
					},
					body: &types.RequestBody{},
					tls: &types.TLSClientConfig{
						SkipVerify: true,
						CACert:     "",
					},
					protocols: []Protocol{ProtocolHTTP11},
				}
			},
		},
		{
			name: "successful request with HTTP/2 over HTTPS",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Scheme:    "https",
				Path:      "/t1",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
				TLS: types.TLSClientConfig{
					SkipVerify: false,
					CACert:     "localhost-ca",
				},
				Protocols: []string{"http1.1", "http2"},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, fnd *appMocks.MockFoundation) services.Service {
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
					scheme:    "https",
					path:      "/t1",
					method:    "POST",
					headers: types.Headers{
						"content-type": "application/json",
					},
					body: &types.RequestBody{},
					tls: &types.TLSClientConfig{
						SkipVerify: false,
						CACert:     "localhost-ca",
					},
					protocols: []Protocol{ProtocolHTTP11, ProtocolHTTP2},
				}
			},
		},
		{
			name: "successful request with HTTP/2 cleartext (h2c)",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Scheme:    "http",
				Path:      "/t1",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
				TLS: types.TLSClientConfig{
					SkipVerify: false,
					CACert:     "",
				},
				Protocols: []string{"http1.1", "http2"},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, fnd *appMocks.MockFoundation) services.Service {
				svc := servicesMocks.NewMockService(t)
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
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
					scheme:    "http",
					path:      "/t1",
					method:    "POST",
					headers: types.Headers{
						"content-type": "application/json",
					},
					body: &types.RequestBody{},
					tls: &types.TLSClientConfig{
						SkipVerify: false,
						CACert:     "",
					},
					protocols: []Protocol{ProtocolHTTP11, ProtocolHTTP2},
				}
			},
		},
		{
			name: "successful request with body and transfer config",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Scheme:    "https",
				Path:      "/upload",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
				Body: types.RequestBody{
					Content:        "test body content",
					RenderTemplate: true,
					Transfer: types.TransferConfig{
						Encoding:      "chunked",
						ChunkSize:     1024,
						ChunkDelay:    100,
						ContentLength: 0,
					},
				},
				TLS: types.TLSClientConfig{
					SkipVerify: false,
					CACert:     "",
				},
				Protocols: []string{"http1.1"},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, fnd *appMocks.MockFoundation) services.Service {
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
					scheme:    "https",
					path:      "/upload",
					method:    "POST",
					headers: types.Headers{
						"content-type": "application/json",
					},
					body: &types.RequestBody{
						Content:        "test body content",
						RenderTemplate: true,
						Transfer: types.TransferConfig{
							Encoding:      "chunked",
							ChunkSize:     1024,
							ChunkDelay:    100,
							ContentLength: 0,
						},
					},
					tls: &types.TLSClientConfig{
						SkipVerify: false,
						CACert:     "",
					},
					protocols: []Protocol{ProtocolHTTP11},
				}
			},
		},
		{
			name: "failure TLS config with HTTP scheme - skip verify",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Scheme:    "http",
				Path:      "/t1",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
				TLS: types.TLSClientConfig{
					SkipVerify: true,
					CACert:     "",
				},
				Protocols: []string{},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, fnd *appMocks.MockFoundation) services.Service {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", "validService").Return(svc, nil)
				return svc
			},
			expectError:      true,
			expectedErrorMsg: "TLS configuration is only valid for HTTPS requests",
		},
		{
			name: "failure TLS config with HTTP scheme - CA cert",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Scheme:    "http",
				Path:      "/t1",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
				TLS: types.TLSClientConfig{
					SkipVerify: false,
					CACert:     "localhost-ca",
				},
				Protocols: []string{},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, fnd *appMocks.MockFoundation) services.Service {
				svc := servicesMocks.NewMockService(t)
				sl.On("Find", "validService").Return(svc, nil)
				return svc
			},
			expectError:      true,
			expectedErrorMsg: "TLS configuration is only valid for HTTPS requests",
		},
		{
			name: "failure request action creation because service not found",
			config: &types.RequestAction{
				Service:   "validService",
				Timeout:   3000,
				When:      "on_success",
				OnFailure: "fail",
				Id:        "new",
				Scheme:    "http",
				Path:      "/t1",
				Method:    "POST",
				Headers: types.Headers{
					"content-type": "application/json",
				},
				TLS: types.TLSClientConfig{
					SkipVerify: false,
					CACert:     "",
				},
				Protocols: []string{},
			},
			defaultTimeout: 5000,
			setupMocks: func(t *testing.T, sl *servicesMocks.MockServiceLocator, fnd *appMocks.MockFoundation) services.Service {
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

			svcs := tt.setupMocks(t, slMock, fndMock)

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
		scheme     string
		path       string
		encodePath bool
		method     string
		headers    types.Headers
		body       *types.RequestBody
		tls        *types.TLSClientConfig
		protocols  []Protocol
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
			name:       "successful execution with path encoding - HTTP with HTTP/1.1",
			id:         "r1",
			scheme:     "http",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: false,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "http://example.com/test"
				svc.On("PublicUrl", "http", "/test").Return(reqUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
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

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "test",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with body content",
			id:         "r1",
			scheme:     "https",
			path:       "/upload",
			encodePath: true,
			method:     "POST",
			headers: types.Headers{
				"content-type": "application/json",
			},
			body: &types.RequestBody{
				Content:        "test body content",
				RenderTemplate: false,
			},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/upload"
				svc.On("PublicUrl", "https", "/upload").Return(reqUrl, nil)
				body := &bodyReader{msg: "ok"}
				header := http.Header{}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == "POST" &&
						req.URL.String() == reqUrl &&
						req.Header.Get("content-type") == "application/json" &&
						req.ContentLength == 17 &&
						req.Body != nil
				})).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "ok",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with chunked encoding and chunk size",
			id:         "r1",
			scheme:     "https",
			path:       "/upload",
			encodePath: true,
			method:     "POST",
			headers: types.Headers{
				"content-type": "application/octet-stream",
			},
			body: &types.RequestBody{
				Content:        "test chunked content",
				RenderTemplate: true,
				Transfer: types.TransferConfig{
					Encoding:  "chunked",
					ChunkSize: 10,
				},
			},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/upload"
				svc.On("PublicUrl", "https", "/upload").Return(reqUrl, nil)
				fnd.On("Sleep", ctx, mock.AnythingOfType("time.Duration")).Return(nil).Maybe()
				body := &bodyReader{msg: "ok"}
				header := http.Header{}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == "POST" &&
						req.URL.String() == reqUrl &&
						req.Header.Get("content-type") == "application/octet-stream" &&
						len(req.TransferEncoding) > 0 &&
						req.TransferEncoding[0] == "chunked" &&
						req.Body != nil
				})).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "ok",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with content-length override",
			id:         "r1",
			scheme:     "https",
			path:       "/upload",
			encodePath: true,
			method:     "POST",
			headers:    types.Headers{},
			body: &types.RequestBody{
				Content: "test content",
				Transfer: types.TransferConfig{
					Encoding:      "none",
					ContentLength: 100,
				},
			},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/upload"
				svc.On("PublicUrl", "https", "/upload").Return(reqUrl, nil)
				body := &bodyReader{msg: "ok"}
				header := http.Header{}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == "POST" &&
						req.URL.String() == reqUrl &&
						req.ContentLength == 100 &&
						req.Body != nil
				})).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "ok",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with chunked encoding but no size or delay",
			id:         "r1",
			scheme:     "https",
			path:       "/upload",
			encodePath: true,
			method:     "POST",
			headers:    types.Headers{},
			body: &types.RequestBody{
				Content: "test content",
				Transfer: types.TransferConfig{
					Encoding: "chunked",
				},
			},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/upload"
				svc.On("PublicUrl", "https", "/upload").Return(reqUrl, nil)
				body := &bodyReader{msg: "ok"}
				header := http.Header{}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", mock.MatchedBy(func(req *http.Request) bool {
					// Verify chunked encoding is set even without size/delay
					return req.Method == "POST" &&
						req.URL.String() == reqUrl &&
						len(req.TransferEncoding) > 0 &&
						req.TransferEncoding[0] == "chunked" &&
						req.Body != nil
				})).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "ok",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with HTTPS and skip verify - HTTP/2",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11, ProtocolHTTP2},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/test"
				svc.On("PublicUrl", "https", "/test").Return(reqUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
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

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						p.SetHTTP2(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "test",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with HTTP/2 only over HTTPS",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "POST",
			headers: types.Headers{
				"content-type": "application/json",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP2},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/test"
				svc.On("PublicUrl", "https", "/test").Return(reqUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "POST", reqUrl, nil)
				assert.Nil(t, err)
				expectedRequest.Header.Add("content-type", "application/json")
				body := &bodyReader{msg: "response"}
				header := http.Header{}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP2(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "response",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with HTTP/2 cleartext (h2c)",
			id:         "r1",
			scheme:     "http",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: false,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11, ProtocolHTTP2},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "http://example.com/test"
				svc.On("PublicUrl", "http", "/test").Return(reqUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
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

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						p.SetUnencryptedHTTP2(true)
						return p
					}(),
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "test",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with HTTP/2 only cleartext (h2c)",
			id:         "r2",
			scheme:     "http",
			path:       "/api/data",
			encodePath: true,
			method:     "PUT",
			headers:    types.Headers{},
			body:       &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: false,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP2},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "http://example.com/api/data"
				svc.On("PublicUrl", "http", "/api/data").Return(reqUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "PUT", reqUrl, nil)
				assert.Nil(t, err)
				body := &bodyReader{msg: "ok"}
				header := http.Header{}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetUnencryptedHTTP2(true)
						return p
					}(),
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r2", ResponseData{
					Body:    "ok",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with HTTPS and CA cert",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: false,
				CACert:     "localhost-ca",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/test"
				svc.On("PublicUrl", "https", "/test").Return(reqUrl, nil)

				mockCert := certificatesMocks.NewMockCertificate(t)
				mockCert.On("CertificateData").Return("valid")
				renderedCert := &certificates.RenderedCertificate{Certificate: mockCert}
				svc.On("FindCertificate", "localhost-ca").Return(renderedCert, nil)

				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
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

				certPool := &x509.CertPool{}
				caCertPool := appMocks.NewMockX509CertPool(t)
				caCertPool.On("AppendCertFromPEM", "valid").Return(true)
				caCertPool.On("CertPool").Return(certPool)
				fnd.On("X509CertPool").Return(caCertPool)

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: false,
						RootCAs:            certPool,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "test",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution with HTTPS, CA cert and HTTP/2",
			id:         "r3",
			scheme:     "https",
			path:       "/secure",
			encodePath: true,
			method:     "GET",
			headers:    types.Headers{},
			body:       &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: false,
				CACert:     "my-ca",
			},
			protocols: []Protocol{ProtocolHTTP11, ProtocolHTTP2},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://secure.example.com/secure"
				svc.On("PublicUrl", "https", "/secure").Return(reqUrl, nil)

				mockCert := certificatesMocks.NewMockCertificate(t)
				mockCert.On("CertificateData").Return("ca-data")
				renderedCert := &certificates.RenderedCertificate{Certificate: mockCert}
				svc.On("FindCertificate", "my-ca").Return(renderedCert, nil)

				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
				assert.Nil(t, err)
				body := &bodyReader{msg: "secure data"}
				header := http.Header{}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}

				certPool := &x509.CertPool{}
				caCertPool := appMocks.NewMockX509CertPool(t)
				caCertPool.On("AppendCertFromPEM", "ca-data").Return(true)
				caCertPool.On("CertPool").Return(certPool)
				fnd.On("X509CertPool").Return(caCertPool)

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						p.SetHTTP2(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: false,
						RootCAs:            certPool,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r3", ResponseData{
					Body:    "secure data",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution without path encoding - HTTPS",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: false,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				publicUrl := "https://example.com/test"
				svc.On("PublicUrl", "https", "/test").Return(publicUrl, nil)
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

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r1", ResponseData{
					Body:    "test",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "successful execution without path encoding - HTTP with h2c",
			id:         "r4",
			scheme:     "http",
			path:       "/unencoded/path",
			encodePath: false,
			method:     "DELETE",
			headers:    types.Headers{},
			body:       &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: false,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP2},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				publicUrl := "http://example.com/unencoded/path"
				svc.On("PublicUrl", "http", "/unencoded/path").Return(publicUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "DELETE", publicUrl, nil)
				assert.Nil(t, err)
				expectedRequest.URL = &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Opaque: "//example.com/unencoded/path",
				}
				body := &bodyReader{msg: "deleted"}
				header := http.Header{}
				resp := &http.Response{
					Body:   body,
					Header: header,
				}

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetUnencryptedHTTP2(true)
						return p
					}(),
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
				rd.On("Store", "response/r4", ResponseData{
					Body:    "deleted",
					Headers: header,
				}).Return(nil)
			},
			want: true,
		},
		{
			name:       "failed execution due to CA certificate not found",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: false,
				CACert:     "nonexistent-ca",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				svc.On("FindCertificate", "nonexistent-ca").Return(nil, errors.New("cert not found"))
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "CA certificate nonexistent-ca not found",
		},
		{
			name:       "failed execution due to invalid CA certificate",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: false,
				CACert:     "invalid-ca",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				mockCert := certificatesMocks.NewMockCertificate(t)
				mockCert.On("CertificateData").Return("invalid certificate data")
				renderedCert := &certificates.RenderedCertificate{Certificate: mockCert}
				svc.On("FindCertificate", "invalid-ca").Return(renderedCert, nil)

				caCertPool := appMocks.NewMockX509CertPool(t)
				caCertPool.On("AppendCertFromPEM", "invalid certificate data").Return(false)
				fnd.On("X509CertPool").Return(caCertPool)
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "failed to parse CA certificate",
		},
		{
			name:       "failed execution due to failed storing",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/test"
				svc.On("PublicUrl", "https", "/test").Return(reqUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
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

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
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
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/test"
				svc.On("PublicUrl", "https", "/test").Return(reqUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
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

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(resp, nil)
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "failed read",
		},
		{
			name:       "failed execution due to context being done",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/test"
				svc.On("PublicUrl", "https", "/test").Return(reqUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
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

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
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
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/test"
				svc.On("PublicUrl", "https", "/test").Return(reqUrl, nil)
				expectedRequest, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
				assert.Nil(t, err)
				expectedRequest.Header.Add("content-type", "application/json")
				expectedRequest.Header.Add("user-agent", "wst")

				expectedTransport := &http.Transport{
					Protocols: func() *http.Protocols {
						p := new(http.Protocols)
						p.SetHTTP1(true)
						return p
					}(),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := appMocks.NewMockHttpClient(t)
				fnd.On("HttpClient", expectedTransport).Return(client)
				client.On("Do", expectedRequest).Return(nil, errors.New("client fail"))
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "client fail",
		},
		{
			name:       "failed execution due to invalid request",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "=",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				reqUrl := "https://example.com/test"
				svc.On("PublicUrl", "https", "/test").Return(reqUrl, nil)
			},
			want:             false,
			expectError:      true,
			expectedErrorMsg: "invalid method",
		},
		{
			name:       "failed execution due to public url failing",
			id:         "r1",
			scheme:     "https",
			path:       "/test",
			encodePath: true,
			method:     "GET",
			headers: types.Headers{
				"content-type": "application/json",
				"user-agent":   "wst",
			},
			body: &types.RequestBody{},
			tls: &types.TLSClientConfig{
				SkipVerify: true,
				CACert:     "",
			},
			protocols: []Protocol{ProtocolHTTP11},
			setupMocks: func(
				t *testing.T,
				ctx context.Context,
				rd *runtimeMocks.MockData,
				fnd *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
			) {
				svc.On("PublicUrl", "https", "/test").Return("", errors.New("pub url"))
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
				scheme:     tt.scheme,
				path:       tt.path,
				encodePath: tt.encodePath,
				method:     tt.method,
				headers:    tt.headers,
				body:       tt.body,
				tls:        tt.tls,
				protocols:  tt.protocols,
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
