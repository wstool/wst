package kubernetes

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
)

func TestCombinedReader_Read(t *testing.T) {
	cases := []struct {
		name        string
		inputs      []string
		expected    string
		expectError bool
	}{
		{
			name:     "Single reader",
			inputs:   []string{"Hello, world!"},
			expected: "Hello, world!",
		},
		{
			name:     "Multiple readers",
			inputs:   []string{"Hello, ", "world", "!"},
			expected: "Hello, world!",
		},
		{
			name:     "Empty reader",
			inputs:   []string{"", "Hello", "", " world!"},
			expected: "Hello world!",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a slice of io.ReadCloser from strings
			var readers []io.ReadCloser
			for _, input := range tc.inputs {
				readers = append(readers, io.NopCloser(strings.NewReader(input)))
			}

			// Create a combined reader
			cr := &CombinedReader{readers: readers}

			// Read from the combined reader

			buf := new(strings.Builder)
			_, err := io.Copy(buf, cr)
			assert.Nil(t, cr.Close())
			if !tc.expectError {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, buf.String())
			} else {
				require.Error(t, err)
			}

			// Ensure all readers are properly closed
			assert.NoError(t, cr.Close(), "should close all readers without error")
		})
	}
}

type failingReader struct {
	io.Reader
}

func (fr *failingReader) Close() error {
	return fmt.Errorf("close error")
}

func TestCombinedReader_Close(t *testing.T) {
	tests := []struct {
		name               string
		readers            []io.ReadCloser
		expectCloseErr     bool
		expectedCloseError string
	}{
		{
			name: "Error on closing one of the readers",
			readers: []io.ReadCloser{
				io.NopCloser(strings.NewReader("Test")),
				&failingReader{Reader: strings.NewReader("This should fail on close")},
				io.NopCloser(strings.NewReader("Another test")),
			},
			expectCloseErr:     true,
			expectedCloseError: "close error",
		},
		{
			name: "All readers close without error",
			readers: []io.ReadCloser{
				io.NopCloser(strings.NewReader("Test")),
				io.NopCloser(strings.NewReader("Another test")),
			},
			expectCloseErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cr := &CombinedReader{readers: test.readers}

			// Use io.ReadAll to consume the reader fully (not necessary but included for completeness)
			_, _ = io.ReadAll(cr)

			err := cr.Close()
			if test.expectCloseErr {
				require.Error(t, err, "Expected an error when closing the CombinedReader")
				assert.Contains(t, err.Error(), test.expectedCloseError, "Error message should contain the expected text")
			} else {
				assert.NoError(t, err, "Expected no error when closing the CombinedReader")
			}
		})
	}
}
