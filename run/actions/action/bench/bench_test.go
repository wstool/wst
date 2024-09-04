package bench

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	runtimeMocks "github.com/bukka/wst/mocks/generated/run/instances/runtime"
	servicesMocks "github.com/bukka/wst/mocks/generated/run/services"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	vegeta "github.com/tsenart/vegeta/v12/lib"
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
		name             string
		config           *types.BenchAction
		defaultTimeout   int
		expectedTimeout  time.Duration
		expectedDuration time.Duration
		locatorErr       error
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful action creation with default timeout",
			config: &types.BenchAction{
				Service:   "validService",
				Duration:  3000,
				Frequency: 1,
				Id:        "testAction",
				Path:      "/test",
				Method:    "GET",
				Headers:   types.Headers{"Content-Type": "application/json"},
			},
			defaultTimeout:   5000,
			expectedTimeout:  5000 * time.Millisecond,
			expectedDuration: 3000 * time.Millisecond,
		},
		{
			name: "successful action creation with calculated timeout",
			config: &types.BenchAction{
				Service:   "validService",
				Duration:  6000,
				Frequency: 1,
				Id:        "testAction",
				Path:      "/test",
				Method:    "GET",
				Headers:   types.Headers{"Content-Type": "application/json"},
			},
			defaultTimeout:   5000,
			expectedTimeout:  11000 * time.Millisecond,
			expectedDuration: 6000 * time.Millisecond,
		},
		{
			name: "successful action creation with calculated duration",
			config: &types.BenchAction{
				Service:   "validService",
				Duration:  0,
				Timeout:   10000,
				Frequency: 1,
				Id:        "testAction",
				Path:      "/test",
				Method:    "GET",
				Headers:   types.Headers{"Content-Type": "application/json"},
			},
			defaultTimeout:   5000,
			expectedTimeout:  10000 * time.Millisecond,
			expectedDuration: 9900 * time.Millisecond,
		},
		{
			name: "successful action creation with zero timeout and duration",
			config: &types.BenchAction{
				Service:   "validService",
				Duration:  0,
				Timeout:   0,
				Frequency: 1,
				Id:        "testAction",
				Path:      "/test",
				Method:    "GET",
				Headers:   types.Headers{"Content-Type": "application/json"},
			},
			defaultTimeout:   0,
			expectedTimeout:  0,
			expectedDuration: 0,
		},
		{
			name:             "service locator failure",
			config:           &types.BenchAction{Service: "invalidService"},
			defaultTimeout:   5000,
			locatorErr:       errors.New("service not found"),
			expectError:      true,
			expectedErrorMsg: "service not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			m := &ActionMaker{
				fnd: fndMock,
			}
			svcMock := servicesMocks.NewMockService(t)
			slMock := servicesMocks.NewMockServiceLocator(t)
			if tt.locatorErr != nil {
				svcMock = nil
			}
			slMock.On("Find", tt.config.Service).Return(svcMock, tt.locatorErr)
			got, err := m.Make(tt.config, slMock, tt.defaultTimeout)
			assert := assert.New(t)
			if tt.expectError {
				assert.Error(err)
				assert.Nil(got)
				assert.Contains(err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(err)
				assert.NotNil(got)

				action, ok := got.(*Action)
				assert.True(ok)
				assert.Equal(svcMock, action.service)
				assert.Equal(tt.expectedTimeout, action.Timeout())
				assert.Equal(tt.expectedDuration, action.duration)
				assert.Equal(tt.config.Frequency, action.freq)
				assert.Equal(tt.config.Id, action.id)
				assert.Equal(tt.config.Path, action.path)
				assert.Equal(tt.config.Method, action.method)
				assert.Equal(tt.config.Headers, action.headers)
			}
		})
	}
}

