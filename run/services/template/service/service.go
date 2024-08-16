package service

import "github.com/bukka/wst/run/sandboxes/dir"

// TemplateService defines template specific service subset
type TemplateService interface {
	Address() string
	PrivateUrl() (string, error)
	Pid() (int, error)
	Dirs() map[dir.DirType]string
	Group() string
	User() string
	EnvironmentConfigPaths() map[string]string
}
