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

func NewRealHttpClient() HttpClient {
	return &RealHttpClient{
		client: &http.Client{},
	}
}
