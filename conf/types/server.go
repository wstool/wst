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

type ServerConfig struct {
	File       string     `wst:"file,path"`
	Parameters Parameters `wst:"parameters,factory=createParameters"`
}

type ServerOutputExpectation struct {
	Parameters Parameters        `wst:"parameters,factory=createParameters"`
	Output     OutputExpectation `wst:"output"`
}

type ServerResponseExpectation struct {
	Parameters Parameters          `wst:"parameters,factory=createParameters"`
	Response   ResponseExpectation `wst:"response"`
}

type ServerExpectationAction interface {
}

type ServerActions struct {
	Expect map[string]ServerExpectationAction `wst:"expect,factory=createServerExpectation"`
}

type ServerTemplate struct {
	File string `wst:"file,path"`
}

type Server struct {
	Name       string                    `wst:"name"`
	Extends    string                    `wst:"extends"`
	Configs    map[string]ServerConfig   `wst:"configs"`
	Templates  map[string]ServerTemplate `wst:"templates"`
	Sandboxes  map[string]Sandbox        `wst:"sandboxes,factory=createSandboxes"`
	Parameters Parameters                `wst:"parameters,factory=createParameters"`
	Actions    ServerActions             `wst:"actions"`
}
