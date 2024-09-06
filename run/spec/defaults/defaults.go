package defaults

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/parameters"
)

type Defaults struct {
	Service    ServiceDefaults
	Timeouts   TimeoutsDefaults
	Parameters parameters.Parameters
}

type ServiceDefaults struct {
	Sandbox string
	Server  ServiceServerDefaults
}

type ServiceServerDefaults struct {
	Tag string
}

type TimeoutsDefaults struct {
	Actions int
	Action  int
}

type Maker interface {
	Make(config *types.SpecDefaults) (*Defaults, error)
}

type nativeMaker struct {
	fnd             app.Foundation
	parametersMaker parameters.Maker
}

func CreateMaker(fnd app.Foundation, parametersMaker parameters.Maker) Maker {
	return &nativeMaker{
		fnd:             fnd,
		parametersMaker: parametersMaker,
	}
}

func (m *nativeMaker) Make(config *types.SpecDefaults) (*Defaults, error) {
	params, err := m.parametersMaker.Make(config.Parameters)
	if err != nil {
		return nil, err
	}
	return &Defaults{
		Service: ServiceDefaults{
			Sandbox: config.Service.Sandbox,
			Server: ServiceServerDefaults{
				Tag: config.Service.Server.Tag,
			},
		},
		Timeouts: TimeoutsDefaults{
			Actions: config.Timeouts.Actions,
			Action:  config.Timeouts.Action,
		},
		Parameters: params,
	}, nil
}
