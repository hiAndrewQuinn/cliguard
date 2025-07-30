package executor

import (
	"os/exec"
)

// CommandExecutor is an interface for executing system commands
type CommandExecutor interface {
	Command(name string, args ...string) Command
}

// Command represents an executable command
type Command interface {
	SetDir(dir string)
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
}

// OSExecutor is the real implementation using os/exec
type OSExecutor struct{}

// Command creates a new command
func (e *OSExecutor) Command(name string, args ...string) Command {
	return &osCommand{cmd: exec.Command(name, args...)}
}

// osCommand wraps exec.Cmd to implement our Command interface
type osCommand struct {
	cmd *exec.Cmd
}

// SetDir sets the working directory for the command
func (c *osCommand) SetDir(dir string) {
	c.cmd.Dir = dir
}

// Output runs the command and returns its standard output
func (c *osCommand) Output() ([]byte, error) {
	return c.cmd.Output()
}

// CombinedOutput runs the command and returns combined stdout and stderr
func (c *osCommand) CombinedOutput() ([]byte, error) {
	return c.cmd.CombinedOutput()
}