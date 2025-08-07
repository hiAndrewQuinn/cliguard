package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeoutExecutor_Command_NoTimeout(t *testing.T) {
	mockExec := &MockExecutor{
		Results: map[string]MockResult{
			"echo hello": {Output: []byte("hello\n"), Error: nil},
		},
	}

	// Create timeout executor with zero timeout (no timeout)
	timeoutExec := NewTimeoutExecutor(mockExec, 0)
	cmd := timeoutExec.Command("echo", "hello")
	cmd.SetDir("/tmp")

	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(output))
}

func TestTimeoutExecutor_Command_WithTimeout_Success(t *testing.T) {
	mockExec := &MockExecutor{
		Results: map[string]MockResult{
			"echo hello": {Output: []byte("hello\n"), Error: nil},
		},
	}

	// Create timeout executor with 1 second timeout
	timeoutExec := NewTimeoutExecutor(mockExec, 1*time.Second)
	cmd := timeoutExec.Command("echo", "hello")

	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(output))
}

func TestTimeoutExecutor_CommandContext(t *testing.T) {
	mockExec := &MockExecutor{
		Results: map[string]MockResult{
			"echo hello": {Output: []byte("hello\n"), Error: nil},
		},
	}

	timeoutExec := NewTimeoutExecutor(mockExec, 1*time.Second)
	ctx := context.Background()
	cmd := timeoutExec.CommandContext(ctx, "echo", "hello")

	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(output))
}

func TestTimeoutExecutor_CombinedOutput(t *testing.T) {
	mockExec := &MockExecutor{
		Results: map[string]MockResult{
			"echo hello": {Output: []byte("hello\n"), Error: nil},
		},
	}

	timeoutExec := NewTimeoutExecutor(mockExec, 1*time.Second)
	cmd := timeoutExec.Command("echo", "hello")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(output))
}

// SlowMockExecutor simulates slow command execution
type SlowMockExecutor struct {
	*MockExecutor
	delay time.Duration
}

func (s *SlowMockExecutor) Command(name string, args ...string) Command {
	return &slowMockCommand{
		mockCommand: s.MockExecutor.Command(name, args...).(*mockCommand),
		delay:       s.delay,
	}
}

func (s *SlowMockExecutor) CommandContext(ctx context.Context, name string, args ...string) Command {
	return &slowMockCommand{
		mockCommand: s.MockExecutor.CommandContext(ctx, name, args...).(*mockCommand),
		delay:       s.delay,
		ctx:         ctx,
	}
}

type slowMockCommand struct {
	*mockCommand
	delay time.Duration
	ctx   context.Context
}

func (s *slowMockCommand) Output() ([]byte, error) {
	// Simulate slow execution
	if s.ctx != nil {
		// Respect context cancellation
		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		case <-time.After(s.delay):
			// Continue to normal execution
		}
	} else {
		time.Sleep(s.delay)
	}

	return s.mockCommand.Output()
}

func (s *slowMockCommand) CombinedOutput() ([]byte, error) {
	// Simulate slow execution
	if s.ctx != nil {
		// Respect context cancellation
		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		case <-time.After(s.delay):
			// Continue to normal execution
		}
	} else {
		time.Sleep(s.delay)
	}

	return s.mockCommand.CombinedOutput()
}

func TestTimeoutExecutor_Command_Timeout(t *testing.T) {
	mockExec := &SlowMockExecutor{
		MockExecutor: &MockExecutor{
			Results: map[string]MockResult{
				"sleep 5": {Output: []byte("done\n"), Error: nil},
			},
		},
		delay: 2 * time.Second, // Simulate 2 second delay
	}

	// Create timeout executor with 500ms timeout
	timeoutExec := NewTimeoutExecutor(mockExec, 500*time.Millisecond)
	cmd := timeoutExec.Command("sleep", "5")

	start := time.Now()
	output, err := cmd.Output()
	elapsed := time.Since(start)

	// Should timeout quickly (within 1 second, allowing some buffer)
	assert.True(t, elapsed < 1*time.Second, "Command should have timed out quickly, but took %v", elapsed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command timed out after 500ms")
	assert.Nil(t, output)
}

func TestTimeoutExecutor_CombinedOutput_Timeout(t *testing.T) {
	mockExec := &SlowMockExecutor{
		MockExecutor: &MockExecutor{
			Results: map[string]MockResult{
				"sleep 5": {Output: []byte("done\n"), Error: nil},
			},
		},
		delay: 2 * time.Second, // Simulate 2 second delay
	}

	// Create timeout executor with 500ms timeout
	timeoutExec := NewTimeoutExecutor(mockExec, 500*time.Millisecond)
	cmd := timeoutExec.Command("sleep", "5")

	start := time.Now()
	output, err := cmd.CombinedOutput()
	elapsed := time.Since(start)

	// Should timeout quickly (within 1 second, allowing some buffer)
	assert.True(t, elapsed < 1*time.Second, "Command should have timed out quickly, but took %v", elapsed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command timed out after 500ms")
	assert.Nil(t, output)
}

func TestTimeoutExecutor_Command_Error(t *testing.T) {
	mockExec := &MockExecutor{
		Results: map[string]MockResult{
			"false": {Output: nil, Error: errors.New("exit status 1")},
		},
	}

	timeoutExec := NewTimeoutExecutor(mockExec, 1*time.Second)
	cmd := timeoutExec.Command("false")

	output, err := cmd.Output()
	assert.Error(t, err)
	assert.Equal(t, "exit status 1", err.Error())
	assert.Nil(t, output)
}

func TestTimeoutCommand_SetDir(t *testing.T) {
	mockExec := &MockExecutor{
		Results: map[string]MockResult{
			"pwd": {Output: []byte("/test\n"), Error: nil},
		},
	}

	timeoutExec := NewTimeoutExecutor(mockExec, 1*time.Second)
	cmd := timeoutExec.Command("pwd")
	cmd.SetDir("/test")

	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "/test\n", string(output))

	// Verify the directory was set on the underlying command
	require.Len(t, mockExec.Commands, 1)
	assert.Equal(t, "/test", mockExec.Commands[0].Dir)
}

func TestNewTimeoutExecutor(t *testing.T) {
	mockExec := &MockExecutor{}
	timeout := 30 * time.Second

	timeoutExec := NewTimeoutExecutor(mockExec, timeout)

	assert.NotNil(t, timeoutExec)
	assert.Equal(t, mockExec, timeoutExec.executor)
	assert.Equal(t, timeout, timeoutExec.timeout)
}

// Integration test using real commands (skipped in CI)
func TestTimeoutExecutor_Integration_RealCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	realExec := &OSExecutor{}
	timeoutExec := NewTimeoutExecutor(realExec, 100*time.Millisecond)

	// This should succeed quickly
	cmd := timeoutExec.Command("echo", "hello")
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(output))
}

// Benchmark to ensure timeout wrapper doesn't add significant overhead
func BenchmarkTimeoutExecutor_FastCommand(b *testing.B) {
	mockExec := &MockExecutor{
		Results: map[string]MockResult{
			"echo test": {Output: []byte("test\n"), Error: nil},
		},
	}

	timeoutExec := NewTimeoutExecutor(mockExec, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := timeoutExec.Command("echo", "test")
		_, err := cmd.Output()
		if err != nil {
			b.Fatal(err)
		}
	}
}