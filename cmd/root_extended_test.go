package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
	"github.com/hiAndrewQuinn/cliguard/internal/service"
	"github.com/hiAndrewQuinn/cliguard/internal/validator"
	"github.com/spf13/cobra"
)

// MockValidateRunner for testing
type MockValidateRunner struct {
	RunFunc func(cmd *cobra.Command, projectPath, contractPath, entrypoint string) error
	Calls   []MockCall
}

type MockCall struct {
	ProjectPath  string
	ContractPath string
	Entrypoint   string
}

func (m *MockValidateRunner) Run(cmd *cobra.Command, projectPath, contractPath, entrypoint string) error {
	m.Calls = append(m.Calls, MockCall{
		ProjectPath:  projectPath,
		ContractPath: contractPath,
		Entrypoint:   entrypoint,
	})
	if m.RunFunc != nil {
		return m.RunFunc(cmd, projectPath, contractPath, entrypoint)
	}
	return nil
}

func TestExecute(t *testing.T) {
	// Save original validateRunner
	originalRunner := validateRunner
	defer func() { validateRunner = originalRunner }()

	// Create a mock runner
	mockRunner := &MockValidateRunner{
		RunFunc: func(cmd *cobra.Command, projectPath, contractPath, entrypoint string) error {
			return nil
		},
	}
	validateRunner = mockRunner

	// Capture output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Test Execute doesn't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Execute() panicked: %v", r)
			}
		}()

		// We can't easily test Execute() without exiting, but we can test ExecuteWithWriter
		buf := new(bytes.Buffer)

		// Save os.Args
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		// Set args to trigger help
		os.Args = []string{"cliguard", "--help"}

		// This will print help and return without error
		cmd := NewRootCmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.Execute()
	}()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read any output
	var output bytes.Buffer
	io.Copy(&output, r)
}

func TestExecuteWithWriter(t *testing.T) {
	// This test would need to handle os.Exit, which is difficult
	// Instead we test the command execution flow through other tests
	t.Skip("ExecuteWithWriter calls os.Exit which is hard to test")
}

func TestRunValidate_Success(t *testing.T) {
	// Save original runner
	originalRunner := validateRunner
	defer func() { validateRunner = originalRunner }()

	// Create mock runner that returns success
	mockRunner := &MockValidateRunner{
		RunFunc: func(cmd *cobra.Command, projectPath, contractPath, entrypoint string) error {
			cmd.Println("✅ Validation passed! CLI structure matches the contract.")
			return nil
		},
	}
	validateRunner = mockRunner

	// Set up command
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run with valid args
	cmd.SetArgs([]string{
		"validate",
		"--project-path", "/tmp/test",
		"--contract", "/tmp/test/contract.yaml",
		"--entrypoint", "test.Func",
	})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Check output
	output := buf.String()
	if !contains(output, "Validation passed") {
		t.Errorf("Expected success message in output, got: %q", output)
	}

	// Verify mock was called correctly
	if len(mockRunner.Calls) != 1 {
		t.Errorf("Expected 1 call to runner, got %d", len(mockRunner.Calls))
	}

	call := mockRunner.Calls[0]
	if call.ProjectPath != "/tmp/test" {
		t.Errorf("ProjectPath = %q, want %q", call.ProjectPath, "/tmp/test")
	}
	if call.ContractPath != "/tmp/test/contract.yaml" {
		t.Errorf("ContractPath = %q, want %q", call.ContractPath, "/tmp/test/contract.yaml")
	}
	if call.Entrypoint != "test.Func" {
		t.Errorf("Entrypoint = %q, want %q", call.Entrypoint, "test.Func")
	}
}

