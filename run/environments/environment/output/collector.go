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

package output

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/pkg/errors"
	"io"
	"strings"
	"sync"
)

type Collector interface {
	Close() error
	AnyReader(ctx context.Context) io.Reader
	StderrReader(ctx context.Context) io.Reader
	StdoutReader(ctx context.Context) io.Reader
	Start(stdoutPipe, stderrPipe io.ReadCloser) error
	StdoutWriter() io.Writer
	StderrWriter() io.Writer
	LogOutput()
	Wait()
}

// blockingBufferReader is a custom reader that blocks until data is available in the buffer.
type blockingBufferReader struct {
	buffer    *bytes.Buffer
	dataCh    chan struct{}
	closeCh   chan struct{}
	closed    bool
	mu        sync.Mutex
	closeOnce sync.Once
}

// newBlockingBufferReader creates a new blockingBufferReader.
func newBlockingBufferReader(buffer *bytes.Buffer) *blockingBufferReader {
	return &blockingBufferReader{
		buffer:  buffer,
		dataCh:  make(chan struct{}, 1),
		closeCh: make(chan struct{}),
	}
}

// Write appends data to the buffer and signals readers.
func (r *blockingBufferReader) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, err := r.buffer.Write(data)
	select {
	case r.dataCh <- struct{}{}: // Notify readers that data is available
	default: // Non-blocking send
	}
	return n, err
}

func (r *blockingBufferReader) readWithContext(ctx context.Context, p []byte) (int, error) {
	for {
		r.mu.Lock()
		if r.buffer.Len() > 0 {
			n, err := r.buffer.Read(p)
			r.mu.Unlock()
			return n, err
		}
		if r.closed && r.buffer.Len() == 0 {
			r.mu.Unlock()
			return 0, io.EOF
		}
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return 0, errors.Errorf("read cancelled: %v", ctx.Err())
		case <-r.dataCh: // Wait for data to be available
			continue
		case <-r.closeCh: // Wait for the reader to be closed
			return 0, io.EOF
		}
	}
}

// Read reads data from the buffer, blocking if no data is available.
func (r *blockingBufferReader) Read(p []byte) (int, error) {
	return r.readWithContext(context.Background(), p)
}

// Close marks the reader as closed and signals any waiting readers.
func (r *blockingBufferReader) Close() error {
	r.closeOnce.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.closed = true
		close(r.closeCh) // Close the channel to notify readers
	})
	return nil
}

// BufferedCollector collects stdout, stderr, and mixed logs.
type BufferedCollector struct {
	fnd          app.Foundation
	tid          string
	stdoutPipe   io.ReadCloser
	stderrPipe   io.ReadCloser
	stdoutBuffer *blockingBufferReader
	stderrBuffer *blockingBufferReader
	mixedBuffer  *blockingBufferReader
	wg           sync.WaitGroup
}

// NewBufferedCollector initializes and returns a new BufferedCollector.
func NewBufferedCollector(fnd app.Foundation, tid string) *BufferedCollector {
	stdoutBuffer := newBlockingBufferReader(new(bytes.Buffer))
	stderrBuffer := newBlockingBufferReader(new(bytes.Buffer))
	mixedBuffer := newBlockingBufferReader(new(bytes.Buffer))

	return &BufferedCollector{
		fnd:          fnd,
		tid:          tid,
		stdoutBuffer: stdoutBuffer,
		stderrBuffer: stderrBuffer,
		mixedBuffer:  mixedBuffer,
	}
}

// Start collects stdout, stderr, and mixed logs from the given command.
func (bc *BufferedCollector) Start(stdoutPipe, stderrPipe io.ReadCloser) error {
	bc.wg.Add(2)

	bc.stdoutPipe = stdoutPipe
	bc.stderrPipe = stderrPipe

	go bc.collectOutput(stdoutPipe, bc.stdoutBuffer)
	go bc.collectOutput(stderrPipe, bc.stderrBuffer)

	return nil
}

