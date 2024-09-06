package expectations

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/metrics"
	"testing"
)

func Test_nativeMaker_MakeMetricsExpectation(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.MetricsExpectation
		expectError bool
		expected    *MetricsExpectation
	}{
		{
			name: "valid configuration",
			config: &types.MetricsExpectation{
				Id: "test-id",
				Rules: []types.MetricRule{
					{
						Metric:   "cpu",
						Operator: "gt",
						Value:    75.0,
					},
				},
			},
			expectError: false,
			expected: &MetricsExpectation{
				Id: "test-id",
				Rules: []MetricRule{
					{
						Metric:   "cpu",
						Operator: metrics.MetricGtOperator,
						Value:    75.0,
					},
				},
			},
		},
		{
			name: "invalid operator",
			config: &types.MetricsExpectation{
				Id: "test-id",
				Rules: []types.MetricRule{
					{
						Metric:   "memory",
						Operator: "unknown", // Assuming this is invalid
						Value:    50.0,
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maker := &nativeMaker{}
			result, err := maker.MakeMetricsExpectation(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
