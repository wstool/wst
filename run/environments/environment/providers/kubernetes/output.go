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
