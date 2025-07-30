package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates v2 of the CLI with breaking changes
func NewRootCmd() *cobra.Command {
	var settings string  // BREAKING: renamed from 'config'
	var debug bool       // BREAKING: renamed from 'verbose'
	var format string    // BREAKING: renamed from 'output'

	rootCmd := &cobra.Command{
		Use:   "breaking-test",
		Short: "Test CLI for breaking changes v2",
		Long:  "Version 2 of the CLI with breaking changes for testing.",
	}

	// BREAKING: Changed flag names and shortcuts
	rootCmd.PersistentFlags().StringVarP(&settings, "settings", "s", "", "Settings file path") // was --config/-c
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug output")      // was --verbose/-v

	// BREAKING: Removed 'test' command, renamed 'deploy' to 'release'
	rootCmd.AddCommand(newReleaseCmd(&format)) // was 'deploy'
	rootCmd.AddCommand(newBuildCmd())
	// 'test' command removed

	return rootCmd
}

func newReleaseCmd(format *string) *cobra.Command {
	var env string       // BREAKING: renamed from 'environment'
	var skipChecks bool  // BREAKING: renamed from 'force'
	// BREAKING: 'dry-run' flag removed

	cmd := &cobra.Command{
		Use:   "release [target]", // BREAKING: renamed from 'deploy'
		Short: "Release the application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Releasing to %s\n", args[0])
		},
	}

	// BREAKING: Changed flag names and shortcuts
	cmd.Flags().StringVar(&env, "env", "production", "Target environment")                   // was --environment/-e
	cmd.Flags().BoolVar(&skipChecks, "skip-checks", false, "Skip pre-release checks")       // was --force/-f
	cmd.Flags().StringVarP(format, "format", "f", "json", "Output format")                  // was --output/-o, changed shortcut
	// 'dry-run' flag removed

	return cmd
}

func newBuildCmd() *cobra.Command {
	var platform string // BREAKING: renamed from 'target'
	var fast bool       // BREAKING: changed from 'optimize' with inverted logic

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the application",
		// BREAKING: Removed 'compile' alias
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Building application...")
		},
	}

	// BREAKING: Changed flag names and logic
	cmd.Flags().StringVarP(&platform, "platform", "p", "linux", "Build platform")  // was --target/-t
	cmd.Flags().BoolVar(&fast, "fast", false, "Fast build (no optimizations)")     // was --optimize with default true

	// BREAKING: Removed 'cache' subcommand

	return cmd
}