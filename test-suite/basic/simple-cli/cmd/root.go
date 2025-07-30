package cmd

import "github.com/spf13/cobra"

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "simple",
		Short: "A simple CLI for testing",
		Long:  "This is the simplest possible Cobra CLI with just a root command.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Hello from simple CLI!")
		},
	}
}