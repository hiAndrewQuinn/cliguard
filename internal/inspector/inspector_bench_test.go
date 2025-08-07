package inspector_test

import (
	"testing"

	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// createTestCLI creates a test CLI with the specified number of commands
func createTestCLI(numCommands int) *cobra.Command {
	root := &cobra.Command{
		Use:   "test-cli",
		Short: "Test CLI for benchmarking",
	}

	// Add global flags
	root.PersistentFlags().String("config", "", "Config file")
	root.PersistentFlags().Bool("verbose", false, "Verbose output")
	root.PersistentFlags().Bool("debug", false, "Debug mode")

	for i := 0; i < numCommands; i++ {
		cmd := &cobra.Command{
			Use:   "command" + string(rune(i+'0')),
			Short: "Test command",
		}

		// Add flags
		cmd.Flags().String("input", "", "Input file")
		cmd.Flags().String("output", "", "Output file")
		cmd.Flags().Bool("force", false, "Force operation")
		cmd.Flags().Int("count", 10, "Count value")

		// Add subcommands for some commands
		if i%3 == 0 && i > 0 {
			for j := 0; j < 2; j++ {
				subcmd := &cobra.Command{
					Use:   "sub" + string(rune(j+'0')),
					Short: "Test subcommand",
				}
				subcmd.Flags().String("option", "", "Option value")
				cmd.AddCommand(subcmd)
			}
		}

		root.AddCommand(cmd)
	}

	return root
}

// convertCobraToInspected simulates the conversion process
func convertCobraToInspected(cmd *cobra.Command) *inspector.InspectedCLI {
	cli := &inspector.InspectedCLI{
		Use:      cmd.Use,
		Short:    cmd.Short,
		Long:     cmd.Long,
		Flags:    make([]inspector.InspectedFlag, 0),
		Commands: make([]inspector.InspectedCommand, 0),
	}

	// Convert persistent flags
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		cli.Flags = append(cli.Flags, inspector.InspectedFlag{
			Name:       flag.Name,
			Shorthand:  flag.Shorthand,
			Usage:      flag.Usage,
			Type:       flag.Value.Type(),
			Persistent: true,
		})
	})

	// Convert commands
	for _, subcmd := range cmd.Commands() {
		inspectedCmd := convertCobraCommand(subcmd)
		cli.Commands = append(cli.Commands, inspectedCmd)
	}

	return cli
}

func convertCobraCommand(cmd *cobra.Command) inspector.InspectedCommand {
	inspectedCmd := inspector.InspectedCommand{
		Use:      cmd.Use,
		Short:    cmd.Short,
		Long:     cmd.Long,
		Flags:    make([]inspector.InspectedFlag, 0),
		Commands: make([]inspector.InspectedCommand, 0),
	}

	// Convert local flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		inspectedCmd.Flags = append(inspectedCmd.Flags, inspector.InspectedFlag{
			Name:      flag.Name,
			Shorthand: flag.Shorthand,
			Usage:     flag.Usage,
			Type:      flag.Value.Type(),
		})
	})

	// Convert subcommands
	for _, subcmd := range cmd.Commands() {
		subInspectedCmd := convertCobraCommand(subcmd)
		inspectedCmd.Commands = append(inspectedCmd.Commands, subInspectedCmd)
	}

	return inspectedCmd
}

// Benchmark tests for CLI structure conversion

func BenchmarkConvertSmallCLI(b *testing.B) {
	cli := createTestCLI(10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = convertCobraToInspected(cli)
	}
}

func BenchmarkConvertMediumCLI(b *testing.B) {
	cli := createTestCLI(50)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = convertCobraToInspected(cli)
	}
}

func BenchmarkConvertLargeCLI(b *testing.B) {
	cli := createTestCLI(100)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = convertCobraToInspected(cli)
	}
}

func BenchmarkConvertExtraLargeCLI(b *testing.B) {
	cli := createTestCLI(500)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = convertCobraToInspected(cli)
	}
}

// Benchmark flag extraction specifically
func BenchmarkExtractFlags(b *testing.B) {
	root := &cobra.Command{
		Use: "test",
	}

	// Add many flags
	for i := 0; i < 100; i++ {
		root.Flags().String("flag"+string(rune(i)), "", "Test flag")
		root.PersistentFlags().Bool("pflag"+string(rune(i)), false, "Persistent flag")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		flags := make([]inspector.InspectedFlag, 0)

		root.Flags().VisitAll(func(flag *pflag.Flag) {
			flags = append(flags, inspector.InspectedFlag{
				Name:      flag.Name,
				Shorthand: flag.Shorthand,
				Usage:     flag.Usage,
				Type:      flag.Value.Type(),
			})
		})

		root.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			flags = append(flags, inspector.InspectedFlag{
				Name:       flag.Name,
				Shorthand:  flag.Shorthand,
				Usage:      flag.Usage,
				Type:       flag.Value.Type(),
				Persistent: true,
			})
		})
	}
}
