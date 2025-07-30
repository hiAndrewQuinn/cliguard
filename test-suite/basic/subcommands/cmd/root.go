package cmd

import (
	"github.com/spf13/cobra"
)

var (
	globalVerbose bool
	globalConfig  string
)

// NewRootCmd creates the root command with subcommands
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "subcmd-test",
		Short: "A CLI with subcommands for testing",
		Long:  "This CLI demonstrates various subcommand patterns for testing cliguard.",
	}

	// Global persistent flags
	rootCmd.PersistentFlags().BoolVarP(&globalVerbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&globalConfig, "config", "", "Config file path")

	// Add subcommands
	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newDeleteCmd())
	rootCmd.AddCommand(newConfigCmd())

	return rootCmd
}

func newCreateCmd() *cobra.Command {
	var createType string
	var createName string
	var force bool

	cmd := &cobra.Command{
		Use:   "create [resource]",
		Short: "Create a new resource",
		Long:  "Create various types of resources in the system.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Creating %s resource: %s\n", createType, args[0])
		},
	}

	cmd.Flags().StringVarP(&createType, "type", "t", "default", "Resource type to create")
	cmd.Flags().StringVar(&createName, "name", "", "Name for the resource")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation even if exists")

	// Add subcommands to create
	cmd.AddCommand(newCreateUserCmd())
	cmd.AddCommand(newCreateProjectCmd())

	return cmd
}

func newCreateUserCmd() *cobra.Command {
	var email string
	var admin bool

	cmd := &cobra.Command{
		Use:   "user [username]",
		Short: "Create a new user",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Creating user: %s\n", args[0])
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "User email address")
	cmd.Flags().BoolVar(&admin, "admin", false, "Grant admin privileges")
	cmd.MarkFlagRequired("email")

	return cmd
}

func newCreateProjectCmd() *cobra.Command {
	var template string
	var private bool

	cmd := &cobra.Command{
		Use:   "project [name]",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Creating project: %s\n", args[0])
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "default", "Project template to use")
	cmd.Flags().BoolVar(&private, "private", false, "Make project private")

	return cmd
}

func newListCmd() *cobra.Command {
	var format string
	var limit int
	var all bool

	cmd := &cobra.Command{
		Use:     "list [resource]",
		Short:   "List resources",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Listing resources...")
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, yaml)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Maximum number of items to list")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "List all items including archived")

	return cmd
}

func newDeleteCmd() *cobra.Command {
	var force bool
	var cascade bool

	cmd := &cobra.Command{
		Use:     "delete [resource]",
		Short:   "Delete resources",
		Aliases: []string{"del", "rm"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Deleting: %v\n", args)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion without confirmation")
	cmd.Flags().BoolVar(&cascade, "cascade", false, "Delete dependent resources")

	return cmd
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "View and modify configuration settings.",
	}

	// Add config subcommands
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())

	return cmd
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get configuration value",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Config %s = <value>\n", args[0])
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set configuration value",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Setting %s = %s\n", args[0], args[1])
		},
	}
}