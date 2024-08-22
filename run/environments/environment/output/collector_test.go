package output

import (
	"bytes"
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

	go func() {
		defer wg.Done()
		mixedData := new(bytes.Buffer)

		// Continuously read until EOF
		buf := make([]byte, 1024)
		for {
			n, err := collector.AnyReader().Read(buf)
			if n > 0 {
				mixedData.Write(buf[:n])
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				break
			}
			// Short sleep to allow more data to arrive
			time.Sleep(10 * time.Millisecond)
		}

		expected := "stdout line 1\nstdout line 2\nstderr line 1\n"
		assert.Equal(t, expected, mixedData.String())
	}()

	time.Sleep(600 * time.Millisecond)
	collector.Close()

	wg.Wait()
	collector.Wait()
}

func TestBufferedCollector_StdoutReader(t *testing.T) {
	// Define the events with delays
	stdoutEvents := []event{
		{delay: 100 * time.Millisecond, message: "stdout line 1\n"},
		{delay: 200 * time.Millisecond, message: "stdout line 2\n"},
	}
	stderrEvents := []event{
		{delay: 300 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser(stdoutEvents)
	stderrMock := newMockReadCloser(stderrEvents)

	collector := NewBufferedCollector()

	err := collector.Start(stdoutMock, stderrMock)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		mixedData := new(bytes.Buffer)

		// Continuously read until EOF
		buf := make([]byte, 1024)
		for {
			n, err := collector.StdoutReader().Read(buf)
			if n > 0 {
				mixedData.Write(buf[:n])
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				break
			}
			// Short sleep to allow more data to arrive
			time.Sleep(10 * time.Millisecond)
		}

		expected := "stdout line 1\nstdout line 2\n"
		assert.Equal(t, expected, mixedData.String())
	}()

	time.Sleep(400 * time.Millisecond)
	collector.Close()

	wg.Wait()
	collector.Wait()
}

func TestBufferedCollector_StderrReader(t *testing.T) {
	// Define the events with delays
	stdoutEvents := []event{
		{delay: 100 * time.Millisecond, message: "stdout line 1\n"},
		{delay: 200 * time.Millisecond, message: "stdout line 2\n"},
	}
	stderrEvents := []event{
		{delay: 300 * time.Millisecond, message: "stderr line 1\n"},
	}

	stdoutMock := newMockReadCloser(stdoutEvents)
	stderrMock := newMockReadCloser(stderrEvents)

	collector := NewBufferedCollector()

	err := collector.Start(stdoutMock, stderrMock)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		mixedData := new(bytes.Buffer)

		// Continuously read until EOF
		buf := make([]byte, 1024)
		for {
			n, err := collector.StderrReader().Read(buf)
			if n > 0 {
				mixedData.Write(buf[:n])
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				break
			}
			// Short sleep to allow more data to arrive
			time.Sleep(10 * time.Millisecond)
		}

		expected := "stderr line 1\n"
		assert.Equal(t, expected, mixedData.String())
	}()

	time.Sleep(400 * time.Millisecond)
	collector.Close()

	wg.Wait()
	collector.Wait()
}
