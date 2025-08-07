package contract_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"gopkg.in/yaml.v3"
)

// createLargeContract creates a YAML contract file with the specified number of commands
func createLargeContract(b *testing.B, numCommands int) string {
	b.Helper()
	
	tmpDir := b.TempDir()
	contractPath := filepath.Join(tmpDir, "contract.yaml")
	
	c := &contract.Contract{
		Use:      "benchmark-cli",
		Short:    "Large contract for benchmarking",
		Commands: make([]contract.Command, 0, numCommands),
		Flags: []contract.Flag{
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
		},
	}
	
	// Generate commands
	for i := 0; i < numCommands; i++ {
		cmd := contract.Command{
			Use:   fmt.Sprintf("command%d", i),
			Short: fmt.Sprintf("Test command %d with description", i),
			Flags: make([]contract.Flag, 0, 10),
		}
		
		// Add flags to each command
		for j := 0; j < 10; j++ {
			flag := contract.Flag{
				Name:  fmt.Sprintf("flag%d", j),
				Type:  getRandomType(j),
				Usage: fmt.Sprintf("Test flag %d description", j),
			}
			cmd.Flags = append(cmd.Flags, flag)
		}
		
		// Add subcommands to every 5th command
		if i%5 == 0 && i > 0 {
			for k := 0; k < 5; k++ {
				subcmd := contract.Command{
					Use:   fmt.Sprintf("subcommand%d", k),
					Short: fmt.Sprintf("Subcommand %d description", k),
					Flags: make([]contract.Flag, 0, 3),
				}
				
				for l := 0; l < 3; l++ {
					flag := contract.Flag{
						Name:  fmt.Sprintf("subflag%d", l),
						Type:  getRandomType(l),
						Usage: fmt.Sprintf("Subcommand flag %d", l),
					}
					subcmd.Flags = append(subcmd.Flags, flag)
				}
				
				cmd.Commands = append(cmd.Commands, subcmd)
			}
		}
		
		c.Commands = append(c.Commands, cmd)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		b.Fatal(err)
	}
	
	if err := os.WriteFile(contractPath, data, 0644); err != nil {
		b.Fatal(err)
	}
	
	return contractPath
}

func getRandomType(index int) string {
	types := []string{"string", "bool", "int", "float64", "int64"}
	return types[index%len(types)]
}

// createComplexContract creates a contract with complex nested structure
func createComplexContract(b *testing.B) string {
	b.Helper()
	
	tmpDir := b.TempDir()
	contractPath := filepath.Join(tmpDir, "complex-contract.yaml")
	
	c := &contract.Contract{
		Use:   "complex-cli",
		Short: "Complex contract with nested commands",
		Commands: []contract.Command{
			{
				Use:   "deploy",
				Short: "Deploy application",
				Flags: []contract.Flag{
					{Name: "environment", Type: "string", Usage: "Deployment environment"},
					{Name: "replicas", Type: "int", Usage: "Number of replicas"},
					{Name: "tags", Type: "string", Usage: "Deployment tags"},
				},
				Commands: []contract.Command{
					{
						Use:   "rollback",
						Short: "Rollback deployment",
						Flags: []contract.Flag{
							{Name: "version", Type: "string", Usage: "Version to rollback to"},
						},
					},
				},
			},
		},
		Flags: []contract.Flag{
			{Name: "profile", Type: "string", Usage: "AWS profile"},
		},
	}
	
	// Add more commands with various complexity
	for i := 0; i < 20; i++ {
		c.Commands = append(c.Commands, generateComplexCommand(i))
	}
	
	data, err := yaml.Marshal(c)
	if err != nil {
		b.Fatal(err)
	}
	
	if err := os.WriteFile(contractPath, data, 0644); err != nil {
		b.Fatal(err)
	}
	
	return contractPath
}

func generateComplexCommand(index int) contract.Command {
	return contract.Command{
		Use:   fmt.Sprintf("complex-cmd-%d", index),
		Short: fmt.Sprintf("Complex command %d with flags", index),
		Flags: []contract.Flag{
			{Name: "input-file", Type: "string", Usage: "Input file path"},
			{Name: "timeout", Type: "int", Usage: "Operation timeout in seconds"},
			{Name: "memory-limit", Type: "string", Usage: "Memory limit"},
		},
	}
}

// Benchmark tests

func BenchmarkLoadSmallContract(b *testing.B) {
	contractPath := createLargeContract(b, 10)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := contract.Load(contractPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadMediumContract(b *testing.B) {
	contractPath := createLargeContract(b, 50)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := contract.Load(contractPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadLargeContract(b *testing.B) {
	contractPath := createLargeContract(b, 100)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := contract.Load(contractPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadExtraLargeContract(b *testing.B) {
	contractPath := createLargeContract(b, 500)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := contract.Load(contractPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadComplexContract(b *testing.B) {
	contractPath := createComplexContract(b)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := contract.Load(contractPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark contract serialization (for caching scenarios)
func BenchmarkSerializeContract(b *testing.B) {
	c := &contract.Contract{
		Use:      "bench-cli",
		Short:    "Benchmark CLI",
		Commands: make([]contract.Command, 100),
	}
	
	for i := 0; i < 100; i++ {
		c.Commands[i] = generateComplexCommand(i)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := yaml.Marshal(c)
		if err != nil {
			b.Fatal(err)
		}
	}
}