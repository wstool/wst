package types

type Instance struct {
	Name      string             `wst:"name"`
	Resources Resources          `wst:"resources"`
	Services  map[string]Service `wst:"services,loadable"`
	Actions   []Action           `wst:"actions,factory=createActions"`
}

type Spec struct {
	Workspace string     `wst:"workspace"`
	Instances []Instance `wst:"instances,loadable"`
}
