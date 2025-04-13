package output

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/wstool/wst/mocks/authored/external"
	appMocks "github.com/wstool/wst/mocks/generated/app"
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

	fndMock := appMocks.NewMockFoundation(t)
	collector := NewBufferedCollector(fndMock, "tid")

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

	fndMock := appMocks.NewMockFoundation(t)
	collector := NewBufferedCollector(fndMock, "tid")

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

	fndMock := appMocks.NewMockFoundation(t)
	collector := NewBufferedCollector(fndMock, "tid")

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

	fndMock := appMocks.NewMockFoundation(t)
	collector := NewBufferedCollector(fndMock, "tid")

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

func TestBufferedCollector_Reader_dispatch(t *testing.T) {
	stdoutEvents := []event{
		{delay: 50 * time.Millisecond, message: "stdout line 1\n"},
	}
	stderrEvents := []event{
		{delay: 50 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser(stdoutEvents)
	stderrMock := newMockReadCloser(stderrEvents)

	fndMock := appMocks.NewMockFoundation(t)
	collector := NewBufferedCollector(fndMock, "tid")
	err := collector.Start(stdoutMock, stderrMock)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		outputType  Type
		expectErr   bool
		expectTexts []string // allow multiple valid results
	}{
		{
			name:        "Reader with Stdout",
			outputType:  Stdout,
			expectTexts: []string{"stdout line 1\n"},
		},
		{
			name:        "Reader with Stderr",
			outputType:  Stderr,
			expectTexts: []string{"stderr line 1\n"},
		},
		{
			name:       "Reader with Any",
			outputType: Any,
			expectTexts: []string{
				"stdout line 1\nstderr line 1\n",
				"stderr line 1\nstdout line 1\n",
			},
		},
		{
			name:       "Reader with unknown type",
			outputType: Type(8),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)

			if tt.expectErr {
				reader, err := collector.Reader(context.Background(), tt.outputType)
				assert.Error(t, err)
				assert.Nil(t, reader)
				wg.Done()
			} else {
				go func() {
					defer wg.Done()
					r, err := collector.Reader(context.Background(), tt.outputType)
					assert.NoError(t, err)
					buf, err := io.ReadAll(r)
					assert.NoError(t, err)

					output := string(buf)
					assert.Contains(t, tt.expectTexts, output,
						"actual output %q not in expected list %v", output, tt.expectTexts)
				}()
			}

			time.Sleep(100 * time.Millisecond)
			collector.Close()
			wg.Wait()
		})
	}

	collector.Wait()
}

func TestBufferedCollector_stderrBuffer_Read(t *testing.T) {
	// Define the events with delays
	stderrEvents := []event{
		{delay: 150 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser([]event{})
	stderrMock := newMockReadCloser(stderrEvents)

	fndMock := appMocks.NewMockFoundation(t)
	collector := NewBufferedCollector(fndMock, "tid")

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

	fndMock := appMocks.NewMockFoundation(t)
	collector := NewBufferedCollector(fndMock, "tid")

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
			fndMock := appMocks.NewMockFoundation(t)
			collector := NewBufferedCollector(fndMock, "tid")
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

func TestBufferedCollector_StdoutWriter(t *testing.T) {
	// Initialize the BufferedCollector
	fndMock := appMocks.NewMockFoundation(t)
	collector := NewBufferedCollector(fndMock, "tid")

	// Write some data using StdoutWriter
	stdoutWriter := collector.StdoutWriter()
	_, err := stdoutWriter.Write([]byte("stdout log 1\n"))
	assert.NoError(t, err)

	_, err = stdoutWriter.Write([]byte("stdout log 2\n"))
	assert.NoError(t, err)

	// Validate that the data written to StdoutWriter is in both stdoutBuffer and mixedBuffer
	expected := "stdout log 1\nstdout log 2\n"

	assert.Equal(t, expected, collector.stdoutBuffer.buffer.String())
	assert.Equal(t, expected, collector.mixedBuffer.buffer.String())
}

func TestBufferedCollector_StderrWriter(t *testing.T) {
	// Initialize the BufferedCollector
	fndMock := appMocks.NewMockFoundation(t)
	collector := NewBufferedCollector(fndMock, "tid")

	// Write some data using StderrWriter
	stderrWriter := collector.StderrWriter()
	_, err := stderrWriter.Write([]byte("stderr log 1\n"))
	assert.NoError(t, err)

	_, err = stderrWriter.Write([]byte("stderr log 2\n"))
	assert.NoError(t, err)

	// Validate that the data written to StderrWriter is in both stderrBuffer and mixedBuffer
	expected := "stderr log 1\nstderr log 2\n"

	assert.Equal(t, expected, collector.stderrBuffer.buffer.String())
	assert.Equal(t, expected, collector.mixedBuffer.buffer.String())
}

func TestBufferedCollector_LogOutput(t *testing.T) {
	tid := "abc"
	tests := []struct {
		name            string
		stdoutMessages  []string
		stderrMessages  []string
		expectedMessage []string
	}{
		{
			name:           "stdout not empty and stderr not empty",
			stdoutMessages: []string{"stdout log 1\n", "stdout log 2\n"},
			stderrMessages: []string{"stderr log 1\n", "stderr log 2\n"},
			expectedMessage: []string{
				"Task abc stdout: stdout log 1",
				"Task abc stdout: stdout log 2",
				"Task abc stderr: stderr log 1",
				"Task abc stderr: stderr log 2",
			},
		},
		{
			name:           "stdout not empty and stderr empty",
			stdoutMessages: []string{"stdout log 1\n", "stdout log 2\n"},
			stderrMessages: []string{},
			expectedMessage: []string{
				"Task abc stdout: stdout log 1",
				"Task abc stdout: stdout log 2",
				"Task abc stderr - nothing logged",
			},
		},
		{
			name:           "stdout empty and stderr not empty",
			stdoutMessages: []string{},
			stderrMessages: []string{"stderr log 1\n", "stderr log 2\n"},
			expectedMessage: []string{
				"Task abc stdout - nothing logged",
				"Task abc stderr: stderr log 1",
				"Task abc stderr: stderr log 2",
			},
		},
		{
			name:           "stdout empty and stderr empty",
			stdoutMessages: []string{},
			stderrMessages: []string{},
			expectedMessage: []string{
				"Task abc stdout - nothing logged",
				"Task abc stderr - nothing logged",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			mockLogger := external.NewMockLogger()
			fndMock.On("Logger").Return(mockLogger.SugaredLogger)
			collector := NewBufferedCollector(fndMock, tid)

			for _, msg := range tt.stdoutMessages {
				_, err := collector.stdoutBuffer.Write([]byte(msg))
				assert.NoError(t, err)
			}
			for _, msg := range tt.stderrMessages {
				_, err := collector.stderrBuffer.Write([]byte(msg))
				assert.NoError(t, err)
			}

			collector.LogOutput()

			assert.Equal(t, tt.expectedMessage, mockLogger.Messages())
		})
	}
}
