package cmd

import (
	"github.com/bukka/wst/app"
	"strings"
)

// getOverwrites handles overwrites from both the command-line flag and the environment variable
func getOverwrites(overwriteValues []string, noEnvs bool, env app.Env) map[string]string {
	var overwrites = make(map[string]string)

	// Collect overwrites from flags
	for _, arg := range overwriteValues {
		pair := strings.SplitN(arg, "=", 2)
		if len(pair) != 2 {
			env.Logger().Warn("Invalid key-value pair: ", arg)
			continue
		}
		overwrites[pair[0]] = pair[1]
	}

	// Overwrite with environment variables if not disable with --no-envs
	if !noEnvs {
		if val, ok := env.LookupEnvVar("WST_OVERWRITE"); ok {
			envVars := strings.Split(val, ",")
			for _, arg := range envVars {
				pair := strings.SplitN(arg, "=", 2)
				if len(pair) != 2 {
					env.Logger().Warnf("Invalid environment key-value pair: %s", arg)
					continue
				}
				overwrites[pair[0]] = pair[1]
			}
		}
	}

	return overwrites
}
