// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
