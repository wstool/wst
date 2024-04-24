package parameters

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/parameters/parameter"
)

type Parameters map[string]parameter.Parameter

type Maker interface {
	Make(config types.Parameters) (Parameters, error)
}

type nativeMaker struct {
	fnd            app.Foundation
	parameterMaker parameter.Maker
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{
		fnd:            fnd,
		parameterMaker: parameter.CreateMaker(fnd),
	}
}

func (m *nativeMaker) Make(config types.Parameters) (Parameters, error) {
	params := make(map[string]parameter.Parameter)
	for key, elem := range config {
		if paramElem, err := m.parameterMaker.Make(elem); err == nil {
			params[key] = paramElem
		} else {
			return nil, err
		}
	}

	return params, nil
}

func (p Parameters) Inherit(parameters Parameters) Parameters {
	for key, value := range parameters {
		if _, ok := p[key]; !ok {
			p[key] = value
		}
	}
	return p
}
