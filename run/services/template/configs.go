package template

// Config holds the main service configuration
type Config struct {
	User  string
	Group string
	Dirs  struct {
		Run    string
		Script string
	}
}

// Configs are used for config paths
type Configs map[string]string
