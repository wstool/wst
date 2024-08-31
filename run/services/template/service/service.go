package service

// TemplateService defines template specific service subset
type TemplateService interface {
	Address() string
	PrivateUrl() (string, error)
	Executable() (string, error)
	Pid() (int, error)
	ConfDir() (string, error)
	RunDir() (string, error)
	ScriptDir() (string, error)
	Group() string
	User() string
	EnvironmentConfigPaths() map[string]string
}
