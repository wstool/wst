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

package request

import (
	"context"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions/action"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"io"
	"net/http"
	"time"
)

type Maker interface {
	Make(
		config *types.RequestAction,
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
	config *types.RequestAction,
	sl services.ServiceLocator,
	defaultTimeout int,
) (action.Action, error) {
	svc, err := sl.Find(config.Service)
	if err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return &Action{
		fnd:     m.fnd,
		service: svc,
		timeout: time.Duration(config.Timeout * 1e6),
		when:    action.When(config.When),
		id:      config.Id,
		path:    config.Path,
		method:  config.Method,
		headers: config.Headers,
	}, nil
}

// ResponseData holds the response body and headers from an HTTP request.
type ResponseData struct {
	Body    string
	Headers http.Header
}

type Action struct {
	fnd     app.Foundation
	service services.Service
	timeout time.Duration
	when    action.When
	id      string
	path    string
	method  string
	headers types.Headers
}

func (a *Action) When() action.When {
	return a.when
}

func (a *Action) Timeout() time.Duration {
	return a.timeout
}

func (a *Action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing request action")

	url, err := a.service.PublicUrl(a.path)
	if err != nil {
		return false, err
	}

	// Create the HTTP request.
	req, err := http.NewRequestWithContext(ctx, a.method, url, nil)
	if err != nil {
		return false, err
	}

	// Add headers to the request.
	for key, value := range a.headers {
		req.Header.Add(key, value)
	}
	a.fnd.Logger().Debugf("Sending request: %v", req)

	// Send the request.
	client := a.fnd.HttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := readResponse(ctx, resp.Body)
	if err != nil {
		return false, err
	}

	// Create a ResponseData instance to hold both body and headers.
	responseData := ResponseData{
		Body:    body,
		Headers: resp.Header,
	}

	// Store the ResponseData in runData.
	key := fmt.Sprintf("response/%s", a.id)
	a.fnd.Logger().Debugf("Storing response %s: %v", key, responseData)
	if err := runData.Store(key, responseData); err != nil {
		return false, err
	}

	return true, nil
}

func readResponse(ctx context.Context, body io.ReadCloser) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		content, err := io.ReadAll(body)
		return string(content), err
	}
}
