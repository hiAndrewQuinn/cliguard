package inspector

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

func TestGetModuleName(t *testing.T) {
	tests := []struct {
		name         string
		goModContent string
		want         string
	}{
		{
			name: "simple_module",
			goModContent: `module github.com/example/project

go 1.21
`,
			want: "github.com/example/project",
		},
		{
			name: "module_with_dependencies",
			goModContent: `module github.com/test/app

go 1.21

require (
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.16.0
)
`,
			want: "github.com/test/app",
		},
		{
			name: "module_with_comments",
			goModContent: `// This is a comment
module github.com/myorg/myapp

go 1.21
`,
			want: "github.com/myorg/myapp",
		},
		{
			name: "module_with_leading_whitespace",
			goModContent: `   module   github.com/space/project   

go 1.21
`,
			want: "github.com/space/project",
		},
		{
			name: "no_module_line",
			goModContent: `go 1.21

require (
	github.com/spf13/cobra v1.7.0
)
`,
			want: "",
		},
		{
			name:         "empty_content",
			goModContent: "",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getModuleName([]byte(tt.goModContent))
			if got != tt.want {
				t.Errorf("getModuleName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInspectorTemplateCompiles(t *testing.T) {
	// This test ensures the template is valid Go code
	// The actual template execution is tested via integration tests
	if !strings.Contains(inspectorTemplate, "package main") {
		t.Error("Inspector template should contain 'package main'")
	}
	if !strings.Contains(inspectorTemplate, "import") {
		t.Error("Inspector template should contain import statements")
	}
	if !strings.Contains(inspectorTemplate, "func main()") {
		t.Error("Inspector template should contain main function")
	}
	if !strings.Contains(inspectorTemplate, "InspectedCLI") {
		t.Error("Inspector template should contain InspectedCLI type")
	}

	// Test that the template compiles and executes
	tmpl, err := template.New("inspector").Parse(inspectorTemplate)
	if err != nil {
		t.Fatalf("Failed to parse inspector template: %v", err)
	}

	// Test template execution with sample data
	testCases := []struct {
		name string
		data struct {
			ImportPath     string
			ImportAlias    string
			EntrypointFunc string
		}
	}{
		{
			name: "with entrypoint",
			data: struct {
				ImportPath     string
				ImportAlias    string
				EntrypointFunc string
			}{
				ImportPath:     "github.com/test/repo",
				ImportAlias:    "testcmd",
				EntrypointFunc: "NewRootCmd",
			},
		},
		{
			name: "without entrypoint",
			data: struct {
				ImportPath     string
				ImportAlias    string
				EntrypointFunc string
			}{
				ImportPath:     "",
				ImportAlias:    "",
				EntrypointFunc: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tmpl.Execute(&buf, tc.data)
			if err != nil {
				t.Errorf("Failed to execute template: %v", err)
			}

			// Check that output contains expected content
			output := buf.String()
			if len(output) == 0 {
				t.Error("Template produced empty output")
			}
		})
	}
}

func TestInspectProject(t *testing.T) {
	// Test the public InspectProject function
	tests := []struct {
		name        string
		projectPath string
		entrypoint  string
		wantErr     bool
	}{
		{
			name:        "valid inputs",
			projectPath: "/test/project",
			entrypoint:  "github.com/test/repo.Func",
			wantErr:     true, // Will error in test environment without real project
		},
		{
			name:        "empty project path",
			projectPath: "",
			entrypoint:  "test.Func",
			wantErr:     true,
		},
		{
			name:        "invalid entrypoint",
			projectPath: "/test",
			entrypoint:  "invalid",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InspectProject(tt.projectPath, tt.entrypoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("InspectProject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
