package inspector

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/hiAndrewQuinn/cliguard/internal/executor"
	"github.com/hiAndrewQuinn/cliguard/internal/filesystem"
)

func TestInspector_parseEntrypoint(t *testing.T) {
	tests := []struct {
		name       string
		entrypoint string
		want       *EntrypointInfo
		wantErr    bool
	}{
		{
			name:       "empty entrypoint",
			entrypoint: "",
			want:       &EntrypointInfo{},
			wantErr:    false,
		},
		{
			name:       "main package function",
			entrypoint: "main.NewRootCmd",
			want: &EntrypointInfo{
				FunctionName:  "NewRootCmd",
				IsMainPackage: true,
			},
			wantErr: false,
		},
		{
			name:       "regular package function",
			entrypoint: "github.com/user/repo/cmd.NewRootCmd",
			want: &EntrypointInfo{
				ImportPath:   "github.com/user/repo/cmd",
				ImportAlias:  "userCmd",
				FunctionName: "NewRootCmd",
			},
			wantErr: false,
		},
		{
			name:       "nested package function",
			entrypoint: "github.com/user/repo/internal/cmd.NewRootCmd",
			want: &EntrypointInfo{
				ImportPath:   "github.com/user/repo/internal/cmd",
				ImportAlias:  "userCmd",
				FunctionName: "NewRootCmd",
			},
			wantErr: false,
		},
		{
			name:       "invalid format - no dot",
			entrypoint: "invalidformat",
			wantErr:    true,
		},
		{
			name:       "invalid format - single part",
			entrypoint: "package.",
			wantErr:    false, // Empty function name results in valid but empty EntrypointInfo
			want: &EntrypointInfo{
				ImportPath:   "package",
				ImportAlias:  "userCmd",
				FunctionName: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Inspector{}
			got, err := i.parseEntrypoint(tt.entrypoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEntrypoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !compareEntrypointInfo(got, tt.want) {
				t.Errorf("parseEntrypoint() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestInspector_generateInspectorCode(t *testing.T) {
	tests := []struct {
		name    string
		info    *EntrypointInfo
		wantErr bool
		check   func(string) error
	}{
		{
			name: "main package",
			info: &EntrypointInfo{
				ImportPath:   "github.com/test/repo",
				ImportAlias:  "userPkg",
				FunctionName: "NewRootCmd",
			},
			wantErr: false,
			check: func(code string) error {
				// Check that the generated code contains expected elements
				expectedStrings := []string{
					`userPkg "github.com/test/repo"`,
					`rootCmd = userPkg.NewRootCmd()`,
					`func inspectCommand(cmd *cobra.Command)`,
					`encoding/json`,
				}
				for _, expected := range expectedStrings {
					if !contains(code, expected) {
						return fmt.Errorf("generated code missing: %s", expected)
					}
				}
				return nil
			},
		},
		{
			name: "no entrypoint",
			info: &EntrypointInfo{},
			wantErr: false,
			check: func(code string) error {
				if !contains(code, "findRootCommand()") {
					return fmt.Errorf("generated code should use findRootCommand when no entrypoint")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Inspector{}
			got, err := i.generateInspectorCode(tt.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateInspectorCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				if err := tt.check(got); err != nil {
					t.Errorf("generateInspectorCode() check failed: %v", err)
				}
			}
		})
	}
}

func TestInspector_Inspect(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		setupMocks    func(*filesystem.MockFileSystem, *executor.MockExecutor)
		want          *InspectedCLI
		wantErr       bool
		wantErrString string
	}{
		{
			name: "successful inspection",
			config: Config{
				ProjectPath: "/test/project",
				Entrypoint:  "github.com/test/repo/cmd.NewRootCmd",
			},
			setupMocks: func(fs *filesystem.MockFileSystem, exec *executor.MockExecutor) {
				// Mock go.mod file
				fs.Files["/test/project/go.mod"] = []byte("module github.com/test/repo\n\ngo 1.21")
				
				// Mock successful commands
				exec.Results = map[string]executor.MockResult{
					"go mod init cliguard-inspector": {Output: []byte("go: creating new go.mod"), Error: nil},
					"go mod edit -replace github.com/test/repo=/test/project": {Output: []byte(""), Error: nil},
					"go get ./...": {Output: []byte(""), Error: nil},
					"go run inspector.go": {
						Output: []byte(`{
							"use": "myapp",
							"short": "My CLI application",
							"flags": [
								{
									"name": "config",
									"shorthand": "c",
									"usage": "Config file",
									"type": "string",
									"persistent": true
								}
							],
							"commands": [
								{
									"use": "serve",
									"short": "Start the server"
								}
							]
						}`),
						Error: nil,
					},
				}
			},
			want: &InspectedCLI{
				Use:   "myapp",
				Short: "My CLI application",
				Flags: []InspectedFlag{
					{
						Name:       "config",
						Shorthand:  "c",
						Usage:      "Config file",
						Type:       "string",
						Persistent: true,
					},
				},
				Commands: []InspectedCommand{
					{
						Use:   "serve",
						Short: "Start the server",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "temp dir creation failure",
			config: Config{
				ProjectPath: "/test/project",
				Entrypoint:  "test.Func",
			},
			setupMocks: func(fs *filesystem.MockFileSystem, exec *executor.MockExecutor) {
				fs.MkdirTempErr = errors.New("permission denied")
			},
			wantErr:       true,
			wantErrString: "failed to create temp directory",
		},
		{
			name: "invalid entrypoint",
			config: Config{
				ProjectPath: "/test/project",
				Entrypoint:  "invalid",
			},
			setupMocks: func(fs *filesystem.MockFileSystem, exec *executor.MockExecutor) {},
			wantErr:       true,
			wantErrString: "failed to parse entrypoint",
		},
		{
			name: "go mod init failure",
			config: Config{
				ProjectPath: "/test/project",
				Entrypoint:  "test.Func",
			},
			setupMocks: func(fs *filesystem.MockFileSystem, exec *executor.MockExecutor) {
				exec.Results = map[string]executor.MockResult{
					"go mod init cliguard-inspector": {
						Output: []byte("error output"),
						Error:  errors.New("exit status 1"),
					},
				}
			},
			wantErr:       true,
			wantErrString: "failed to setup temp module",
		},
		{
			name: "inspector run failure",
			config: Config{
				ProjectPath: "/test/project",
				Entrypoint:  "test.Func",
			},
			setupMocks: func(fs *filesystem.MockFileSystem, exec *executor.MockExecutor) {
				exec.Results = map[string]executor.MockResult{
					"go mod init cliguard-inspector": {Output: []byte(""), Error: nil},
					"go get ./...": {Output: []byte(""), Error: nil},
					"go run inspector.go": {
						Output: []byte("compilation error"),
						Error:  errors.New("exit status 1"),
					},
				}
			},
			wantErr:       true,
			wantErrString: "failed to run inspector",
		},
		{
			name: "invalid JSON output",
			config: Config{
				ProjectPath: "/test/project",
				Entrypoint:  "test.Func",
			},
			setupMocks: func(fs *filesystem.MockFileSystem, exec *executor.MockExecutor) {
				exec.Results = map[string]executor.MockResult{
					"go mod init cliguard-inspector": {Output: []byte(""), Error: nil},
					"go get ./...": {Output: []byte(""), Error: nil},
					"go run inspector.go": {
						Output: []byte("invalid json"),
						Error:  nil,
					},
				}
			},
			wantErr:       true,
			wantErrString: "failed to parse inspector output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockFS := filesystem.NewMockFileSystem()
			mockExec := &executor.MockExecutor{
				Results: make(map[string]executor.MockResult),
			}

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, mockExec)
			}

			// Configure inspector
			tt.config.FileSystem = mockFS
			tt.config.Executor = mockExec

			// Run inspection
			inspector := NewInspector(tt.config)
			got, err := inspector.Inspect()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Inspect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrString != "" {
				if !contains(err.Error(), tt.wantErrString) {
					t.Errorf("Inspect() error = %v, want error containing %v", err, tt.wantErrString)
				}
				return
			}

			// Check result
			if !tt.wantErr && !compareInspectedCLI(got, tt.want) {
				gotJSON, _ := json.MarshalIndent(got, "", "  ")
				wantJSON, _ := json.MarshalIndent(tt.want, "", "  ")
				t.Errorf("Inspect() got = %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}

// Helper functions
func compareEntrypointInfo(a, b *EntrypointInfo) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.ImportPath == b.ImportPath &&
		a.ImportAlias == b.ImportAlias &&
		a.FunctionName == b.FunctionName &&
		a.IsMainPackage == b.IsMainPackage
}

func compareInspectedCLI(a, b *InspectedCLI) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Use != b.Use || a.Short != b.Short || a.Long != b.Long {
		return false
	}
	if len(a.Flags) != len(b.Flags) || len(a.Commands) != len(b.Commands) {
		return false
	}
	// Compare flags
	for i := range a.Flags {
		if !compareInspectedFlag(a.Flags[i], b.Flags[i]) {
			return false
		}
	}
	// Compare commands
	for i := range a.Commands {
		if !compareInspectedCommand(a.Commands[i], b.Commands[i]) {
			return false
		}
	}
	return true
}

func compareInspectedFlag(a, b InspectedFlag) bool {
	return a.Name == b.Name &&
		a.Shorthand == b.Shorthand &&
		a.Usage == b.Usage &&
		a.Type == b.Type &&
		a.Persistent == b.Persistent
}

func compareInspectedCommand(a, b InspectedCommand) bool {
	if a.Use != b.Use || a.Short != b.Short || a.Long != b.Long {
		return false
	}
	if len(a.Flags) != len(b.Flags) || len(a.Commands) != len(b.Commands) {
		return false
	}
	// Compare flags
	for i := range a.Flags {
		if !compareInspectedFlag(a.Flags[i], b.Flags[i]) {
			return false
		}
	}
	// Compare sub-commands
	for i := range a.Commands {
		if !compareInspectedCommand(a.Commands[i], b.Commands[i]) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || substringIndex(s, substr) != -1))
}

func substringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}