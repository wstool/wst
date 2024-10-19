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
	"github.com/pkg/errors"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/run/parameters"
	"github.com/wstool/wst/run/servers/templates"
	"github.com/wstool/wst/run/services/template/service"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

// DefaultIncludeMaxDepth is a max depth set to not allow too big slow down.
const DefaultIncludeMaxDepth = 100

type Template interface {
	RenderToWriter(content string, parameters parameters.Parameters, writer io.Writer) error
	RenderToFile(content string, parameters parameters.Parameters, filePath string, perm os.FileMode) error
	RenderToString(content string, parameters parameters.Parameters) (string, error)
}

type Maker interface {
	Make(
		service service.TemplateService,
		services Services,
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
	services Services,
	serverTemplates templates.Templates,
) Template {
	return &nativeTemplate{
		fnd:             m.fnd,
		service:         service,
		services:        services,
		serverTemplates: serverTemplates,
		maxIncludeDepth: DefaultIncludeMaxDepth,
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
	// Inclusion depth
	includeDepth int
	// Max inclusion depth
	maxIncludeDepth int
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
		return errors.Errorf("error parsing main template: %v", err)
	}
	configs := t.service.EnvironmentConfigPaths()
	if configs == nil {
		return errors.Errorf("service configs are not set")
	}

	data := &Data{
		Configs:    configs,
		Service:    t.service,
		Services:   t.services,
		Parameters: NewParameters(params, t),
	}
	if err = mainTmpl.Execute(writer, data); err != nil {
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

	dirPath := filepath.Dir(filePath)
	err := fs.MkdirAll(dirPath, 0755)
	if err != nil {
		return errors.Errorf("creating directory %s failed: %v", dirPath, err)
	}

	file, err := fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		return errors.Errorf("creating / opening file %s failed: %v", filePath, err)
	}
	defer file.Close()

	if err = t.RenderToWriter(content, params, file); err != nil {
		return errors.Errorf("rendering file %s failed: %v", filePath, err)
	}

	return nil
}

func (t *nativeTemplate) RenderToString(content string, params parameters.Parameters) (string, error) {
	var buf bytes.Buffer
	if err := t.RenderToWriter(content, params, &buf); err != nil {
		return "", errors.Errorf("rendering template string fialed: %v", err)
	}

	return buf.String(), nil
}
