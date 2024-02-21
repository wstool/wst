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

type CustomExpectationAction struct {
	Service    string     `wst:"service"`
	Parameters Parameters `wst:"parameters,factory=createParameters"`
}

type OutputExpectation struct {
	Order          string   `wst:"order,enum=fixed|random,default=fixed"`
	Match          string   `wst:"match,enum=exact|regexp,default=exact"`
	Type           string   `wst:"type,enum=stdout|stderr|any,default=any"`
	RenderTemplate bool     `wst:"render_template,default=true"`
	Messages       []string `wst:"messages"`
}

type OutputExpectationAction struct {
	Service string            `wst:"service"`
	Timeout int               `wst:"timeout"`
	Output  OutputExpectation `wst:"output"`
}

type Headers map[string]string

type ResponseBody struct {
	Content        string `wst:"content"`
	Match          string `wst:"match,enum=exact|regexp,default=exact"`
	RenderTemplate bool   `wst:"render_template,default=true"`
}

type ResponseExpectation struct {
	Request string       `wst:"request,default=last"`
	Headers Headers      `wst:"headers"`
	Body    ResponseBody `wst:"content,string=Content"`
}

type ResponseExpectationAction struct {
	Service  string              `wst:"service"`
	Timeout  int                 `wst:"timeout"`
	Response ResponseExpectation `wst:"response"`
}

type ExpectationAction interface {
	Action
}

type RequestAction struct {
	Service string  `wst:"service"`
	Timeout int     `wst:"timeout"`
	Id      string  `wst:"id,default=last"`
	Path    string  `wst:"path"`
	Method  string  `wst:"method,enum=GET|HEAD|DELETE|POST|PUT|PATCH|PURGE,default=GET"`
	Headers Headers `wst:"headers"`
}

type ParallelAction struct {
	Actions []Action `wst:"actions"`
	Timeout int      `wst:"timeout"`
}

type NotAction struct {
	Action  Action `wst:"action"`
	Timeout int    `wst:"timeout"`
}

type StartAction struct {
	Service  string   `wst:"service"`
	Services []string `wst:"service"`
	Timeout  int      `wst:"timeout"`
}

type RestartAction struct {
	Service  string   `wst:"service"`
	Services []string `wst:"service"`
	Reload   bool     `wst:"reload"`
	Timeout  int      `wst:"timeout"`
}

type StopAction struct {
	Service  string   `wst:"service"`
	Services []string `wst:"service"`
	Timeout  int      `wst:"timeout"`
}

type Action interface {
}