// collectOutput reads from the given pipe and writes to the corresponding buffer.
func (bc *BufferedCollector) collectOutput(pipe io.ReadCloser, buffer *blockingBufferReader) (int64, error) {
	defer bc.wg.Done()

	reader := io.TeeReader(pipe, buffer)
	written, err := io.Copy(bc.mixedBuffer, reader)

	return written, err
}

// Wait blocks until all logs are collected.
func (bc *BufferedCollector) Wait() {
	bc.wg.Wait()
}

// StdoutReader returns an io.Reader for the collected stdout logs.
func (bc *BufferedCollector) StdoutReader(ctx context.Context) io.Reader {
	return newContextAwareReader(ctx, bc.stdoutBuffer)
}

// StderrReader returns an io.Reader for the collected stderr logs.
func (bc *BufferedCollector) StderrReader(ctx context.Context) io.Reader {
	return newContextAwareReader(ctx, bc.stderrBuffer)
}

// AnyReader returns an io.Reader for the mixed logs in the order they were collected.
func (bc *BufferedCollector) AnyReader(ctx context.Context) io.Reader {
	return newContextAwareReader(ctx, bc.mixedBuffer)
}

// StdoutWriter returns a MultiWriter that writes to both stdoutBuffer and mixedBuffer
func (bc *BufferedCollector) StdoutWriter() io.Writer {
	return io.MultiWriter(bc.stdoutBuffer, bc.mixedBuffer)
}

// StderrWriter returns a MultiWriter that writes to both stderrBuffer and mixedBuffer
func (bc *BufferedCollector) StderrWriter() io.Writer {
	return io.MultiWriter(bc.stderrBuffer, bc.mixedBuffer)
}

// Close closes all pipes.
func (bc *BufferedCollector) Close() error {
	var errStrings []string

	// Attempt to close stdoutPipe if set
	if bc.stdoutPipe != nil {
		if err := bc.stdoutPipe.Close(); err != nil {
			errStrings = append(errStrings, fmt.Sprintf("stdoutPipe close error: %v", err))
		}
	}

	// Attempt to close stderrPipe if set
	if bc.stderrPipe != nil {
		if err := bc.stderrPipe.Close(); err != nil {
			errStrings = append(errStrings, fmt.Sprintf("stderrPipe close error: %v", err))
		}
	}

	bc.stdoutBuffer.Close()
	bc.stderrBuffer.Close()
	bc.mixedBuffer.Close()

	// Aggregate errors if any occurred
	if len(errStrings) > 0 {
		return errors.New(strings.Join(errStrings, "; "))
	}

	// Return nil if all close operations succeeded
	return nil
}

func (bc *BufferedCollector) logBuffer(name string, buf *bytes.Buffer) {
	logger := bc.fnd.Logger()
	if buf.Len() > 0 {
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			logger.Debugf("Task %s %s: %s", bc.tid, name, scanner.Text())
		}
	} else {
		logger.Debugf("Task %s %s - nothing logged", bc.tid, name)
	}
}

// LogOutput logs data of all buffers. It should be used only after closing the collector otherwise races possible.
func (bc *BufferedCollector) LogOutput() {
	bc.logBuffer("stdout", bc.stdoutBuffer.buffer)
	bc.logBuffer("stderr", bc.stderrBuffer.buffer)
}

// contextAwareReader wraps an io.Reader and respects context cancellation.
type contextAwareReader struct {
	ctx    context.Context
	reader *blockingBufferReader
}

// NewContextAwareReader creates a new contextAwareReader.
func newContextAwareReader(ctx context.Context, reader *blockingBufferReader) *contextAwareReader {
	return &contextAwareReader{
		ctx:    ctx,
		reader: reader,
	}
}

// Read reads from the underlying reader with context.
func (r *contextAwareReader) Read(p []byte) (n int, err error) {
	return r.reader.readWithContext(r.ctx, p)
}
