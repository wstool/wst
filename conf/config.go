package conf

import (
	"fmt"
	"github.com/bukka/wst/app"
)

type Options struct {
	Configs    []string
	Overwrites map[string]string
	Instances  []string
	DryRun     bool
}

func ExecuteConfigs(options Options, env *app.Env) error {
	fmt.Println(options)
	return nil
}
