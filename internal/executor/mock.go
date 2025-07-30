package executor

import (
	"fmt"
	"strings"
)

// MockExecutor is a mock implementation for testing
type MockExecutor struct {
	Commands []MockCommand
	Results  map[string]MockResult
}

// MockCommand represents a recorded command execution
type MockCommand struct {
	Name string
	Args []string
	Dir  string
}

// MockResult represents the result to return for a command
type MockResult struct {
	Output []byte
	Error  error
}

// Command creates a new mock command
func (m *MockExecutor) Command(name string, args ...string) Command {
	cmd := &mockCommand{
		executor: m,
		name:     name,
		args:     args,
	}
	return cmd
}

// mockCommand implements the Command interface for testing
type mockCommand struct {
	executor *MockExecutor
	name     string
	args     []string
	dir      string
}

// SetDir sets the working directory
func (c *mockCommand) SetDir(dir string) {
	c.dir = dir
}

// Output returns the mocked output
func (c *mockCommand) Output() ([]byte, error) {
	c.executor.Commands = append(c.executor.Commands, MockCommand{
		Name: c.name,
		Args: c.args,
		Dir:  c.dir,
	})

	key := c.commandKey()
	if result, ok := c.executor.Results[key]; ok {
		return result.Output, result.Error
	}

	return nil, fmt.Errorf("no mock result configured for command: %s", key)
}

// CombinedOutput returns the mocked combined output
func (c *mockCommand) CombinedOutput() ([]byte, error) {
	// For simplicity, we'll use the same behavior as Output
	return c.Output()
}

// commandKey generates a unique key for the command
func (c *mockCommand) commandKey() string {
	parts := []string{c.name}
	parts = append(parts, c.args...)
	return strings.Join(parts, " ")
}
