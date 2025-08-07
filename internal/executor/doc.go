// Package executor provides an abstraction layer for executing system commands
// and managing external processes in cliguard.
//
// The executor package is designed to make system command execution testable
// and predictable. It provides interfaces and implementations for running
// external commands, capturing output, and handling errors in a consistent way.
//
// # Basic Usage
//
// The default executor runs commands directly:
//
//	exec := executor.NewCommandExecutor()
//	output, err := exec.Execute("go", "build", "./...")
//	if err != nil {
//	    return fmt.Errorf("build failed: %w", err)
//	}
//	fmt.Println(output)
//
// # Testing Support
//
// The executor interface makes it easy to mock command execution in tests:
//
//	type MockExecutor struct{}
//
//	func (m *MockExecutor) Execute(name string, args ...string) (string, error) {
//	    if name == "go" && args[0] == "build" {
//	        return "Build successful", nil
//	    }
//	    return "", fmt.Errorf("unexpected command: %s", name)
//	}
//
// # Command Execution
//
// The executor handles:
//   - Command execution with arguments
//   - Output capture (stdout and stderr)
//   - Error handling and exit codes
//   - Working directory management
//   - Environment variable handling
//
// # Error Handling
//
// The executor provides detailed error information:
//   - Command not found errors
//   - Non-zero exit codes with stderr output
//   - Timeout errors (if configured)
//   - Permission errors
//
// # Security Considerations
//
// The executor:
//   - Does not use shell interpretation by default
//   - Properly escapes arguments
//   - Validates command paths
//   - Provides safe defaults for execution
//
// # Advanced Features
//
// Custom executors can be implemented for:
//   - Command logging and auditing
//   - Dry-run modes
//   - Command interception
//   - Resource limiting
//   - Parallel execution management
package executor
