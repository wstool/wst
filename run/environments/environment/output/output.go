package output

import (
	"github.com/bukka/wst/app"
)

type Type int

const (
	Stdout Type = 1
	Stderr      = 2
	Any         = 3
)

type Maker interface {
	MakeCollector() Collector
}

type nativeMaker struct {
	fnd app.Foundation
}

func (n nativeMaker) MakeCollector() Collector {
	return NewBufferedCollector()
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{fnd: fnd}
}
