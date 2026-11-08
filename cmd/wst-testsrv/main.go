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

// Command wst-testsrv is a small HTTP server used by the WST integration tests.
//
// It is not a real web server. Its purpose is to exercise WST features (configuration rendering, start/stop/reload
// hooks, output expectations, requests, ports, unix sockets, TLS, failure handling) in a deterministic way without
// depending on third-party servers such as php-fpm, httpd or nginx.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

// config is the server configuration. All values are strings so the file can be rendered from WST templates
// using only GetString; the server parses them.
type config struct {
	// Listen is either a TCP address (host:port) or a unix socket in the form unix:/path/to/socket.
	Listen string `json:"listen"`
	// ScriptDir is the directory from which plain files are served for all non-special paths.
	ScriptDir string `json:"script_dir"`
	// AccessLog selects where the access log line for each request goes: stdout, stderr or off.
	AccessLog string `json:"access_log"`
	// StartupDelayMs delays the readiness of the server.
	StartupDelayMs string `json:"startup_delay_ms"`
	// FailStart makes the server log an error and exit with a non-zero code instead of starting.
	FailStart string `json:"fail_start"`
	// IgnoreSigterm makes the server log and ignore SIGTERM, so it has to be killed.
	IgnoreSigterm string `json:"ignore_sigterm"`
	// ShutdownTimeoutMs bounds the graceful shutdown.
	ShutdownTimeoutMs string `json:"shutdown_timeout_ms"`
	// TLSCert and TLSKey switch the server to HTTPS when both are set.
	TLSCert string `json:"tls_cert"`
	TLSKey  string `json:"tls_key"`
}

func (c *config) intValue(raw string, name string, def int) (int, error) {
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q: %w", name, raw, err)
	}
	return v, nil
}

func (c *config) boolValue(raw string, name string) (bool, error) {
	switch strings.ToLower(raw) {
	case "", "0", "false", "no", "off":
		return false, nil
	case "1", "true", "yes", "on":
		return true, nil
	}
	return false, fmt.Errorf("invalid %s value %q", name, raw)
}

type server struct {
	cfg         *config
	accessLog   io.Writer
	requests    atomic.Int64
	reloads     atomic.Int64
	connections atomic.Int64
	startedAt   time.Time
	httpServer  *http.Server
}

// connIDKey is the context key under which the connection ID is stored in the request context.
type connIDKey struct{}

// connContext assigns a sequential ID to every accepted connection. The ID is reported in the X-Conn-Id response
// header so that the tests can observe which requests share a connection (keep-alive) and which open a new one.
func (s *server) connContext(ctx context.Context, _ net.Conn) context.Context {
	return context.WithValue(ctx, connIDKey{}, s.connections.Add(1))
}

func notice(format string, args ...any) {
	logLine("NOTICE", format, args...)
}

func errorLog(format string, args ...any) {
	logLine("ERROR", format, args...)
}

func logLine(level string, format string, args ...any) {
	ts := time.Now().Format("02-Jan-2006 15:04:05")
	fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", ts, level, fmt.Sprintf(format, args...))
}

func main() {
	configPath := flag.String("c", "", "path to the JSON configuration file")
	flag.Parse()
	if *configPath == "" {
		errorLog("configuration file is required (-c <path>)")
		os.Exit(2)
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		errorLog("failed to load configuration: %v", err)
		os.Exit(2)
	}

	if err := run(cfg); err != nil {
		errorLog("%v", err)
		os.Exit(1)
	}
}

func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Listen == "" {
		return nil, errors.New("listen address is required")
	}
	return cfg, nil
}

