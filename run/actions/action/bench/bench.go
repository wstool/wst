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
	"context"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"time"
)

type Maker interface {
	Make(
		config *types.BenchAction,
		sl services.ServiceLocator,
		defaultTimeout int,
	) (action.Action, error)
}

type ActionMaker struct {
	fnd app.Foundation
}

func CreateActionMaker(fnd app.Foundation) *ActionMaker {
	return &ActionMaker{
		fnd: fnd,
	}
}

func (m *ActionMaker) Make(
	config *types.BenchAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	svc, err := sl.Find(config.Service)
	if err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		if defaultTimeout > config.Duration {
			config.Timeout = defaultTimeout
		} else {
			config.Timeout = config.Duration + 5000
		}
	}

	return &Action{
		fnd:      m.fnd,
		service:  svc,
		timeout:  time.Duration(config.Timeout * 1e6),
		duration: time.Duration(config.Duration * 1e6),
		freq:     config.Frequency,
		id:       config.Id,
		path:     config.Path,
		method:   config.Method,
		headers:  config.Headers,
	}, nil
}

type Action struct {
	fnd      app.Foundation
	service  services.Service
	timeout  time.Duration
	duration time.Duration
	freq     int
	id       string
	path     string
	method   string
	headers  types.Headers
}

func (a *Action) Timeout() time.Duration {
	return a.timeout
}

func (a *Action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing bench action")
	url, err := a.service.PublicUrl(a.path)
	if err != nil {
		return false, err
	}
	rate := vegeta.Rate{Freq: a.freq, Per: time.Second}
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: a.method,
		URL:    url,
	})
	attacker := a.fnd.VegetaAttacker()

	results := attacker.Attack(targeter, rate, a.duration, a.service.Name())

	metrics := a.fnd.VegetaMetrics()
	done := make(chan struct{})
	errChan := make(chan error)
	go func() {
		defer close(done)
		for res := range results {
			metrics.Add(res)
		}
		metrics.Close()

		key := fmt.Sprintf("metrics/%s", a.id)
		metricsData := &Metrics{
			metrics: metrics,
		}
		a.fnd.Logger().Debugf("Storing response %s: %v", key, metricsData)
		if err = runData.Store(key, metricsData); err != nil {
			a.fnd.Logger().Errorf("Error storing metrics data: %v", err)
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.fnd.Logger().Infof("Cancelling attack due to context cancellation.")
		return false, ctx.Err()
	case err = <-errChan:
		return false, err
	case <-done:
		// Attack completed
		return true, nil
	}
}
