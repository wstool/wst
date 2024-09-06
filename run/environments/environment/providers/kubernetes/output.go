// Copyright 2024 Jakub Zelenka and The WST Authors
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

package kubernetes

import "io"

type CombinedReader struct {
	readers []io.ReadCloser
}

func (cr *CombinedReader) Read(p []byte) (n int, err error) {
	readers := make([]io.Reader, len(cr.readers))
	for i, reader := range cr.readers {
		readers[i] = reader
	}
	return io.MultiReader(readers...).Read(p)
}

func (cr *CombinedReader) Close() error {
	var err error
	for _, reader := range cr.readers {
		if e := reader.Close(); e != nil {
			err = e
		}
	}
	return err
}
