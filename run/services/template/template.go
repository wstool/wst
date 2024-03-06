package template

import (
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
)

type Template struct {
	// All services.
	svcs services.Services
	// Current svc.
	svc services.Service
}

func NewTemplate(svc services.Service, svcs services.Services) *Template {
	return &Template{
		svc:  svc,
		svcs: svcs,
	}
}

func (t *Template) Render(content string, parameters parameters.Parameters) (string, error) {
	panic("implement me")
}
