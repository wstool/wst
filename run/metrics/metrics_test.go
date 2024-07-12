package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertToOperator(t *testing.T) {
	tests := []struct {
		name           string
		operator       string
		expectedOp     MetricOperator
		expectingError bool
	}{
		{"Valid Eq", "eq", MetricEqOperator, false},
		{"Valid Ne", "ne", MetricNeOperator, false},
		{"Valid Gt", "gt", MetricGtOperator, false},
		{"Valid Ge", "ge", MetricGeOperator, false},
		{"Valid Lt", "lt", MetricLtOperator, false},
		{"Valid Le", "le", MetricLeOperator, false},
		{"Invalid Operator", "unknown", "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op, err := ConvertToOperator(test.operator)
			if test.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedOp, op)
			}
		})
	}
}

func TestGenericMetric_Compare(t *testing.T) {
	tests := []struct {
		name           string
		value          float64
		operator       MetricOperator
		compareValue   float64
		expectedResult bool
		expectingError bool
	}{
		{"Equal True", 100, MetricEqOperator, 100, true, false},
		{"Equal False", 100, MetricEqOperator, 101, false, false},
		{"Greater Than True", 200, MetricGtOperator, 100, true, false},
		{"Greater Than False", 100, MetricGtOperator, 200, false, false},
		{"Greater Or Equal True", 100, MetricGeOperator, 100, true, false},
		{"Greater Or Equal True 2", 200, MetricGeOperator, 100, true, false},
		{"Less Than True", 100, MetricLtOperator, 200, true, false},
		{"Less Than False", 200, MetricLtOperator, 100, false, false},
		{"Less Or Equal True", 100, MetricLeOperator, 100, true, false},
		{"Less Or Equal True 2", 100, MetricLeOperator, 200, true, false},
		{"Invalid Operator", 100, "invalid", 100, false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metric := GenericMetric[float64]{Value: test.value}
			result, err := metric.Compare(test.operator, test.compareValue)
			if test.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResult, result)
			}
		})
	}
}
