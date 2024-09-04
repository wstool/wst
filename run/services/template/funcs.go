package template

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"text/template"
)

func (t *nativeTemplate) include(tmplName string, data interface{}) (string, error) {
	serverTemplate, found := t.serverTemplates[tmplName]
	if !found {
		return "", fmt.Errorf("failed to find template: %s", tmplName)
	}
	fs := t.fnd.Fs()
	file, err := fs.Open(serverTemplate.FilePath())
	if err != nil {
		return "", fmt.Errorf("failed to open template file: %w", err)
	}
	defer file.Close()

	tmplContent, err := afero.ReadAll(file)
	if err != nil {
		return "", errors.Errorf("failed to read template file: %v", err)
	}

	tmpl, err := template.New(tmplName).Funcs(t.funcs()).Parse(string(tmplContent))
	if err != nil {
		return "", errors.Errorf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", errors.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

func (t *nativeTemplate) funcs() template.FuncMap {
	funcMap := sprig.FuncMap()
	funcMap["include"] = t.include
	return funcMap
}
