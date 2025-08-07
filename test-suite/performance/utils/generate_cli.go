package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CLISize represents the size categories for CLI testing
type CLISize string

const (
	Small  CLISize = "small"  // 5-10 commands
	Medium CLISize = "medium" // 50-100 commands
	Large  CLISize = "large"  // 500+ commands
)

// GenerateCLIProject creates a Go CLI project with the specified number of commands
func GenerateCLIProject(dir string, size CLISize) error {
	numCommands := getCommandCount(size)
	
	// Create directory structure
	if err := os.MkdirAll(filepath.Join(dir, "cmd"), 0755); err != nil {
		return err
	}
	
	// Generate go.mod
	goModContent := `module test-cli

go 1.21

require (
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
)
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return err
	}
	
	// Generate main.go
	mainContent := `package main

import (
	"os"
	"test-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainContent), 0644); err != nil {
		return err
	}
	
	// Generate root command
	rootContent := generateRootCommand(numCommands)
	if err := os.WriteFile(filepath.Join(dir, "cmd", "root.go"), []byte(rootContent), 0644); err != nil {
		return err
	}
	
	// Generate individual command files
	for i := 0; i < numCommands; i++ {
		cmdContent := generateCommand(i, size == Large)
		filename := filepath.Join(dir, "cmd", fmt.Sprintf("cmd_%03d.go", i))
		if err := os.WriteFile(filename, []byte(cmdContent), 0644); err != nil {
			return err
		}
	}
	
	return nil
}

func getCommandCount(size CLISize) int {
	switch size {
	case Small:
		return 10
	case Medium:
		return 75
	case Large:
		return 500
	default:
		return 10
	}
}

func generateRootCommand(numCommands int) string {
	template := `package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "test-cli",
	Short: "A test CLI with %d commands",
	Long:  "This is a test CLI application for performance benchmarking with %d commands",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.test-cli.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("output", "json", "output format (json, yaml, text)")
	
	// Add all commands
%s}

func NewRootCmd() *cobra.Command {
	return rootCmd
}
`
	
	var cmdRegistrations string
	for i := 0; i < numCommands; i++ {
		cmdRegistrations += fmt.Sprintf("\trootCmd.AddCommand(newCmd%03d())\n", i)
	}
	
	return fmt.Sprintf(template, numCommands, numCommands, cmdRegistrations)
}

func generateCommand(index int, addMoreFlags bool) string {
	flagCount := 5
	if addMoreFlags {
		flagCount = 15
	}
	
	template := `package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func newCmd%03d() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command-%03d",
		Short: "Command %d for performance testing",
		Long:  "This is command %d of the test CLI, used for performance benchmarking and testing at scale",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simulate command execution
			fmt.Println("Executing command %d")
			return nil
		},
	}
	
	// Add flags
%s
	
	// Mark some flags as required
	_ = cmd.MarkFlagRequired("input")
	
%s
	
	return cmd
}
`
	
	var flags string
	for i := 0; i < flagCount; i++ {
		flags += generateFlag(i)
	}
	
	// Add subcommands for every 10th command
	var subcommands string
	if index%10 == 0 && index > 0 {
		subcommands = generateSubcommands(index)
	}
	
	return fmt.Sprintf(template, index, index, index, index, index, flags, subcommands)
}

func generateFlag(index int) string {
	flagTypes := []string{
		`cmd.Flags().String("flag-%d", "", "String flag %d")`,
		`cmd.Flags().Bool("flag-%d", false, "Boolean flag %d")`,
		`cmd.Flags().Int("flag-%d", 0, "Integer flag %d")`,
		`cmd.Flags().Float64("flag-%d", 0.0, "Float flag %d")`,
		`cmd.Flags().StringSlice("flag-%d", nil, "String slice flag %d")`,
	}
	
	// Special flags
	if index == 0 {
		return fmt.Sprintf("\tcmd.Flags().String(\"input\", \"\", \"Input file path\")\n")
	}
	if index == 1 {
		return fmt.Sprintf("\tcmd.Flags().String(\"output\", \"\", \"Output file path\")\n")
	}
	if index == 2 {
		return fmt.Sprintf("\tcmd.Flags().Bool(\"force\", false, \"Force operation without confirmation\")\n")
	}
	
	flagType := flagTypes[index%len(flagTypes)]
	return fmt.Sprintf("\t"+flagType+"\n", index, index)
}

func generateSubcommands(parentIndex int) string {
	template := `	// Add subcommands
	for i := 0; i < 5; i++ {
		subcmd := &cobra.Command{
			Use:   fmt.Sprintf("sub-%%d", i),
			Short: fmt.Sprintf("Subcommand %%d of command %d", i),
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		subcmd.Flags().String("sub-option", "", "Subcommand option")
		subcmd.Flags().Bool("sub-flag", false, "Subcommand flag")
		cmd.AddCommand(subcmd)
	}
`
	return fmt.Sprintf(template, parentIndex)
}