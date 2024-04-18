package bench

import (
	"context"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	appMocks "github.com/bukka/wst/mocks/app"
	"github.com/bukka/wst/mocks/external"
	runtimeMocks "github.com/bukka/wst/mocks/run/instances/runtime"
	servicesMocks "github.com/bukka/wst/mocks/run/services"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/services"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"reflect"
	"testing"
	"time"
)

func TestActionMaker_Make(t *testing.T) {
	type fields struct {
		fnd app.Foundation
	}
	type args struct {
		config         *types.BenchAction
		sl             services.ServiceLocator
		defaultTimeout int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    action.Action
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ActionMaker{
				fnd: tt.fields.fnd,
			}
			got, err := m.Make(tt.args.config, tt.args.sl, tt.args.defaultTimeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("Make() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Make() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAction_Execute(t *testing.T) {
	serviceName := "svc"
	duration := time.Duration(5000000)
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
					results <- result
				}()
				vegetaAttackerMock.On(
					"Attack",
					mock.MatchedBy(func(targeter vegeta.Targeter) bool {
						// We are checking the type using the argument type
						return true
					}),
					duration,
					serviceName,
				).Return(results)
				vegetaMetricsMock := appMocks.NewMockVegetaMetrics(t)
				vegetaMetricsMock.On("Add", result).Return()
				vegetaMetricsMock.On("Close").Return()
				fnd.On("VegetaAttacker").Return(vegetaAttackerMock)
				fnd.On("VegetaMetrics").Return(vegetaMetricsMock)
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
				fnd.On("VegetaAttacker").Return(vegetaAttackerMock) // Simulate attack results
				svc.On("PublicUrl", mock.Anything).Return("http://example.com", nil)
				runData.On("Store", mock.Anything, mock.Anything).Return(nil)
			},
			contextSetup: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Immediately cancel the context
				return ctx
			},
			expectSuccess: false,
			expectErr:     true,
			expectedErr:   "context canceled",
		},
		{
			name: "error fetching public URL",
			setupMocks: func(t *testing.T, fnd *appMocks.MockFoundation, svc *servicesMocks.MockService, runData *runtimeMocks.MockData) {
				vegetaAttackerMock := appMocks.NewMockVegetaAttacker(t)
				fnd.On("VegetaAttacker").Return(vegetaAttackerMock)
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

			svcMock.On("Name").Return(serviceName)
			fndMock.On("Logger").Return(mockLogger.SugaredLogger)

			tt.setupMocks(t, fndMock, svcMock, runDataMock)

			action := &Action{
				fnd:      fndMock,
				service:  svcMock,
				duration: duration,
				id:       "sid",
			}

			ctx := tt.contextSetup()
			success, err := action.Execute(ctx, runDataMock)

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
	type fields struct {
		fnd      app.Foundation
		service  services.Service
		timeout  time.Duration
		duration time.Duration
		freq     int
		id       string
		path     string
		method   string
		headers  types.Headers
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
			a := &Action{
				fnd:      tt.fields.fnd,
				service:  tt.fields.service,
				timeout:  tt.fields.timeout,
				duration: tt.fields.duration,
				freq:     tt.fields.freq,
				id:       tt.fields.id,
				path:     tt.fields.path,
				method:   tt.fields.method,
				headers:  tt.fields.headers,
			}
			if got := a.Timeout(); got != tt.want {
				t.Errorf("Timeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateActionMaker(t *testing.T) {
	type args struct {
		fnd app.Foundation
	}
	tests := []struct {
		name string
		args args
		want *ActionMaker
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateActionMaker(tt.args.fnd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateActionMaker() = %v, want %v", got, tt.want)
			}
		})
	}
}
