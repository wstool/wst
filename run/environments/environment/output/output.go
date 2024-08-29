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
	MakeCollector(tid string) Collector
}

type nativeMaker struct {
	fnd app.Foundation
}

func (m *nativeMaker) MakeCollector(tid string) Collector {
	return NewBufferedCollector(m.fnd, tid)
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{fnd: fnd}
}
