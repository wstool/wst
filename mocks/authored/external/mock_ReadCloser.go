package external

import (
	"github.com/stretchr/testify/mock"
)

// Mock for io.ReadCloser
type MockReadCloser struct {
	mock.Mock
}

// NewMockReadCloser initializes a MockReadCloser with some data
func NewMockReadCloser() *MockReadCloser {
	return &MockReadCloser{}
}

// Read simulates the Read method, integrates with testify expectations
func (m *MockReadCloser) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	if args.Get(0) != nil {
		n = args.Int(0)
	}
	err = args.Error(1)
	return
}

// Close simulates the Close method, integrates with testify expectations
func (m *MockReadCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}
