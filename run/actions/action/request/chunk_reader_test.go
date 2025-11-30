package request

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appMocks "github.com/wstool/wst/mocks/generated/app"
)

func TestChunkControlledReader(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		chunkSize      int
		chunkDelay     time.Duration
		setupMocks     func(*testing.T, context.Context, *appMocks.MockFoundation)
		contextSetup   func() context.Context
		expectedChunks []string
		expectError    bool
		errorMsg       string
	}{
		{
			name:       "read all content in one chunk when chunk size not specified",
			content:    "test content here",
			chunkSize:  0,
			chunkDelay: 0,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// No mocks needed
			},
			expectedChunks: []string{"test content here"},
		},
		{
			name:       "read content in specified chunk sizes",
			content:    "0123456789",
			chunkSize:  3,
			chunkDelay: 0,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// No mocks needed
			},
			expectedChunks: []string{"012", "345", "678", "9"},
		},
		{
			name:       "read content with delays between chunks",
			content:    "abcdefgh",
			chunkSize:  4,
			chunkDelay: 10 * time.Millisecond,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// First read - no delay
				// Second read - delay called
				fnd.On("Sleep", ctx, 10*time.Millisecond).Return(nil).Once()
			},
			expectedChunks: []string{"abcd", "efgh"},
		},
		{
			name:       "read multiple chunks with delays",
			content:    "12345678901234",
			chunkSize:  5,
			chunkDelay: 5 * time.Millisecond,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// First read - no delay
				// Second read - delay
				// Third read - delay
				fnd.On("Sleep", ctx, 5*time.Millisecond).Return(nil).Times(2)
			},
			expectedChunks: []string{"12345", "67890", "1234"},
		},
		{
			name:       "context cancelled during first read",
			content:    "test content",
			chunkSize:  5,
			chunkDelay: 0,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// No mocks needed
			},
			contextSetup: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expectError: true,
			errorMsg:    "context canceled",
		},
		{
			name:       "context cancelled during delay",
			content:    "test content",
			chunkSize:  5,
			chunkDelay: 100 * time.Millisecond,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// First read succeeds
				// Second read - Sleep returns context error
				fnd.On("Sleep", ctx, 100*time.Millisecond).Return(context.Canceled).Once()
			},
			expectedChunks: []string{"test "},
			expectError:    true,
			errorMsg:       "context canceled",
		},
		{
			name:       "chunk size larger than content",
			content:    "short",
			chunkSize:  100,
			chunkDelay: 0,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// No mocks needed
			},
			expectedChunks: []string{"short"},
		},
		{
			name:       "empty content",
			content:    "",
			chunkSize:  10,
			chunkDelay: 0,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// No mocks needed
			},
			expectedChunks: nil,
		},
		{
			name:       "chunk size of 1 (one byte per chunk)",
			content:    "abc",
			chunkSize:  1,
			chunkDelay: 0,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// No mocks needed
			},
			expectedChunks: []string{"a", "b", "c"},
		},
		{
			name:       "delay only without chunk size",
			content:    "test content",
			chunkSize:  0,
			chunkDelay: 10 * time.Millisecond,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// With chunk size 0, it uses buffer size, which in our test is large enough
				// So only one read, no delay
			},
			expectedChunks: []string{"test content"},
		},
		{
			name:       "no delay before EOF read",
			content:    "12345678",
			chunkSize:  4,
			chunkDelay: 100 * time.Millisecond,
			setupMocks: func(t *testing.T, ctx context.Context, fnd *appMocks.MockFoundation) {
				// First read (4 bytes) - no delay
				// Second read (4 bytes) - delay
				// Third read (EOF) - NO DELAY
				fnd.On("Sleep", ctx, 100*time.Millisecond).Return(nil).Once()
			},
			expectedChunks: []string{"1234", "5678"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)

			var ctx context.Context
			if tt.contextSetup != nil {
				ctx = tt.contextSetup()
			} else {
				ctx = context.Background()
			}

			tt.setupMocks(t, ctx, fndMock)

			reader := &chunkControlledReader{
				ctx:        ctx,
				fnd:        fndMock,
				data:       []byte(tt.content),
				chunkSize:  tt.chunkSize,
				chunkDelay: tt.chunkDelay,
				offset:     0,
				readCount:  0,
			}

			var chunks []string
			buf := make([]byte, 1024) // Large buffer to test chunk size control

			for {
				n, err := reader.Read(buf)

				if n > 0 {
					chunks = append(chunks, string(buf[:n]))
				}

				if err == io.EOF {
					break
				}

				if err != nil {
					if tt.expectError {
						assert.Contains(t, err.Error(), tt.errorMsg)
						// Verify we got expected chunks before error
						assert.Equal(t, tt.expectedChunks, chunks)
						return
					}
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if tt.expectError {
				t.Fatal("expected error but got none")
			}

			assert.Equal(t, tt.expectedChunks, chunks)
		})
	}
}

func TestChunkControlledReader_SmallBuffer(t *testing.T) {
	// Test the case where buffer size is smaller than chunk size
	// This covers the "toRead > len(p)" branch
	tests := []struct {
		name           string
		content        string
		chunkSize      int
		bufferSize     int
		expectedChunks []string
	}{
		{
			name:       "buffer smaller than chunk size",
			content:    "1234567890",
			chunkSize:  5,
			bufferSize: 3,
			// Should read min(chunkSize, bufferSize) = 3 bytes at a time
			expectedChunks: []string{"123", "456", "789", "0"},
		},
		{
			name:       "buffer size 1 with chunk size 10",
			content:    "abcde",
			chunkSize:  10,
			bufferSize: 1,
			// Should read 1 byte at a time (buffer limit)
			expectedChunks: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:       "buffer size 2 with chunk size 100",
			content:    "testdata",
			chunkSize:  100,
			bufferSize: 2,
			// Should read 2 bytes at a time (buffer limit)
			expectedChunks: []string{"te", "st", "da", "ta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fndMock := appMocks.NewMockFoundation(t)
			ctx := context.Background()

			reader := &chunkControlledReader{
				ctx:        ctx,
				fnd:        fndMock,
				data:       []byte(tt.content),
				chunkSize:  tt.chunkSize,
				chunkDelay: 0,
				offset:     0,
				readCount:  0,
			}

			var chunks []string
			buf := make([]byte, tt.bufferSize) // Small buffer to trigger the len(p) check

			for {
				n, err := reader.Read(buf)

				if n > 0 {
					chunks = append(chunks, string(buf[:n]))
				}

				if err == io.EOF {
					break
				}

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			assert.Equal(t, tt.expectedChunks, chunks)
		})
	}
}