func run(cfg *config) error {
	notice("starting, pid %d", os.Getpid())

	startupDelay, err := cfg.intValue(cfg.StartupDelayMs, "startup_delay_ms", 0)
	if err != nil {
		return err
	}
	failStart, err := cfg.boolValue(cfg.FailStart, "fail_start")
	if err != nil {
		return err
	}
	ignoreSigterm, err := cfg.boolValue(cfg.IgnoreSigterm, "ignore_sigterm")
	if err != nil {
		return err
	}
	shutdownTimeout, err := cfg.intValue(cfg.ShutdownTimeoutMs, "shutdown_timeout_ms", 5000)
	if err != nil {
		return err
	}

	srv := &server{cfg: cfg, startedAt: time.Now()}
	switch strings.ToLower(cfg.AccessLog) {
	case "", "stdout":
		srv.accessLog = os.Stdout
	case "stderr":
		srv.accessLog = os.Stderr
	case "off":
		srv.accessLog = io.Discard
	default:
		return fmt.Errorf("invalid access_log value %q", cfg.AccessLog)
	}

	if startupDelay > 0 {
		notice("delaying startup by %d ms", startupDelay)
		time.Sleep(time.Duration(startupDelay) * time.Millisecond)
	}
	if failStart {
		return errors.New("failed to start: failure requested by configuration")
	}

	listener, err := listen(cfg.Listen)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", cfg.Listen, err)
	}
	notice("listening on %s", cfg.Listen)

	srv.httpServer = &http.Server{Handler: srv.handler(), ConnContext: srv.connContext}
	serveErr := make(chan error, 1)
	go func() {
		var err error
		if cfg.TLSCert != "" && cfg.TLSKey != "" {
			err = srv.httpServer.ServeTLS(listener, cfg.TLSCert, cfg.TLSKey)
		} else {
			err = srv.httpServer.Serve(listener)
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()
	notice("ready to handle connections")

	signals := make(chan os.Signal, 4)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGUSR2)

	for {
		select {
		case err := <-serveErr:
			if err != nil {
				return fmt.Errorf("server failed: %w", err)
			}
			return nil
		case sig := <-signals:
			switch sig {
			case syscall.SIGHUP, syscall.SIGUSR2:
				notice("reloading (%s)", sig)
				srv.requests.Store(0)
				srv.reloads.Add(1)
				notice("reload done")
			case syscall.SIGTERM, syscall.SIGINT:
				if ignoreSigterm && sig == syscall.SIGTERM {
					notice("ignoring %s", sig)
					continue
				}
				notice("terminating (%s) ...", sig)
				ctx, cancel := context.WithTimeout(context.Background(),
					time.Duration(shutdownTimeout)*time.Millisecond)
				err := srv.httpServer.Shutdown(ctx)
				cancel()
				if err != nil {
					errorLog("graceful shutdown failed: %v", err)
				}
				notice("exiting, bye-bye!")
				return nil
			}
		}
	}
}

func listen(address string) (net.Listener, error) {
	if path, ok := strings.CutPrefix(address, "unix:"); ok {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return net.Listen("unix", path)
	}
	return net.Listen("tcp", address)
}

// statusRecorder captures the status code for the access log.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// Unwrap exposes the underlying ResponseWriter so http.ResponseController can reach its optional interfaces.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/_/ping", s.handlePing)
	mux.HandleFunc("/_/echo", s.handleEcho)
	mux.HandleFunc("/_/status/", s.handleStatus)
	mux.HandleFunc("/_/delay", s.handleDelay)
	mux.HandleFunc("/_/stream", s.handleStream)
	mux.HandleFunc("/_/upload", s.handleUpload)
	mux.HandleFunc("/_/stats", s.handleStats)
	mux.HandleFunc("/_/exit", s.handleExit)
	mux.HandleFunc("/", s.handleFile)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.requests.Add(1)
		if connID, ok := r.Context().Value(connIDKey{}).(int64); ok {
			w.Header().Set("X-Conn-Id", strconv.FormatInt(connID, 10))
		}
		rec := &statusRecorder{ResponseWriter: w}
		mux.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = http.StatusOK
		}
		fmt.Fprintf(s.accessLog, "\"%s %s\" %d\n", r.Method, r.URL.RequestURI(), rec.status)
	})
}

func (s *server) handlePing(w http.ResponseWriter, r *http.Request) {
	writeText(w, http.StatusOK, "pong")
}

