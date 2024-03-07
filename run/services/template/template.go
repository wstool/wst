package template

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
)

type Template interface {
	Render(content string, parameters parameters.Parameters) (string, error)
}

type Maker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd: fnd,
	}
}

func (m *Maker) Make(svc services.Service, svcs services.Services) Template {
	return &nativeTemplate{
		svc:  svc,
		svcs: svcs,
	}
}

type nativeTemplate struct {
	// All services.
	svcs services.Services
	// Current svc.
	svc services.Service
}

func (t *nativeTemplate) Render(content string, params parameters.Parameters) (string, error) {

}
