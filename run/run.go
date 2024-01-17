package run

import (
	"log"
	"os"
	"path/filepath"
)

type Options struct {
	ConfigPaths     []string
	IncludeAll      bool
	ParameterValues []string
	NoEnvs          bool
	DryRun          bool
}

func Execute(options *Options) {
	if options.IncludeAll {
		newPaths := GetConfigPaths()
		options.ConfigPaths = append(options.ConfigPaths, newPaths...)
	}
	options.ConfigPaths = removeDuplicates(options.ConfigPaths)
}

func GetConfigPaths() []string {
	var paths []string
	home, _ := os.UserHomeDir()

	validateAndAppendPath("wst.yaml", &paths)
	validateAndAppendPath(filepath.Join(home, ".wst/wst.yaml"), &paths)
	validateAndAppendPath(filepath.Join(home, ".config/wst/wst.yaml"), &paths)

	return paths
}

func isPathInPaths(path string, paths []string) bool {
	for _, p := range paths {
		if p == path {
			return true
		}
	}
	return false
}

func validateAndAppendPath(path string, paths *[]string) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if !isPathInPaths(path, *paths) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				log.Println(err)
				return
			}
			*paths = append(*paths, absPath)
		}
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
