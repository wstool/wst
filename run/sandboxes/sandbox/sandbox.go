package sandbox

import (
	"github.com/bukka/wst/run/sandboxes/sandbox/hooks"
)

type Sandbox interface {
	ExecuteCommand(command *hooks.HookCommand) error
	ExecuteSignal(signal *hooks.HookSignal) error
}

type Type string

const (
	LocalType      Type = "local"
	DockerType          = "docker"
	KubernetesType      = "kubernetes"
)
