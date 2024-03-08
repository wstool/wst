package template

import (
	"bytes"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"log"
	"text/template"
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
	// Current service.
	svc services.Service
}

type Data struct {
	Configs    map[string]string
	Service    Service
	Services   Services
	Parameters Parameters
}

func (t *nativeTemplate) Render(content string, params parameters.Parameters) (string, error) {
	mainTmpl, err := template.New("main").Funcs(t.funcs()).Parse(content)
	if err != nil {
		return "", fmt.Errorf("error parsing main template: %v", err)
	}

	data := &Data{
		Configs: t.svc.Server().ConfigPaths(),
		Service: Service{
			service: t.svc,
		},
		Services:   *NewServices(t.svcs),
		Parameters: nil,
	}
	var buf bytes.Buffer
	if err := mainTmpl.Execute(&buf, data); err != nil {
		log.Fatalf("error executing main template: %v", err)
	}

	return buf.String(), nil
}
