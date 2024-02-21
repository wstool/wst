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

package types

type Script struct {
	Content string `wst:"content"`
	Path    string `wst:"path"`
	Mode    string `wst:"mode"`
}

type Resources struct {
	Scripts map[string]Script `wst:"scripts,string=Content"`
}

type ServiceConfig struct {
	Parameters          Parameters `wst:"parameters,factory=createParameters"`
	OverwriteParameters bool       `wst:"overwrite_parameters"`
}

type ServiceScripts struct {
	IncludeAll  bool
	IncludeList []string
}

type ServiceResources struct {
	Scripts ServiceScripts `wst:"scripts,factory=createServiceScripts"`
}

type Service struct {
	Server    string                   `wst:"server"`
	Sandbox   string                   `wst:"sandbox,enum=local|docker|kubernetes,default=local"`
	Resources ServiceResources         `wst:"resources"`
	Configs   map[string]ServiceConfig `wst:"configs"`
}