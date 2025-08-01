package cmd

import (
	"fmt"
	
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
			// Get port from parent command
			parentPort, _ := cmd.Parent().Flags().GetInt("port")
			fmt.Printf("Starting server on port %d\n", parentPort)
			if configFile != "" {
				fmt.Printf("Using config file: %s\n", configFile)
			}
			fmt.Println("Server configuration complete. (In a real implementation, the server would start here)")
		},
	}

	serverCmd.AddCommand(startCmd)
	rootCmd.AddCommand(serverCmd)

	return rootCmd
}
