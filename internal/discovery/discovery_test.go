package discovery

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockFileSystem implements filesystem.FileSystem for testing
type MockFileSystem struct {
	Files map[string][]byte
}

func (m *MockFileSystem) MkdirTemp(dir, pattern string) (string, error) {
	return "/tmp/test", nil
}

func (m *MockFileSystem) RemoveAll(path string) error {
	return nil
}

func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	m.Files[name] = data
	return nil
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if data, ok := m.Files[name]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if _, ok := m.Files[name]; ok {
		return nil, nil
	}
	return nil, os.ErrNotExist
}

func TestDiscoverEntrypoints(t *testing.T) {
	tests := []struct {
		name              string
		files             map[string]string
		expectedCount     int
		expectedFirst     string
		expectedFramework string
	}{
		{
			name: "cobra CLI with NewRootCmd",
			files: map[string]string{
				"/project/go.mod": `module github.com/test/project

go 1.21
`,
				"/project/cmd/root.go": `package cmd

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "A test CLI",
	}
	return rootCmd
}
`,
				"/project/main.go": `package main

import "github.com/test/project/cmd"

func main() {
	cmd.Execute()
}
`,
			},
			expectedCount:     2, // NewRootCmd and rootCmd initialization
			expectedFirst:     "func NewRootCmd() *cobra.Command",
			expectedFramework: "cobra",
		},
		{
			name: "urfave/cli framework",
			files: map[string]string{
				"/project/go.mod": `module github.com/test/project

go 1.21
`,
				"/project/main.go": `package main

import (
	"os"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "test",
		Usage: "make an explosive entrance",
	}
	
	app.Run(os.Args)
}
`,
			},
			expectedCount:     2, // app initialization and app.Run
			expectedFirst:     "app := &cli.App{",
			expectedFramework: "urfave/cli",
		},
		{
			name: "standard flag package",
			files: map[string]string{
				"/project/go.mod": `module github.com/test/project

go 1.21
`,
				"/project/main.go": `package main

import "flag"

func main() {
	var name string
	flag.StringVar(&name, "name", "", "your name")
	flag.Parse()
}
`,
			},
			expectedCount:     1, // flag.Parse only (flag definitions aren't counted)
			expectedFirst:     "flag.Parse()",
			expectedFramework: "flag",
		},
		{
			name: "no CLI framework",
			files: map[string]string{
				"/project/go.mod": `module github.com/test/project

go 1.21
`,
				"/project/main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`,
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock filesystem
			mockFS := &MockFileSystem{
				Files: make(map[string][]byte),
			}

			for path, content := range tt.files {
				mockFS.Files[path] = []byte(content)
			}

			// Create a temporary test directory structure
			tempDir := t.TempDir()
			for path, content := range tt.files {
				// Create relative path from /project
				relPath := strings.TrimPrefix(path, "/project/")
				fullPath := filepath.Join(tempDir, relPath)

				// Create directory if needed
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory %s: %v", dir, err)
				}

				// Write file
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", fullPath, err)
				}
			}

			// Create discoverer with real filesystem (since filepath.Walk needs real files)
			discoverer := NewDiscoverer(tempDir, nil)

			// Discover entrypoints
			candidates, err := discoverer.DiscoverEntrypoints()
			if err != nil {
				t.Fatalf("DiscoverEntrypoints() error = %v", err)
			}

			// Check count
			if len(candidates) != tt.expectedCount {
				t.Errorf("Expected %d candidates, got %d", tt.expectedCount, len(candidates))
				for i, c := range candidates {
					t.Logf("Candidate %d: %s (line %d, pattern: %s)",
						i+1, c.Line, c.LineNumber, c.Pattern)
				}
			}

			// Check first candidate if expected
			if tt.expectedCount > 0 {
				if !strings.Contains(candidates[0].Line, tt.expectedFirst) {
					t.Errorf("Expected first candidate to contain %q, got %q",
						tt.expectedFirst, candidates[0].Line)
				}

				if candidates[0].Framework != tt.expectedFramework {
					t.Errorf("Expected framework %q, got %q",
						tt.expectedFramework, candidates[0].Framework)
				}
			}
		})
	}
}

