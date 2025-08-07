// Package validator provides the core validation logic for comparing
// CLI contracts against actual CLI implementations.
//
// The validator package is responsible for comparing the expected CLI structure
// (defined in contracts) against the actual CLI structure (discovered through
// inspection) and reporting any discrepancies.
//
// # Basic Usage
//
// To validate a CLI against its contract:
//
//	contract, err := contract.Load("cliguard.yaml")
//	if err != nil {
//	    return err
//	}
//
//	actual, err := inspector.InspectProject(".", "cmd.NewRootCmd")
//	if err != nil {
//	    return err
//	}
//
//	result := validator.Validate(contract, actual)
//	if !result.IsValid() {
//	    result.PrintReport()
//	    return fmt.Errorf("validation failed")
//	}
//
// # Validation Rules
//
// The validator checks for:
//
// Command Validation:
//   - Command names must match exactly
//   - Short descriptions must match
//   - Long descriptions must match (if specified)
//   - All expected subcommands must exist
//
// Flag Validation:
//   - Flag names must match exactly
//   - Flag types must match (string, bool, int, etc.)
//   - Shorthands must match (if specified)
//   - Usage strings must match (if specified)
//   - Default values must match (if specified)
//   - Required flags must be marked as required
//
// # Validation Modes
//
// The validator supports different validation modes:
//
// Strict Mode (default):
//   - All contract specifications must match exactly
//   - Extra flags or commands in the implementation are allowed
//   - Missing required elements cause validation to fail
//
// # Error Reporting
//
// The validator provides detailed error messages:
//
//	result := validator.Validate(contract, actual)
//	for _, err := range result.Errors {
//	    fmt.Printf("Error at %s: %s\n", err.Path, err.Message)
//	    if err.Expected != "" {
//	        fmt.Printf("  Expected: %s\n", err.Expected)
//	        fmt.Printf("  Got: %s\n", err.Got)
//	    }
//	}
//
// # Extensibility
//
// The validator is designed to be extensible. Custom validation rules
// can be added by implementing additional validation functions that
// operate on the Contract and InspectedCLI structures.
package validator
