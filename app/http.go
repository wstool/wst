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
	"bytes"
	"io"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type DryRunHttpClient struct {
}

func (c *DryRunHttpClient) Do(req *http.Request) (*http.Response, error) {
	dummyResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`OK`)),
	}
	return dummyResponse, nil
}

func NewDryRunHttpClient() HttpClient {
	return &DryRunHttpClient{}
}

type RealHttpClient struct {
	client *http.Client
}

func (c *RealHttpClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func NewRealHttpClient(tr *http.Transport) HttpClient {
	return &RealHttpClient{
		client: &http.Client{
			Transport: tr,
		},
	}
}