func TestAction_Execute(t *testing.T) {
	serviceName := "svc"
	duration := time.Duration(5000000)
	freq := 30
	tests := []struct {
		name          string
		setupMocks    func(*testing.T, *appMocks.MockFoundation, *servicesMocks.MockService, *runtimeMocks.MockData)
		contextSetup  func() context.Context
		expectSuccess bool
		expectErr     bool
		expectedErr   string
	}{
		{
			name: "successful execution",
			setupMocks: func(t *testing.T, fnd *appMocks.MockFoundation, svc *servicesMocks.MockService, runData *runtimeMocks.MockData) {
				vegetaAttackerMock := appMocks.NewMockVegetaAttacker(t)
				result := &vegeta.Result{}
				results := make(chan *vegeta.Result)
				go func() {
					defer close(results)
					results <- result
				}()
				vegetaAttackerMock.On(
					"Attack",
					mock.MatchedBy(func(targeter vegeta.Targeter) bool {
						// We are checking the type using the argument type
						return true
					}),
					mock.MatchedBy(func(rate vegeta.Rate) bool {
						return assert.Equal(t, freq, rate.Freq) && assert.Equal(t, time.Second, rate.Per)
					}),
					duration,
					serviceName,
				).Return((<-chan *vegeta.Result)(results))
				vegetaMetricsMock := appMocks.NewMockVegetaMetrics(t)
				vegetaMetricsMock.On("Add", result).Return()
				vegetaMetricsMock.On("Close").Return()
				fnd.On("VegetaAttacker").Return(vegetaAttackerMock)
				fnd.On("VegetaMetrics").Return(vegetaMetricsMock)
				svc.On("Name").Return(serviceName)
				svc.On("PublicUrl", mock.Anything).Return("http://example.com", nil)
				runData.On("Store", "metrics/sid", mock.MatchedBy(func(metrics *Metrics) bool {
					return assert.Equal(t, vegetaMetricsMock, metrics.metrics)
				})).Return(nil)
			},
			contextSetup: func() context.Context {
				return context.Background()
			},
			expectSuccess: true,
			expectErr:     false,
		},
		{
			name: "execution with context cancellation",
			setupMocks: func(t *testing.T, fnd *appMocks.MockFoundation, svc *servicesMocks.MockService, runData *runtimeMocks.MockData) {
				vegetaAttackerMock := appMocks.NewMockVegetaAttacker(t)
				result := &vegeta.Result{}
				results := make(chan *vegeta.Result)
				go func() {
					defer close(results)
					results <- result
				}()
				vegetaAttackerMock.On(
					"Attack",
					mock.MatchedBy(func(targeter vegeta.Targeter) bool {
						// We are checking the type using the argument type
						return true
					}),
					mock.MatchedBy(func(rate vegeta.Rate) bool {
						return assert.Equal(t, freq, rate.Freq) && assert.Equal(t, time.Second, rate.Per)
					}),
					duration,
					serviceName,
				).Return((<-chan *vegeta.Result)(results))
				vegetaMetricsMock := appMocks.NewMockVegetaMetrics(t)
				vegetaMetricsMock.On("Add", result).Maybe().Return()
				vegetaMetricsMock.On("Close").Maybe().Return()
				fnd.On("VegetaAttacker").Return(vegetaAttackerMock)
				fnd.On("VegetaMetrics").Return(vegetaMetricsMock)
				svc.On("Name").Return(serviceName)
				svc.On("PublicUrl", mock.Anything).Return("http://example.com", nil)
				runData.On("Store", "metrics/sid", mock.MatchedBy(func(metrics *Metrics) bool {
					return assert.Equal(t, vegetaMetricsMock, metrics.metrics)
				})).Maybe().Return(nil)
			},
			contextSetup: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expectSuccess: false,
			expectErr:     true,
			expectedErr:   "context canceled",
		},
		{
			name: "execution with store failure",
			setupMocks: func(t *testing.T, fnd *appMocks.MockFoundation, svc *servicesMocks.MockService, runData *runtimeMocks.MockData) {
				vegetaAttackerMock := appMocks.NewMockVegetaAttacker(t)
				result := &vegeta.Result{}
				results := make(chan *vegeta.Result)
				go func() {
					defer close(results)
					results <- result
				}()
				vegetaAttackerMock.On(
					"Attack",
					mock.MatchedBy(func(targeter vegeta.Targeter) bool {
						// We are checking the type using the argument type
						return true
					}),
					mock.MatchedBy(func(rate vegeta.Rate) bool {
						return assert.Equal(t, freq, rate.Freq) && assert.Equal(t, time.Second, rate.Per)
					}),
					duration,
					serviceName,
				).Return((<-chan *vegeta.Result)(results))
				vegetaMetricsMock := appMocks.NewMockVegetaMetrics(t)
				vegetaMetricsMock.On("Add", result).Return()
				vegetaMetricsMock.On("Close").Return()
				fnd.On("VegetaAttacker").Return(vegetaAttackerMock)
				fnd.On("VegetaMetrics").Return(vegetaMetricsMock)
				svc.On("Name").Return(serviceName)
				svc.On("PublicUrl", mock.Anything).Return("http://example.com", nil)
				runData.On("Store", "metrics/sid", mock.MatchedBy(func(metrics *Metrics) bool {
					return assert.Equal(t, vegetaMetricsMock, metrics.metrics)
				})).Return(errors.New("stored failed"))
			},
			contextSetup: func() context.Context {
				return context.Background()
			},
			expectSuccess: false,
			expectErr:     true,
			expectedErr:   "stored failed",
		},
		{
			name: "error fetching public URL",
			setupMocks: func(t *testing.T, fnd *appMocks.MockFoundation, svc *servicesMocks.MockService, runData *runtimeMocks.MockData) {
				svc.On("PublicUrl", mock.Anything).Return("", errors.New("public url error"))
			},
			contextSetup: func() context.Context {
				return context.Background()
			},
			expectSuccess: false,
			expectErr:     true,
			expectedErr:   "public url error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			svcMock := servicesMocks.NewMockService(t)
			runDataMock := runtimeMocks.NewMockData(t)
			mockLogger := external.NewMockLogger()

			fndMock.On("Logger").Return(mockLogger.SugaredLogger)

			tt.setupMocks(t, fndMock, svcMock, runDataMock)

			a := &Action{
				fnd:      fndMock,
				service:  svcMock,
				duration: duration,
				id:       "sid",
				freq:     freq,
			}

			ctx := tt.contextSetup()
			success, err := a.Execute(ctx, runDataMock)

			if tt.expectSuccess {
				assert.True(t, success)
			}
			if tt.expectErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, fndMock, svcMock, runDataMock)
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