func TestPrintCandidates(t *testing.T) {
	tests := []struct {
		name       string
		candidates []EntrypointCandidate
		wantOutput []string
	}{
		{
			name:       "no candidates",
			candidates: []EntrypointCandidate{},
			wantOutput: []string{
				"No CLI entrypoints found.",
				"Try specifying the entrypoint manually with --entrypoint flag.",
			},
		},
		{
			name: "single candidate with high confidence",
			candidates: []EntrypointCandidate{
				{
					FilePath:          "cmd/root.go",
					LineNumber:        10,
					Line:              "func NewRootCmd() *cobra.Command {",
					Framework:         "cobra",
					Pattern:           "Function returning root cobra.Command",
					Confidence:        95,
					FunctionSignature: "func NewRootCmd() *cobra.Command",
					PackagePath:       "github.com/test/project/cmd",
				},
			},
			wantOutput: []string{
				"Found 1 potential CLI entrypoint(s):",
				"1. cobra (confidence: 95%)",
				"File: cmd/root.go:10",
				"Pattern: Function returning root cobra.Command",
				"Code: func NewRootCmd() *cobra.Command {",
				"Function: func NewRootCmd() *cobra.Command",
				"Package: github.com/test/project/cmd",
				"Suggested entrypoint:",
				"--entrypoint github.com/test/project/cmd.NewRootCmd",
			},
		},
		{
			name: "multiple candidates",
			candidates: []EntrypointCandidate{
				{
					FilePath:    "main.go",
					LineNumber:  20,
					Line:        "app := &cli.App{",
					Framework:   "urfave/cli",
					Pattern:     "CLI app initialization",
					Confidence:  90,
					PackagePath: "github.com/test/project",
				},
				{
					FilePath:    "main.go",
					LineNumber:  25,
					Line:        "flag.Parse()",
					Framework:   "flag",
					Pattern:     "Flag parsing call",
					Confidence:  70,
					PackagePath: "github.com/test/project",
				},
			},
			wantOutput: []string{
				"Found 2 potential CLI entrypoint(s):",
				"1. urfave/cli (confidence: 90%)",
				"2. flag (confidence: 70%)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintCandidates(&buf, tt.candidates, ".", false)

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s",
						want, output)
				}
			}
		})
	}
}

func TestGetModulePath(t *testing.T) {
	mockFS := &MockFileSystem{
		Files: map[string][]byte{
			"/project/go.mod": []byte(`module github.com/test/myproject

go 1.21

require (
	github.com/spf13/cobra v1.8.0
)`),
		},
	}

	discoverer := &Discoverer{
		fs:          mockFS,
		projectPath: "/project",
	}

	modulePath, err := discoverer.getModulePath()
	if err != nil {
		t.Fatalf("getModulePath() error = %v", err)
	}

	expected := "github.com/test/myproject"
	if modulePath != expected {
		t.Errorf("Expected module path %q, got %q", expected, modulePath)
	}
}

func TestCalculatePackagePath(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		filePath   string
		expected   string
	}{
		{
			name:       "root package",
			modulePath: "github.com/test/project",
			filePath:   "main.go",
			expected:   "github.com/test/project",
		},
		{
			name:       "cmd subdirectory",
			modulePath: "github.com/test/project",
			filePath:   "cmd/root.go",
			expected:   "github.com/test/project/cmd",
		},
		{
			name:       "nested subdirectory",
			modulePath: "github.com/test/project",
			filePath:   "internal/cli/commands/root.go",
			expected:   "github.com/test/project/internal/cli/commands",
		},
		{
			name:       "empty module path",
			modulePath: "",
			filePath:   "cmd/root.go",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discoverer := &Discoverer{}
			result := discoverer.calculatePackagePath(tt.modulePath, tt.filePath)

			if result != tt.expected {
				t.Errorf("Expected package path %q, got %q", tt.expected, result)
			}
		})
	}
}
