package template

import (
	"bytes"
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/run/parameters"
	"github.com/bukka/wst/run/servers/templates"
	"github.com/bukka/wst/run/services/template/service"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

type Template interface {
	RenderToWriter(content string, parameters parameters.Parameters, writer io.Writer) error
	RenderToFile(content string, parameters parameters.Parameters, filePath string, perm os.FileMode) error
	RenderToString(content string, parameters parameters.Parameters) (string, error)
}

type Maker interface {
	Make(
		service service.TemplateService,
		services map[string]service.TemplateService,
		serverTemplates templates.Templates,
	) Template
}

type nativeMaker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) Maker {
	return &nativeMaker{
		fnd: fnd,
	}
}

func (m *nativeMaker) Make(
	service service.TemplateService,
	services map[string]service.TemplateService,
	serverTemplates templates.Templates,
) Template {
	return &nativeTemplate{
		fnd:             m.fnd,
		service:         service,
		services:        services,
		serverTemplates: serverTemplates,
	}
}

type nativeTemplate struct {
	// Fondation
	fnd app.Foundation
	// Services.
	services Services
	// Current service.
	service service.TemplateService
	// Server templates
	serverTemplates templates.Templates
}

type Data struct {
	Configs    map[string]string
	Service    service.TemplateService
	Services   Services
	Parameters Parameters
}

func (t *nativeTemplate) RenderToWriter(content string, params parameters.Parameters, writer io.Writer) error {
	mainTmpl, err := template.New("main").Funcs(t.funcs()).Parse(content)
	if err != nil {
		return fmt.Errorf("error parsing main template: %v", err)
	}
	configs := t.service.EnvironmentConfigPaths()
	if configs == nil {
		return fmt.Errorf("configs are not set")
	}

	data := &Data{
		Configs:    configs,
		Service:    t.service,
		Services:   t.services,
		Parameters: NewParameters(params, t),
	}
	if err := mainTmpl.Execute(writer, data); err != nil {
		return err
	}

	return nil
}

func (t *nativeTemplate) RenderToFile(
	content string,
	params parameters.Parameters,
	filePath string,
	perm os.FileMode,
) error {
	fs := t.fnd.Fs()

	err := fs.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return err
	}

	file, err := fs.OpenFile(filePath, os.O_RDWR, perm)
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
