package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hiAndrewQuinn/cliguard/internal/discovery"
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
	interactive  bool
	force        bool
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

	validateCmd.Flags().StringVar(&projectPath, "project-path", "", "Path to the root of the target Go project (defaults to current directory)")
	validateCmd.Flags().StringVar(&contractPath, "contract", "", "Path to the contract file (defaults to cliguard.yaml in project path)")
	validateCmd.Flags().StringVar(&entrypoint, "entrypoint", "", "The function that returns the root command (e.g., github.com/user/repo/cmd.NewRootCmd)")
	validateCmd.Flags().BoolVar(&force, "force", false, "Force operation even with unsupported CLI frameworks")

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

	generateCmd.Flags().StringVar(&projectPath, "project-path", "", "Path to the root of the target Go project (defaults to current directory)")
	generateCmd.Flags().StringVar(&entrypoint, "entrypoint", "", "The function that returns the root command (e.g., github.com/user/repo/cmd.NewRootCmd)")
	generateCmd.Flags().BoolVar(&force, "force", false, "Force operation even with unsupported CLI frameworks")

	rootCmd.AddCommand(generateCmd)

	// Discover command
	discoverCmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover CLI entrypoints in a Go project",
		Long: `Discover searches a Go project for potential CLI entrypoints by analyzing
common patterns used by various CLI frameworks (Cobra, urfave/cli, flag, etc.).
This helps you quickly identify where commands are defined in unfamiliar codebases.`,
		RunE: runDiscover,
	}

	discoverCmd.Flags().StringVar(&projectPath, "project-path", "", "Path to the root of the target Go project (required)")
	discoverCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode: prompt to select from multiple candidates")
	discoverCmd.Flags().BoolVar(&force, "force", false, "Force operation even with unsupported CLI frameworks")

	_ = discoverCmd.MarkFlagRequired("project-path")

	rootCmd.AddCommand(discoverCmd)

	return rootCmd
}

// ValidateRunner interface for dependency injection
type ValidateRunner interface {
	Run(cmd *cobra.Command, projectPath, contractPath, entrypoint string, force bool) error
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
func (r *DefaultValidateRunner) Run(cmd *cobra.Command, projectPath, contractPath, entrypoint string, force bool) error {
	// Check if entrypoint is provided and detect framework
	if entrypoint != "" {
		framework, err := discovery.DetectEntrypointFramework(projectPath, entrypoint, nil)
		if err == nil && framework != "" && framework != "cobra" {
			if !force {
				return fmt.Errorf("Error: cliguard currently only supports Cobra CLIs. Support for %s is coming soon!\nUse --force to proceed anyway (may produce unexpected results)", framework)
			}
			cmd.Printf("⚠️  Warning: Proceeding with unsupported framework %s. Results may be unreliable.\n\n", framework)
		}
	}

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
	// Default to current directory if no project path specified
	path := projectPath
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}
	return validateRunner.Run(cmd, path, contractPath, entrypoint, force)
}

// GenerateRunner interface for dependency injection
type GenerateRunner interface {
	Run(cmd *cobra.Command, projectPath, entrypoint string, force bool) error
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
func (r *DefaultGenerateRunner) Run(cmd *cobra.Command, projectPath, entrypoint string, force bool) error {
	// Check if entrypoint is provided and detect framework
	if entrypoint != "" {
		framework, err := discovery.DetectEntrypointFramework(projectPath, entrypoint, nil)
		if err == nil && framework != "" && framework != "cobra" {
			if !force {
				return fmt.Errorf("Error: cliguard currently only supports Cobra CLIs. Support for %s is coming soon!\nUse --force to proceed anyway (may produce unexpected results)", framework)
			}
			cmd.Printf("⚠️  Warning: Proceeding with unsupported framework %s. Results may be unreliable.\n\n", framework)
		}
	}

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
	// Default to current directory if no project path specified
	path := projectPath
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}
	return generateRunner.Run(cmd, path, entrypoint, force)
}

// DiscoverRunner interface for dependency injection
type DiscoverRunner interface {
	Run(cmd *cobra.Command, projectPath string, interactive bool, force bool) error
}

// DefaultDiscoverRunner is the default implementation
type DefaultDiscoverRunner struct{}

// NewDefaultDiscoverRunner creates a new default runner
func NewDefaultDiscoverRunner() *DefaultDiscoverRunner {
	return &DefaultDiscoverRunner{}
}

// Run executes the discovery
func (r *DefaultDiscoverRunner) Run(cmd *cobra.Command, projectPath string, interactive bool, force bool) error {
	// Convert to absolute path if needed
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	discoverer := discovery.NewDiscoverer(absPath, nil)

	fmt.Fprintf(cmd.OutOrStdout(), "Searching for CLI entrypoints in: %s\n\n", projectPath)

	candidates, err := discoverer.DiscoverEntrypoints()
	if err != nil {
		return fmt.Errorf("failed to discover entrypoints: %w", err)
	}

	// Handle interactive mode
	if interactive && len(candidates) > 1 {
		selector := discovery.NewInteractiveSelector(cmd.InOrStdin(), cmd.OutOrStdout())
		selected, err := selector.SelectCandidate(candidates)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nSelected entrypoint:\n%s\n",
			discovery.FormatSelectedEntrypoint(selected))
		return nil
	}

	discovery.PrintCandidates(cmd.OutOrStdout(), candidates, projectPath, force)
	return nil
}

// Global runner for testing
var discoverRunner DiscoverRunner = NewDefaultDiscoverRunner()

func runDiscover(cmd *cobra.Command, args []string) error {
	return discoverRunner.Run(cmd, projectPath, interactive, force)
}
