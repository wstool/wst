package expect

import (
	"context"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/mocks/authored/external"
	appMocks "github.com/bukka/wst/mocks/generated/app"
	expectationsMocks "github.com/bukka/wst/mocks/generated/run/expectations"
	runtimeMocks "github.com/bukka/wst/mocks/generated/run/instances/runtime"
	metricsMocks "github.com/bukka/wst/mocks/generated/run/metrics"
	parametersMocks "github.com/bukka/wst/mocks/generated/run/parameters"
	parameterMocks "github.com/bukka/wst/mocks/generated/run/parameters/parameter"
	servicesMocks "github.com/bukka/wst/mocks/generated/run/services"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/expectations"
	"github.com/bukka/wst/run/metrics"
	"github.com/bukka/wst/run/parameters"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestExpectationActionMaker_MakeMetricsAction(t *testing.T) {
	tests := []struct {
		name           string
		config         *types.MetricsExpectationAction
		defaultTimeout int
		setupMocks     func(
			*testing.T,
			*servicesMocks.MockServiceLocator,
			*servicesMocks.MockService,
			*expectationsMocks.MockMaker,
			*types.MetricsExpectationAction,
		) *expectations.MetricsExpectation
		getExpectedAction func(
			*appMocks.MockFoundation,
			*servicesMocks.MockService,
			*expectations.MetricsExpectation,
		) *metricsAction
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful metrics action creation",
			config: &types.MetricsExpectationAction{
				Service: "validService",
				When:    "on_success",
				Metrics: types.MetricsExpectation{
					Id: "eid1",
					Rules: []types.MetricRule{
						{
							Metric:   "test",
							Operator: "eq",
							Value:    0,
						},
					},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(
				t *testing.T,
				sl *servicesMocks.MockServiceLocator,
				svc *servicesMocks.MockService,
				expectationMaker *expectationsMocks.MockMaker,
				config *types.MetricsExpectationAction,
			) *expectations.MetricsExpectation {
				sl.On("Find", "validService").Return(svc, nil)
				metricExpectation := &expectations.MetricsExpectation{
					Id: "test",
					Rules: []expectations.MetricRule{
						{
							Metric:   "test",
							Operator: metrics.MetricEqOperator,
							Value:    10.,
						},
					},
				}
				expectationMaker.On("MakeMetricsExpectation", &config.Metrics).Return(metricExpectation, nil)
				return metricExpectation
			},
			getExpectedAction: func(
				fndMock *appMocks.MockFoundation,
				svc *servicesMocks.MockService,
				expectation *expectations.MetricsExpectation,
			) *metricsAction {
				return &metricsAction{
					CommonExpectation: &CommonExpectation{
						fnd:     fndMock,
						service: svc,
						timeout: 5000 * 1e6,
						when:    action.OnSuccess,
					},
					MetricsExpectation: expectation,
					parameters:         parameters.Parameters{},
				}
			},
		},
		{
			name: "failed metrics action creation because no service found",
			config: &types.MetricsExpectationAction{
				Service: "invalidService",
				When:    "on_success",
				Metrics: types.MetricsExpectation{
					Id: "eid1",
					Rules: []types.MetricRule{
						{
							Metric:   "test",
							Operator: "eq",
							Value:    0,
						},
					},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(
				t *testing.T,
				sl *servicesMocks.MockServiceLocator,
				svc *servicesMocks.MockService,
				expectationMaker *expectationsMocks.MockMaker,
				config *types.MetricsExpectationAction,
			) *expectations.MetricsExpectation {
				sl.On("Find", "invalidService").Return(nil, errors.New("svc not found"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "svc not found",
		},
		{
			name: "failed metrics action creation because metrics expectation creation failed",
			config: &types.MetricsExpectationAction{
				Service: "validService",
				When:    "on_success",
				Metrics: types.MetricsExpectation{
					Id: "eid1",
					Rules: []types.MetricRule{
						{
							Metric:   "test",
							Operator: "eq",
							Value:    0,
						},
					},
				},
			},
			defaultTimeout: 5000,
			setupMocks: func(
				t *testing.T,
				sl *servicesMocks.MockServiceLocator,
				svc *servicesMocks.MockService,
				expectationMaker *expectationsMocks.MockMaker,
				config *types.MetricsExpectationAction,
			) *expectations.MetricsExpectation {
				sl.On("Find", "validService").Return(svc, nil)
				expectationMaker.On("MakeMetricsExpectation", &config.Metrics).Return(nil, errors.New("metrics failed"))
				return nil
			},
			expectError:      true,
			expectedErrorMsg: "metrics failed",
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

			metricsExpectation := tt.setupMocks(t, slMock, svcMock, expectationsMakerMock, tt.config)

			got, err := m.MakeMetricsAction(tt.config, slMock, tt.defaultTimeout)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				actualAction, ok := got.(*metricsAction)
				assert.True(t, ok)
				expectedAction := tt.getExpectedAction(fndMock, svcMock, metricsExpectation)
				assert.Equal(t, expectedAction, actualAction)
			}
		})
	}
}

func Test_metricsAction_Execute(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(
			*testing.T,
			*appMocks.MockFoundation,
			context.Context,
			*runtimeMocks.MockData,
		)
		rules            []expectations.MetricRule
		id               string
		want             bool
		expectErr        bool
		expectedErrorMsg string
	}{
		{
			name: "successful metrics comparison match",
			id:   "mid",
			rules: []expectations.MetricRule{
				{
					Metric:   "latency",
					Operator: metrics.MetricLtOperator,
					Value:    12.0,
				},
			},
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				metricsMock := metricsMocks.NewMockMetrics(t)
				metricMock := metricsMocks.NewMockMetric(t)
				rd.On("Load", "metrics/mid").Return(metricsMock, true)
				metricsMock.On("Find", "latency").Return(metricMock, nil)
				metricMock.On("Compare", metrics.MetricLtOperator, 12.0).Return(true, nil)
			},
			want: true,
		},
		{
			name: "successful metrics comparison not match dry run",
			id:   "mid",
			rules: []expectations.MetricRule{
				{
					Metric:   "latency",
					Operator: metrics.MetricLtOperator,
					Value:    12.0,
				},
			},
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(true)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				metricsMock := metricsMocks.NewMockMetrics(t)
				metricMock := metricsMocks.NewMockMetric(t)
				rd.On("Load", "metrics/mid").Return(metricsMock, true)
				metricsMock.On("Find", "latency").Return(metricMock, nil)
				metricMock.On("Compare", metrics.MetricLtOperator, 12.0).Return(false, nil)
			},
			want: true,
		},
		{
			name: "successful metrics comparison not match normal run",
			id:   "mid",
			rules: []expectations.MetricRule{
				{
					Metric:   "latency",
					Operator: metrics.MetricLtOperator,
					Value:    12.0,
				},
			},
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("DryRun").Return(false)
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				metricsMock := metricsMocks.NewMockMetrics(t)
				metricMock := metricsMocks.NewMockMetric(t)
				rd.On("Load", "metrics/mid").Return(metricsMock, true)
				metricsMock.On("Find", "latency").Return(metricMock, nil)
				metricMock.On("Compare", metrics.MetricLtOperator, 12.0).Return(false, nil)
			},
			want: false,
		},
		{
			name: "failed metrics comparison because of metric compare",
			id:   "mid",
			rules: []expectations.MetricRule{
				{
					Metric:   "latency",
					Operator: metrics.MetricLtOperator,
					Value:    12.0,
				},
			},
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				metricsMock := metricsMocks.NewMockMetrics(t)
				metricMock := metricsMocks.NewMockMetric(t)
				rd.On("Load", "metrics/mid").Return(metricsMock, true)
				metricsMock.On("Find", "latency").Return(metricMock, nil)
				metricMock.On("Compare", metrics.MetricLtOperator, 12.0).Return(false, errors.New("compare fail"))
			},
			want:             false,
			expectErr:        true,
			expectedErrorMsg: "compare fail",
		},
		{
			name: "failed metrics comparison because of metrics find fail",
			id:   "mid",
			rules: []expectations.MetricRule{
				{
					Metric:   "latency",
					Operator: metrics.MetricLtOperator,
					Value:    12.0,
				},
			},
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				metricsMock := metricsMocks.NewMockMetrics(t)
				rd.On("Load", "metrics/mid").Return(metricsMock, true)
				metricsMock.On("Find", "latency").Return(nil, errors.New("find fail"))
			},
			want:             false,
			expectErr:        true,
			expectedErrorMsg: "find fail",
		},

		{
			name: "failed metrics comparison because of invalid metrics",
			id:   "mid",
			rules: []expectations.MetricRule{
				{
					Metric:   "latency",
					Operator: metrics.MetricLtOperator,
					Value:    12.0,
				},
			},
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				rd.On("Load", "metrics/mid").Return("data", true)
			},
			want:             false,
			expectErr:        true,
			expectedErrorMsg: "invalid metrics data type",
		},

		{
			name: "failed metrics comparison because of no metrics",
			id:   "mid",
			rules: []expectations.MetricRule{
				{
					Metric:   "latency",
					Operator: metrics.MetricLtOperator,
					Value:    12.0,
				},
			},
			setupMocks: func(
				t *testing.T,
				fnd *appMocks.MockFoundation,
				ctx context.Context,
				rd *runtimeMocks.MockData,
			) {
				mockLogger := external.NewMockLogger()
				fnd.On("Logger").Return(mockLogger.SugaredLogger)
				rd.On("Load", "metrics/mid").Return(nil, false)
			},
			want:             false,
			expectErr:        true,
			expectedErrorMsg: "metrics data for key metrics/mid not found",
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

			tt.setupMocks(t, fndMock, ctx, dataMock)

			a := &metricsAction{
				CommonExpectation: &CommonExpectation{
					fnd:     fndMock,
					service: svcMock,
					timeout: 20 * 1e6,
				},
				MetricsExpectation: &expectations.MetricsExpectation{
					Id:    tt.id,
					Rules: tt.rules,
				},
				parameters: params,
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

func Test_metricsAction_Timeout(t *testing.T) {
	timeout := time.Duration(50 * 1e6)
	a := &metricsAction{
		CommonExpectation: &CommonExpectation{
			fnd:     nil,
			service: nil,
			timeout: timeout,
		},
		MetricsExpectation: &expectations.MetricsExpectation{
			Id: "test",
		},
	}
	assert.Equal(t, timeout, a.Timeout())
}
