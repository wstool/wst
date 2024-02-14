package types

type ServerConfig struct {
	File       string     `wst:"file,path"`
	Parameters Parameters `wst:"parameters,factory=createParameters"`
}

type ServerOutputExpectation struct {
	Parameters Parameters        `wst:"parameters,factory=createParameters"`
	Output     OutputExpectation `wst:"output"`
}

type ServerResponseExpectation struct {
	Parameters Parameters          `wst:"parameters,factory=createParameters"`
	Response   ResponseExpectation `wst:"response"`
}

type ServerActions struct {
	Expect map[string]ExpectationAction `wst:"expect,factory=createServerExpectation"`
}

type ServerTemplate struct {
	File string `wst:"file,path"`
}

type Server struct {
	Name       string                    `wst:"name"`
	Extends    string                    `wst:"extends"`
	Configs    map[string]ServerConfig   `wst:"configs"`
	Templates  map[string]ServerTemplate `wst:"templates"`
	Sandboxes  map[string]Sandbox        `wst:"sandboxes"`
	Parameters Parameters                `wst:"parameters,factory=createParameters"`
	Actions    ServerActions             `wst:"actions"`
}
