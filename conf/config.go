package conf

import (
	"fmt"
	"github.com/spf13/afero"
)

func ExecuteConfigs(configs []string, overwrites map[string]string, instances []string, dryRun bool, fs afero.Fs) error {
	fmt.Println(configs, overwrites, instances, dryRun)
	return nil
}
