package cmd

import (
	"fmt"
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
			// Add execution code here.
			fmt.Println("Running command...")
		},
	}

	runCmd.PersistentFlags().StringP("config", "c", "wst.yaml",
		"Path to the configuration file")
	runCmd.PersistentFlags().BoolP("all", "a", false,
		"Include additional configuration files")
	runCmd.PersistentFlags().StringP("parameter", "p", "",
		"Define specific parameters")
	runCmd.PersistentFlags().Bool("no-envs", false,
		"Prevent environment variables from superseding parameters")
	runCmd.PersistentFlags().Bool("dry-run", false,
		"Activate dry-run mode")

	var rootCmd = &cobra.Command{Use: "wst"}
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false,
		"Provide a more detailed output by logging additional debugging information")
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
