package types

type EnvironmentType string

const (
	CommonEnvironmentType     EnvironmentType = "common"
	LocalEnvironmentType                      = "local"
	ContainerEnvironmentType                  = "container"
	DockerEnvironmentType                     = "docker"
	KubernetesEnvironmentType                 = "kubernetes"
)

func EnvironmentTypes() []EnvironmentType {
	return []EnvironmentType{
		CommonEnvironmentType,
		LocalEnvironmentType,
		ContainerEnvironmentType,
		DockerEnvironmentType,
		KubernetesEnvironmentType,
	}
}

type Environment interface {
}

type CommonEnvironmentPorts struct {
	Start int32 `wst:"start"`
	End   int32 `wst:"end"`
}

type CommonEnvironment struct {
	Ports CommonEnvironmentPorts `wst:"ports"`
}

type LocalEnvironment struct {
	CommonEnvironment
}

type ContainerEnvironment struct {
	CommonEnvironment
	Registry ContainerRegistry `wst:"registry"`
}

type DockerEnvironment struct {
	ContainerEnvironment
	NamePrefix string `wst:"name_prefix"`
}

type KubernetesEnvironment struct {
	ContainerEnvironment
	Namespace  string `wst:"namespace"`
	Kubeconfig string `wst:"kubeconfig,path"`
}
