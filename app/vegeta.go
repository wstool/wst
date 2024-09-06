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

package app

import (
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"time"
)

type VegetaAttacker interface {
	Attack(targeter vegeta.Targeter, rate vegeta.Rate, duration time.Duration, name string) <-chan *vegeta.Result
}

type RealVegetaAttacker struct {
	attacker *vegeta.Attacker
}

func (a *RealVegetaAttacker) Attack(
	targeter vegeta.Targeter,
	rate vegeta.Rate,
	duration time.Duration,
	name string,
) <-chan *vegeta.Result {
	return a.attacker.Attack(targeter, rate, duration, name)
}

func NewRealVegetaAttacker() VegetaAttacker {
	return &RealVegetaAttacker{
		attacker: vegeta.NewAttacker(),
	}
}

type DryRunVegetaAttacker struct{}

func (a *DryRunVegetaAttacker) Attack(
	targeter vegeta.Targeter,
	rate vegeta.Rate,
	duration time.Duration,
	name string,
) <-chan *vegeta.Result {
	results := make(chan *vegeta.Result, 1) // Buffer of 1 to avoid blocking when sending the result

	go func() {
		defer close(results)
		result := &vegeta.Result{
			Code:      200,
			Latency:   10 * time.Millisecond,
			BytesOut:  512,
			BytesIn:   1024,
			Timestamp: time.Now(),
		}
		results <- result
	}()

	return results
}

func NewDryRunVegetaAttacker() VegetaAttacker {
	return &DryRunVegetaAttacker{}
}

type VegetaMetrics interface {
	Add(r *vegeta.Result)
	Close()
	Metrics() *vegeta.Metrics
}

type DefaultVegetaMetrics struct {
	metrics *vegeta.Metrics
}

func (m DefaultVegetaMetrics) Add(r *vegeta.Result) {
	m.metrics.Add(r)
}

func (m DefaultVegetaMetrics) Close() {
	m.metrics.Close()
}

func (m DefaultVegetaMetrics) Metrics() *vegeta.Metrics {
	return m.metrics
}

func NewDefaultVegetaMetrics() VegetaMetrics {
	return &DefaultVegetaMetrics{
		metrics: &vegeta.Metrics{},
	}
}
