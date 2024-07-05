package local

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

type errorReaderCloser struct {
	io.Reader
	err error
}

func (erc *errorReaderCloser) Read(p []byte) (int, error) {
	if erc.err != nil {
		return 0, erc.err
	}
	return erc.Reader.Read(p)
}

func (erc *errorReaderCloser) Close() error {
	return nil // For simplicity, no error on close
}

func newErrorReaderCloser(data string, err error) io.ReadCloser {
	return &errorReaderCloser{
		Reader: strings.NewReader(data),
		err:    err,
	}
}

func TestBufferedOutputCollector(t *testing.T) {
	tests := []struct {
		name           string
		inputs         func() []io.ReadCloser
		expectedOutput string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "single stream without error",
			inputs: func() []io.ReadCloser {
				r := io.NopCloser(strings.NewReader("Hello, world!"))
				return []io.ReadCloser{r}
			},
			expectedOutput: "Hello, world!",
			expectError:    false,
		},
		{
			name: "multiple streams without error",
			inputs: func() []io.ReadCloser {
				r1 := io.NopCloser(strings.NewReader("out"))
				r2 := io.NopCloser(strings.NewReader("out"))
				return []io.ReadCloser{r1, r2}
			},
			expectedOutput: "outout",
			expectError:    false,
		},
		{
			name: "stream with read error",
			inputs: func() []io.ReadCloser {
				r := newErrorReaderCloser("Hello", fmt.Errorf("read error"))
				return []io.ReadCloser{r}
			},
			expectError:    true,
			expectedErrMsg: "read error",
		},
		{
			name: "multiple streams with one error",
			inputs: func() []io.ReadCloser {
				r1 := io.NopCloser(strings.NewReader("Hello, "))
				r2 := newErrorReaderCloser("world!", fmt.Errorf("read error"))
				return []io.ReadCloser{r1, r2}
			},
			expectError:    true,
			expectedErrMsg: "read error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewBufferedOutputCollector()
			readers := tt.inputs()

			// Start collecting output
			collector.collectOutput(readers...)

			// Create buffer to read into
			buf := new(strings.Builder)
			_, err := io.Copy(buf, collector)

			// Check for error expectations
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, buf.String())
			}
		})
	}
}
