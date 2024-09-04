package expect

import (
	"bufio"
	"context"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	expectationsMocks "github.com/bukka/wst/mocks/generated/run/expectations"
	runtimeMocks "github.com/bukka/wst/mocks/generated/run/instances/runtime"
	parametersMocks "github.com/bukka/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	servicesMocks "github.com/bukka/wst/mocks/generated/run/services"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/environments/environment/output"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/parameters"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestExpectationActionMaker_MakeOutputAction(t *testing.T) {
	tests := []struct {
		name           string
		config         *types.OutputExpectationAction
		defaultTimeout int
		setupMocks     func(
			*testing.T,
			*servicesMocks.MockServiceLocator,
			*servicesMocks.MockService,
			*expectationsMocks.MockMaker,
			*types.OutputExpectationAction,
		) *expectations.OutputExpectation
		getExpectedAction func(
			*appMocks.MockFoundation,
			*servicesMocks.MockService,
			*expectations.OutputExpectation,
		) *outputAction
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful output action creation",
			config: &types.OutputExpectationAction{
				Service: "validService",
				When:    "on_success",
				Output: types.OutputExpectation{
					Order:          "fixed",
					Match:          "exact",
					Type:           "any",
					RenderTemplate: true,
					Messages:       []string{"test"},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(
				t *testing.T,
				sl *servicesMocks.MockServiceLocator,
				svc *servicesMocks.MockService,
				expectationMaker *expectationsMocks.MockMaker,
				config *types.OutputExpectationAction,
			) *expectations.OutputExpectation {
				sl.On("Find", "validService").Return(svc, nil)
				outputExpectation := &expectations.OutputExpectation{
					OrderType:      expectations.OrderTypeFixed,
					MatchType:      expectations.MatchTypeExact,
					OutputType:     expectations.OutputTypeAny,
					RenderTemplate: true,
					Messages:       []string{"test"},
				}
				expectationMaker.On("MakeOutputExpectation", &config.Output).Return(outputExpectation, nil)
				return outputExpectation
			},
			getExpectedAction: func(
				fndMock *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				expectation *expectations.OutputExpectation,
			) *outputAction {
				return &outputAction{
					CommonExpectation: &CommonExpectation{
						fnd:     fndMock,
						service: svc,
						timeout: 5000 * 1e6,
						when:    action.OnSuccess,
					},
					OutputExpectation: expectation,
					parameters:        parameters.Parameters{},
				}
			},
		},
		{
			name: "failed output action creation because no service found",
			config: &types.OutputExpectationAction{
				Service: "invalidService",
				When:    "on_success",
				Output: types.OutputExpectation{
					Order:          "fixed",
					Match:          "exact",
					Type:           "any",
					RenderTemplate: true,
					Messages:       []string{"test"},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(
				t *testing.T,
				sl *servicesMocks.MockServiceLocator,
				svc *servicesMocks.MockService,
				expectationMaker *expectationsMocks.MockMaker,
				config *types.OutputExpectationAction,
			) *expectations.OutputExpectation {
				sl.On("Find", "invalidService").Return(nil, errors.New("svc not found"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "svc not found",
		},
		{
			name: "failed output action creation because output expectation creation failed",
			config: &types.OutputExpectationAction{
				Service: "validService",
				When:    "on_success",
				Output: types.OutputExpectation{
					Order:          "fixed",
					Match:          "exact",
					Type:           "any",
					RenderTemplate: true,
					Messages:       []string{"test"},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(
				t *testing.T,
				sl *servicesMocks.MockServiceLocator,
				svc *servicesMocks.MockService,
				expectationMaker *expectationsMocks.MockMaker,
				config *types.OutputExpectationAction,
			) *expectations.OutputExpectation {
				sl.On("Find", "validService").Return(svc, nil)
				expectationMaker.On("MakeOutputExpectation", &config.Output).Return(nil, errors.New("output failed"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "output failed",
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

			got, err := m.MakeOutputAction(tt.config, slMock, tt.defaultTimeout)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				actualAction, ok := got.(*outputAction)
				assert.True(t, ok)
				expectedAction := tt.getExpectedAction(fndMock, svcMock, outputExpectation)
				assert.Equal(t, expectedAction, actualAction)
			}
		})
	}
}

type errorReader struct {
	msg        string
	first_read bool
}

func (e *errorReader) Read(p []byte) (int, error) {
	if e.first_read {
		// Fill the buffer with some data (e.g., "data")
		data := []byte("data")

		// Copy data into the provided buffer `p`
		n := copy(p, data)

		// Mark first_read as false, indicating the first read has been completed
		e.first_read = false

		// Return the number of bytes written to the buffer and no error
		return n, nil
	}

	// On subsequent reads, return an error
	return 0, errors.New(e.msg)
}

func Test_outputAction_Execute(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(
			t *testing.T,
			fnd *appMocks.MockFoundation,
			ctx context.Context,
			svc *servicesMocks.MockService,
			params parameters.Parameters,
			outputType output.Type,
		)
		expectation      *expectations.OutputExpectation
		outputType       output.Type
		want             bool
		expectErr        bool
		expectedErrorMsg string
	}{
		{
			name: "successful output fixed exact match with template rendering",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
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
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeFixed,
				MatchType:      expectations.MatchTypeExact,
				OutputType:     expectations.OutputTypeAny,
				RenderTemplate: true,
				Messages:       []string{"test"},
			},
			outputType: output.Any,
			want:       true,
		},
		{
			name: "successful output fixed exact match without template rendering",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				r := strings.NewReader("test")
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeFixed,
				MatchType:      expectations.MatchTypeExact,
				OutputType:     expectations.OutputTypeStdout,
				RenderTemplate: false,
				Messages:       []string{"test"},
			},
			outputType: output.Stdout,
			want:       true,
		},
		{
			name: "successful output random regexp match without template rendering",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				r := strings.NewReader("test")
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeRandom,
				MatchType:      expectations.MatchTypeRegexp,
				OutputType:     expectations.OutputTypeStderr,
				RenderTemplate: false,
				Messages:       []string{"te.*"},
			},
			outputType: output.Stderr,
			want:       true,
		},
		{
			name: "successful output ordered regexp match without template rendering",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				r := strings.NewReader("[01-Sep-2024 19:13:14] NOTICE: fpm is running, pid 174924")
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeRandom,
				MatchType:      expectations.MatchTypeRegexp,
				OutputType:     expectations.OutputTypeStderr,
				RenderTemplate: false,
				Messages: []string{
					"\\[.*\\] NOTICE: fpm is running, pid 174924",
				},
			},
			outputType: output.Stderr,
			want:       true,
		},
		{
			name: "successful output ordered regexp match without template rendering",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				r := strings.NewReader("2024/09/01 18:33:51 [notice] 164024#164024: start worker process 16402")
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeRandom,
				MatchType:      expectations.MatchTypeRegexp,
				OutputType:     expectations.OutputTypeStderr,
				RenderTemplate: false,
				Messages: []string{
					"\\d{4}/\\d{2}/\\d{2} \\d{2}:\\d{2}:\\d{2} \\[notice\\] \\d+#\\d+: start worker process \\d+",
				},
			},
			outputType: output.Stderr,
			want:       true,
		},
		{
			name: "successful no match without dry run",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				fnd.On("DryRun").Return(false)
				r := strings.NewReader("test")
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeFixed,
				MatchType:      expectations.MatchTypeExact,
				OutputType:     expectations.OutputTypeStdout,
				RenderTemplate: false,
				Messages:       []string{"what"},
			},
			outputType: output.Stdout,
			want:       false,
		},
		{
			name: "successful no match with dry run",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				fnd.On("DryRun").Return(true)
				r := strings.NewReader("test")
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeFixed,
				MatchType:      expectations.MatchTypeExact,
				OutputType:     expectations.OutputTypeStdout,
				RenderTemplate: false,
				Messages:       []string{"what"},
			},
			outputType: output.Stdout,
			want:       true,
		},
		{
			name: "successful false due scanner context deadline error",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				errReader := &errorReader{msg: "context deadline exceeded"}
				scanner := bufio.NewScanner(errReader)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeRandom,
				MatchType:      expectations.MatchTypeRegexp,
				OutputType:     expectations.OutputTypeStderr,
				RenderTemplate: false,
				Messages:       []string{"te.*"},
			},
			outputType: output.Stderr,
			want:       false,
			expectErr:  false,
		},
		{
			name: "failed match due to scanner internal error",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				errReader := &errorReader{msg: "scanner internal error", first_read: true}
				scanner := bufio.NewScanner(errReader)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeRandom,
				MatchType:      expectations.MatchTypeRegexp,
				OutputType:     expectations.OutputTypeStderr,
				RenderTemplate: false,
				Messages:       []string{"te.*"},
			},
			outputType:       output.Stderr,
			expectErr:        true,
			expectedErrorMsg: "scanner internal error",
		},
		{
			name: "failed due to invalid fixed match",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				r := strings.NewReader("test")
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeFixed,
				MatchType:      expectations.MatchType("x"),
				OutputType:     expectations.OutputTypeStdout,
				RenderTemplate: false,
				Messages:       []string{"what"},
			},
			outputType:       output.Stdout,
			expectErr:        true,
			expectedErrorMsg: "unknown match type x",
		},
		{
			name: "failed due to invalid random match",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				r := strings.NewReader("test")
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeRandom,
				MatchType:      expectations.MatchType("x"),
				OutputType:     expectations.OutputTypeStdout,
				RenderTemplate: false,
				Messages:       []string{"what"},
			},
			outputType:       output.Stdout,
			expectErr:        true,
			expectedErrorMsg: "unknown match type x",
		},
		{
			name: "failed due to invalid order",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				r := strings.NewReader("test")
				scanner := bufio.NewScanner(r)
				svc.On("OutputScanner", ctx, outputType).Return(scanner, nil)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderType("y"),
				MatchType:      expectations.MatchTypeExact,
				OutputType:     expectations.OutputTypeStdout,
				RenderTemplate: false,
				Messages:       []string{"what"},
			},
			outputType:       output.Stdout,
			expectErr:        true,
			expectedErrorMsg: "unknown order type y",
		},
		{
			name: "failed due to invalid scanner",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				svc.On("OutputScanner", ctx, outputType).Return(nil, errors.New("invalid scanner"))
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeFixed,
				MatchType:      expectations.MatchTypeExact,
				OutputType:     expectations.OutputTypeStdout,
				RenderTemplate: false,
				Messages:       []string{"what"},
			},
			outputType:       output.Stdout,
			expectErr:        true,
			expectedErrorMsg: "invalid scanner",
		},
		{
			name: "failed due to invalid message rendering",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				svc.On("RenderTemplate", "test", params).Return("", errors.New("render err"))
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeFixed,
				MatchType:      expectations.MatchTypeExact,
				OutputType:     expectations.OutputTypeStdout,
				RenderTemplate: true,
				Messages:       []string{"test"},
			},
			outputType:       output.Stdout,
			expectErr:        true,
			expectedErrorMsg: "render err",
		},
		{
			name: "failed due to invalid output type",
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				svc *servicesMocks.MockService,
				params parameters.Parameters,
				outputType output.Type,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
			},
			expectation: &expectations.OutputExpectation{
				OrderType:      expectations.OrderTypeFixed,
				MatchType:      expectations.MatchTypeExact,
				OutputType:     expectations.OutputType("z"),
				RenderTemplate: true,
				Messages:       []string{"test"},
			},
			outputType:       output.Stdout,
			expectErr:        true,
			expectedErrorMsg: "unknown output type z",
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

			tt.setupMocks(t, fndMock, ctx, svcMock, params, tt.outputType)

			a := &outputAction{
				CommonExpectation: &CommonExpectation{
					fnd:     fndMock,
					service: svcMock,
					timeout: 20 * 1e6,
				},
				OutputExpectation: tt.expectation,
				parameters:        params,
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

func Test_outputAction_Timeout(t *testing.T) {
	timeout := time.Duration(50 * 1e6)
	a := &outputAction{
		CommonExpectation: &CommonExpectation{
			fnd:     nil,
			service: nil,
			timeout: timeout,
		},
		OutputExpectation: &expectations.OutputExpectation{
			OrderType:      expectations.OrderTypeFixed,
			MatchType:      expectations.MatchTypeExact,
			OutputType:     expectations.OutputTypeAny,
			RenderTemplate: true,
			Messages:       []string{"test"},
		},
	}
	assert.Equal(t, timeout, a.Timeout())
}

func Test_outputAction_When(t *testing.T) {
	fndMock := appMocks.NewMockFoundation(t)
	a := &outputAction{
		CommonExpectation: &CommonExpectation{
			fnd:     fndMock,
			service: nil,
			when:    action.OnSuccess,
		},
	}
	assert.Equal(t, action.OnSuccess, a.When())
}
