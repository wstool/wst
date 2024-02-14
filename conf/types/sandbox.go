package types

type SandboxHookNative struct {
	Type string `wst:"type,enum=start|restart|stop"`
}

type SandboxHookShellCommand struct {
	Command string `wst:"command"`
	Shell   string `wst:"shell"`
}

type SandboxHookCommand struct {
	Executable string   `wst:"executable"`
	Args       []string `wst:"args"`
}

type SandboxHook interface {
}

type CommonSandbox struct {
	Dirs  map[string]string      `wst:"dirs,keys=conf|run|script"`
	Hooks map[string]SandboxHook `wst:"hooks,factory=createHooks"`
}

type LocalSandbox struct {
	CommonSandbox
}

type ContainerImage struct {
	Name string `wst:"name"`
	Tag  string `wst:"tag"`
}

type ContainerRegistryAuth struct {
	Username string `wst:"username"`
	Password string `wst:"password"`
}

type ContainerRegistry struct {
	Auth ContainerRegistryAuth `wst:"auth"`
}

type ContainerSandbox struct {
	CommonSandbox
	Image    ContainerImage    `wst:"image,factory=createContainerImage"`
	Registry ContainerRegistry `wst:"registry"`
}

type DockerSandbox struct {
	ContainerSandbox
}

type KubernetesAuth struct {
	Kubeconfig string `wst:"kubeconfig"`
}

type KubernetesSandbox struct {
	ContainerSandbox
	Auth KubernetesAuth `wst:"auth"`
}

type Sandbox interface {
}
