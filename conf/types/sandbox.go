package types

type SandboxType string

const (
	CommonSandboxType     SandboxType = "common"
	LocalSandboxType                  = "local"
	ContainerSandboxType              = "container"
	DockerSandboxType                 = "docker"
	KubernetesSandboxType             = "kubernetes"
)

type SandboxHookType string

const (
	StartSandboxHookType  SandboxHookType = "start"
	StopSandboxHookType                   = "stop"
	ReloadSandboxHookType                 = "reload"
)

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

type SandboxHookSignal struct {
	IsString    bool
	StringValue string
	IntValue    int
}

type SandboxHook interface {
}

type CommonSandbox struct {
	Dirs  map[string]string      `wst:"dirs,keys=conf|run|script"`
	Hooks map[string]SandboxHook `wst:"hooks,factory=createHooks"`
}

func (s *CommonSandbox) GetType() SandboxType {
	return CommonSandboxType
}

type LocalSandbox struct {
	CommonSandbox
}

func (s *LocalSandbox) GetType() SandboxType {
	return KubernetesSandboxType
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

func (s *ContainerSandbox) GetType() SandboxType {
	return KubernetesSandboxType
}

type DockerSandbox struct {
	ContainerSandbox
}

func (s *DockerSandbox) GetType() SandboxType {
	return KubernetesSandboxType
}

type KubernetesAuth struct {
	Kubeconfig string `wst:"kubeconfig,path"`
}

type KubernetesSandbox struct {
	ContainerSandbox
	Auth KubernetesAuth `wst:"auth"`
}

func (s *KubernetesSandbox) GetType() SandboxType {
	return KubernetesSandboxType
}

type Sandbox interface {
	GetType() SandboxType
}
