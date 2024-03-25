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

// Metric is a generic interface for metrics of different types.
type Metric interface {
	String() string
	// Additional methods for comparison or manipulation can be added here
}

// NumericMetric holds numeric values (int, uint64, float64).
type NumericMetric struct {
	Value interface{}
}

func (n NumericMetric) String() string {
	return fmt.Sprintf("%v", n.Value)
}

// TimeMetric holds time.Time values.
type TimeMetric struct {
	Value time.Time
}

func (t TimeMetric) String() string {
	return t.Value.String()
}

// DurationMetric holds time.Duration values.
type DurationMetric struct {
	Value time.Duration
}

func (d DurationMetric) String() string {
	return d.Value.String()
}

// MapMetric holds map[string]int values (e.g., status codes).
type MapMetric struct {
	Value map[string]int
}

func (m MapMetric) String() string {
	return fmt.Sprintf("%v", m.Value)
}

// SliceMetric holds []string values (e.g., errors).
type SliceMetric struct {
	Value []string
}

func (s SliceMetric) String() string {
	return fmt.Sprintf("%v", s.Value)
}

type Metrics interface {
	Find(name string) (Metric, error)
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
