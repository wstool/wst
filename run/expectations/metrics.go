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

package expectations

import (
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/metrics"
)

func (m *Maker) MakeMetricsExpectation(
	config *types.MetricsExpectation,
) (*MetricsExpectation, error) {
	rules := make([]MetricRule, 0, len(config.Rules))
	for _, configRule := range config.Rules {
		operator, err := metrics.ConvertToOperator(configRule.Operator)
		if err != nil {
			return nil, err
		}
		rules = append(rules, MetricRule{
			Metric:   configRule.Metric,
			Operator: operator,
			Value:    configRule.Value,
		})
	}

	return &MetricsExpectation{
		Id:    config.Id,
		Rules: rules,
	}, nil
}

type MetricRule struct {
	Metric   string
	Operator metrics.MetricOperator
	Value    float64
}

type MetricsExpectation struct {
	Id    string
	Rules []MetricRule
}
