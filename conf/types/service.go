package types

type Script struct {
	Content string `wst:"content"`
	Path    string `wst:"path"`
	Mode    string `wst:"mode"`
}

type Resources struct {
	Scripts map[string]Script `wst:"scripts,string=Content"`
}

type ServiceConfig struct {
	Parameters          Parameters `wst:"parameters,factory=createParameters"`
	OverwriteParameters bool       `wst:"overwrite_parameters"`
}

type ServiceScripts struct {
	IncludeAll  bool
	IncludeList []string
}

type ServiceResources struct {
	Scripts ServiceScripts `wst:"scripts,factory=createServiceScripts"`
}

type Service struct {
	Server    string                   `wst:"server"`
	Sandbox   string                   `wst:"sandbox,enum=local|docker|kubernetes,default=local"`
	Resources ServiceResources         `wst:"resources"`
	Configs   map[string]ServiceConfig `wst:"configs"`
}
