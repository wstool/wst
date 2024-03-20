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

package scripts

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/parameters"
	"os"
	"strconv"
)

type Script interface {
	Path() string
	Content() string
	Mode() os.FileMode
	Parameters() parameters.Parameters
}

type Scripts map[string]Script

type Maker struct {
	fnd             app.Foundation
	parametersMaker *parameters.Maker
}

func CreateMaker(fnd app.Foundation, parametersMaker *parameters.Maker) *Maker {
	return &Maker{
		fnd:             fnd,
		parametersMaker: parametersMaker,
	}
}

func (m *Maker) Make(config map[string]types.Script) (Scripts, error) {
	scripts := make(Scripts)
	for scriptName, scriptConfig := range config {
		mode, err := strconv.ParseUint(scriptConfig.Mode, 8, 32)
		if err != nil {
			return nil, fmt.Errorf("error parsing file mode for script %s: %v", scriptName, err)
		}
		scriptParameters, err := m.parametersMaker.Make(scriptConfig.Parameters)
		if err != nil {
			return nil, fmt.Errorf("error parsing parameters for script %s: %v", scriptName, err)
		}
		script := &nativeScript{
			content:    scriptConfig.Content,
			path:       scriptConfig.Path,
			mode:       os.FileMode(mode),
			parameters: scriptParameters,
		}
		scripts[scriptName] = script
	}
	return scripts, nil
}

type nativeScript struct {
	content    string
	path       string
	mode       os.FileMode
	parameters parameters.Parameters
}

func (s *nativeScript) Parameters() parameters.Parameters {
	return s.parameters
}

func (s *nativeScript) Path() string {
	return s.path
}

func (s *nativeScript) Content() string {
	return s.content
}

func (s *nativeScript) Mode() os.FileMode {
	return s.mode
}
