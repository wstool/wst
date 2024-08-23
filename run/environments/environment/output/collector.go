package output

import (
	"bytes"
	"context"
	"fmt"
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

// Read reads data from the buffer, blocking if no data is available.
func (r *blockingBufferReader) Read(p []byte) (int, error) {
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
		case <-r.dataCh: // Wait for data to be available
		case <-r.closeCh: // Wait for the reader to be closed
			return 0, io.EOF
		}
	}
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
	stdoutPipe   io.ReadCloser
	stderrPipe   io.ReadCloser
	stdoutBuffer *blockingBufferReader
	stderrBuffer *blockingBufferReader
	mixedBuffer  *blockingBufferReader
	wg           sync.WaitGroup
}

// NewBufferedCollector initializes and returns a new BufferedCollector.
func NewBufferedCollector() *BufferedCollector {
	stdoutBuffer := newBlockingBufferReader(new(bytes.Buffer))
	stderrBuffer := newBlockingBufferReader(new(bytes.Buffer))
	mixedBuffer := newBlockingBufferReader(new(bytes.Buffer))

	return &BufferedCollector{
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
func (bc *BufferedCollector) collectOutput(pipe io.ReadCloser, buffer *blockingBufferReader) {
	defer bc.wg.Done()

	reader := io.TeeReader(pipe, buffer)

	// Continuously read from the pipe
	io.Copy(bc.mixedBuffer, reader)
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

// Close closes all pipes.
func (bc *BufferedCollector) Close() error {
	var errStrings []string

	// Attempt to close stdoutPipe
	if err := bc.stdoutPipe.Close(); err != nil {
		errStrings = append(errStrings, fmt.Sprintf("stdoutPipe close error: %v", err))
	}

	// Attempt to close stderrPipe
	if err := bc.stderrPipe.Close(); err != nil {
		errStrings = append(errStrings, fmt.Sprintf("stderrPipe close error: %v", err))
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

// Read reads from the underlying reader or returns early if the context is done.
func (r *contextAwareReader) Read(p []byte) (n int, err error) {
	for {
		select {
		case <-r.ctx.Done():
			return 0, errors.Errorf("read cancelled: %v", r.ctx.Err())
		case <-r.reader.dataCh: // Data is available
			return r.reader.Read(p)
		case <-r.reader.closeCh: // Reader is closed
			return 0, io.EOF
		}
	}
}
