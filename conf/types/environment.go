package types

type LocalEnvironmentPorts struct {
	From int16 `wst:"from"`
	To   int16 `wst:"to"`
}

type LocalEnvironment struct {
	Ports LocalEnvironmentPorts `wst:"ports"`
}

type DockerEnvironment struct {
	NamePrefix string `wst:"name_prefix"`
}

type KubernetesEnvironment struct {
	Namespace  string `wst:"name_prefix"`
	Kubeconfig string `wst:"kubeconfig,path"`
}

type Environment interface {
}
