package app

import "github.com/spf13/afero"

type File interface {
	afero.File
}

type Fs interface {
	afero.Fs
}
