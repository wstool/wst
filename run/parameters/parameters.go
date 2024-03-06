package parameters

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/parameters/parameter"
)

type Parameters map[string]parameter.Parameter

type Maker struct {
	fnd            app.Foundation
	parameterMaker *parameter.Maker
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd:            fnd,
		parameterMaker: parameter.CreateMaker(fnd),
	}
}

func (m *Maker) Make(config types.Parameters) (Parameters, error) {
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
