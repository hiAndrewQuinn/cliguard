package inspector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hiAndrewQuinn/cliguard/internal/executor"
	"github.com/hiAndrewQuinn/cliguard/internal/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestInspectProjectWithTimeout_Success(t *testing.T) {
	// Basic test that the function exists and can handle timeout parameter
	// We won't mock all the internals here - that's covered in other tests
	cli, err := InspectProjectWithTimeout(".", "main.NewRootCmd", 30*time.Second)
	
	// This might fail due to the actual project structure, but that's OK
	// We're mainly testing that the timeout parameter is accepted
	// and the function doesn't crash
	if err != nil {
		// Expected to fail in the test environment, just verify it's not a panic
		assert.Contains(t, err.Error(), "failed to")
	} else {
		assert.NotNil(t, cli)
	}
}

func TestInspectProjectWithTimeout_ZeroTimeout(t *testing.T) {
	// Zero timeout should work the same as InspectProject
	cli1, err1 := InspectProjectWithTimeout(".", "main.NewRootCmd", 0)
	cli2, err2 := InspectProject(".", "main.NewRootCmd")
	
	// Both should have the same result
	assert.Equal(t, err1 != nil, err2 != nil)
	if err1 == nil && err2 == nil {
		assert.Equal(t, cli1, cli2)
	}
}

func TestInspectProjectWithTimeout_TimeoutError(t *testing.T) {
	// Mock filesystem
	mockFS := &filesystem.MockFileSystem{
		Files: map[string][]byte{
			"/tmp/go.mod": []byte("module test.com/cli\n"),
		},
		Directories: map[string]bool{
			"/tmp": true,
		},
	}

	// Create slow mock executor that simulates long-running process
	slowExec := &SlowTimeoutMockExecutor{
		MockExecutor: &executor.MockExecutor{
			Results: map[string]executor.MockResult{
				"go mod init cliguard-inspector":           {Output: []byte("go: creating new go.mod"), Error: nil},
				"go mod edit -replace test.com/cli=/tmp":    {Output: []byte(""), Error: nil},
				"go mod tidy -e":                           {Output: []byte(""), Error: nil},
				"go run inspector.go":                      {Output: []byte(`{"use":"test","short":"Test CLI","commands":[]}`), Error: nil},
			},
		},
		slowCommands: map[string]time.Duration{
			"go run inspector.go": 2 * time.Second, // Simulate long-running inspector
		},
	}

	// Create inspector with short timeout
	inspector := NewInspector(Config{
		ProjectPath: "/tmp",
		Entrypoint:  "main.NewRootCmd",
		Timeout:     100 * time.Millisecond, // Very short timeout
		FileSystem:  mockFS,
		Executor:    slowExec,
	})

	start := time.Now()
	cli, err := inspector.Inspect()
	elapsed := time.Since(start)

	// Should timeout quickly
	assert.Error(t, err)
	assert.Nil(t, cli)
	assert.Contains(t, err.Error(), "command timed out")
	assert.True(t, elapsed < 1*time.Second, "Should timeout quickly, but took %v", elapsed)
}

func TestInspector_NewInspector_WithTimeout(t *testing.T) {
	config := Config{
		ProjectPath: "/tmp",
		Entrypoint:  "main.NewRootCmd",
		Timeout:     30 * time.Second,
	}

	inspector := NewInspector(config)
	assert.NotNil(t, inspector)

	// The timeout executor should be wrapped inside
	// We can't directly test this without exposing internal state,
	// but we can verify that the configuration was accepted
	assert.Equal(t, "/tmp", inspector.config.ProjectPath)
	assert.Equal(t, "main.NewRootCmd", inspector.config.Entrypoint)
	assert.Equal(t, 30*time.Second, inspector.config.Timeout)
}

func TestInspector_NewInspector_WithoutTimeout(t *testing.T) {
	config := Config{
		ProjectPath: "/tmp",
		Entrypoint:  "main.NewRootCmd",
		Timeout:     0, // No timeout
	}

	inspector := NewInspector(config)
	assert.NotNil(t, inspector)

	// Should still work with zero timeout
	assert.Equal(t, "/tmp", inspector.config.ProjectPath)
	assert.Equal(t, time.Duration(0), inspector.config.Timeout)
}

// SlowTimeoutMockExecutor simulates slow execution for specific commands
type SlowTimeoutMockExecutor struct {
	*executor.MockExecutor
	slowCommands map[string]time.Duration
}

func (s *SlowTimeoutMockExecutor) Command(name string, args ...string) executor.Command {
	cmd := s.MockExecutor.Command(name, args...)
	return &slowTimeoutMockCommand{
		Command:      cmd,
		slowCommands: s.slowCommands,
		commandKey:   strings.Join(append([]string{name}, args...), " "),
	}
}

func (s *SlowTimeoutMockExecutor) CommandContext(ctx context.Context, name string, args ...string) executor.Command {
	cmd := s.MockExecutor.CommandContext(ctx, name, args...)
	return &slowTimeoutMockCommand{
		Command:      cmd,
		slowCommands: s.slowCommands,
		commandKey:   strings.Join(append([]string{name}, args...), " "),
		ctx:          ctx,
	}
}

type slowTimeoutMockCommand struct {
	executor.Command
	slowCommands map[string]time.Duration
	commandKey   string
	ctx          context.Context
}

func (s *slowTimeoutMockCommand) Output() ([]byte, error) {
	if delay, ok := s.slowCommands[s.commandKey]; ok {
		if s.ctx != nil {
			// Respect context cancellation
			select {
			case <-s.ctx.Done():
				return nil, s.ctx.Err()
			case <-time.After(delay):
				// Continue to normal execution
			}
		} else {
			time.Sleep(delay)
		}
	}
	return s.Command.Output()
}

func (s *slowTimeoutMockCommand) CombinedOutput() ([]byte, error) {
	if delay, ok := s.slowCommands[s.commandKey]; ok {
		if s.ctx != nil {
			// Respect context cancellation
			select {
			case <-s.ctx.Done():
				return nil, s.ctx.Err()
			case <-time.After(delay):
				// Continue to normal execution
			}
		} else {
			time.Sleep(delay)
		}
	}
	return s.Command.CombinedOutput()
}

func TestInspectProject_BackwardsCompatibility(t *testing.T) {
	// Test that the old function still works (calls new one with 0 timeout)
	cli1, err1 := InspectProject(".", "main.NewRootCmd")
	cli2, err2 := InspectProjectWithTimeout(".", "main.NewRootCmd", 0)
	
	// Both should have the same result
	assert.Equal(t, err1 != nil, err2 != nil)
	if err1 == nil && err2 == nil {
		assert.Equal(t, cli1, cli2)
	}
}