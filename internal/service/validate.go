package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"github.com/hiAndrewQuinn/cliguard/internal/errors"
	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
	"github.com/hiAndrewQuinn/cliguard/internal/validator"
)

// ValidateService provides validation functionality
type ValidateService struct {
	// Dependencies can be injected for testing
	ContractLoader func(string) (*contract.Contract, error)
	Inspector      func(string, string) (*inspector.InspectedCLI, error)
}

// NewValidateService creates a new validation service with default dependencies
func NewValidateService() *ValidateService {
	return &ValidateService{
		ContractLoader: contract.Load,
		Inspector:      inspector.InspectProject,
	}
}

// ValidateOptions contains the options for validation
type ValidateOptions struct {
	ProjectPath  string
	ContractPath string
	Entrypoint   string
}

// ValidateResult contains the result of validation
type ValidateResult struct {
	Success bool
	Result  *validator.ValidationResult
	Error   error
}

// Validate performs the validation
func (s *ValidateService) Validate(opts ValidateOptions) (*ValidateResult, error) {
	// Resolve project path
	absProjectPath, err := filepath.Abs(opts.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path '%s': %w", opts.ProjectPath, err)
	}

	// Check if project path exists
	if _, err := os.Stat(absProjectPath); os.IsNotExist(err) {
		return nil, errors.ProjectNotFoundError{Path: absProjectPath}
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
		return nil, errors.InspectionError{
			ProjectPath: absProjectPath,
			Entrypoint:  opts.Entrypoint,
			Err:         err,
		}
	}

	// Validate the actual structure against the contract
	result := validator.Validate(contractSpec, actualStructure)

	return &ValidateResult{
		Success: result.IsValid(),
		Result:  result,
		Error:   nil,
	}, nil
}
