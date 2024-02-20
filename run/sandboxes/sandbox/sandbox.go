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
	GetOutputScanner(outputType OutputType) *bufio.Scanner
	ExecuteCommand(command *hooks.HookCommand) error
	ExecuteSignal(signal *hooks.HookSignal) error
}

type Type string

const (
	LocalType      Type = "local"
	DockerType          = "docker"
	KubernetesType      = "kubernetes"
)
