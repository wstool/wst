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

type InstanceTimeouts struct {
	Action  int `wst:"action,default=30000"`
	Actions int `wst:"actions,default=0"`
}

type InstanceExtends struct {
	Name       string     `wst:"name"`
	Parameters Parameters `wst:"parameters,factory=createParameters"`
}

type Instance struct {
	Name         string                 `wst:"name"`
	Title        string                 `wst:"title"`
	Description  string                 `wst:"description"`
	Labels       []string               `wst:"labels"`
	Abstract     bool                   `wst:"abstract,default=false"`
	Extends      InstanceExtends        `wst:"extends,string=name"`
	Parameters   Parameters             `wst:"parameters,factory=createParameters"`
	Resources    Resources              `wst:"resources"`
	Services     map[string]Service     `wst:"services,loadable"`
	Timeouts     InstanceTimeouts       `wst:"timeouts"`
	Environments map[string]Environment `wst:"environments,loadable,factory=createEnvironments"`
	Actions      []Action               `wst:"actions,factory=createActions"`
}
