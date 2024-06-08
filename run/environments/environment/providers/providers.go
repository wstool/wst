package providers

type Type string

const (
	LocalType      Type = "local"
	DockerType     Type = "docker"
	KubernetesType Type = "kubernetes"
)
