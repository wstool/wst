package template

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"text/template"
)

func (t *nativeTemplate) include(tmplName string, data interface{}) (string, error) {
	serverTemplate, found := t.server.Template(tmplName)
	if !found {
		return "", fmt.Errorf("failed to find template: %s", tmplName)
	}
	tmpl, err := template.ParseFiles(serverTemplate.FilePath())
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (t *nativeTemplate) funcs() template.FuncMap {
	funcMap := sprig.FuncMap()
	funcMap["include"] = t.include
	return funcMap
}
