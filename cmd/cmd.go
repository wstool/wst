package cmd

import (
	"fmt"
	"github.com/bukka/wst/run"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var debug bool
var overwriteValues []string

func Run() {
	var runCmd = &cobra.Command{
		Use:   "run [instance]...",
		Short: "Executes the predefined configuration",
		Long:  "Constructs the final configuration and executes actions in order",
		Run: func(cmd *cobra.Command, args []string) {
			configPaths, _ := cmd.Flags().GetStringSlice("config")
			includeAll, _ := cmd.Flags().GetBool("all")
			noEnvs, _ := cmd.Flags().GetBool("no-envs")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			options := run.Options{
				ConfigPaths: configPaths,
				IncludeAll:  includeAll,
				Overwrites:  getOverwrites(noEnvs),
				NoEnvs:      noEnvs,
				DryRun:      dryRun,
				Instances:   args,
			}
			// Add execution code here.
			run.Execute(&options, run.DefaultsFs)
		},
	}

	runCmd.Flags().StringSliceP("config", "c", []string{}, "List of paths to configuration files")
	runCmd.PersistentFlags().BoolP("all", "a", false, "Include additional configuration files")
	runCmd.Flags().StringSliceVarP(&overwriteValues, "overwrite", "o", nil, "Overwrite configuration values")
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

// getOverwrites handles overwrites from both the command-line flag and the environment variable
func getOverwrites(noEnvs bool) map[string]string {
	var overwrites = make(map[string]string)

	// Collect overwrites from flags
	for _, arg := range overwriteValues {
		pair := strings.SplitN(arg, "=", 2)
		if len(pair) != 2 {
			fmt.Println("Invalid key-value pair:", arg)
			continue
		}
		overwrites[pair[0]] = pair[1]
	}

	// Overwrite with environment variables if not disable with --no-envs
	if !noEnvs {
		if val, ok := os.LookupEnv("WST_OVERWRITE"); ok {
			envVars := strings.Split(val, ",")
			for _, arg := range envVars {
				pair := strings.SplitN(arg, "=", 2)
				if len(pair) != 2 {
					fmt.Printf("Invalid environment key-value pair: %s\n", arg)
					continue
				}
				overwrites[pair[0]] = pair[1]
			}
		}
	}

	return overwrites
}
