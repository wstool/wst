package run

import (
	"github.com/bukka/wst/conf"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

type Options struct {
	ConfigPaths     []string
	IncludeAll      bool
	ParameterValues []string
	NoEnvs          bool
	DryRun          bool
	Instances       []string
}

var DefaultsFs = afero.NewOsFs()

func Execute(options *Options, fs afero.Fs) {
	var configPaths []string
	if options.IncludeAll {
		extraPaths := GetConfigPaths(fs)
		configPaths = append(options.ConfigPaths, extraPaths...)
	} else {
		configPaths = options.ConfigPaths
	}
	configPaths = removeDuplicates(configPaths)

	err := conf.ExecuteConfigs(configPaths, options.Instances, options.ParameterValues, options.DryRun, fs)
	if err != nil {
		return
	}
}

func GetConfigPaths(fs afero.Fs) []string {
	var paths []string
	home, _ := os.UserHomeDir()

	validateAndAppendPath("wst.yaml", &paths, fs)
	validateAndAppendPath(filepath.Join(home, ".wst/wst.yaml"), &paths, fs)
	validateAndAppendPath(filepath.Join(home, ".config/wst/wst.yaml"), &paths, fs)

	return paths
}

func validateAndAppendPath(path string, paths *[]string, fs afero.Fs) {
	if _, err := fs.Stat(path); !os.IsNotExist(err) {
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
