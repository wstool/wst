package containers

type ContainerConfig struct {
	ImageName        string
	ImageTag         string
	RegistryUsername string
	RegistryPassword string
}

func (c *ContainerConfig) Image() string {
	return c.ImageName + ":" + c.ImageTag
}
