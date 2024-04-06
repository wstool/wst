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

type EnvironmentPorts struct {
	Start int32 `wst:"start"`
	End   int32 `wst:"end"`
}

type CommonEnvironment struct {
	Ports EnvironmentPorts `wst:"ports"`
}

type LocalEnvironment struct {
	Ports EnvironmentPorts `wst:"ports"`
}

type ContainerEnvironment struct {
	Ports    EnvironmentPorts  `wst:"ports"`
	Registry ContainerRegistry `wst:"registry"`
}

type DockerEnvironment struct {
	Ports      EnvironmentPorts  `wst:"ports"`
	Registry   ContainerRegistry `wst:"registry"`
	NamePrefix string            `wst:"name_prefix"`
}

type KubernetesEnvironment struct {
	Ports      EnvironmentPorts  `wst:"ports"`
	Registry   ContainerRegistry `wst:"registry"`
	Namespace  string            `wst:"namespace"`
	Kubeconfig string            `wst:"kubeconfig,path"`
}
