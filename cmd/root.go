package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
	"github.com/hiAndrewQuinn/cliguard/internal/validator"
	"github.com/spf13/cobra"
)

// Execute runs the root command
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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

	return rootCmd
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Resolve project path
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	// Check if project path exists
	if _, err := os.Stat(absProjectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", absProjectPath)
	}

	// Determine contract path
	if contractPath == "" {
		contractPath = filepath.Join(absProjectPath, "cliguard.yaml")
	} else {
		contractPath, err = filepath.Abs(contractPath)
		if err != nil {
			return fmt.Errorf("failed to resolve contract path: %w", err)
		}
	}

	// Load the contract
	fmt.Printf("Loading contract from: %s\n", contractPath)
	contractSpec, err := contract.Load(contractPath)
	if err != nil {
		return fmt.Errorf("failed to load contract: %w", err)
	}

	// Generate and run the inspector
	fmt.Printf("Inspecting CLI structure in: %s\n", absProjectPath)
	actualStructure, err := inspector.InspectProject(absProjectPath, entrypoint)
	if err != nil {
		return fmt.Errorf("failed to inspect project: %w", err)
	}

	// Validate the actual structure against the contract
	fmt.Println("Validating CLI structure against contract...")
	result := validator.Validate(contractSpec, actualStructure)

	// Report results
	if result.IsValid() {
		fmt.Println("✅ Validation passed! CLI structure matches the contract.")
		return nil
	}

	// Print validation errors
	fmt.Println("❌ Validation failed!")
	fmt.Println()
	result.PrintReport()

	os.Exit(1)
	return nil
}
