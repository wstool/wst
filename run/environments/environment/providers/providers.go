package providers

type Type string

const (
	LocalType      Type = "local"
	DockerType          = "docker"
	KubernetesType      = "kubernetes"
)

func Types() []Type {
	return []Type{LocalType, DockerType, KubernetesType}
}
