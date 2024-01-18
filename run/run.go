package run

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

type Options struct {
	ConfigPaths []string
	IncludeAll  bool
	Overwrites  map[string]string
	NoEnvs      bool
	DryRun      bool
	Instances   []string
}

var DefaultsFs = afero.NewOsFs()

func Execute(options *Options, env *app.Env) {
	var configPaths []string
	if options.IncludeAll {
		extraPaths := GetConfigPaths(env)
		configPaths = append(options.ConfigPaths, extraPaths...)
	} else {
		configPaths = options.ConfigPaths
	}
	configPaths = removeDuplicates(configPaths)

	confOptions := conf.Options{
		Configs:    configPaths,
		Overwrites: options.Overwrites,
		Instances:  options.Instances,
		DryRun:     options.DryRun,
	}
	err := conf.ExecuteConfigs(confOptions, env)
	if err != nil {
		return
	}
}

func GetConfigPaths(env *app.Env) []string {
	var paths []string
	home, _ := env.GetUserHomeDir()
	validateAndAppendPath("wst.yaml", &paths, env)
	validateAndAppendPath(filepath.Join(home, ".wst/wst.yaml"), &paths, env)
	validateAndAppendPath(filepath.Join(home, ".config/wst/wst.yaml"), &paths, env)

	return paths
}

func validateAndAppendPath(path string, paths *[]string, env *app.Env) {
	if _, err := env.Fs.Stat(path); !os.IsNotExist(err) {
		*paths = append(*paths, path)
	}
}

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	var result []string

	for v := range elements {
		if !encountered[elements[v]] {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}
	return result
}
