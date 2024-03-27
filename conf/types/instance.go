package types

type InstanceTimeouts struct {
	Action  int `wst:"action,default=30000"`
	Actions int `wst:"actions,default=0"`
}

type Instance struct {
	Name         string                 `wst:"name"`
	Resources    Resources              `wst:"resources"`
	Services     map[string]Service     `wst:"services,loadable"`
	Timeouts     InstanceTimeouts       `wst:"timeouts"`
	Environments map[string]Environment `wst:"environments,loadable,factory=createEnvironments"`
	Actions      []Action               `wst:"actions,factory=createActions"`
}
