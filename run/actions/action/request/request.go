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
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/conf/types"
	"github.com/wstool/wst/run/actions/action"
	"github.com/wstool/wst/run/instances/runtime"
	"github.com/wstool/wst/run/services"
)

type Protocol string

const (
	ProtocolHTTP11 Protocol = "http1.1"
	ProtocolHTTP2  Protocol = "http2"
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

	if config.Scheme != "https" && (config.TLS.SkipVerify || config.TLS.CACert != "") {
		return nil, errors.New("TLS configuration is only valid for HTTPS requests")
	}

	// Set default protocols if not specified
	protocols := config.Protocols
	if len(protocols) == 0 {
		if config.Scheme == "https" {
			// Default for HTTPS: allow both HTTP/1.1 and HTTP/2
			protocols = []string{string(ProtocolHTTP11), string(ProtocolHTTP2)}
		} else {
			// Default for HTTP: only HTTP/1.1 (h2c is not commonly supported)
			protocols = []string{string(ProtocolHTTP11)}
		}
	}

	// Convert strings to Protocol type and validate
	validatedProtocols := make([]Protocol, 0, len(protocols))
	for _, protoStr := range protocols {
		proto := Protocol(protoStr)
		if proto == ProtocolHTTP2 && config.Scheme != "https" {
			m.fnd.Logger().Infof("Using unencrypted HTTP/2 (h2c) over plain HTTP")
		}
		validatedProtocols = append(validatedProtocols, proto)
	}

	return &Action{
		fnd:        m.fnd,
		service:    svc,
		timeout:    time.Duration(config.Timeout * 1e6),
		when:       action.When(config.When),
		onFailure:  action.OnFailureType(config.OnFailure),
		id:         config.Id,
		scheme:     config.Scheme,
		path:       config.Path,
		encodePath: config.EncodePath,
		method:     config.Method,
		headers:    config.Headers,
		body:       &config.Body,
		tls:        &config.TLS,
		protocols:  validatedProtocols,
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
	scheme     string
	path       string
	encodePath bool
	method     string
	headers    types.Headers
	body       *types.RequestBody
	tls        *types.TLSClientConfig
	protocols  []Protocol
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
	a.fnd.Logger().Infof("Executing request action with HTTP protocols: %v", a.protocols)

	// Create transport
	tr := &http.Transport{}

	// Configure protocols using the Protocols API
	protocolConfig := a.buildProtocolConfig()
	tr.Protocols = protocolConfig

	if a.scheme == "https" {
		tlsConfig, err := a.buildTLSConfig()
		if err != nil {
			return false, err
		}
		tr.TLSClientConfig = tlsConfig
	}

	a.fnd.Logger().Debugf("Protocol configuration: HTTP/1=%t, HTTP/2=%t, UnencryptedHTTP/2=%t",
		protocolConfig.HTTP1(), protocolConfig.HTTP2(), protocolConfig.UnencryptedHTTP2())

	publicUrl, err := a.service.PublicUrl(a.scheme, a.path)
	if err != nil {
		return false, err
	}

	// Create a request body reader
	var bodyReader io.Reader
	if a.body != nil && a.body.Content != "" {
		bodyReader = a.createBodyReader(ctx)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, a.method, publicUrl, bodyReader)
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

	// Add headers to the request
	for key, value := range a.headers {
		req.Header.Add(key, value)
	}

	// Handle transfer configuration
	if a.body != nil && a.body.Content != "" {
		a.applyTransferConfig(req)
	}

	a.fnd.Logger().Debugf("Sending request: %s", requestToString(req))

	// Send the request
	client := a.fnd.HttpClient(tr)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := readResponse(ctx, resp.Body)
	if err != nil {
		return false, err
	}

	// Create a ResponseData instance to hold both body and headers
	responseData := ResponseData{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Proto:      resp.Proto,
		Body:       body,
		Headers:    resp.Header,
	}

	// Store the ResponseData in runData
	key := fmt.Sprintf("response/%s", a.id)
	a.fnd.Logger().Debugf("Storing response %s: %s (protocol: %s)", key, responseData, resp.Proto)
	if err := runData.Store(key, responseData); err != nil {
		return false, err
	}

	return true, nil
}

// createBodyReader creates an io.Reader for the request body based on transfer configuration
func (a *Action) createBodyReader(ctx context.Context) io.Reader {
	content := []byte(a.body.Content)

	// If chunked encoding with chunk size or delay specified, use a custom reader
	if a.body.Transfer.Encoding == "chunked" && (a.body.Transfer.ChunkSize > 0 || a.body.Transfer.ChunkDelay > 0) {
		return &chunkControlledReader{
			ctx:        ctx,
			fnd:        a.fnd,
			data:       content,
			chunkSize:  a.body.Transfer.ChunkSize,
			chunkDelay: time.Duration(a.body.Transfer.ChunkDelay) * time.Millisecond,
			offset:     0,
		}
	}

	// For normal transfers or chunked without size/delay control, use bytes.Reader
	return bytes.NewReader(content)
}

// applyTransferConfig applies transfer configuration to the request
func (a *Action) applyTransferConfig(req *http.Request) {
	if a.body.Transfer.Encoding == "chunked" {
		req.TransferEncoding = []string{"chunked"}
		if a.body.Transfer.ChunkSize > 0 {
			a.fnd.Logger().Debugf("Using chunked encoding with chunk size: %d", a.body.Transfer.ChunkSize)
		}
	}

	// Override Content-Length if specified (allows mismatches for testing unfinished uploads)
	if a.body.Transfer.ContentLength > 0 {
		req.ContentLength = int64(a.body.Transfer.ContentLength)
		a.fnd.Logger().Debugf("Setting Content-Length to: %d (actual body length: %d)",
			a.body.Transfer.ContentLength, len(a.body.Content))
	} else if a.body.Transfer.Encoding != "chunked" {
		// Set actual content length if not chunked and not overridden
		req.ContentLength = int64(len(a.body.Content))
	}
}

func (a *Action) buildProtocolConfig() *http.Protocols {
	config := new(http.Protocols)

	for _, proto := range a.protocols {
		switch proto {
		case ProtocolHTTP11:
			config.SetHTTP1(true)
		case ProtocolHTTP2:
			if a.scheme == "https" {
				// HTTP/2 over TLS
				config.SetHTTP2(true)
			} else {
				// HTTP/2 cleartext (h2c) over plain HTTP
				config.SetUnencryptedHTTP2(true)
			}
		}
	}

	return config
}

func (a *Action) buildTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: a.tls.SkipVerify,
	}

	if a.tls.CACert != "" {
		caCert, err := a.service.FindCertificate(a.tls.CACert)
		if err != nil {
			return nil, errors.Errorf("CA certificate %s not found", a.tls.CACert)
		}
		caCertPool := a.fnd.X509CertPool()
		if !caCertPool.AppendCertFromPEM(caCert.Certificate.CertificateData()) {
			return nil, errors.New("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool.CertPool()
	}

	return tlsConfig, nil
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
