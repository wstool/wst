package cmd

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/run"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"strings"
)

var debug bool
var overwriteValues []string
var logger *zap.Logger

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

			var err error
			if debug {
				logger, err = zap.NewDevelopment()
			} else {
				logger, err = zap.NewProduction()
			}
			if err != nil {
				panic(fmt.Sprintf("Cannot initialize zap logger: %v", err))
			}

			appEnv := &app.Env{
				Logger: logger.Sugar(),
				Fs:     run.DefaultsFs,
			}

			options := &run.Options{
				ConfigPaths: configPaths,
				IncludeAll:  includeAll,
				Overwrites:  getOverwrites(noEnvs, appEnv),
				NoEnvs:      noEnvs,
				DryRun:      dryRun,
				Instances:   args,
			}
			// Add execution code here.
			run.Execute(options, appEnv)
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
func getOverwrites(noEnvs bool, env *app.Env) map[string]string {
	var overwrites = make(map[string]string)

	// Collect overwrites from flags
	for _, arg := range overwriteValues {
		pair := strings.SplitN(arg, "=", 2)
		if len(pair) != 2 {
			env.Logger.Warn("Invalid key-value pair: ", arg)
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
					env.Logger.Warnf("Invalid environment key-value pair: %s", arg)
					continue
				}
				overwrites[pair[0]] = pair[1]
			}
		}
	}

	return overwrites
}
