package app

import "github.com/spf13/afero"

type Fs interface {
	afero.Fs
}
