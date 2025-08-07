// Package inspector provides functionality for analyzing Go CLI applications
// built with cobra to extract their command structure, flags, and metadata.
//
// The inspector package is responsible for dynamically analyzing Go source code
// and extracting the actual CLI structure that will be validated against contracts.
// It uses Go's build and reflection capabilities to understand the CLI's structure
// without requiring manual specification.
//
// # Basic Usage
//
// To inspect a CLI application:
//
//	cli, err := InspectProject("./path/to/project", "cmd.NewRootCmd")
//	if err != nil {
//	    return err
//	}
//	
//	// Access the inspected structure
//	fmt.Printf("CLI: %s\n", cli.Use)
//	for _, flag := range cli.Flags {
//	    fmt.Printf("  Flag: --%s (%s)\n", flag.Name, flag.Type)
//	}
//
// # Inspection Process
//
// The inspector works by:
//
// 1. Building the target Go project
// 2. Executing a temporary program that imports the target
// 3. Using reflection to extract the cobra command structure
// 4. Converting the structure to an InspectedCLI representation
//
// # Supported Features
//
// The inspector can extract:
//   - Command names, descriptions, and aliases
//   - Flag names, types, shorthands, and usage strings
//   - Default values and required flags
//   - Nested subcommands
//   - Persistent flags vs local flags
//
// # Limitations
//
// The inspector requires:
//   - The target project must be a valid Go module
//   - The command constructor must be exported
//   - The project must use github.com/spf13/cobra
//   - Build dependencies must be available
//
// # Error Handling
//
// The inspector provides detailed error messages for common issues:
//   - Build failures with compilation errors
//   - Missing or incorrect constructor functions
//   - Module resolution problems
//   - Reflection errors
package inspector