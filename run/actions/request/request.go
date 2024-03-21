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
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/actions"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"io"
	"net/http"
	"time"
)

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
	svcs services.Services,
	defaultTimeout int,
) (actions.Action, error) {
	svc, err := svcs.FindService(config.Service)
	if err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	return &action{
		fnd:     m.fnd,
		service: svc,
		timeout: time.Duration(config.Timeout),
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

type action struct {
	fnd     app.Foundation
	service services.Service
	timeout time.Duration
	id      string
	path    string
	method  string
	headers types.Headers
}

func (a *action) Timeout() time.Duration {
	return a.timeout
}

func (a *action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing request action")
	// Construct the request URL from the service and path.
	baseUrl, err := a.service.PublicUrl()
	if err != nil {
		return false, err
	}
	url := baseUrl + a.path

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
	a.fnd.Logger().Debugf("Storing response %s: %v", a.id, responseData)
	if err := runData.Store(a.id, responseData); err != nil {
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
