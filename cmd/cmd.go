package cmd

import (
	"fmt"
	"github.com/bukka/wst/run"
	"github.com/spf13/cobra"
	"os"
)

func Run() {
	var debug bool
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Executes the predefined configuration",
		Long:  "Constructs the final configuration and executes actions in order",
		Run: func(cmd *cobra.Command, args []string) {
			configPaths, _ := cmd.Flags().GetStringSlice("config")
			includeAll, _ := cmd.Flags().GetBool("all")
			parameterValues, _ := cmd.Flags().GetStringSlice("parameter")
			noEnvs, _ := cmd.Flags().GetBool("no-envs")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			options := run.Options{
				ConfigPaths:     configPaths,
				IncludeAll:      includeAll,
				ParameterValues: parameterValues,
				NoEnvs:          noEnvs,
				DryRun:          dryRun,
			}
			// Add execution code here.
			run.Execute(options)
		},
	}

	runCmd.Flags().StringSliceP("config", "c", []string{}, "List of paths to configuration files")
	runCmd.PersistentFlags().BoolP("all", "a", false, "Include additional configuration files")
	runCmd.PersistentFlags().StringP("parameter", "p", "", "Define specific parameters")
	runCmd.PersistentFlags().Bool("no-envs", false, "Prevent environment variables from superseding parameters")
	runCmd.PersistentFlags().Bool("dry-run", false, "Activate dry-run mode")

	var rootCmd = &cobra.Command{Use: "wst"}
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false,
		"Provide a more detailed output by logging additional debugging information")
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
