package executor

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"
)

// TimeoutExecutor wraps a CommandExecutor with timeout capabilities
type TimeoutExecutor struct {
	executor CommandExecutor
	timeout  time.Duration
}

// NewTimeoutExecutor creates a new timeout-aware executor
func NewTimeoutExecutor(executor CommandExecutor, timeout time.Duration) *TimeoutExecutor {
	return &TimeoutExecutor{
		executor: executor,
		timeout:  timeout,
	}
}

// Command creates a new command (uses timeout if configured)
func (t *TimeoutExecutor) Command(name string, args ...string) Command {
	if t.timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
		return &timeoutCommand{
			command: t.executor.CommandContext(ctx, name, args...),
			cancel:  cancel,
			timeout: t.timeout,
		}
	}
	return t.executor.Command(name, args...)
}

// CommandContext creates a new command with the provided context
func (t *TimeoutExecutor) CommandContext(ctx context.Context, name string, args ...string) Command {
	return t.executor.CommandContext(ctx, name, args...)
}

// timeoutCommand wraps a command with timeout and graceful shutdown capabilities
type timeoutCommand struct {
	command Command
	cancel  context.CancelFunc
	timeout time.Duration
}

// SetDir sets the working directory for the command
func (t *timeoutCommand) SetDir(dir string) {
	t.command.SetDir(dir)
}

// Output runs the command and returns its standard output with timeout protection
func (t *timeoutCommand) Output() ([]byte, error) {
	defer t.cancel()

	// Create a channel to capture the result
	type result struct {
		output []byte
		err    error
	}
	resultChan := make(chan result, 1)

	// Run the command in a goroutine
	go func() {
		output, err := t.command.Output()
		resultChan <- result{output: output, err: err}
	}()

	// Wait for either completion or timeout
	select {
	case res := <-resultChan:
		return res.output, res.err
	case <-time.After(t.timeout):
		// Attempt graceful shutdown
		if err := t.gracefulShutdown(); err != nil {
			return nil, fmt.Errorf("command timed out after %v and failed to terminate gracefully: %w", t.timeout, err)
		}
		return nil, fmt.Errorf("command timed out after %v (terminated gracefully)", t.timeout)
	}
}

// CombinedOutput runs the command and returns combined stdout and stderr with timeout protection
func (t *timeoutCommand) CombinedOutput() ([]byte, error) {
	defer t.cancel()

	// Create a channel to capture the result
	type result struct {
		output []byte
		err    error
	}
	resultChan := make(chan result, 1)

	// Run the command in a goroutine
	go func() {
		output, err := t.command.CombinedOutput()
		resultChan <- result{output: output, err: err}
	}()

	// Wait for either completion or timeout
	select {
	case res := <-resultChan:
		return res.output, res.err
	case <-time.After(t.timeout):
		// Attempt graceful shutdown
		if err := t.gracefulShutdown(); err != nil {
			return nil, fmt.Errorf("command timed out after %v and failed to terminate gracefully: %w", t.timeout, err)
		}
		return nil, fmt.Errorf("command timed out after %v (terminated gracefully)", t.timeout)
	}
}

// gracefulShutdown attempts to terminate the process gracefully
func (t *timeoutCommand) gracefulShutdown() error {
	// Try to get the underlying process
	if osCmd, ok := t.command.(*osCommand); ok && osCmd.cmd.Process != nil {
		// First try SIGTERM for graceful shutdown
		if err := osCmd.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			// If SIGTERM fails, try SIGKILL immediately
			return osCmd.cmd.Process.Signal(os.Kill)
		}

		// Give the process 5 seconds to shut down gracefully
		done := make(chan error, 1)
		go func() {
			_, err := osCmd.cmd.Process.Wait()
			done <- err
		}()

		select {
		case <-done:
			// Process terminated gracefully
			return nil
		case <-time.After(5 * time.Second):
			// Process didn't terminate, force kill
			return osCmd.cmd.Process.Signal(os.Kill)
		}
	}

	// If we can't access the underlying process, the context cancellation
	// should have handled termination
	return nil
}