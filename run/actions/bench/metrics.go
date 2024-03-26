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
	"golang.org/x/exp/constraints"
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

type Metric interface {
	Compare(operator MetricOperator, value any) (bool, error)
}

// NumericOrDuration is a constraint that limits the types to numeric types or time.Duration.
type NumericOrDuration interface {
	constraints.Integer | constraints.Float | time.Duration
}

// GenericMetric is a generic struct for metrics that can handle numeric values and time.Duration.
type GenericMetric[T NumericOrDuration] struct {
	Value T
}

// Compare performs a comparison operation between the GenericMetric's value and another value.
func (g GenericMetric[T]) Compare(operator MetricOperator, value any) (bool, error) {
	typedValue, ok := value.(T)
	if !ok {
		return false, fmt.Errorf("invalid metric type %t", value)
	}
	switch operator {
	case MetricEqOperator:
		return g.Value == typedValue, nil
	case MetricGtOperator:
		return g.Value > typedValue, nil
	case MetricGeOperator:
		return g.Value >= typedValue, nil
	case MetricLtOperator:
		return g.Value < typedValue, nil
	case MetricLeOperator:
		return g.Value <= typedValue, nil
	default:
		return false, fmt.Errorf("invalid metric operator %s", operator)
	}
}

type vegataMetrics struct {
	metrics vegeta.Metrics
}

func (vm *vegataMetrics) Find(name string) (Metric, error) {
	switch name {
	case "Requests":
		return GenericMetric[uint64]{Value: vm.metrics.Requests}, nil
	case "Rate":
		return GenericMetric[float64]{Value: vm.metrics.Rate}, nil
	case "Throughput":
		return GenericMetric[float64]{Value: vm.metrics.Throughput}, nil
	case "Duration":
		return GenericMetric[time.Duration]{Value: vm.metrics.Duration}, nil
	case "Success":
		return GenericMetric[float64]{Value: vm.metrics.Success}, nil
	case "LatencyTotal":
		return GenericMetric[time.Duration]{Value: vm.metrics.Latencies.Total}, nil
	case "LatencyMean":
		return GenericMetric[time.Duration]{Value: vm.metrics.Latencies.Mean}, nil
	case "LatencyP50":
		return GenericMetric[time.Duration]{Value: vm.metrics.Latencies.P50}, nil
	case "LatencyP90":
		return GenericMetric[time.Duration]{Value: vm.metrics.Latencies.P90}, nil
	case "LatencyP95":
		return GenericMetric[time.Duration]{Value: vm.metrics.Latencies.P95}, nil
	case "LatencyP99":
		return GenericMetric[time.Duration]{Value: vm.metrics.Latencies.P99}, nil
	case "LatencyMax":
		return GenericMetric[time.Duration]{Value: vm.metrics.Latencies.Max}, nil
	case "LatencyMin":
		return GenericMetric[time.Duration]{Value: vm.metrics.Latencies.Min}, nil
	default:
		return nil, fmt.Errorf("metric %s not found", name)
	}
}
