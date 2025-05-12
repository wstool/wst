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
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/services"
	"io"
	"net/http"
	"net/url"
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
		fnd:        m.fnd,
		service:    svc,
		timeout:    time.Duration(config.Timeout * 1e6),
		when:       action.When(config.When),
		onFailure:  action.OnFailureType(config.OnFailure),
		id:         config.Id,
		path:       config.Path,
		encodePath: config.EncodePath,
		method:     config.Method,
		headers:    config.Headers,
	}, nil
}

// ResponseData holds the response body and headers from an HTTP request.
type ResponseData struct {
	Status     string
	StatusCode int
	Proto      string
	Body       string
	Headers    http.Header
}

func (r ResponseData) String() string {
	var headers string
	for name, values := range r.Headers {
		for _, value := range values {
			headers += fmt.Sprintf("\n%s: %s", name, value)
		}
	}

	body := ""
	if r.Body != "" {
		body = "\n\n" + r.Body
	}

	return fmt.Sprintf("%s %s%s%s", r.Proto, r.Status, headers, body)
}

type Action struct {
	fnd        app.Foundation
	service    services.Service
	timeout    time.Duration
	when       action.When
	onFailure  action.OnFailureType
	id         string
	path       string
	encodePath bool
	method     string
	headers    types.Headers
}

func (a *Action) When() action.When {
	return a.when
}

func (a *Action) OnFailure() action.OnFailureType {
	return a.onFailure
}

func (a *Action) Timeout() time.Duration {
	return a.timeout
}

func (a *Action) Execute(ctx context.Context, runData runtime.Data) (bool, error) {
	a.fnd.Logger().Infof("Executing request action")

	publicUrl, err := a.service.PublicUrl(a.path)
	if err != nil {
		return false, err
	}

	// Create the HTTP request.
	req, err := http.NewRequestWithContext(ctx, a.method, publicUrl, nil)
	if err != nil {
		return false, err
	}
	if !a.encodePath {
		// Error is ignored because this can never fail (invalid URL would fail in NewRequestWithContext call).
		parsedUrl, _ := url.Parse(publicUrl)
		req.URL = &url.URL{
			Scheme: parsedUrl.Scheme,
			Host:   parsedUrl.Host,
			Opaque: fmt.Sprintf("//%s%s", parsedUrl.Host, a.path),
		}
	}

	// Add headers to the request.
	for key, value := range a.headers {
		req.Header.Add(key, value)
	}
	a.fnd.Logger().Debugf("Sending request: %s", requestToString(req))

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
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Proto:      resp.Proto,
		Body:       body,
		Headers:    resp.Header,
	}

	// Store the ResponseData in runData.
	key := fmt.Sprintf("response/%s", a.id)
	a.fnd.Logger().Debugf("Storing response %s: %s", key, responseData)
	if err := runData.Store(key, responseData); err != nil {
		return false, err
	}

	return true, nil
}

func requestToString(req *http.Request) string {
	var headers string
	for name, values := range req.Header {
		for _, value := range values {
			headers += fmt.Sprintf("\n%s: %s", name, value)
		}
	}

	return fmt.Sprintf("%s %s %s%s", req.Method, req.URL, req.Proto, headers)
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
