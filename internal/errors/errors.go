package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ContractNotFoundError indicates the contract file could not be found
type ContractNotFoundError struct {
	Path         string
	SearchedPath string
}

func (e ContractNotFoundError) Error() string {
	return fmt.Sprintf(`Cannot load contract file '%s': file not found

Ensure the contract file exists or specify a different path with --contract flag.
Example: cliguard validate --project-path . --contract ./mycontract.yaml

To generate a new contract file:
  cliguard generate --project-path . > cliguard.yaml`, e.Path)
}

// ContractParseError indicates the contract file could not be parsed
type ContractParseError struct {
	Path    string
	Err     error
	Content string
}

func (e ContractParseError) Error() string {
	// Try to extract line number from YAML error
	lineInfo := ""
	if e.Err != nil {
		errStr := e.Err.Error()
		if strings.Contains(errStr, "line") {
			lineInfo = "\n\n" + errStr
		}
	}

	return fmt.Sprintf(`Failed to parse contract YAML file '%s'%s

Common issues:
  - Incorrect indentation (YAML requires consistent spaces, not tabs)
  - Missing quotes around special characters
  - Invalid flag types (check supported types in documentation)

To validate your YAML syntax:
  cat %s | yq eval . -`, e.Path, lineInfo, e.Path)
}

// InvalidContractError indicates the contract has validation errors
type InvalidContractError struct {
	Path    string
	Message string
}

func (e InvalidContractError) Error() string {
	return fmt.Sprintf(`Contract validation failed in '%s':
%s

Please fix the contract file and try again.`, e.Path, e.Message)
}

// ProjectNotFoundError indicates the project path does not exist
type ProjectNotFoundError struct {
	Path string
}

func (e ProjectNotFoundError) Error() string {
	return fmt.Sprintf(`Project path does not exist: '%s'

Please ensure the path is correct or use --project-path to specify a different location.
Current directory: %s`, e.Path, getCurrentDir())
}

// EntrypointParseError indicates the entrypoint format is invalid
type EntrypointParseError struct {
	Entrypoint string
	Reason     string
}

func (e EntrypointParseError) Error() string {
	return fmt.Sprintf(`Failed to parse entrypoint '%s': %s

Expected format: package.Function or github.com/user/repo/package.Function
Examples:
  - main.NewRootCmd
  - github.com/spf13/cobra/cmd.Execute
  - internal/cmd.NewRootCommand`, e.Entrypoint, e.Reason)
}

// InspectionError indicates the project inspection failed
type InspectionError struct {
	ProjectPath string
	Entrypoint  string
	Err         error
}

func (e InspectionError) Error() string {
	msg := fmt.Sprintf(`Failed to inspect project at '%s'`, e.ProjectPath)

	if e.Entrypoint != "" {
		msg += fmt.Sprintf(` with entrypoint '%s'`, e.Entrypoint)
	}

	msg += fmt.Sprintf(`: %v

Common causes:
  - Project is not a valid Go module (missing go.mod)
  - Entrypoint function does not exist or is not exported
  - Build errors in the project
  - Missing dependencies

To debug:
  1. Ensure 'go build' works in your project directory
  2. Verify the entrypoint function exists and returns *cobra.Command
  3. Try running 'go mod tidy' to resolve dependencies`, e.Err)

	return msg
}

// TempDirError indicates temporary directory operations failed
type TempDirError struct {
	Operation string
	Err       error
}

func (e TempDirError) Error() string {
	return fmt.Sprintf(`Failed to %s temporary directory: %v

This might be due to:
  - Insufficient disk space
  - Permission issues in temp directory
  - System temp directory not accessible

Try setting TMPDIR environment variable to a writable directory:
  export TMPDIR=/path/to/writable/directory`, e.Operation, e.Err)
}

// DependencyError indicates Go module dependency resolution failed
type DependencyError struct {
	Operation string
	Output    string
	Err       error
}

func (e DependencyError) Error() string {
	return fmt.Sprintf(`Failed to resolve Go module dependencies: %v

Operation: %s

This might be due to:
  - Network connectivity issues
  - Private repository access problems
  - Incompatible dependency versions

To debug:
  1. Run 'go mod download' in your project
  2. Check GOPROXY and GOPRIVATE settings
  3. Ensure all dependencies are accessible

Output:
%s`, e.Err, e.Operation, e.Output)
}

// FlagTypeError indicates an unsupported flag type
type FlagTypeError struct {
	FlagName    string
	InvalidType string
	ValidTypes  []string
}

func (e FlagTypeError) Error() string {
	return fmt.Sprintf(`Invalid flag type '%s' for flag '%s'

Supported types:
%s

Example flag definition:
  flags:
    - name: %s
      type: string  # Change to one of the supported types
      description: "Your flag description"`,
		e.InvalidType,
		e.FlagName,
		formatValidTypes(e.ValidTypes),
		e.FlagName)
}

// Helper functions

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "<unable to determine>"
	}
	return dir
}

func formatValidTypes(types []string) string {
	var result []string
	for _, t := range types {
		result = append(result, fmt.Sprintf("  - %s", t))
	}
	return strings.Join(result, "\n")
}

// IsContractNotFound checks if an error is a ContractNotFoundError
func IsContractNotFound(err error) bool {
	_, ok := err.(ContractNotFoundError)
	return ok
}

// IsProjectNotFound checks if an error is a ProjectNotFoundError
func IsProjectNotFound(err error) bool {
	_, ok := err.(ProjectNotFoundError)
	return ok
}

// IsEntrypointParseError checks if an error is an EntrypointParseError
func IsEntrypointParseError(err error) bool {
	_, ok := err.(EntrypointParseError)
	return ok
}

// WrapContractNotFound wraps a file not found error as ContractNotFoundError
func WrapContractNotFound(path string, err error) error {
	if os.IsNotExist(err) {
		absPath, _ := filepath.Abs(path)
		return ContractNotFoundError{
			Path:         path,
			SearchedPath: absPath,
		}
	}
	return err
}

