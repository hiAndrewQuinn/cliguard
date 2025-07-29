package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var configFile string

	rootCmd := &cobra.Command{
		Use:   "simple-cli",
		Short: "A simple test CLI",
		Long:  "This is a simple CLI for testing cliguard",
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Server management commands",
	}

	var port int
	serverCmd.Flags().IntVarP(&port, "port", "p", 8080, "server port")

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the server",
		Run: func(cmd *cobra.Command, args []string) {
			// Implementation here
		},
	}

	serverCmd.AddCommand(startCmd)
	rootCmd.AddCommand(serverCmd)

	return rootCmd
}
