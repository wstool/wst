// Copyright 2025 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package request

import (
	"context"
	"io"
	"time"

	"github.com/wstool/wst/app"
)

// chunkControlledReader implements io.Reader with control over chunk sizes and delays.
// Each Read() call returns at most chunkSize bytes, which forces the HTTP client to
// make multiple Read() calls, effectively controlling the chunk sizes sent over the wire
// when using chunked transfer encoding.
type chunkControlledReader struct {
	ctx        context.Context
	fnd        app.Foundation
	data       []byte
	chunkSize  int
	chunkDelay time.Duration
	offset     int
	readCount  int
}

func (r *chunkControlledReader) Read(p []byte) (n int, err error) {
	// Check context cancellation
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
	}

	// If all data has been read, return EOF (no delay before EOF)
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}

	// Add delay before each chunk (except the first one)
	if r.readCount > 0 && r.chunkDelay > 0 {
		if err := r.fnd.Sleep(r.ctx, r.chunkDelay); err != nil {
			return 0, err
		}
	}
	r.readCount++

	// Determine how much to read in this chunk
	chunkSize := r.chunkSize
	if chunkSize <= 0 {
		// If chunk size not specified, use the full buffer provided by HTTP client
		chunkSize = len(p)
	}

	// Calculate how much data to copy
	remaining := len(r.data) - r.offset
	toRead := chunkSize
	if toRead > remaining {
		toRead = remaining
	}
	if toRead > len(p) {
		// Never read more than the provided buffer can hold
		toRead = len(p)
	}

	// Copy data to buffer
	n = copy(p, r.data[r.offset:r.offset+toRead])
	r.offset += n

	return n, nil
}
