package output

import (
	"bytes"
	"io"
	"sync"
)

type Collector interface {
	Close()
	AnyReader() io.Reader
	StderrReader() io.Reader
	StdoutReader() io.Reader
	Start(stdoutPipe, stderrPipe io.ReadCloser) error
	Wait()
}

// blockingBufferReader is a custom reader that blocks until data is available in the buffer.
type blockingBufferReader struct {
	buffer *bytes.Buffer
	cond   *sync.Cond
	closed bool
	mu     sync.Mutex
}

// newBlockingBufferReader creates a new blockingBufferReader.
func newBlockingBufferReader(buffer *bytes.Buffer) *blockingBufferReader {
	reader := &blockingBufferReader{
		buffer: buffer,
	}
	reader.cond = sync.NewCond(&reader.mu)
	return reader
}

// Write appends data to the buffer and signals readers.
func (r *blockingBufferReader) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, err := r.buffer.Write(data)
	r.cond.Broadcast() // Signal that data is available
	return n, err
}

// Read reads data from the buffer, blocking if no data is available.
func (r *blockingBufferReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for r.buffer.Len() == 0 && !r.closed {
		r.cond.Wait() // Wait for data to be available
	}

	if r.buffer.Len() == 0 && r.closed {
		return 0, io.EOF
	}

	return r.buffer.Read(p)
}

// Close marks the reader as closed and signals any waiting readers.
func (r *blockingBufferReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	r.cond.Broadcast() // Signal all waiting readers
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
func (bc *BufferedCollector) StdoutReader() io.Reader {
	return bc.stdoutBuffer
}

// StderrReader returns an io.Reader for the collected stderr logs.
func (bc *BufferedCollector) StderrReader() io.Reader {
	return bc.stderrBuffer
}

// AnyReader returns an io.Reader for the mixed logs in the order they were collected.
func (bc *BufferedCollector) AnyReader() io.Reader {
	return bc.mixedBuffer
}

// Close closes all pipes.
func (bc *BufferedCollector) Close() {
	bc.stdoutPipe.Close()
	bc.stderrPipe.Close()

	// Close the custom readers to signal EOF
	bc.stdoutBuffer.Close()
	bc.stderrBuffer.Close()
	bc.mixedBuffer.Close()
}
