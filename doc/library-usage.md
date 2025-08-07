# Using Cliguard as a Library

Cliguard can be used programmatically in your Go applications, test suites, or CI/CD pipelines to validate CLI structures against contracts.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Advanced Usage](#advanced-usage)
- [Integration Examples](#integration-examples)
- [API Reference](#api-reference)
- [Error Handling](#error-handling)
- [Testing with Cliguard](#testing-with-cliguard)
- [Best Practices](#best-practices)

## Installation

```bash
go get github.com/hiAndrewQuinn/cliguard
```

## Basic Usage

### Simple Validation

The most common use case is validating a CLI against a contract:

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/hiAndrewQuinn/cliguard/internal/contract"
    "github.com/hiAndrewQuinn/cliguard/internal/inspector"
    "github.com/hiAndrewQuinn/cliguard/internal/validator"
)

func main() {
    // Load the contract specification
    contractSpec, err := contract.Load("cliguard.yaml")
    if err != nil {
        log.Fatalf("Failed to load contract: %v", err)
    }
    
    // Inspect the actual CLI structure
    actualCLI, err := inspector.InspectProject(".", "cmd.NewRootCmd")
    if err != nil {
        log.Fatalf("Failed to inspect CLI: %v", err)
    }
    
    // Validate the CLI against the contract
    result := validator.Validate(contractSpec, actualCLI)
    
    // Check the results
    if !result.IsValid() {
        fmt.Println("❌ Validation failed:")
        for _, err := range result.Errors {
            fmt.Printf("  • %s at %s\n", err.Message, err.Path)
            if err.Expected != "" {
                fmt.Printf("    Expected: %s\n", err.Expected)
                fmt.Printf("    Got: %s\n", err.Got)
            }
        }
        os.Exit(1)
    }
    
    fmt.Println("✅ CLI matches contract")
}
```

### Using the Service Layer

For a higher-level API, use the service layer:

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/hiAndrewQuinn/cliguard/internal/service"
)

func main() {
    // Create a validation service
    svc := service.NewValidateService()
    
    // Configure validation options
    opts := service.ValidateOptions{
        ProjectPath:  "./my-cli",
        ContractPath: "./my-cli/cliguard.yaml",
        Entrypoint:   "cmd.NewRootCmd",
    }
    
    // Run validation
    result, err := svc.Validate(opts)
    if err != nil {
        log.Fatalf("Validation failed to run: %v", err)
    }
    
    // Check results
    if !result.Success {
        fmt.Println("Validation failed:")
        fmt.Println(result.Result.FormatReport())
    } else {
        fmt.Println("Validation passed!")
    }
}
```

## Advanced Usage

### Custom Contract Loading

You can implement custom contract loading logic:

```go
package main

import (
    "github.com/hiAndrewQuinn/cliguard/internal/contract"
    "github.com/hiAndrewQuinn/cliguard/internal/service"
)

func customContractLoader(path string) (*contract.Contract, error) {
    // Load from database, API, or custom format
    return &contract.Contract{
        Use:   "myapp",
        Short: "My application",
        Flags: []contract.Flag{
            {
                Name:  "config",
                Type:  "string",
                Usage: "Config file path",
            },
        },
    }, nil
}

func main() {
    svc := &service.ValidateService{
        ContractLoader: customContractLoader,
        Inspector:      inspector.InspectProject,
    }
    
    // Use the service with custom loader
    result, err := svc.Validate(service.ValidateOptions{
        ProjectPath: ".",
        Entrypoint:  "cmd.NewRootCmd",
    })
    // ... handle result
}
```

### Programmatic Contract Creation

Create contracts programmatically instead of using YAML:

```go
package main

import (
    "github.com/hiAndrewQuinn/cliguard/internal/contract"
)

func createContract() *contract.Contract {
    return &contract.Contract{
        Use:   "myapp",
        Short: "My application description",
        Long:  "A longer description of my application...",
        Flags: []contract.Flag{
            {
                Name:       "verbose",
                Shorthand:  "v",
                Type:       "bool",
                Usage:      "Enable verbose output",
                Persistent: true,
            },
            {
                Name:  "config",
                Type:  "string",
                Usage: "Config file path",
            },
        },
        Commands: []contract.Command{
            {
                Use:   "serve",
                Short: "Start the server",
                Flags: []contract.Flag{
                    {
                        Name:      "port",
                        Shorthand: "p",
                        Type:      "int",
                        Usage:     "Port to listen on",
                    },
                },
            },
            {
                Use:   "migrate",
                Short: "Run database migrations",
                Commands: []contract.Command{
                    {
                        Use:   "up",
                        Short: "Migrate database up",
                    },
                    {
                        Use:   "down",
                        Short: "Migrate database down",
                    },
                },
            },
        },
    }
}
```

### Generating Contracts from Existing CLIs

Generate a contract from an existing CLI implementation:

```go
package main

import (
    "log"
    "gopkg.in/yaml.v3"
    
    "github.com/hiAndrewQuinn/cliguard/internal/inspector"
)

func generateContract(projectPath, entrypoint string) error {
    // Inspect the CLI
    cli, err := inspector.InspectProject(projectPath, entrypoint)
    if err != nil {
        return err
    }
    
    // Convert to contract format (you'd implement this conversion)
    contract := inspectedToContract(cli)
    
    // Save as YAML
    data, err := yaml.Marshal(contract)
    if err != nil {
        return err
    }
    
    return os.WriteFile("cliguard.yaml", data, 0644)
}

func inspectedToContract(cli *inspector.InspectedCLI) *contract.Contract {
    // Implementation to convert InspectedCLI to Contract
    // This is a simplified example
    return &contract.Contract{
        Use:      cli.Use,
        Short:    cli.Short,
        Long:     cli.Long,
        Flags:    convertFlags(cli.Flags),
        Commands: convertCommands(cli.Commands),
    }
}
```

## Integration Examples

### GitHub Actions Integration

Use cliguard in your CI/CD pipeline:

```go
// ci/validate_cli.go
package main

import (
    "fmt"
    "os"
    
    "github.com/hiAndrewQuinn/cliguard/internal/service"
)

func main() {
    svc := service.NewValidateService()
    
    result, err := svc.Validate(service.ValidateOptions{
        ProjectPath:  os.Getenv("GITHUB_WORKSPACE"),
        ContractPath: "cliguard.yaml",
        Entrypoint:   "cmd.NewRootCmd",
    })
    
    if err != nil {
        fmt.Printf("::error::Failed to run validation: %v\n", err)
        os.Exit(1)
    }
    
    if !result.Success {
        for _, err := range result.Result.Errors {
            fmt.Printf("::error file=cliguard.yaml::%s at %s\n", 
                err.Message, err.Path)
        }
        os.Exit(1)
    }
    
    fmt.Println("::notice::CLI validation passed ✅")
}
```

### Test Suite Integration

Integrate cliguard into your Go test suite:

```go
package cli_test

import (
    "testing"
    
    "github.com/hiAndrewQuinn/cliguard/internal/contract"
    "github.com/hiAndrewQuinn/cliguard/internal/inspector"
    "github.com/hiAndrewQuinn/cliguard/internal/validator"
)

func TestCLIContract(t *testing.T) {
    // Load contract
    expected, err := contract.Load("../cliguard.yaml")
    if err != nil {
        t.Fatalf("Failed to load contract: %v", err)
    }
    
    // Inspect CLI
    actual, err := inspector.InspectProject("..", "cmd.NewRootCmd")
    if err != nil {
        t.Fatalf("Failed to inspect CLI: %v", err)
    }
    
    // Validate
    result := validator.Validate(expected, actual)
    
    // Assert
    if !result.IsValid() {
        t.Errorf("CLI does not match contract:")
        for _, err := range result.Errors {
            t.Errorf("  - %s: %s", err.Path, err.Message)
        }
    }
}

func TestSpecificCommand(t *testing.T) {
    tests := []struct {
        name     string
        command  string
        wantFlags []string
    }{
        {
            name:    "serve command has required flags",
            command: "serve",
            wantFlags: []string{"port", "host", "config"},
        },
        {
            name:    "migrate command has subcommands",
            command: "migrate",
            wantFlags: []string{"database", "verbose"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Pre-commit Hook

Create a pre-commit hook that validates CLI changes:

```go
// scripts/pre-commit.go
package main

import (
    "fmt"
    "os"
    "os/exec"
    
    "github.com/hiAndrewQuinn/cliguard/internal/service"
)

func main() {
    // Check if CLI files were modified
    cmd := exec.Command("git", "diff", "--cached", "--name-only")
    output, _ := cmd.Output()
    
    if !containsCLIFiles(string(output)) {
        os.Exit(0) // No CLI changes
    }
    
    // Run validation
    svc := service.NewValidateService()
    result, err := svc.Validate(service.ValidateOptions{
        ProjectPath:  ".",
        ContractPath: "cliguard.yaml",
        Entrypoint:   "cmd.NewRootCmd",
    })
    
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Validation error: %v\n", err)
        os.Exit(1)
    }
    
    if !result.Success {
        fmt.Fprintln(os.Stderr, "❌ CLI contract validation failed:")
        fmt.Fprintln(os.Stderr, result.Result.FormatReport())
        fmt.Fprintln(os.Stderr, "\nPlease update cliguard.yaml or fix the CLI")
        os.Exit(1)
    }
    
    fmt.Println("✅ CLI contract validation passed")
}
```

## API Reference

### Core Packages

#### `contract` Package

Handles contract specifications:

```go
// Load a contract from YAML
contract, err := contract.Load("path/to/contract.yaml")

// Contract structure
type Contract struct {
    Use      string    // Command name
    Short    string    // Short description
    Long     string    // Long description
    Flags    []Flag    // Command flags
    Commands []Command // Subcommands
}
```

#### `inspector` Package

Analyzes Go CLI projects:

```go
// Inspect a project
cli, err := inspector.InspectProject(projectPath, entrypoint)

// InspectedCLI structure
type InspectedCLI struct {
    Use      string             // Command name
    Short    string             // Short description
    Long     string             // Long description
    Flags    []InspectedFlag    // Discovered flags
    Commands []InspectedCommand // Discovered commands
}
```

#### `validator` Package

Compares contracts with actual CLIs:

```go
// Validate CLI against contract
result := validator.Validate(contract, inspectedCLI)

// Check if valid
if result.IsValid() {
    // Validation passed
}

// Access errors
for _, err := range result.Errors {
    fmt.Printf("%s: %s\n", err.Type, err.Message)
}
```

#### `service` Package

High-level orchestration:

```go
// Create service
svc := service.NewValidateService()

// Configure options
opts := service.ValidateOptions{
    ProjectPath:  "./project",
    ContractPath: "./contract.yaml",
    Entrypoint:   "cmd.NewRootCmd",
}

// Run validation
result, err := svc.Validate(opts)
```

## Error Handling

### Validation Errors

Cliguard distinguishes between operational errors and validation failures:

```go
result, err := svc.Validate(opts)

// Operational error (couldn't run validation)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "not found"):
        // Handle missing files
    case strings.Contains(err.Error(), "build failed"):
        // Handle build errors
    default:
        // Handle other errors
    }
}

// Validation failure (CLI doesn't match contract)
if !result.Success {
    for _, validationErr := range result.Result.Errors {
        switch validationErr.Type {
        case validator.ErrorTypeMissing:
            // Handle missing elements
        case validator.ErrorTypeMismatch:
            // Handle mismatches
        case validator.ErrorTypeUnexpected:
            // Handle unexpected elements
        }
    }
}
```

### Common Error Types

```go
const (
    // Element is missing in implementation
    ErrorTypeMissing = "missing"
    
    // Element exists but doesn't match specification
    ErrorTypeMismatch = "mismatch"
    
    // Element exists but isn't in specification
    ErrorTypeUnexpected = "unexpected"
    
    // Type mismatch for flags
    ErrorTypeInvalidType = "invalid_type"
)
```

## Testing with Cliguard

### Unit Testing

Mock dependencies for unit tests:

```go
func TestValidationService(t *testing.T) {
    svc := &service.ValidateService{
        ContractLoader: func(path string) (*contract.Contract, error) {
            return &contract.Contract{
                Use:   "test",
                Short: "Test CLI",
            }, nil
        },
        Inspector: func(path, entrypoint string) (*inspector.InspectedCLI, error) {
            return &inspector.InspectedCLI{
                Use:   "test",
                Short: "Test CLI",
            }, nil
        },
    }
    
    result, err := svc.Validate(service.ValidateOptions{
        ProjectPath: ".",
        Entrypoint:  "test",
    })
    
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

### Table-Driven Tests

Test multiple scenarios:

```go
func TestContractValidation(t *testing.T) {
    tests := []struct {
        name        string
        contract    *contract.Contract
        inspected   *inspector.InspectedCLI
        wantValid   bool
        wantErrors  int
    }{
        {
            name: "matching CLI",
            contract: &contract.Contract{
                Use:   "app",
                Short: "Application",
            },
            inspected: &inspector.InspectedCLI{
                Use:   "app",
                Short: "Application",
            },
            wantValid:  true,
            wantErrors: 0,
        },
        {
            name: "mismatched description",
            contract: &contract.Contract{
                Use:   "app",
                Short: "Application",
            },
            inspected: &inspector.InspectedCLI{
                Use:   "app",
                Short: "Different",
            },
            wantValid:  false,
            wantErrors: 1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validator.Validate(tt.contract, tt.inspected)
            
            if got := result.IsValid(); got != tt.wantValid {
                t.Errorf("IsValid() = %v, want %v", got, tt.wantValid)
            }
            
            if got := len(result.Errors); got != tt.wantErrors {
                t.Errorf("Errors count = %d, want %d", got, tt.wantErrors)
            }
        })
    }
}
```

## Best Practices

### 1. Version Your Contracts

Keep contracts in version control alongside your code:

```yaml
# cliguard.yaml
# Version: 1.0.0
# Last Updated: 2024-01-15
use: myapp
short: My application
```

### 2. Automate Validation

Add validation to your CI/CD pipeline:

```yaml
# .github/workflows/validate.yml
name: Validate CLI Contract
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go run scripts/validate_cli.go
```

### 3. Document Contract Changes

When updating contracts, document the changes:

```go
// Contract v2.0.0 - Added new 'export' command
// Breaking changes:
// - Renamed 'dump' command to 'export'
// - Changed 'format' flag type from string to stringSlice
```

### 4. Use Semantic Versioning

Version your contracts and CLI together:

```go
const (
    CLIVersion      = "2.0.0"
    ContractVersion = "2.0.0"
)
```

### 5. Test Contract Evolution

Test that your CLI remains backward compatible:

```go
func TestBackwardCompatibility(t *testing.T) {
    // Load old contract
    oldContract, _ := contract.Load("contracts/v1.0.0.yaml")
    
    // Inspect current CLI
    currentCLI, _ := inspector.InspectProject(".", "cmd.NewRootCmd")
    
    // Validate with compatibility checks
    result := validateWithCompatibility(oldContract, currentCLI)
    
    if !result.BackwardCompatible {
        t.Error("Breaking changes detected")
    }
}
```

### 6. Handle Optional Features

Design contracts to handle optional features:

```go
type Contract struct {
    // Required fields
    Use   string `yaml:"use" validate:"required"`
    Short string `yaml:"short" validate:"required"`
    
    // Optional fields
    Long     string    `yaml:"long,omitempty"`
    Flags    []Flag    `yaml:"flags,omitempty"`
    Commands []Command `yaml:"commands,omitempty"`
}
```

### 7. Custom Validation Rules

Implement custom validation for specific requirements:

```go
func validateCustomRules(contract *contract.Contract, cli *inspector.InspectedCLI) []error {
    var errors []error
    
    // Custom rule: All commands must have descriptions
    for _, cmd := range cli.Commands {
        if cmd.Short == "" {
            errors = append(errors, 
                fmt.Errorf("command %s missing description", cmd.Use))
        }
    }
    
    // Custom rule: Dangerous flags must have confirmation
    for _, flag := range cli.Flags {
        if isDangerous(flag.Name) && !hasConfirmation(cli, flag) {
            errors = append(errors,
                fmt.Errorf("dangerous flag %s needs confirmation", flag.Name))
        }
    }
    
    return errors
}
```

## Troubleshooting

### Common Issues

1. **"Failed to inspect project"**
   - Ensure the project has a go.mod file
   - Verify the entrypoint function exists and is exported
   - Check that the project builds successfully

2. **"Contract validation failed unexpectedly"**
   - Verify the contract YAML syntax
   - Ensure flag types match exactly (e.g., "string" not "String")
   - Check for invisible characters in descriptions

3. **"Cannot find cobra commands"**
   - Ensure the project uses github.com/spf13/cobra
   - Verify the entrypoint returns a *cobra.Command

### Debug Mode

Enable detailed logging for debugging:

```go
func debugValidation() {
    // Set up detailed logging
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    
    // Log each step
    log.Println("Loading contract...")
    contract, err := contract.Load("cliguard.yaml")
    if err != nil {
        log.Printf("Contract load error: %+v", err)
        return
    }
    log.Printf("Contract loaded: %+v", contract)
    
    log.Println("Inspecting project...")
    cli, err := inspector.InspectProject(".", "cmd.NewRootCmd")
    if err != nil {
        log.Printf("Inspection error: %+v", err)
        return
    }
    log.Printf("CLI inspected: %+v", cli)
    
    log.Println("Validating...")
    result := validator.Validate(contract, cli)
    log.Printf("Validation result: %+v", result)
}
```

## Contributing

See the main cliguard repository for contribution guidelines. When adding library features:

1. Maintain backward compatibility
2. Add comprehensive godoc comments
3. Include usage examples
4. Write unit tests
5. Update this documentation

## License

Cliguard is released under the MIT License. See the LICENSE file for details.