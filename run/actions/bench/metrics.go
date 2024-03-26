// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bench

import (
	"fmt"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"time"
)

type MetricOperator string

const (
	MetricEqOperator MetricOperator = "eq"
	MetricGtOperator                = "gt"
	MetricGeOperator                = "ge"
	MetricLtOperator                = "lt"
	MetricLeOperator                = "le"
)

// Metric is a generic interface for metrics of different types.
type Metric interface {
	Compare(operator MetricOperator, value any) bool
}

// NumericMetric holds numeric values (float64 for simplification, but can be adapted).
type NumericMetric struct {
	Value float64
}

func (n NumericMetric) Compare(operator MetricOperator, value any) bool {
	val, ok := value.(float64)
	if !ok {
		return false
	}
	switch operator {
	case MetricEqOperator:
		return n.Value == val
	case MetricGtOperator:
		return n.Value > val
	case MetricGeOperator:
		return n.Value >= val
	case MetricLtOperator:
		return n.Value < val
	case MetricLeOperator:
		return n.Value <= val
	default:
		return false
	}
}

// TimeMetric holds time.Time values.
type TimeMetric struct {
	Value time.Time
}

func (t TimeMetric) Compare(operator MetricOperator, value any) bool {
	val, ok := value.(time.Time)
	if !ok {
		return false
	}
	switch operator {
	case MetricEqOperator:
		return t.Value.Equal(val)
	case MetricGtOperator:
		return t.Value.After(val)
	case MetricGeOperator:
		return t.Value.After(val) || t.Value.Equal(val)
	case MetricLtOperator:
		return t.Value.Before(val)
	case MetricLeOperator:
		return t.Value.Before(val) || t.Value.Equal(val)
	default:
		return false
	}
}

// DurationMetric holds time.Duration values.
type DurationMetric struct {
	Value time.Duration
}

func (d DurationMetric) Compare(operator MetricOperator, value any) bool {
	val, ok := value.(time.Duration)
	if !ok {
		return false
	}
	switch operator {
	case MetricEqOperator:
		return d.Value == val
	case MetricGtOperator:
		return d.Value > val
	case MetricGeOperator:
		return d.Value >= val
	case MetricLtOperator:
		return d.Value < val
	case MetricLeOperator:
		return d.Value <= val
	default:
		return false
	}
}

type vegataMetrics struct {
	metrics vegeta.Metrics
}

func (vm *vegataMetrics) Find(name string) (Metric, error) {
	switch name {
	case "latencies":
		return NumericMetric{Value: vm.metrics.Latencies.Mean}, nil
	case "bytes_in":
		return NumericMetric{Value: vm.metrics.BytesIn.Total}, nil
	case "bytes_out":
		return NumericMetric{Value: vm.metrics.BytesOut.Total}, nil
	case "earliest":
		return TimeMetric{Value: vm.metrics.Earliest}, nil
	case "latest":
		return TimeMetric{Value: vm.metrics.Latest}, nil
	case "end":
		return TimeMetric{Value: vm.metrics.End}, nil
	case "duration":
		return DurationMetric{Value: vm.metrics.Duration}, nil
	case "wait":
		return DurationMetric{Value: vm.metrics.Wait}, nil
	case "requests":
		return NumericMetric{Value: vm.metrics.Requests}, nil
	case "rate":
		return NumericMetric{Value: vm.metrics.Rate}, nil
	case "throughput":
		return NumericMetric{Value: vm.metrics.Throughput}, nil
	case "success":
		return NumericMetric{Value: vm.metrics.Success}, nil
	case "status_codes":
		return MapMetric{Value: vm.metrics.StatusCodes}, nil
	case "errors":
		return SliceMetric{Value: vm.metrics.Errors}, nil
	default:
		return nil, fmt.Errorf("metric %s not found", name)
	}
}
