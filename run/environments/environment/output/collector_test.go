package output

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"
	"time"
)

type event struct {
	delay   time.Duration
	message string
}

type mockReadCloser struct {
	buffer *bytes.Buffer
	events []event
	closed bool
	mu     sync.Mutex
	cond   *sync.Cond
}

func newMockReadCloser(events []event) *mockReadCloser {
	mrc := &mockReadCloser{
		buffer: new(bytes.Buffer),
		events: events,
	}
	mrc.cond = sync.NewCond(&mrc.mu)
	go mrc.generateEvents()
	return mrc
}

func (m *mockReadCloser) generateEvents() {
	for _, e := range m.events {
		time.Sleep(e.delay)
		m.mu.Lock()
		m.buffer.WriteString(e.message)
		m.cond.Broadcast() // Notify that new data is available
		m.mu.Unlock()
	}
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for m.buffer.Len() == 0 && !m.closed {
		m.cond.Wait() // Wait for data to be written to the buffer
	}

	if m.closed && m.buffer.Len() == 0 {
		return 0, io.EOF
	}

	return m.buffer.Read(p)
}

func (m *mockReadCloser) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.cond.Broadcast() // Notify all waiting readers
	return nil
}

func (m *mockReadCloser) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func readWorker(t *testing.T, wg *sync.WaitGroup, expectError bool, expected string, read func(buf []byte) (int, error)) {
	defer wg.Done()
	mixedData := new(bytes.Buffer)

	// Continuously read until EOF
	buf := make([]byte, 1024)
	for {
		n, err := read(buf)
		if n > 0 {
			mixedData.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			if expectError {
				assert.Contains(t, err.Error(), expected)
			} else {
				t.Errorf("unexpected error: %v", err)
			}
			break
		}
		// Short sleep to allow more data to arrive
		time.Sleep(10 * time.Millisecond)
	}
	if !expectError {
		assert.Equal(t, expected, mixedData.String())
	}
}

func TestBufferedCollector_AnyReader(t *testing.T) {
	// Define the events with delays
	stdoutEvents := []event{
		{delay: 100 * time.Millisecond, message: "stdout line 1\n"},
		{delay: 200 * time.Millisecond, message: "stdout line 2\n"},
	}
	stderrEvents := []event{
		{delay: 400 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser(stdoutEvents)
	stderrMock := newMockReadCloser(stderrEvents)

	collector := NewBufferedCollector()

	err := collector.Start(stdoutMock, stderrMock)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go readWorker(t, &wg, false, "stdout line 1\nstdout line 2\nstderr line 1\n",
		func(buf []byte) (int, error) {
			return collector.AnyReader(context.Background()).Read(buf)
		})

	time.Sleep(600 * time.Millisecond)
	collector.Close()

	wg.Wait()
	collector.Wait()
}

func TestBufferedCollector_StdoutReader(t *testing.T) {
	// Define the events with delays
	stdoutEvents := []event{
		{delay: 50 * time.Millisecond, message: "stdout line 1\n"},
		{delay: 100 * time.Millisecond, message: "stdout line 2\n"},
	}
	stderrEvents := []event{
		{delay: 150 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser(stdoutEvents)
	stderrMock := newMockReadCloser(stderrEvents)

	collector := NewBufferedCollector()

	err := collector.Start(stdoutMock, stderrMock)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go readWorker(t, &wg, false, "stdout line 1\nstdout line 2\n",
		func(buf []byte) (int, error) {
			return collector.StdoutReader(context.Background()).Read(buf)
		})

	time.Sleep(200 * time.Millisecond)
	collector.Close()

	wg.Wait()
	collector.Wait()
}

func TestBufferedCollector_StderrReader(t *testing.T) {
	// Define the events with delays
	stdoutEvents := []event{
		{delay: 50 * time.Millisecond, message: "stdout line 1\n"},
		{delay: 100 * time.Millisecond, message: "stdout line 2\n"},
	}
	stderrEvents := []event{
		{delay: 150 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser(stdoutEvents)
	stderrMock := newMockReadCloser(stderrEvents)

	collector := NewBufferedCollector()

	err := collector.Start(stdoutMock, stderrMock)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go readWorker(t, &wg, false, "stderr line 1\n", func(buf []byte) (int, error) {
		return collector.StderrReader(context.Background()).Read(buf)
	})

	time.Sleep(250 * time.Millisecond)
	collector.Close()

	wg.Wait()
	collector.Wait()
}

func TestBufferedCollector_StderrReader_cancelled(t *testing.T) {
	// Define the events with delays
	stderrEvents := []event{
		{delay: 150 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser([]event{})
	stderrMock := newMockReadCloser(stderrEvents)

	collector := NewBufferedCollector()

	err := collector.Start(stdoutMock, stderrMock)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go readWorker(t, &wg, true, "read cancelled", func(buf []byte) (int, error) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return collector.StderrReader(ctx).Read(buf)
	})

	time.Sleep(250 * time.Millisecond)
	collector.Close()

	wg.Wait()
	collector.Wait()
}

func TestBufferedCollector_stderrBuffer_Read(t *testing.T) {
	// Define the events with delays
	stderrEvents := []event{
		{delay: 150 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser([]event{})
	stderrMock := newMockReadCloser(stderrEvents)

	collector := NewBufferedCollector()

	err := collector.Start(stdoutMock, stderrMock)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go readWorker(t, &wg, false, "stderr line 1\n", func(buf []byte) (int, error) {
		return collector.stderrBuffer.Read(buf)
	})

	time.Sleep(250 * time.Millisecond)
	collector.Close()

	wg.Wait()
	collector.Wait()
}

func TestBufferedCollector_stderrBuffer_Close(t *testing.T) {
	// Define the events with delays
	stderrEvents := []event{
		{delay: 150 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser([]event{})
	stderrMock := newMockReadCloser(stderrEvents)

	collector := NewBufferedCollector()

	err := collector.Start(stdoutMock, stderrMock)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go readWorker(t, &wg, false, "", func(buf []byte) (int, error) {
		collector.Close()
		return collector.stderrBuffer.Read(buf)
	})

	time.Sleep(250 * time.Millisecond)
	collector.Close()

	wg.Wait()
	collector.Wait()
}

// MockReadCloser simulates an io.ReadCloser with an optional close error.
type closeErrorReadCloser struct {
	closeErr error
}

func (m *closeErrorReadCloser) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (m *closeErrorReadCloser) Close() error {
	return m.closeErr
}

// TestBufferedCollector_Close tests the Close method of BufferedCollector.
func TestBufferedCollector_Close(t *testing.T) {
	tests := []struct {
		name           string
		stdoutCloseErr error
		stderrCloseErr error
		expectedErr    string
	}{
		{
			name:        "All close operations succeed",
			expectedErr: "",
		},
		{
			name:           "stdoutPipe close fails",
			stdoutCloseErr: errors.New("stdout close error"),
			expectedErr:    "stdoutPipe close error: stdout close error",
		},
		{
			name:           "stderrPipe close fails",
			stderrCloseErr: errors.New("stderr close error"),
			expectedErr:    "stderrPipe close error: stderr close error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock objects with the desired close errors
			stdoutPipe := &closeErrorReadCloser{closeErr: tt.stdoutCloseErr}
			stderrPipe := &closeErrorReadCloser{closeErr: tt.stderrCloseErr}

			// Create a BufferedCollector with the actual blockingBufferReader instances
			collector := NewBufferedCollector()
			collector.stdoutPipe = stdoutPipe
			collector.stderrPipe = stderrPipe

			// Call Close and check the result
			err := collector.Close()
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			}
		})
	}
}
