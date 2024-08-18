package service

// TemplateService defines template specific service subset
type TemplateService interface {
	Address() string
	PrivateUrl() (string, error)
	Pid() (int, error)
	ConfDir() string
	RunDir() string
	ScriptDir() string
	Group() string
	User() string
	EnvironmentConfigPaths() map[string]string
}
