package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/hiAndrewQuinn/cliguard/internal/discovery"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDiscoverRunner struct {
	runFunc func(cmd *cobra.Command, projectPath string, interactive bool, force bool) error
}

func (m *mockDiscoverRunner) Run(cmd *cobra.Command, projectPath string, interactive bool, force bool) error {
	if m.runFunc != nil {
		return m.runFunc(cmd, projectPath, interactive, force)
	}
	return nil
}

func TestDiscoverCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		runner    DiscoverRunner
		wantErr   bool
		errString string
	}{
		{
			name: "successful discovery",
			args: []string{"discover", "--project-path", "/test/path"},
			runner: &mockDiscoverRunner{
				runFunc: func(cmd *cobra.Command, projectPath string, interactive bool, force bool) error {
					assert.Equal(t, "/test/path", projectPath)
					assert.False(t, interactive)
					assert.False(t, force)
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "discovery with interactive mode",
			args: []string{"discover", "--project-path", "/test/path", "--interactive"},
			runner: &mockDiscoverRunner{
				runFunc: func(cmd *cobra.Command, projectPath string, interactive bool, force bool) error {
					assert.Equal(t, "/test/path", projectPath)
					assert.True(t, interactive)
					assert.False(t, force)
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "discovery with force flag",
			args: []string{"discover", "--project-path", "/test/path", "--force"},
			runner: &mockDiscoverRunner{
				runFunc: func(cmd *cobra.Command, projectPath string, interactive bool, force bool) error {
					assert.Equal(t, "/test/path", projectPath)
					assert.False(t, interactive)
					assert.True(t, force)
					return nil
				},
			},
			wantErr: false,
		},
		{
			name:      "missing required project-path",
			args:      []string{"discover"},
			runner:    &mockDiscoverRunner{},
			wantErr:   true,
			errString: "required flag(s) \"project-path\" not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore the global runner
			oldRunner := discoverRunner
			discoverRunner = tt.runner
			defer func() { discoverRunner = oldRunner }()

			cmd := NewRootCmd()
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultDiscoverRunner(t *testing.T) {
	t.Run("with cobra project", func(t *testing.T) {
		// Create a temporary test project
		tempDir := t.TempDir()
		createTestCobraProject(t, tempDir)

		cmd := &cobra.Command{}
		var buf bytes.Buffer
		cmd.SetOut(&buf)

		runner := NewDefaultDiscoverRunner()
		err := runner.Run(cmd, tempDir, false, false)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Found")
		assert.Contains(t, output, "cobra")
	})

	t.Run("with no CLI project", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestNonCLIProject(t, tempDir)

		cmd := &cobra.Command{}
		var buf bytes.Buffer
		cmd.SetOut(&buf)

		runner := NewDefaultDiscoverRunner()
		err := runner.Run(cmd, tempDir, false, false)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "No CLI entrypoints found")
	})

	t.Run("interactive mode with single candidate", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestCobraProject(t, tempDir)

		cmd := &cobra.Command{}
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetIn(strings.NewReader("1\n"))

		runner := NewDefaultDiscoverRunner()
		err := runner.Run(cmd, tempDir, true, false)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Selected entrypoint:")
	})

	t.Run("project path does not exist", func(t *testing.T) {
		cmd := &cobra.Command{}
		runner := NewDefaultDiscoverRunner()
		err := runner.Run(cmd, "/nonexistent/path", false, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})
}

func TestDiscoverIntegration(t *testing.T) {
	t.Run("discover cobra project", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestCobraProject(t, tempDir)

		cmd := NewRootCmd()
		cmd.SetArgs([]string{"discover", "--project-path", tempDir})

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "cobra")
		assert.Contains(t, output, "Suggested entrypoint:")
	})

	t.Run("discover with interactive selection", func(t *testing.T) {
		tempDir := t.TempDir()
		createTestCobraProject(t, tempDir)

		cmd := NewRootCmd()
		cmd.SetArgs([]string{"discover", "--project-path", tempDir, "--interactive"})
		cmd.SetIn(strings.NewReader("1\n"))

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Selected entrypoint:")
	})
}

func TestInteractiveSelection(t *testing.T) {
	tests := []struct {
		name       string
		candidates []discovery.EntrypointCandidate
		input      string
		wantErr    bool
		wantIndex  int
	}{
		{
			name: "select first option",
			candidates: []discovery.EntrypointCandidate{
				{Framework: "cobra", FilePath: "cmd/root.go"},
				{Framework: "flag", FilePath: "main.go"},
			},
			input:     "1\n",
			wantErr:   false,
			wantIndex: 0,
		},
		{
			name: "quit selection",
			candidates: []discovery.EntrypointCandidate{
				{Framework: "cobra", FilePath: "cmd/root.go"},
				{Framework: "flag", FilePath: "main.go"},
			},
			input:   "q\n",
			wantErr: true,
		},
		{
			name: "invalid then valid selection",
			candidates: []discovery.EntrypointCandidate{
				{Framework: "cobra", FilePath: "cmd/root.go"},
				{Framework: "flag", FilePath: "main.go"},
			},
			input:     "invalid\n2\n",
			wantErr:   false,
			wantIndex: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdin := strings.NewReader(tt.input)
			stdout := &bytes.Buffer{}

			selector := discovery.NewInteractiveSelector(stdin, stdout)
			selected, err := selector.SelectCandidate(tt.candidates)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.candidates[tt.wantIndex], *selected)
			}
		})
	}
}

// Helper functions to create test projects
func createTestCobraProject(t *testing.T, dir string) {
	// Create a simple Cobra project structure
	cmdDir := dir + "/cmd"
	require.NoError(t, os.MkdirAll(cmdDir, 0755))

	rootFile := cmdDir + "/root.go"
	content := `package cmd

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "A test CLI",
	}
	return rootCmd
}

func Execute() {
	NewRootCmd().Execute()
}
`
	require.NoError(t, os.WriteFile(rootFile, []byte(content), 0644))

	// Create go.mod
	goMod := dir + "/go.mod"
	modContent := `module test-project

go 1.21

require github.com/spf13/cobra v1.8.0
`
	require.NoError(t, os.WriteFile(goMod, []byte(modContent), 0644))
}

func createTestNonCLIProject(t *testing.T, dir string) {
	mainFile := dir + "/main.go"
	content := `package main

import "fmt"

func main() {
	fmt.Println("Not a CLI")
}
`
	require.NoError(t, os.WriteFile(mainFile, []byte(content), 0644))

	goMod := dir + "/go.mod"
	modContent := `module test-project

go 1.21
`
	require.NoError(t, os.WriteFile(goMod, []byte(modContent), 0644))
}
