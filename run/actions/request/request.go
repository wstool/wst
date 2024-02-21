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
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/instances/runtime"
	"github.com/bukka/wst/run/services"
	"io"
	"net/http"
)

type Action struct {
	Service services.Service
	Id      string
	Path    string
	Method  string
	Headers types.Headers
}

type ActionMaker struct {
	env app.Env
}

func CreateActionMaker(env app.Env) *ActionMaker {
	return &ActionMaker{
		env: env,
	}
}

func (m *ActionMaker) Make(
	config *types.RequestAction,
	svcs services.Services,
) (*Action, error) {
	svc, err := svcs.GetService(config.Service)
	if err != nil {
		return nil, err
	}

	return &Action{
		Service: svc,
		Id:      config.Id,
		Path:    config.Path,
		Method:  config.Method,
		Headers: config.Headers,
	}, nil
}

// ResponseData holds the response body and headers from an HTTP request.
type ResponseData struct {
	Body    string
	Headers http.Header
}

func (a Action) Execute(runData runtime.Data, dryRun bool) (bool, error) {
	// Construct the request URL from the Service and Path.
	url := a.Service.BaseUrl() + a.Path

	// Create the HTTP request.
	req, err := http.NewRequest(a.Method, url, nil)
	if err != nil {
		return false, err
	}

	// Add headers to the request.
	for key, value := range a.Headers {
		req.Header.Add(key, value)
	}

	// Send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// Create a ResponseData instance to hold both body and headers.
	responseData := ResponseData{
		Body:    string(body),
		Headers: resp.Header,
	}

	// Store the ResponseData in runData.
	if err := runData.Store(a.Id, responseData); err != nil {
		return false, err
	}

	return true, nil
}
