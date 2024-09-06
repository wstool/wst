package bench

import (
	"github.com/stretchr/testify/assert"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	appMocks "github.com/wstool/wst/mocks/generated/app"
	"github.com/wstool/wst/run/metrics"
	"testing"
	"time"
)

func TestMetrics_Find(t *testing.T) {
	vm := &vegeta.Metrics{
		Latencies: vegeta.LatencyMetrics{
			Total: time.Millisecond * 100,
			Mean:  time.Millisecond * 10,
			P50:   time.Millisecond * 5,
			P90:   time.Millisecond * 15,
			P95:   time.Millisecond * 20,
			P99:   time.Millisecond * 30,
			Max:   time.Millisecond * 40,
			Min:   time.Millisecond * 2,
		},
		Duration:   5 * time.Second,
		Requests:   100,
		Rate:       25.0,
		Throughput: 20.0,
		Success:    0.95,
	}

	tests := []struct {
		name             string
		metricName       string
		expectedMetric   metrics.Metric
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:           "Requests test",
			metricName:     "Requests",
			expectedMetric: metrics.GenericMetric[uint64]{Value: 100},
		},
		{
			name:           "Rate test",
			metricName:     "Rate",
			expectedMetric: metrics.GenericMetric[float64]{Value: 25.0},
		},
		{
			name:           "Throughput test",
			metricName:     "Throughput",
			expectedMetric: metrics.GenericMetric[float64]{Value: 20.0},
		},
		{
			name:           "Duration test",
			metricName:     "Duration",
			expectedMetric: metrics.GenericMetric[time.Duration]{Value: 5 * time.Second},
		},
		{
			name:           "Success test",
			metricName:     "Success",
			expectedMetric: metrics.GenericMetric[float64]{Value: 0.95},
		},
		{
			name:           "Latency Total test",
			metricName:     "LatencyTotal",
			expectedMetric: metrics.GenericMetric[time.Duration]{Value: time.Millisecond * 100},
		},
		{
			name:           "Latency Mean test",
			metricName:     "LatencyMean",
			expectedMetric: metrics.GenericMetric[time.Duration]{Value: time.Millisecond * 10},
		},
		{
			name:           "Latency P50 test",
			metricName:     "LatencyP50",
			expectedMetric: metrics.GenericMetric[time.Duration]{Value: time.Millisecond * 5},
		},
		{
			name:           "Latency P90 test",
			metricName:     "LatencyP90",
			expectedMetric: metrics.GenericMetric[time.Duration]{Value: time.Millisecond * 15},
		},
		{
			name:           "Latency P95 test",
			metricName:     "LatencyP95",
			expectedMetric: metrics.GenericMetric[time.Duration]{Value: time.Millisecond * 20},
		},
		{
			name:           "Latency P99 test",
			metricName:     "LatencyP99",
			expectedMetric: metrics.GenericMetric[time.Duration]{Value: time.Millisecond * 30},
		},
		{
			name:           "Latency Max test",
			metricName:     "LatencyMax",
			expectedMetric: metrics.GenericMetric[time.Duration]{Value: time.Millisecond * 40},
		},
		{
			name:           "Latency Min test",
			metricName:     "LatencyMin",
			expectedMetric: metrics.GenericMetric[time.Duration]{Value: time.Millisecond * 2},
		},
		{
			name:             "Invalid metric test",
			metricName:       "InvalidMetric",
			expectedMetric:   nil,
			expectError:      true,
			expectedErrorMsg: "InvalidMetric not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricsMock := appMocks.NewMockVegetaMetrics(t)
			metricsMock.On("Metrics").Return(vm)
			m := Metrics{metrics: metricsMock}
			result, err := m.Find(tt.metricName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMetric, result)
			}
		})
	}
}
