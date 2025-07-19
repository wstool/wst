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
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/run"
	"go.uber.org/zap"
	"os"
)

func Run() {
	var debug bool
	var runFailed bool
	var overwriteValues []string
	var logger *zap.Logger

	var runCmd = &cobra.Command{
		Use:   "run [instance]...",
		Short: "Executes the predefined configuration",
		Long:  "Constructs the final configuration and executes actions in order",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPaths, _ := cmd.Flags().GetStringSlice("config")
			includeAll, _ := cmd.Flags().GetBool("all")
			preFilter, _ := cmd.Flags().GetBool("pre-filter")
			noEnvs, _ := cmd.Flags().GetBool("no-envs")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			var err error
			if debug {
				logger, err = zap.NewDevelopment()
			} else {
				logger, err = zap.NewProduction()
			}
			if err != nil {
				panic(fmt.Sprintf("Cannot initialize zap logger: %v", err))
			}

			fnd := app.NewFoundation(logger.Sugar(), dryRun)

			options := &run.Options{
				ConfigPaths: configPaths,
				IncludeAll:  includeAll,
				Overwrites:  getOverwrites(overwriteValues, noEnvs, fnd),
				PreFilter:   preFilter,
				NoEnvs:      noEnvs,
				Instances:   args,
			}
			// Add execution code here.
			runFailed = false
			if err = run.CreateRunner(fnd).Execute(options); err != nil {
				runFailed = true
				logger.Error("Unable to execute run operation: ", zap.Error(err))
				if debug {
					fmt.Fprintf(os.Stderr, "\nERROR: %+v\n", err)
				}
			}
			return err
		},
	}

	runCmd.Flags().StringSliceP("config", "c", []string{}, "List of paths to configuration files")
	runCmd.PersistentFlags().BoolP("all", "a", false, "Include additional configuration files")
	runCmd.Flags().StringSliceVarP(&overwriteValues, "overwrite", "o", nil, "Overwrite configuration values")
	runCmd.PersistentFlags().Bool("pre-filter", false, "Whether to filter instances in the initial phase for easier debugging")
	runCmd.PersistentFlags().Bool("no-envs", false, "Prevent environment variables from superseding parameters")
	runCmd.PersistentFlags().Bool("dry-run", false, "Activate dry-run mode")

	var rootCmd = &cobra.Command{Use: "wst"}
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false,
		"Provide a more detailed output by logging additional debugging information")
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		if !runFailed {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
