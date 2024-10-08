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

type Spec struct {
	Environments map[string]Environment `wst:"environments,loadable,factory=createEnvironments"`
	Instances    []Instance             `wst:"instances,loadable"`
	Sandboxes    map[string]Sandbox     `wst:"sandboxes,loadable,factory=createSandboxes"`
	Servers      []Server               `wst:"servers,loadable"`
	Workspace    string                 `wst:"workspace,path=virtual"`
	Defaults     SpecDefaults           `wst:"defaults,loadable"`
}

type SpecDefaults struct {
	Service    SpecServiceDefaults `wst:"service"`
	Timeouts   SpecTimeouts        `wst:"timeouts"`
	Parameters Parameters          `wst:"parameters,factory=createParameters"`
}

type SpecServiceDefaults struct {
	Sandbox string                    `wst:"sandbox,enum=local|docker|kubernetes,default=local"`
	Server  SpecServiceServerDefaults `wst:"server"`
}

type SpecServiceServerDefaults struct {
	Tag string `wst:"tag,default=default"`
}

type SpecTimeouts struct {
	Action  int `wst:"action,default=0"`
	Actions int `wst:"actions,default=0"`
}
