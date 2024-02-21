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

package sandbox

import (
	"bufio"
	"github.com/bukka/wst/run/sandboxes/sandbox/hooks"
)

type OutputType int

const (
	StdoutOutputType = 1
	StderrOutputType = 2
)

type Sandbox interface {
	OutputScanner(outputType OutputType) *bufio.Scanner
	ExecuteCommand(command *hooks.HookCommand) error
	ExecuteSignal(signal *hooks.HookSignal) error
}

type Type string

const (
	LocalType      Type = "local"
	DockerType          = "docker"
	KubernetesType      = "kubernetes"
)