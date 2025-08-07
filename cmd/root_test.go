package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	if cmd.Use != "cliguard" {
		t.Errorf("Root command Use = %q, want %q", cmd.Use, "cliguard")
	}

	if cmd.Short != "A contract-based validation tool for Cobra CLIs" {
		t.Errorf("Root command Short = %q, want %q", cmd.Short, "A contract-based validation tool for Cobra CLIs")
	}

	// Check that validate command exists
	validateCmd, _, err := cmd.Find([]string{"validate"})
	if err != nil {
		t.Errorf("Could not find validate command: %v", err)
	}

	if validateCmd.Use != "validate" {
		t.Errorf("Validate command Use = %q, want %q", validateCmd.Use, "validate")
	}

	// Check required flags
	projectPathFlag := validateCmd.Flag("project-path")
	if projectPathFlag == nil {
		t.Error("project-path flag not found")
	}

	// Check that project-path is NOT required (it's now optional)
	if projectPathFlag.Annotations != nil && projectPathFlag.Annotations[cobra.BashCompOneRequiredFlag] != nil && projectPathFlag.Annotations[cobra.BashCompOneRequiredFlag][0] == "true" {
		t.Error("project-path flag should not be required")
	}

	// Check optional flags
	if validateCmd.Flag("contract") == nil {
		t.Error("contract flag not found")
	}

	if validateCmd.Flag("entrypoint") == nil {
		t.Error("entrypoint flag not found")
	}
}

func TestRunValidate_Errors(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (string, func())
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "project_path_does_not_exist",
			setup: func() (string, func()) {
				return "", func() {}
			},
			args:        []string{"--project-path", "/nonexistent/path"},
			wantErr:     true,
			errContains: "Project path does not exist",
		},
		{
			name: "contract_file_not_found",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				return tmpDir, func() {}
			},
			args:        []string{"--project-path", ""},
			wantErr:     true,
			errContains: "failed to load contract",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath, cleanup := tt.setup()
			defer cleanup()

			// Set up command
			cmd := NewRootCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Update args with actual temp path if needed
			args := make([]string, len(tt.args))
			copy(args, tt.args)
			for i, arg := range args {
				if arg == "" && i > 0 && args[i-1] == "--project-path" {
					args[i] = tmpPath
				}
			}

			// Add validate command
			fullArgs := append([]string{"validate"}, args...)
			cmd.SetArgs(fullArgs)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" {
				output := buf.String()
				if !contains(output, tt.errContains) && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error output = %q, want to contain %q", output, tt.errContains)
				}
			}
		})
	}
}

func TestIntegration_ValidateCommand(t *testing.T) {
	// Skip if short tests are requested
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set up test fixtures
	fixturePath := setupTestFixtures(t)
	contractPath := filepath.Join(fixturePath, "cliguard.yaml")

	// Check if contract exists
	if _, err := os.Stat(contractPath); os.IsNotExist(err) {
		t.Fatalf("Contract file not found at %s", contractPath)
	}

	cmd := NewRootCmd()
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	cmd.SetArgs([]string{
		"validate",
		"--project-path", fixturePath,
		"--contract", contractPath,
		"--entrypoint", "github.com/test/simple-cli/cmd.NewRootCmd",
	})

	err := cmd.Execute()
	if err != nil {
		// For validation failures, err will be nil but exit code would be 1
		// So we only fail on actual execution errors
		if !contains(errBuf.String(), "Validation failed") {
			t.Errorf("Execute() error = %v", err)
		}
	}

	// Check both stdout and stderr for output
	output := outBuf.String() + errBuf.String()
	if !contains(output, "Validation passed") && !contains(output, "Validation failed") {
		t.Errorf("Expected validation result in output, got stdout: %q, stderr: %q", outBuf.String(), errBuf.String())
	}
}

func TestValidateCommand_DefaultProjectPath(t *testing.T) {
	// Test that validate command uses current directory when project-path is not provided
	
	// Save original runner and restore after test
	originalRunner := validateRunner
	defer func() { validateRunner = originalRunner }()
	
	// Mock runner to capture the project path
	var capturedPath string
	mockRunner := &mockValidateRunner{
		runFunc: func(cmd *cobra.Command, projectPath, contractPath, entrypoint string, force bool) error {
			capturedPath = projectPath
			return nil
		},
	}
	validateRunner = mockRunner
	
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"validate"})
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	
	// Get expected current directory
	expectedPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	
	if capturedPath != expectedPath {
		t.Errorf("projectPath = %q, want %q (current directory)", capturedPath, expectedPath)
	}
}

// mockValidateRunner for testing
type mockValidateRunner struct {
	runFunc func(cmd *cobra.Command, projectPath, contractPath, entrypoint string, force bool) error
}

func (m *mockValidateRunner) Run(cmd *cobra.Command, projectPath, contractPath, entrypoint string, force bool) error {
	if m.runFunc != nil {
		return m.runFunc(cmd, projectPath, contractPath, entrypoint, force)
	}
	return nil
}
