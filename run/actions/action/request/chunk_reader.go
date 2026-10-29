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
	"errors"
	"io"
	"sync/atomic"
	"time"

	"github.com/wstool/wst/app"
)

// errUploadAborted is returned by chunkControlledReader once the configured content has been sent and the
// upload should not be completed (stall or close transfer options). It makes the HTTP client close the connection.
var errUploadAborted = errors.New("request body upload aborted by client")

// chunkControlledReader implements io.Reader with control over chunk sizes and delays.
// Each Read() call returns at most chunkSize bytes, which forces the HTTP client to
// make multiple Read() calls, effectively controlling the chunk sizes sent over the wire
// when using chunked transfer encoding.
//
// If incomplete is set, the reader never signals the end of the body. After all data has been read, it optionally
// stalls (keeping the connection open without sending anything) and then returns errUploadAborted, which closes
// the connection. The context is used to stop the stall early, typically when the server responded already.
type chunkControlledReader struct {
	ctx        context.Context
	fnd        app.Foundation
	data       []byte
	chunkSize  int
	chunkDelay time.Duration
	offset     int
	readCount  int
	incomplete bool
	stall      time.Duration
	stalled    bool
	aborted    atomic.Bool
}

// Aborted reports whether the reader has aborted the upload (after the stall, if any).
func (r *chunkControlledReader) Aborted() bool {
	return r.aborted.Load()
}

func (r *chunkControlledReader) Read(p []byte) (n int, err error) {
	// Check context cancellation
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
	}

	// If all data has been read, return EOF (no delay before EOF) or abort the upload if it is incomplete
	if r.offset >= len(r.data) {
		if !r.incomplete {
			return 0, io.EOF
		}
		return 0, r.abort()
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

// abort stalls for the configured duration (unless the context is done first) and then marks the upload as
// aborted. The returned error closes the connection on the HTTP client side.
func (r *chunkControlledReader) abort() error {
	if r.stall > 0 && !r.stalled {
		r.stalled = true
		// The context error is intentionally ignored: it means the stall was cut short (for example because
		// the server already responded) and the upload is aborted in any case.
		_ = r.fnd.Sleep(r.ctx, r.stall)
	}
	r.aborted.Store(true)
	return errUploadAborted
}
