// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/wstool/wst/app"
	"strings"
)

// getOverwrites handles overwrites from both the command-line flag and the environment variable
func getOverwrites(overwriteValues []string, noEnvs bool, fnd app.Foundation) map[string]string {
	var overwrites = make(map[string]string)

	// Collect overwrites from flags
	for _, arg := range overwriteValues {
		pair := strings.SplitN(arg, "=", 2)
		if len(pair) != 2 {
			fnd.Logger().Warn("Invalid key-value pair: ", arg)
			continue
		}
		overwrites[pair[0]] = pair[1]
	}

	// Overwrite with environment variables if not disable with --no-envs
	if !noEnvs {
		if val, ok := fnd.LookupEnvVar("WST_OVERWRITE"); ok {
			envVars := strings.Split(val, ":")
			for _, arg := range envVars {
				pair := strings.SplitN(arg, "=", 2)
				if len(pair) != 2 {
					fnd.Logger().Warnf("Invalid environment key-value pair: %s", arg)
					continue
				}
				overwrites[pair[0]] = pair[1]
			}
		}
	}

	return overwrites
}