type echoResponse struct {
	Method           string            `json:"method"`
	Proto            string            `json:"proto"`
	Host             string            `json:"host"`
	Path             string            `json:"path"`
	RawQuery         string            `json:"raw_query"`
	Headers          map[string]string `json:"headers"`
	ContentLength    int64             `json:"content_length"`
	TransferEncoding string            `json:"transfer_encoding"`
	Body             string            `json:"body"`
	BodyLength       int               `json:"body_length"`
	// BodyChunks lists the sizes of the individual reads of the request body as they arrived from the network,
	// for example "25,25" when a 50 bytes body arrived in two pieces.
	BodyChunks string `json:"body_chunks"`
	// BodyReadMs is the time in milliseconds between the first and the last read of the request body.
	BodyReadMs int64 `json:"body_read_ms"`
}

// readBodyChunks reads the whole body and records the size of each read and the time it took.
func readBodyChunks(body io.Reader) ([]byte, []int, time.Duration, error) {
	var data []byte
	var chunks []int
	var start, end time.Time
	buf := make([]byte, 64*1024)
	for {
		n, err := body.Read(buf)
		if n > 0 {
			if start.IsZero() {
				start = time.Now()
			}
			end = time.Now()
			data = append(data, buf[:n]...)
			chunks = append(chunks, n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return data, chunks, end.Sub(start), err
		}
	}
	if start.IsZero() {
		return data, chunks, 0, nil
	}
	return data, chunks, end.Sub(start), nil
}

func (s *server) handleEcho(w http.ResponseWriter, r *http.Request) {
	body, chunks, readDuration, err := readBodyChunks(r.Body)
	if err != nil {
		writeText(w, http.StatusBadRequest, "failed to read body: %v", err)
		return
	}
	chunkSizes := make([]string, len(chunks))
	for i, n := range chunks {
		chunkSizes[i] = strconv.Itoa(n)
	}
	headers := make(map[string]string, len(r.Header))
	for name, values := range r.Header {
		headers[name] = strings.Join(values, ", ")
	}
	resp := echoResponse{
		Method:           r.Method,
		Proto:            r.Proto,
		Host:             r.Host,
		Path:             r.URL.Path,
		RawQuery:         r.URL.RawQuery,
		Headers:          headers,
		ContentLength:    r.ContentLength,
		TransferEncoding: strings.Join(r.TransferEncoding, ","),
		Body:             string(body),
		BodyLength:       len(body),
		BodyChunks:       strings.Join(chunkSizes, ","),
		BodyReadMs:       readDuration.Milliseconds(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimPrefix(r.URL.Path, "/_/status/")
	code, err := strconv.Atoi(raw)
	if err != nil || code < 100 || code > 599 {
		writeText(w, http.StatusBadRequest, "invalid status code %q", raw)
		return
	}
	writeText(w, code, "status %d", code)
}

func (s *server) handleDelay(w http.ResponseWriter, r *http.Request) {
	ms, err := queryInt(r, "ms", 0)
	if err != nil {
		writeText(w, http.StatusBadRequest, "%v", err)
		return
	}
	select {
	case <-time.After(time.Duration(ms) * time.Millisecond):
		writeText(w, http.StatusOK, "delayed %d ms", ms)
	case <-r.Context().Done():
	}
}

func (s *server) handleStream(w http.ResponseWriter, r *http.Request) {
	chunks, err := queryInt(r, "chunks", 3)
	if err != nil {
		writeText(w, http.StatusBadRequest, "%v", err)
		return
	}
	delay, err := queryInt(r, "delay_ms", 100)
	if err != nil {
		writeText(w, http.StatusBadRequest, "%v", err)
		return
	}
	size, err := queryInt(r, "size", 16)
	if err != nil {
		writeText(w, http.StatusBadRequest, "%v", err)
		return
	}
	flusher, _ := w.(http.Flusher)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	chunk := strings.Repeat("x", size)
	for i := 0; i < chunks; i++ {
		if i > 0 {
			select {
			case <-time.After(time.Duration(delay) * time.Millisecond):
			case <-r.Context().Done():
				return
			}
		}
		fmt.Fprintf(w, "%d:%s\n", i, chunk)
		if flusher != nil {
			flusher.Flush()
		}
	}
}

// handleUpload reads the request body like a server with a body read timeout would. If no body data arrives for
// timeout_ms milliseconds (default 1000), it responds with 408 Request Timeout without waiting for the rest of
// the body. If the body ends with an error (for example the client closed the connection), it responds with
// 400 Bad Request. Otherwise it responds with 200 OK and the number of bytes received.
func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	timeout, err := queryInt(r, "timeout_ms", 1000)
	if err != nil {
		writeText(w, http.StatusBadRequest, "%v", err)
		return
	}
	// By default the Go server drains the unread request body before writing the response headers, which
	// would block on a stalled client. Full duplex allows responding while the body is still unread, like a
	// server with a request body timeout does.
	if err := http.NewResponseController(w).EnableFullDuplex(); err != nil {
		errorLog("failed to enable full duplex for upload: %v", err)
	}

	type readResult struct {
		n   int
		err error
	}
	results := make(chan readResult)
	go func() {
		defer close(results)
		buf := make([]byte, 64*1024)
		for {
			n, err := r.Body.Read(buf)
			results <- readResult{n: n, err: err}
			if err != nil {
				return
			}
		}
	}()

	received := 0
	timer := time.NewTimer(time.Duration(timeout) * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case res, ok := <-results:
			if !ok {
				writeText(w, http.StatusOK, "received %d bytes", received)
				return
			}
			received += res.n
			if res.err == io.EOF {
				writeText(w, http.StatusOK, "received %d bytes", received)
				return
			}
			if res.err != nil {
				notice("upload failed after %d bytes: %v", received, res.err)
				writeText(w, http.StatusBadRequest, "request body error after %d bytes: %v", received, res.err)
				return
			}
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(time.Duration(timeout) * time.Millisecond)
		case <-timer.C:
			notice("upload timed out after %d bytes", received)
			writeText(w, http.StatusRequestTimeout, "request body timeout after %d bytes", received)
			return
		case <-r.Context().Done():
			// The client went away (the server cancels the request context when the connection is closed).
			// The response cannot be delivered but the status is still recorded in the access log.
			notice("upload aborted by client after %d bytes", received)
			writeText(w, http.StatusBadRequest, "request body aborted after %d bytes", received)
			return
		}
	}
}

func (s *server) handleStats(w http.ResponseWriter, r *http.Request) {
	writeText(w, http.StatusOK, "pid: %d\nrequests: %d\nreloads: %d\nuptime_seconds: %d\n",
		os.Getpid(), s.requests.Load(), s.reloads.Load(), int(time.Since(s.startedAt).Seconds()))
}

func (s *server) handleExit(w http.ResponseWriter, r *http.Request) {
	code, err := queryInt(r, "code", 1)
	if err != nil {
		writeText(w, http.StatusBadRequest, "%v", err)
		return
	}
	writeText(w, http.StatusOK, "exiting with code %d", code)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		notice("exiting on request with code %d", code)
		os.Exit(code)
	}()
}

func (s *server) handleFile(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ScriptDir == "" {
		writeText(w, http.StatusNotFound, "script dir is not configured")
		return
	}
	root, err := filepath.Abs(s.cfg.ScriptDir)
	if err != nil {
		writeText(w, http.StatusInternalServerError, "invalid script dir: %v", err)
		return
	}
	cleaned := filepath.Clean("/" + r.URL.Path)
	fullPath := filepath.Join(root, cleaned)
	if fullPath != root && !strings.HasPrefix(fullPath, root+string(filepath.Separator)) {
		writeText(w, http.StatusForbidden, "forbidden")
		return
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		writeText(w, http.StatusNotFound, "not found: %s", r.URL.Path)
		return
	}
	if info.IsDir() {
		writeText(w, http.StatusForbidden, "directory listing is not allowed")
		return
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		writeText(w, http.StatusInternalServerError, "failed to read file: %v", err)
		return
	}
	contentType := mime.TypeByExtension(filepath.Ext(fullPath))
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func queryInt(r *http.Request, name string, def int) (int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, fmt.Errorf("invalid %s value %q", name, raw)
	}
	return v, nil
}

func writeText(w http.ResponseWriter, status int, format string, args ...any) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, format, args...)
}
