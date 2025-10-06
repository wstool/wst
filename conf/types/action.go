// Copyright 2024-2025 Jakub Zelenka and The WST Authors
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
	Service   string            `wst:"service"`
	Timeout   int               `wst:"timeout"`
	When      string            `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string            `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
	Custom    CustomExpectation `wst:"custom"`
}

type OutputExpectation struct {
	Command        string   `wst:"command"`
	Order          string   `wst:"order,enum=fixed|random,default=fixed"`
	Match          string   `wst:"match,enum=exact|regexp|prefix|suffix|infix,default=exact"`
	Type           string   `wst:"type,enum=stdout|stderr|any,default=any"`
	RenderTemplate bool     `wst:"render_template,default=true"`
	Messages       []string `wst:"messages"`
}

type OutputExpectationAction struct {
	Service   string            `wst:"service"`
	Timeout   int               `wst:"timeout"`
	When      string            `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string            `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
	Output    OutputExpectation `wst:"output"`
}

type Headers map[string]string

type ResponseBody struct {
	Content        string `wst:"content"`
	Match          string `wst:"match,enum=exact|regexp|prefix|suffix|infix,default=exact"`
	RenderTemplate bool   `wst:"render_template,default=true"`
}

type ResponseExpectation struct {
	Request string       `wst:"request,default=last"`
	Headers Headers      `wst:"headers"`
	Body    ResponseBody `wst:"body,string=Content"`
	Status  int          `wst:"status"`
}

type ResponseExpectationAction struct {
	Service   string              `wst:"service"`
	Timeout   int                 `wst:"timeout"`
	When      string              `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string              `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
	Response  ResponseExpectation `wst:"response"`
}

type MetricRule struct {
	Metric   string  `wst:"metric"`
	Operator string  `wst:"operator,enum=eq|ne|gt|lt|ge|le"`
	Value    float64 `wst:"value"`
}

type MetricsExpectation struct {
	Id    string       `wst:"id,default=last"`
	Rules []MetricRule `wst:"rules"`
}

type MetricsExpectationAction struct {
	Service   string             `wst:"service"`
	Timeout   int                `wst:"timeout"`
	When      string             `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string             `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
	Metrics   MetricsExpectation `wst:"metrics"`
}

type ShellCommand struct {
	Command string
}

type ArgsCommand struct {
	Args []string
}

type Command interface{}

type ExecuteAction struct {
	Service        string            `wst:"service"`
	Timeout        int               `wst:"timeout"`
	When           string            `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure      string            `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
	Id             string            `wst:"id,default=last"`
	Command        Command           `wst:"command,factory=createCommand"`
	RenderTemplate bool              `wst:"render_template,default=true"`
	Shell          string            `wst:"shell,default=/bin/sh"`
	Env            map[string]string `wst:"env"`
}

type TLSClientConfig struct {
	SkipVerify bool   `wst:"skip_verify,default=false"`
	CACert     string `wst:"ca_certificate"`
}

type RequestAction struct {
	Service    string          `wst:"service"`
	Timeout    int             `wst:"timeout"`
	When       string          `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure  string          `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
	Id         string          `wst:"id,default=last"`
	Scheme     string          `wst:"scheme,enum=http|https,default=http"`
	Protocols  []string        `wst:"protocols,enum=http1.1|http2"`
	Path       string          `wst:"path"`
	EncodePath bool            `wst:"encode_path,default=true"`
	Method     string          `wst:"method,enum=GET|HEAD|DELETE|POST|PUT|PATCH|PURGE,default=GET"`
	Headers    Headers         `wst:"headers"`
	TLS        TLSClientConfig `wst:"tls"`
}

type BenchAction struct {
	Service   string  `wst:"service"`
	Timeout   int     `wst:"timeout"`
	When      string  `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string  `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
	Id        string  `wst:"id,default=last"`
	Scheme    string  `wst:"scheme,enum=http|https,default=http"`
	Path      string  `wst:"path"`
	Method    string  `wst:"method,enum=GET|HEAD|DELETE|POST|PUT|PATCH|PURGE,default=GET"`
	Headers   Headers `wst:"headers"`
	Frequency int     `wst:"frequency"`
	Duration  int     `wst:"duration"`
}

type ParallelAction struct {
	Actions   []Action `wst:"actions,factory=createActions"`
	Timeout   int      `wst:"timeout"`
	When      string   `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string   `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
}

type SequentialAction struct {
	Actions   []Action `wst:"actions,factory=createActions"`
	Service   string   `wst:"service"`
	Timeout   int      `wst:"timeout"`
	Name      string   `wst:"name"`
	When      string   `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string   `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
}

type NotAction struct {
	Action    Action `wst:"action,factory=createAction"`
	Timeout   int    `wst:"timeout"`
	When      string `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
}

type StartAction struct {
	Service   string   `wst:"service"`
	Services  []string `wst:"services"`
	Timeout   int      `wst:"timeout"`
	When      string   `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string   `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
}

type ReloadAction struct {
	Service   string   `wst:"service"`
	Services  []string `wst:"services"`
	Timeout   int      `wst:"timeout"`
	When      string   `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string   `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
}

type RestartAction struct {
	Service   string   `wst:"service"`
	Services  []string `wst:"services"`
	Timeout   int      `wst:"timeout"`
	When      string   `wst:"when,enum=always|on_success|on_failure,default=on_success"`
	OnFailure string   `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
}

type StopAction struct {
	Service   string   `wst:"service"`
	Services  []string `wst:"services"`
	Timeout   int      `wst:"timeout"`
	When      string   `wst:"when,enum=always|on_success|on_failure,default=always"`
	OnFailure string   `wst:"on_failure,enum=fail|ignore|skip,default=fail"`
}

type Action interface {
}
