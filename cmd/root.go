package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/hiAndrewQuinn/cliguard/internal/service"
	"github.com/spf13/cobra"
)

// Execute runs the root command
func Execute() {
	ExecuteWithWriter(os.Stderr)
}

// ExecuteWithWriter runs the root command with a custom writer for testing
func ExecuteWithWriter(errWriter io.Writer) {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(errWriter, err)
		os.Exit(1)
	}
}

var (
	projectPath  string
	contractPath string
	entrypoint   string
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cliguard",
		Short: "A contract-based validation tool for Cobra CLIs",
		Long: `Cliguard validates Cobra command structures against a YAML contract file.
It ensures your CLI commands, flags, and structure remain consistent over time.`,
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a Cobra CLI against a contract file",
		Long: `Validate inspects a Go project's Cobra command structure and validates
it against a YAML contract file. This ensures the CLI's structure, commands,
and flags match the expected specification.`,
		RunE: runValidate,
	}

	validateCmd.Flags().StringVar(&projectPath, "project-path", "", "Path to the root of the target Go project (required)")
	validateCmd.Flags().StringVar(&contractPath, "contract", "", "Path to the contract file (defaults to cliguard.yaml in project path)")
	validateCmd.Flags().StringVar(&entrypoint, "entrypoint", "", "The function that returns the root command (e.g., github.com/user/repo/cmd.NewRootCmd)")

	validateCmd.MarkFlagRequired("project-path")

	rootCmd.AddCommand(validateCmd)

	// Generate command
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a contract file from a Cobra CLI",
		Long: `Generate inspects a Go project's Cobra command structure and generates
a YAML contract file that can be used for validation. This is useful for
creating an initial contract from an existing CLI.`,
		RunE: runGenerate,
	}

	generateCmd.Flags().StringVar(&projectPath, "project-path", "", "Path to the root of the target Go project (required)")
	generateCmd.Flags().StringVar(&entrypoint, "entrypoint", "", "The function that returns the root command (e.g., github.com/user/repo/cmd.NewRootCmd)")

	generateCmd.MarkFlagRequired("project-path")

	rootCmd.AddCommand(generateCmd)

	return rootCmd
}

// ValidateRunner interface for dependency injection
type ValidateRunner interface {
	Run(cmd *cobra.Command, projectPath, contractPath, entrypoint string) error
}

// DefaultValidateRunner is the default implementation
type DefaultValidateRunner struct {
	service *service.ValidateService
}

// NewDefaultValidateRunner creates a new default runner
func NewDefaultValidateRunner() *DefaultValidateRunner {
	return &DefaultValidateRunner{
		service: service.NewValidateService(),
	}
}

// Run executes the validation
func (r *DefaultValidateRunner) Run(cmd *cobra.Command, projectPath, contractPath, entrypoint string) error {
	opts := service.ValidateOptions{
		ProjectPath:  projectPath,
		ContractPath: contractPath,
		Entrypoint:   entrypoint,
	}

	// Print progress messages
	if contractPath == "" {
		contractPath = "cliguard.yaml in project path"
	}
	cmd.Printf("Loading contract from: %s\n", contractPath)
	cmd.Printf("Inspecting CLI structure in: %s\n", projectPath)
	cmd.Println("Validating CLI structure against contract...")

	// Run validation
	result, err := r.service.Validate(opts)
	if err != nil {
		return err
	}

	// Report results
	if result.Success {
		cmd.Println("✅ Validation passed! CLI structure matches the contract.")
		return nil
	}

	// Print validation errors
	cmd.Println("❌ Validation failed!")
	cmd.Println()
	result.Result.PrintReport()

	os.Exit(1)
	return nil
}

// Global runner for testing
var validateRunner ValidateRunner = NewDefaultValidateRunner()

func runValidate(cmd *cobra.Command, args []string) error {
	return validateRunner.Run(cmd, projectPath, contractPath, entrypoint)
}

// GenerateRunner interface for dependency injection
type GenerateRunner interface {
	Run(cmd *cobra.Command, projectPath, entrypoint string) error
}

// DefaultGenerateRunner is the default implementation
type DefaultGenerateRunner struct {
	service *service.GenerateService
}

// NewDefaultGenerateRunner creates a new default runner
func NewDefaultGenerateRunner() *DefaultGenerateRunner {
	return &DefaultGenerateRunner{
		service: service.NewGenerateService(),
	}
}

// Run executes the generation
func (r *DefaultGenerateRunner) Run(cmd *cobra.Command, projectPath, entrypoint string) error {
	opts := service.GenerateOptions{
		ProjectPath: projectPath,
		Entrypoint:  entrypoint,
	}

	// Run generation
	yamlContent, err := r.service.Generate(opts)
	if err != nil {
		return err
	}

	// Print YAML to stdout
	fmt.Print(yamlContent)
	return nil
}

// Global runner for testing
var generateRunner GenerateRunner = NewDefaultGenerateRunner()

func runGenerate(cmd *cobra.Command, args []string) error {
	return generateRunner.Run(cmd, projectPath, entrypoint)
}
