package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
	"github.com/hiAndrewQuinn/cliguard/internal/validator"
)

// ValidateService orchestrates the validation process by coordinating
// between contract loading, CLI inspection, and validation.
//
// The service allows dependency injection for testing:
//
//	svc := &ValidateService{
//	    ContractLoader: mockLoader,
//	    Inspector:      mockInspector,
//	}
type ValidateService struct {
	// ContractLoader loads contract specifications from YAML files.
	// Defaults to contract.Load
	ContractLoader func(string) (*contract.Contract, error)
	
	// Inspector analyzes Go projects to extract CLI structure.
	// Defaults to inspector.InspectProject
	Inspector func(string, string) (*inspector.InspectedCLI, error)
}

// NewValidateService creates a new validation service with default dependencies.
//
// Example:
//
//	svc := NewValidateService()
//	result, err := svc.Validate(ValidateOptions{
//	    ProjectPath:  "./my-cli",
//	    ContractPath: "./my-cli/cliguard.yaml",
//	    Entrypoint:   "cmd.NewRootCmd",
//	})
func NewValidateService() *ValidateService {
	return &ValidateService{
		ContractLoader: contract.Load,
		Inspector:      inspector.InspectProject,
	}
}

// ValidateOptions contains the options for validation.
//
// Example:
//
//	opts := ValidateOptions{
//	    ProjectPath:  ".",                    // Current directory
//	    ContractPath: "./cliguard.yaml",      // Contract file
//	    Entrypoint:   "cmd.NewRootCmd",       // CLI constructor function
//	}
type ValidateOptions struct {
	// ProjectPath is the path to the Go project to validate (required).
	// Can be absolute or relative path.
	ProjectPath string
	
	// ContractPath is the path to the contract YAML file (optional).
	// If empty, defaults to "cliguard.yaml" in the project directory.
	ContractPath string
	
	// Entrypoint is the function that creates the root command (required).
	// Format: "package.Function" or "receiver.Method"
	// Example: "cmd.NewRootCmd" or "(*App).NewRootCmd"
	Entrypoint string
}

// ValidateResult contains the result of validation.
//
// Example usage:
//
//	result, err := svc.Validate(opts)
//	if err != nil {
//	    return err // Failed to run validation
//	}
//	if !result.Success {
//	    fmt.Println(result.Result.FormatReport())
//	    os.Exit(1)
//	}
type ValidateResult struct {
	// Success indicates whether validation passed (true) or failed (false)
	Success bool
	
	// Result contains detailed validation results including all errors found
	Result *validator.ValidationResult
	
	// Error contains any error that prevented validation from running
	// (different from validation failures)
	Error error
}

// Validate performs the validation by loading the contract, inspecting the CLI,
// and comparing them for discrepancies.
//
// The validation process:
// 1. Loads the contract from the specified YAML file
// 2. Inspects the Go project to extract the actual CLI structure
// 3. Validates the actual structure against the contract
// 4. Returns detailed results including all validation errors
//
// Example:
//
//	svc := NewValidateService()
//	result, err := svc.Validate(ValidateOptions{
//	    ProjectPath:  "./examples/simple-cli",
//	    ContractPath: "./examples/simple-cli/cliguard.yaml",
//	    Entrypoint:   "main.NewRootCmd",
//	})
//	if err != nil {
//	    log.Fatal(err) // Failed to run validation
//	}
//	if !result.Success {
//	    fmt.Println("Validation failed:")
//	    for _, err := range result.Result.Errors {
//	        fmt.Printf("  - %s: %s\n", err.Path, err.Message)
//	    }
//	}
//
// Returns an error if validation cannot be performed (e.g., file not found,
// build failure). Validation failures are indicated by Success=false in the
// result, not by returning an error.
func (s *ValidateService) Validate(opts ValidateOptions) (*ValidateResult, error) {
	// Resolve project path
	absProjectPath, err := filepath.Abs(opts.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %w", err)
	}

	// Check if project path exists
	if _, err := os.Stat(absProjectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project path does not exist: %s", absProjectPath)
	}

	// Determine contract path
	contractPath := opts.ContractPath
	if contractPath == "" {
		contractPath = filepath.Join(absProjectPath, "cliguard.yaml")
	} else {
		contractPath, err = filepath.Abs(contractPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve contract path: %w", err)
		}
	}

	// Load the contract
	contractSpec, err := s.ContractLoader(contractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load contract: %w", err)
	}

	// Inspect the project
	actualStructure, err := s.Inspector(absProjectPath, opts.Entrypoint)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect project: %w", err)
	}

	// Validate the actual structure against the contract
	result := validator.Validate(contractSpec, actualStructure)

	return &ValidateResult{
		Success: result.IsValid(),
		Result:  result,
		Error:   nil,
	}, nil
}