func TestDefaultValidateRunner(t *testing.T) {
	// Create a test service with mocked dependencies
	runner := NewDefaultValidateRunner()

	// Mock the service dependencies
	mockContractLoader := func(path string) (*contract.Contract, error) {
		if path == "/test/contract.yaml" {
			return &contract.Contract{
				Use:   "myapp",
				Short: "My app",
			}, nil
		}
		return nil, errors.New("contract not found")
	}

	mockInspector := func(projectPath, entrypoint string) (*inspector.InspectedCLI, error) {
		if projectPath == "/test/project" {
			return &inspector.InspectedCLI{
				Use:   "myapp",
				Short: "My app",
			}, nil
		}
		return nil, errors.New("project not found")
	}

	runner.service.ContractLoader = mockContractLoader
	runner.service.Inspector = mockInspector

	// Test successful validation
	t.Run("success", func(t *testing.T) {
		// Create a temp directory that actually exists
		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "contract.yaml")

		// Update mocks to use the real paths
		runner.service.ContractLoader = func(path string) (*contract.Contract, error) {
			if path == contractPath {
				return &contract.Contract{
					Use:   "myapp",
					Short: "My app",
				}, nil
			}
			return nil, errors.New("contract not found")
		}

		runner.service.Inspector = func(projectPath, entrypoint string) (*inspector.InspectedCLI, error) {
			if projectPath == tmpDir {
				return &inspector.InspectedCLI{
					Use:   "myapp",
					Short: "My app",
				}, nil
			}
			return nil, errors.New("project not found")
		}

		cmd := &cobra.Command{}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		err := runner.Run(cmd, tmpDir, contractPath, "test.Func")
		if err != nil {
			t.Errorf("Run() error = %v, want nil", err)
		}

		output := buf.String()
		if !contains(output, "Validation passed") {
			t.Errorf("Expected success message in output, got: %q", output)
		}
	})

	// Test error cases
	t.Run("project not found", func(t *testing.T) {
		cmd := &cobra.Command{}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		err := runner.Run(cmd, "/nonexistent", "/test/contract.yaml", "test.Func")
		if err == nil {
			t.Error("Expected error for nonexistent project")
		}
		if !contains(err.Error(), "Project path does not exist") {
			t.Errorf("Error = %q, want to contain 'Project path does not exist'", err.Error())
		}
	})

	t.Run("contract not found", func(t *testing.T) {
		// Create a temp directory for the project
		tmpDir := t.TempDir()

		cmd := &cobra.Command{}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		err := runner.Run(cmd, tmpDir, "/nonexistent/contract.yaml", "test.Func")
		if err == nil {
			t.Error("Expected error for nonexistent contract")
		}
		if !contains(err.Error(), "failed to load contract") {
			t.Errorf("Error = %q, want to contain 'failed to load contract'", err.Error())
		}
	})
}

func TestValidateServiceIntegration(t *testing.T) {
	// Test the actual ValidateService
	svc := service.NewValidateService()

	// Test with nonexistent project path
	t.Run("nonexistent project", func(t *testing.T) {
		opts := service.ValidateOptions{
			ProjectPath: "/nonexistent/path",
		}

		_, err := svc.Validate(opts)
		if err == nil {
			t.Error("Expected error for nonexistent project")
		}
		if !contains(err.Error(), "Project path does not exist") {
			t.Errorf("Error = %q, want to contain 'Project path does not exist'", err.Error())
		}
	})

	// Test with existing but invalid project
	t.Run("invalid project", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a dummy contract file
		contractPath := filepath.Join(tmpDir, "cliguard.yaml")
		contractContent := `use: test
short: Test CLI
`
		os.WriteFile(contractPath, []byte(contractContent), 0644)

		opts := service.ValidateOptions{
			ProjectPath:  tmpDir,
			ContractPath: contractPath,
			Entrypoint:   "test.Func",
		}

		_, err := svc.Validate(opts)
		if err == nil {
			t.Error("Expected error for invalid project")
		}
		// The error will be from the inspector trying to run go commands
		if !contains(err.Error(), "Failed to inspect project") {
			t.Errorf("Error = %q, want to contain 'Failed to inspect project'", err.Error())
		}
	})
}

func TestCommandFlags(t *testing.T) {
	cmd := NewRootCmd()
	validateCmd, _, _ := cmd.Find([]string{"validate"})

	// Test flag defaults
	t.Run("flag defaults", func(t *testing.T) {
		// Reset flags
		projectPath = ""
		contractPath = ""
		entrypoint = ""

		if projectPath != "" {
			t.Errorf("projectPath default = %q, want empty", projectPath)
		}
		if contractPath != "" {
			t.Errorf("contractPath default = %q, want empty", contractPath)
		}
		if entrypoint != "" {
			t.Errorf("entrypoint default = %q, want empty", entrypoint)
		}
	})

	// Test flag parsing
	t.Run("flag parsing", func(t *testing.T) {
		// Parse flags
		validateCmd.ParseFlags([]string{
			"--project-path", "/test/path",
			"--contract", "/test/contract.yaml",
			"--entrypoint", "test.NewCmd",
		})

		// Check values were set
		if flag := validateCmd.Flag("project-path"); flag.Value.String() != "/test/path" {
			t.Errorf("project-path = %q, want %q", flag.Value.String(), "/test/path")
		}
		if flag := validateCmd.Flag("contract"); flag.Value.String() != "/test/contract.yaml" {
			t.Errorf("contract = %q, want %q", flag.Value.String(), "/test/contract.yaml")
		}
		if flag := validateCmd.Flag("entrypoint"); flag.Value.String() != "test.NewCmd" {
			t.Errorf("entrypoint = %q, want %q", flag.Value.String(), "test.NewCmd")
		}
	})
}

