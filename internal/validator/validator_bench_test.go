package validator_test

import (
	"fmt"
	"testing"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
	"github.com/hiAndrewQuinn/cliguard/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// generateContract creates a test contract with the specified number of commands and flags per command
func generateContract(numCommands, flagsPerCommand int) *contract.Contract {
	c := &contract.Contract{
		Use:      "benchmark-cli",
		Short:    "Benchmark CLI for performance testing",
		Commands: make([]contract.Command, 0, numCommands),
	}

	for i := 0; i < numCommands; i++ {
		cmd := contract.Command{
			Use:   fmt.Sprintf("command%d", i),
			Short: fmt.Sprintf("Test command %d", i),
			Flags: make([]contract.Flag, 0, flagsPerCommand),
		}

		for j := 0; j < flagsPerCommand; j++ {
			flag := contract.Flag{
				Name:  fmt.Sprintf("flag%d", j),
				Type:  "string",
				Usage: fmt.Sprintf("Test flag %d", j),
			}
			cmd.Flags = append(cmd.Flags, flag)
		}

		// Add some subcommands for deeper nesting
		if i%5 == 0 && i > 0 {
			for k := 0; k < 3; k++ {
				subcmd := contract.Command{
					Use:   fmt.Sprintf("subcommand%d", k),
					Short: fmt.Sprintf("Test subcommand %d", k),
					Flags: make([]contract.Flag, 0, 2),
				}
				for l := 0; l < 2; l++ {
					flag := contract.Flag{
						Name:  fmt.Sprintf("subflag%d", l),
						Type:  "bool",
						Usage: fmt.Sprintf("Test subflag %d", l),
					}
					subcmd.Flags = append(subcmd.Flags, flag)
				}
				cmd.Commands = append(cmd.Commands, subcmd)
			}
		}

		c.Commands = append(c.Commands, cmd)
	}

	// Add global flags
	c.Flags = []contract.Flag{
		{
			Name:  "config",
			Type:  "string",
			Usage: "Config file path",
		},
		{
			Name:  "verbose",
			Type:  "bool",
			Usage: "Verbose output",
		},
		{
			Name:  "debug",
			Type:  "bool",
			Usage: "Debug mode",
		},
	}

	return c
}

// generateCLI creates a test CLI structure matching the contract
func generateCLI(numCommands, flagsPerCommand int) *inspector.InspectedCLI {
	rootCmd := &cobra.Command{
		Use:   "benchmark-cli",
		Short: "Benchmark CLI for performance testing",
	}

	// Add global flags
	rootCmd.PersistentFlags().String("config", "", "Config file path")
	rootCmd.PersistentFlags().Bool("verbose", false, "Verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "Debug mode")

	for i := 0; i < numCommands; i++ {
		cmd := &cobra.Command{
			Use:   fmt.Sprintf("command%d", i),
			Short: fmt.Sprintf("Test command %d", i),
		}

		// Add flags to command
		for j := 0; j < flagsPerCommand; j++ {
			cmd.Flags().String(fmt.Sprintf("flag%d", j), fmt.Sprintf("default%d", j), fmt.Sprintf("Test flag %d", j))
		}

		// Add subcommands for deeper nesting
		if i%5 == 0 && i > 0 {
			for k := 0; k < 3; k++ {
				subcmd := &cobra.Command{
					Use:   fmt.Sprintf("subcommand%d", k),
					Short: fmt.Sprintf("Test subcommand %d", k),
				}
				for l := 0; l < 2; l++ {
					subcmd.Flags().Bool(fmt.Sprintf("subflag%d", l), false, fmt.Sprintf("Test subflag %d", l))
				}
				cmd.AddCommand(subcmd)
			}
		}

		rootCmd.AddCommand(cmd)
	}

	return convertCobraToInspected(rootCmd)
}

// convertCobraToInspected converts a cobra command to InspectedCLI
func convertCobraToInspected(cmd *cobra.Command) *inspector.InspectedCLI {
	cli := &inspector.InspectedCLI{
		Use:      cmd.Use,
		Short:    cmd.Short,
		Commands: make([]inspector.InspectedCommand, 0),
		Flags:    make([]inspector.InspectedFlag, 0),
	}

	// Extract flags
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		cli.Flags = append(cli.Flags, inspector.InspectedFlag{
			Name:       flag.Name,
			Type:       flag.Value.Type(),
			Usage:      flag.Usage,
			Persistent: true,
		})
	})

	// Extract commands
	for _, subcmd := range cmd.Commands() {
		if subcmd.Hidden || subcmd.Deprecated != "" {
			continue
		}
		cmdInfo := extractInspectedCommand(subcmd)
		cli.Commands = append(cli.Commands, cmdInfo)
	}

	return cli
}

