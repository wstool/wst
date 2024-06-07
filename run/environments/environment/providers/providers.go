package providers

type Type string

const (
	LocalType      Type = "local"
	DockerType          = "docker"
	KubernetesType      = "kubernetes"
)