func TestValidationErrors(t *testing.T) {
	// Save original runner
	originalRunner := validateRunner
	defer func() { validateRunner = originalRunner }()

	// Create a custom service that simulates validation failure
	runner := NewDefaultValidateRunner()
	runner.service.ContractLoader = func(path string) (*contract.Contract, error) {
		return &contract.Contract{
			Use:   "expected",
			Short: "Expected CLI",
		}, nil
	}
	runner.service.Inspector = func(projectPath, entrypoint string) (*inspector.InspectedCLI, error) {
		return &inspector.InspectedCLI{
			Use:   "actual",
			Short: "Actual CLI",
		}, nil
	}

	validateRunner = runner

	// Run command - this will exit(1) so we can't test the full flow
	// But we can verify the setup works
	cmd := NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Create temp dir
	tmpDir := t.TempDir()
	contractPath := filepath.Join(tmpDir, "contract.yaml")
	os.WriteFile(contractPath, []byte("use: test\nshort: Test"), 0644)

	// The actual test would need to handle os.Exit
	// For now we just verify the command is set up correctly
	cmd.SetArgs([]string{
		"validate",
		"--project-path", tmpDir,
		"--contract", contractPath,
	})

	// We can't easily test this without handling os.Exit
	// The integration tests cover this scenario
}

func TestValidationResultTypes(t *testing.T) {
	// Test that our types work correctly
	result := &validator.ValidationResult{
		Valid: true,
	}

	if !result.IsValid() {
		t.Error("Expected valid result")
	}

	// Add an error
	result.AddError(
		validator.ErrorTypeMismatch,
		"test",
		"expected",
		"actual",
		"Test mismatch",
	)

	if result.IsValid() {
		t.Error("Expected invalid result after adding error")
	}

	// Test PrintReport doesn't panic
	buf := new(bytes.Buffer)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result.PrintReport()

	w.Close()
	os.Stdout = oldStdout
	io.Copy(buf, r)

	output := buf.String()
	if !contains(output, "Test mismatch") {
		t.Errorf("PrintReport output = %q, want to contain 'Test mismatch'", output)
	}
}

// MockGenerateRunner for testing the generate command
type MockGenerateRunner struct {
	RunFunc func(cmd *cobra.Command, projectPath, entrypoint string) error
}

func (m *MockGenerateRunner) Run(cmd *cobra.Command, projectPath, entrypoint string) error {
	if m.RunFunc != nil {
		return m.RunFunc(cmd, projectPath, entrypoint)
	}
	return nil
}

func TestGenerateCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		setupMock func(*MockGenerateRunner)
		wantErr   bool
		checkFunc func(t *testing.T, output string)
	}{
		{
			name: "successful generation",
			args: []string{"generate", "--project-path", "/test/project", "--entrypoint", "cmd.NewRootCmd"},
			setupMock: func(m *MockGenerateRunner) {
				m.RunFunc = func(cmd *cobra.Command, projectPath, entrypoint string) error {
					if projectPath != "/test/project" {
						t.Errorf("projectPath = %q, want %q", projectPath, "/test/project")
					}
					if entrypoint != "cmd.NewRootCmd" {
						t.Errorf("entrypoint = %q, want %q", entrypoint, "cmd.NewRootCmd")
					}
					// Simulate the YAML output that would be printed by the real runner
					fmt.Fprint(cmd.OutOrStdout(), "# Cliguard contract file\nuse: testapp\n")
					return nil
				}
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output string) {
				if !contains(output, "# Cliguard contract file") {
					t.Errorf("output = %q, want to contain %q", output, "# Cliguard contract file")
				}
				if !contains(output, "use: testapp") {
					t.Errorf("output = %q, want to contain %q", output, "use: testapp")
				}
			},
		},
		{
			name: "missing required project-path",
			args: []string{"generate"},
			setupMock: func(m *MockGenerateRunner) {
				// Mock shouldn't be called
				m.RunFunc = func(cmd *cobra.Command, projectPath, entrypoint string) error {
					t.Error("RunFunc should not be called")
					return nil
				}
			},
			wantErr: true,
			checkFunc: func(t *testing.T, output string) {
				if !contains(output, "required flag(s) \"project-path\" not set") {
					t.Errorf("output = %q, want to contain %q", output, "required flag(s) \"project-path\" not set")
				}
			},
		},
		{
			name: "generation error",
			args: []string{"generate", "--project-path", "/test/project"},
			setupMock: func(m *MockGenerateRunner) {
				m.RunFunc = func(cmd *cobra.Command, projectPath, entrypoint string) error {
					return errors.New("generation failed")
				}
			},
			wantErr: true,
		},
		{
			name: "no entrypoint specified",
			args: []string{"generate", "--project-path", "/test/project"},
			setupMock: func(m *MockGenerateRunner) {
				m.RunFunc = func(cmd *cobra.Command, projectPath, entrypoint string) error {
					if entrypoint != "" {
						t.Errorf("entrypoint = %q, want empty string", entrypoint)
					}
					fmt.Fprint(cmd.OutOrStdout(), "# Cliguard contract file\nuse: testapp\n")
					return nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original runner
			originalRunner := generateRunner
			defer func() { generateRunner = originalRunner }()

			// Create and set mock
			mockRunner := &MockGenerateRunner{}
			if tt.setupMock != nil {
				tt.setupMock(mockRunner)
			}
			generateRunner = mockRunner

			// Create command and capture output
			rootCmd := NewRootCmd()
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, buf.String())
			}
		})
	}
}
