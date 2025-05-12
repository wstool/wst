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

package expect

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/expectations"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/metrics"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/services"
)

func (m *ExpectationActionMaker) MakeMetricsAction(
	config *types.MetricsExpectationAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	commonExpectation, err := m.MakeCommonExpectation(
		sl, config.Service, config.Timeout, defaultTimeout, config.When, config.OnFailure)
	if err != nil {
		return nil, err
	}

	metricsExpectation, err := m.expectationsMaker.MakeMetricsExpectation(&config.Metrics)
	if err != nil {
		return nil, err
	}

	return &metricsAction{
		CommonExpectation:  commonExpectation,
		MetricsExpectation: metricsExpectation,
		parameters:         commonExpectation.service.ServerParameters(),
	}, nil
}

type metricsAction struct {
	*CommonExpectation
	*expectations.MetricsExpectation
	parameters parameters.Parameters
}

func (a *metricsAction) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing expectation output action")

	metricsKey := fmt.Sprintf("metrics/%s", a.Id)
	data, ok := runData.Load(metricsKey)
	if !ok {
		return false, errors.Errorf("metrics data for key %s not found", metricsKey)
	}

	metricsData, ok := data.(metrics.Metrics)
	if !ok {
		return false, errors.New("invalid metrics data type")
	}
	a.fnd.Logger().Debugf("Checking metrics %s data: %v", a.Id, metricsData)

	for _, rule := range a.Rules {
		metric, err := metricsData.Find(rule.Metric)
		if err != nil {
			return false, fmt.Errorf("failed to find metric %s: %w", rule.Metric, err)
		}

		result, err := metric.Compare(rule.Operator, rule.Value)
		if err != nil {
			return false, fmt.Errorf("failed to compare metric %s: %w", rule.Metric, err)
		}

		if !result {
			if a.fnd.DryRun() {
				return true, nil
			}
			return false, nil
		}
	}
	return true, nil
}