func extractInspectedCommand(cmd *cobra.Command) inspector.InspectedCommand {
	info := inspector.InspectedCommand{
		Use:      cmd.Name(),
		Short:    cmd.Short,
		Flags:    make([]inspector.InspectedFlag, 0),
		Commands: make([]inspector.InspectedCommand, 0),
	}

	// Extract flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flagInfo := inspector.InspectedFlag{
			Name:  flag.Name,
			Type:  flag.Value.Type(),
			Usage: flag.Usage,
		}
		info.Flags = append(info.Flags, flagInfo)
	})

	// Extract subcommands
	for _, subcmd := range cmd.Commands() {
		if subcmd.Hidden || subcmd.Deprecated != "" {
			continue
		}
		subInfo := extractInspectedCommand(subcmd)
		info.Commands = append(info.Commands, subInfo)
	}

	return info
}

// Benchmark tests

func BenchmarkValidateSmallCLI(b *testing.B) {
	contractDef := generateContract(10, 2) // 10 commands, 2 flags each
	cli := generateCLI(10, 2)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(contractDef, cli)
	}
}

func BenchmarkValidateMediumCLI(b *testing.B) {
	contractDef := generateContract(50, 5) // 50 commands, 5 flags each
	cli := generateCLI(50, 5)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(contractDef, cli)
	}
}

func BenchmarkValidateLargeCLI(b *testing.B) {
	contractDef := generateContract(100, 10) // 100 commands, 10 flags each
	cli := generateCLI(100, 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(contractDef, cli)
	}
}

func BenchmarkValidateExtraLargeCLI(b *testing.B) {
	contractDef := generateContract(500, 15) // 500 commands, 15 flags each
	cli := generateCLI(500, 15)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(contractDef, cli)
	}
}

// Benchmark for deeply nested commands
func BenchmarkValidateDeeplyNestedCLI(b *testing.B) {
	// Create a CLI with deep nesting (5 levels deep, 3 commands per level)
	c := &contract.Contract{
		Use:      "deep-cli",
		Short:    "Deeply nested CLI",
		Commands: make([]contract.Command, 0),
	}

	// Helper to create nested commands
	var createNested func(depth int) []contract.Command
	createNested = func(depth int) []contract.Command {
		if depth == 0 {
			return nil
		}
		cmds := make([]contract.Command, 3)
		for i := 0; i < 3; i++ {
			cmds[i] = contract.Command{
				Use:   fmt.Sprintf("level%d-cmd%d", depth, i),
				Short: fmt.Sprintf("Command at level %d", depth),
				Flags: []contract.Flag{
					{Name: "flag1", Type: "string", Usage: "Test flag"},
					{Name: "flag2", Type: "bool", Usage: "Test flag"},
				},
				Commands: createNested(depth - 1),
			}
		}
		return cmds
	}

	c.Commands = createNested(5)

	// Create matching CLI structure
	rootCmd := &cobra.Command{Use: "deep-cli", Short: "Deeply nested CLI"}
	
	var addNested func(*cobra.Command, int)
	addNested = func(parent *cobra.Command, depth int) {
		if depth == 0 {
			return
		}
		for i := 0; i < 3; i++ {
			cmd := &cobra.Command{
				Use:   fmt.Sprintf("level%d-cmd%d", depth, i),
				Short: fmt.Sprintf("Command at level %d", depth),
			}
			cmd.Flags().String("flag1", "", "Test flag")
			cmd.Flags().Bool("flag2", false, "Test flag")
			parent.AddCommand(cmd)
			addNested(cmd, depth-1)
		}
	}
	
	addNested(rootCmd, 5)
	cli := convertCobraToInspected(rootCmd)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(c, cli)
	}
}