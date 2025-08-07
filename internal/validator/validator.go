package validator

import (
	"fmt"
	"strings"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
)

// Validate compares the actual CLI structure against the contract specification
// and returns a detailed report of any discrepancies found.
//
// The validation process checks:
//   - Command names, descriptions, and structure
//   - Flag names, types, shorthands, and persistence
//   - Nested subcommands recursively
//
// Example:
//
//	// Load the contract
//	contractSpec, err := contract.Load("cliguard.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	// Inspect the actual CLI
//	actualCLI, err := inspector.InspectProject(".", "cmd.NewRootCmd")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	// Validate
//	result := validator.Validate(contractSpec, actualCLI)
//	
//	// Check results
//	if !result.IsValid() {
//	    fmt.Println("Validation failed:")
//	    for _, err := range result.Errors {
//	        fmt.Printf("  %s at %s\n", err.Message, err.Path)
//	        if err.Expected != "" {
//	            fmt.Printf("    Expected: %s\n", err.Expected)
//	            fmt.Printf("    Got: %s\n", err.Got)
//	        }
//	    }
//	    os.Exit(1)
//	}
//	fmt.Println("✓ CLI matches contract")
//
// The function returns a ValidationResult containing all validation errors found.
// An empty Errors slice indicates successful validation.
func Validate(expected *contract.Contract, actual *inspector.InspectedCLI) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate root command
	validateRootCommand(expected, actual, result)

	// Validate flags
	validateFlags("", expected.Flags, actual.Flags, result)

	// Validate subcommands
	validateCommands("", expected.Commands, actual.Commands, result)

	return result
}

func validateRootCommand(expected *contract.Contract, actual *inspector.InspectedCLI, result *ValidationResult) {
	// Validate Use field
	if expected.Use != actual.Use {
		result.AddError(ErrorTypeMismatch, "root", expected.Use, actual.Use, "Mismatch in 'use' field")
	}

	// Validate Short description
	if expected.Short != "" && expected.Short != actual.Short {
		result.AddError(ErrorTypeMismatch, "root", expected.Short, actual.Short, "Mismatch in short description")
	}

	// Validate Long description if specified
	if expected.Long != "" && expected.Long != actual.Long {
		result.AddError(ErrorTypeMismatch, "root", expected.Long, actual.Long, "Mismatch in long description")
	}
}

func validateCommands(parentPath string, expected []contract.Command, actual []inspector.InspectedCommand, result *ValidationResult) {
	// Create maps for easier lookup
	expectedMap := make(map[string]*contract.Command)
	for i := range expected {
		expectedMap[expected[i].Use] = &expected[i]
	}

	actualMap := make(map[string]*inspector.InspectedCommand)
	for i := range actual {
		actualMap[actual[i].Use] = &actual[i]
	}

	// Check for missing commands
	for _, exp := range expected {
		cmdPath := joinPath(parentPath, exp.Use)
		if _, found := actualMap[exp.Use]; !found {
			result.AddError(ErrorTypeMissing, cmdPath, exp.Use, "", "command")
		}
	}

	// Check for unexpected commands
	for _, act := range actual {
		cmdPath := joinPath(parentPath, act.Use)
		if _, found := expectedMap[act.Use]; !found {
			result.AddError(ErrorTypeUnexpected, cmdPath, "", act.Use, "command")
		}
	}

	// Validate matching commands
	for use, exp := range expectedMap {
		if act, found := actualMap[use]; found {
			cmdPath := joinPath(parentPath, use)
			validateCommand(cmdPath, exp, act, result)
		}
	}
}

func validateCommand(path string, expected *contract.Command, actual *inspector.InspectedCommand, result *ValidationResult) {
	// Validate Use field (should already match, but just in case)
	if expected.Use != actual.Use {
		result.AddError(ErrorTypeMismatch, path, expected.Use, actual.Use, "Mismatch in 'use' field")
	}

	// Validate Short description
	if expected.Short != "" && expected.Short != actual.Short {
		result.AddError(ErrorTypeMismatch, path, expected.Short, actual.Short, "Mismatch in short description")
	}

	// Validate Long description if specified
	if expected.Long != "" && expected.Long != actual.Long {
		result.AddError(ErrorTypeMismatch, path, expected.Long, actual.Long, "Mismatch in long description")
	}

	// Validate flags
	validateFlags(path, expected.Flags, actual.Flags, result)

	// Validate subcommands recursively
	validateCommands(path, expected.Commands, actual.Commands, result)
}

func validateFlags(parentPath string, expected []contract.Flag, actual []inspector.InspectedFlag, result *ValidationResult) {
	// Create maps for easier lookup
	expectedMap := make(map[string]*contract.Flag)
	for i := range expected {
		expectedMap[expected[i].Name] = &expected[i]
	}

	actualMap := make(map[string]*inspector.InspectedFlag)
	for i := range actual {
		actualMap[actual[i].Name] = &actual[i]
	}

	// Check for missing flags
	for _, exp := range expected {
		flagPath := joinPath(parentPath, "--"+exp.Name)
		if _, found := actualMap[exp.Name]; !found {
			result.AddError(ErrorTypeMissing, flagPath, exp.Name, "", "flag")
		}
	}

	// Check for unexpected flags
	for _, act := range actual {
		flagPath := joinPath(parentPath, "--"+act.Name)
		if _, found := expectedMap[act.Name]; !found {
			result.AddError(ErrorTypeUnexpected, flagPath, "", act.Name, "flag")
		}
	}

	// Validate matching flags
	for name, exp := range expectedMap {
		if act, found := actualMap[name]; found {
			flagPath := joinPath(parentPath, "--"+name)
			validateFlag(flagPath, exp, act, result)
		}
	}
}

func validateFlag(path string, expected *contract.Flag, actual *inspector.InspectedFlag, result *ValidationResult) {
	// Validate shorthand
	if expected.Shorthand != "" && expected.Shorthand != actual.Shorthand {
		result.AddError(ErrorTypeMismatch, path, expected.Shorthand, actual.Shorthand, "Flag shorthand mismatch")
	}

	// Validate usage/description
	if expected.Usage != "" && expected.Usage != actual.Usage {
		result.AddError(ErrorTypeMismatch, path, expected.Usage, actual.Usage, "Flag usage mismatch")
	}

	// Validate type
	if expected.Type != actual.Type {
		result.AddError(ErrorTypeInvalidType, path, expected.Type, actual.Type, "Flag type mismatch")
	}

	// Validate persistence
	if expected.Persistent != actual.Persistent {
		expectedPersistence := "local"
		actualPersistence := "local"
		if expected.Persistent {
			expectedPersistence = "persistent"
		}
		if actual.Persistent {
			actualPersistence = "persistent"
		}
		result.AddError(ErrorTypeMismatch, path, expectedPersistence, actualPersistence, "Flag persistence mismatch")
	}
}

func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return fmt.Sprintf("%s %s", parent, strings.TrimSpace(child))
}
