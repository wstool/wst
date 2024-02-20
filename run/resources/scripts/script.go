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
	"os"
	"strconv"
)

type Script interface {
}

type Scripts map[string]Script

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(config map[string]types.Script) (Scripts, error) {
	scripts := make(Scripts)
	for scriptName, scriptConfig := range config {
		mode, err := strconv.ParseUint(scriptConfig.Mode, 8, 32)
		if err != nil {
			return nil, fmt.Errorf("error parsing file mode for script %s: %v", scriptName, err)
		}
		script := &nativeScript{
			content: scriptConfig.Content,
			path:    scriptConfig.Path,
			mode:    os.FileMode(mode),
		}
		scripts[scriptName] = script
	}
	return scripts, nil
}

type nativeScript struct {
	content string
	path    string
	mode    os.FileMode
}
