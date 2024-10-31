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

package metrics

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"time"
)

type MetricOperator string

const (
	MetricEqOperator MetricOperator = "eq"
	MetricNeOperator MetricOperator = "ne"
	MetricGtOperator MetricOperator = "gt"
	MetricGeOperator MetricOperator = "ge"
	MetricLtOperator MetricOperator = "lt"
	MetricLeOperator MetricOperator = "le"
)

func ConvertToOperator(op string) (MetricOperator, error) {
	mop := MetricOperator(op)
	switch mop {
	case MetricEqOperator:
		fallthrough
	case MetricGtOperator:
		fallthrough
	case MetricGeOperator:
		fallthrough
	case MetricLeOperator:
		fallthrough
	case MetricLtOperator:
		fallthrough
	case MetricNeOperator:
		return mop, nil
	default:
		return "", fmt.Errorf("invalid operator %s", op)
	}
}

type Metric interface {
	Compare(operator MetricOperator, value float64) (bool, error)
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
func (g GenericMetric[T]) Compare(operator MetricOperator, value float64) (bool, error) {
	typedValue := T(value)
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

type Metrics interface {
	Find(name string) (Metric, error)
	String() string
}
