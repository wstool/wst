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

type Config struct {
	Version     string                  `wst:"version,enum=1.0"`
	Name        string                  `wst:"name"`
	Description string                  `wst:"description"`
	Sandboxes   map[SandboxType]Sandbox `wst:"sandboxes,loadable,factory=createSandboxes"`
	Servers     []Server                `wst:"servers,loadable"`
	Spec        Spec                    `wst:"spec"`
}
