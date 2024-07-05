package local

import (
	"bytes"
	"io"
	"sync"
)

// BufferedOutputCollector collects output into an in-memory buffer.
// It provides concurrent-safe writes and an io.Reader to read the collected output.
type BufferedOutputCollector struct {
	buffer  bytes.Buffer
	mu      sync.Mutex
	copyErr error         // Store the first error encountered during copying
	done    chan struct{} // Channel to signal completion of all copies
}

// NewBufferedOutputCollector creates a new BufferedOutputCollector.
func NewBufferedOutputCollector() *BufferedOutputCollector {
	return &BufferedOutputCollector{done: make(chan struct{})}
}

// Write writes data into the output collector's buffer, ensuring thread safety.
func (b *BufferedOutputCollector) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(p)
}

// Read reads data from the output collector's buffer.
// If the buffer is empty and a copy error was recorded, that error is returned.
func (b *BufferedOutputCollector) Read(p []byte) (n int, err error) {
	<-b.done // Ensure all writes are complete
	if b.buffer.Len() == 0 && b.copyErr != nil {
		return 0, b.copyErr
	}
	return b.buffer.Read(p)
}

// collectOutput starts collecting output from the given readers (e.g., stdout and stderr).
func (b *BufferedOutputCollector) collectOutput(readers ...io.ReadCloser) {
	go func() {
		wg := sync.WaitGroup{}
		for _, reader := range readers {
			wg.Add(1)
			go func(r io.ReadCloser) {
				defer wg.Done()
				_, err := io.Copy(b, r)
				b.mu.Lock()
				if err != nil && b.copyErr == nil { // Store the first error encountered
					b.copyErr = err
				}
				b.mu.Unlock()
			}(reader)
		}
		wg.Wait()
		close(b.done) // Close the channel to signal that copying is complete
	}()
}
