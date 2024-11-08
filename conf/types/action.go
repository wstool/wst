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

type CustomExpectation struct {
	Name       string     `wst:"name"`
	Parameters Parameters `wst:"parameters,factory=createParameters"`
}

type CustomExpectationAction struct {
	Service string            `wst:"service"`
	Timeout int               `wst:"timeout"`
	When    string            `wst:"when,enum=always|on_success|on_fail,default=on_success"`
	Custom  CustomExpectation `wst:"custom"`
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
	When    string            `wst:"when,enum=always|on_success|on_fail,default=on_success"`
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
	Body    ResponseBody `wst:"body,string=Content"`
}

type ResponseExpectationAction struct {
	Service  string              `wst:"service"`
	Timeout  int                 `wst:"timeout"`
	When     string              `wst:"when,enum=always|on_success|on_fail,default=on_success"`
	Response ResponseExpectation `wst:"response"`
}

type MetricRule struct {
	Metric   string  `wst:"metric"`
	Operator string  `wst:"operator,enum=eq,ne,gt,lt,ge,le"`
	Value    float64 `wst:"value"`
}

type MetricsExpectation struct {
	Id    string       `wst:"id,default=last"`
	Rules []MetricRule `wst:"rules"`
}

type MetricsExpectationAction struct {
	Service string             `wst:"service"`
	Timeout int                `wst:"timeout"`
	When    string             `wst:"when,enum=always|on_success|on_fail,default=on_success"`
	Metrics MetricsExpectation `wst:"metrics"`
}

type ExpectationAction interface {
	Action
}

type RequestAction struct {
	Service    string  `wst:"service"`
	Timeout    int     `wst:"timeout"`
	When       string  `wst:"when,enum=always|on_success|on_fail,default=on_success"`
	Id         string  `wst:"id,default=last"`
	Path       string  `wst:"path"`
	EncodePath bool    `wst:"encode_path,default=true"`
	Method     string  `wst:"method,enum=GET|HEAD|DELETE|POST|PUT|PATCH|PURGE,default=GET"`
	Headers    Headers `wst:"headers"`
}

type BenchAction struct {
	Service   string  `wst:"service"`
	Timeout   int     `wst:"timeout"`
	When      string  `wst:"when,enum=always|on_success|on_fail,default=on_success"`
	Id        string  `wst:"id,default=last"`
	Path      string  `wst:"path"`
	Method    string  `wst:"method,enum=GET|HEAD|DELETE|POST|PUT|PATCH|PURGE,default=GET"`
	Headers   Headers `wst:"headers"`
	Frequency int     `wst:"frequency"`
	Duration  int     `wst:"duration"`
}

type ParallelAction struct {
	Actions []Action `wst:"actions,factory=createActions"`
	Timeout int      `wst:"timeout"`
	When    string   `wst:"when,enum=always|on_success|on_fail,default=on_success"`
}

type NotAction struct {
	Action  Action `wst:"action,factory=createAction"`
	Timeout int    `wst:"timeout"`
	When    string `wst:"when,enum=always|on_success|on_fail,default=on_success"`
}

type StartAction struct {
	Service  string   `wst:"service"`
	Services []string `wst:"service"`
	Timeout  int      `wst:"timeout"`
	When     string   `wst:"when,enum=always|on_success|on_fail,default=on_success"`
}

type ReloadAction struct {
	Service  string   `wst:"service"`
	Services []string `wst:"service"`
	Timeout  int      `wst:"timeout"`
	When     string   `wst:"when,enum=always|on_success|on_fail,default=on_success"`
}

type RestartAction struct {
	Service  string   `wst:"service"`
	Services []string `wst:"service"`
	Timeout  int      `wst:"timeout"`
	When     string   `wst:"when,enum=always|on_success|on_fail,default=on_success"`
}

type StopAction struct {
	Service  string   `wst:"service"`
	Services []string `wst:"service"`
	Timeout  int      `wst:"timeout"`
	When     string   `wst:"when,enum=always|on_success|on_fail,default=always"`
}

type Action interface {
}
