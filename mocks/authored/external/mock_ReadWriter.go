package external

import (
	"bytes"
	"github.com/stretchr/testify/mock"
)

// MockWriteCloser simulates an io.WriteCloser and integrates with testify.Mock
type MockWriteCloser struct {
	mock.Mock
	Buffer *bytes.Buffer
}

// NewMockWriteCloser initializes a MockWriteCloser
func NewMockWriteCloser() *MockWriteCloser {
	return &MockWriteCloser{
		Buffer: new(bytes.Buffer),
	}
}

// Write simulates the Write method, integrates with testify expectations
func (m *MockWriteCloser) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	if args.Get(0) != nil {
		n = args.Int(0)
	}
	err = args.Error(1)
	return
}

// Close simulates the Close method, integrates with testify expectations
func (m *MockWriteCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}
