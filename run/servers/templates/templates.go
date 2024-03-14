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

package templates

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
)

type Template interface {
	FilePath() string
}

type Templates map[string]Template

func (a Templates) Inherit(parentTemplates Templates) {
	for templateName, template := range parentTemplates {
		_, ok := a[templateName]
		if !ok {
			a[templateName] = template
		}
	}
}

type Maker struct {
	fnd app.Foundation
}

func CreateMaker(fnd app.Foundation) *Maker {
	return &Maker{
		fnd: fnd,
	}
}

func (m *Maker) Make(config map[string]types.ServerTemplate) (Templates, error) {
	configs := make(Templates)
	for name, serverTemplate := range config {
		configs[name] = &nativeTemplate{
			filePath: serverTemplate.File,
		}
	}
	return configs, nil
}

type nativeTemplate struct {
	filePath string
}

func (t *nativeTemplate) FilePath() string {
	return t.filePath
}
