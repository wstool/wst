package template

import (
	"bytes"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/services"
	"io"
	"os"
	"text/template"
)

type Template interface {
	RenderToWriter(content string, parameters parameters.Parameters, writer io.Writer) error
	RenderToFile(content string, parameters parameters.Parameters, filePath string) error
	RenderToString(content string, parameters parameters.Parameters) (string, error)
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

func (t *nativeTemplate) RenderToWriter(content string, params parameters.Parameters, writer io.Writer) error {
	mainTmpl, err := template.New("main").Funcs(t.funcs()).Parse(content)
	if err != nil {
		return fmt.Errorf("error parsing main template: %v", err)
	}
	configs := t.svc.ConfigPaths()
	if configs == nil {
		return fmt.Errorf("configs are not set")
	}

	data := &Data{
		Configs:    configs,
		Service:    *NewService(t.svc),
		Services:   *NewServices(t.svcs),
		Parameters: NewParameters(params, t),
	}
	if err := mainTmpl.Execute(writer, data); err != nil {
		return err
	}

	return nil
}

func (t *nativeTemplate) RenderToFile(content string, params parameters.Parameters, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = t.RenderToWriter(content, params, file); err != nil {
		return err
	}

	return nil
}

func (t *nativeTemplate) RenderToString(content string, params parameters.Parameters) (string, error) {
	var buf bytes.Buffer
	if err := t.RenderToWriter(content, params, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
