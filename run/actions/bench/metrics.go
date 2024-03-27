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
	"github.com/bukka/wst/run/metrics"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"time"
)

type Metrics struct {
	metrics vegeta.Metrics
}

func (vm *Metrics) Find(name string) (metrics.Metric, error) {
	switch name {
	case "Requests":
		return metrics.GenericMetric[uint64]{Value: vm.metrics.Requests}, nil
	case "Rate":
		return metrics.GenericMetric[float64]{Value: vm.metrics.Rate}, nil
	case "Throughput":
		return metrics.GenericMetric[float64]{Value: vm.metrics.Throughput}, nil
	case "Duration":
		return metrics.GenericMetric[time.Duration]{Value: vm.metrics.Duration}, nil
	case "Success":
		return metrics.GenericMetric[float64]{Value: vm.metrics.Success}, nil
	case "LatencyTotal":
		return metrics.GenericMetric[time.Duration]{Value: vm.metrics.Latencies.Total}, nil
	case "LatencyMean":
		return metrics.GenericMetric[time.Duration]{Value: vm.metrics.Latencies.Mean}, nil
	case "LatencyP50":
		return metrics.GenericMetric[time.Duration]{Value: vm.metrics.Latencies.P50}, nil
	case "LatencyP90":
		return metrics.GenericMetric[time.Duration]{Value: vm.metrics.Latencies.P90}, nil
	case "LatencyP95":
		return metrics.GenericMetric[time.Duration]{Value: vm.metrics.Latencies.P95}, nil
	case "LatencyP99":
		return metrics.GenericMetric[time.Duration]{Value: vm.metrics.Latencies.P99}, nil
	case "LatencyMax":
		return metrics.GenericMetric[time.Duration]{Value: vm.metrics.Latencies.Max}, nil
	case "LatencyMin":
		return metrics.GenericMetric[time.Duration]{Value: vm.metrics.Latencies.Min}, nil
	default:
		return nil, fmt.Errorf("metric %s not found", name)
	}
}
