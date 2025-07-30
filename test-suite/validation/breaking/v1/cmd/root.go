package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates v1 of the CLI
func NewRootCmd() *cobra.Command {
	var configFile string
	var verbose bool
	var output string

	rootCmd := &cobra.Command{
		Use:   "breaking-test",
		Short: "Test CLI for breaking changes v1",
		Long:  "Version 1 of the CLI to test breaking change detection.",
	}

	// Global flags that will be removed/changed in v2
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Commands that will be modified in v2
	rootCmd.AddCommand(newDeployCmd(&output))
	rootCmd.AddCommand(newBuildCmd())
	rootCmd.AddCommand(newTestCmd())

	return rootCmd
}

func newDeployCmd(output *string) *cobra.Command {
	var environment string
	var force bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "deploy [target]",
		Short: "Deploy the application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Deploying to %s\n", args[0])
		},
	}

	// These flags will change in v2
	cmd.Flags().StringVarP(&environment, "environment", "e", "production", "Target environment")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deployment")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Perform a dry run")
	cmd.Flags().StringVarP(output, "output", "o", "text", "Output format")

	return cmd
}

func newBuildCmd() *cobra.Command {
	var target string
	var optimize bool

	cmd := &cobra.Command{
		Use:     "build",
		Short:   "Build the application",
		Aliases: []string{"compile"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Building application...")
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "linux", "Build target")
	cmd.Flags().BoolVar(&optimize, "optimize", true, "Enable optimizations")

	// This subcommand will be removed in v2
	cmd.AddCommand(newBuildCacheCmd())

	return cmd
}

func newBuildCacheCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cache",
		Short: "Manage build cache",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Managing build cache...")
		},
	}
}

func newTestCmd() *cobra.Command {
	var coverage bool
	var parallel int

	cmd := &cobra.Command{
		Use:   "test [packages...]",
		Short: "Run tests",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Running tests...")
		},
	}

	// Flag type will change from int to bool in v2
	cmd.Flags().BoolVar(&coverage, "coverage", false, "Generate coverage report")
	cmd.Flags().IntVarP(&parallel, "parallel", "p", 4, "Number of parallel tests")

	return cmd
}