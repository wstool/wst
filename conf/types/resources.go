// Copyright 2025 Jakub Zelenka and The WST Authors
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

type Certificate struct {
	Certificate string `wst:"certificate,path=cert"`
	PrivateKey  string `wst:"private_key,path=cert"`
}

type Script struct {
	Content    string     `wst:"content"`
	Path       string     `wst:"path"`
	Mode       string     `wst:"mode,default=0644"`
	Parameters Parameters `wst:"parameters,factory=createParameters"`
}

type Resources struct {
	Certificates map[string]Certificate `wst:"certificates"`
	Scripts      map[string]Script      `wst:"scripts,string=Content"`
}
