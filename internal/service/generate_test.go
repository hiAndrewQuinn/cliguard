package service

import (
	"os"
	"reflect"
	"testing"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
	"gopkg.in/yaml.v3"
)

func TestGenerateService_inspectedToContract(t *testing.T) {
	service := NewGenerateService()

	tests := []struct {
		name      string
		inspected *inspector.InspectedCLI
		expected  *contract.Contract
	}{
		{
			name: "simple CLI with flags",
			inspected: &inspector.InspectedCLI{
				Use:   "myapp",
				Short: "A simple application",
				Long:  "A simple application with some flags",
				Flags: []inspector.InspectedFlag{
					{
						Name:       "verbose",
						Shorthand:  "v",
						Usage:      "Enable verbose output",
						Type:       "bool",
						Persistent: true,
					},
					{
						Name:  "config",
						Usage: "Config file path",
						Type:  "string",
					},
				},
			},
			expected: &contract.Contract{
				Use:   "myapp",
				Short: "A simple application",
				Long:  "A simple application with some flags",
				Flags: []contract.Flag{
					{
						Name:       "verbose",
						Shorthand:  "v",
						Usage:      "Enable verbose output",
						Type:       "bool",
						Persistent: true,
					},
					{
						Name:  "config",
						Usage: "Config file path",
						Type:  "string",
					},
				},
			},
		},
		{
			name: "CLI with nested commands",
			inspected: &inspector.InspectedCLI{
				Use:   "myapp",
				Short: "An app with commands",
				Commands: []inspector.InspectedCommand{
					{
						Use:   "create",
						Short: "Create resources",
						Commands: []inspector.InspectedCommand{
							{
								Use:   "user",
								Short: "Create a user",
								Flags: []inspector.InspectedFlag{
									{
										Name:  "name",
										Usage: "User name",
										Type:  "string",
									},
								},
							},
						},
					},
					{
						Use:   "delete",
						Short: "Delete resources",
					},
				},
			},
			expected: &contract.Contract{
				Use:   "myapp",
				Short: "An app with commands",
				Commands: []contract.Command{
					{
						Use:   "create",
						Short: "Create resources",
						Commands: []contract.Command{
							{
								Use:   "user",
								Short: "Create a user",
								Flags: []contract.Flag{
									{
										Name:  "name",
										Usage: "User name",
										Type:  "string",
									},
								},
							},
						},
					},
					{
						Use:   "delete",
						Short: "Delete resources",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.inspectedToContract(tt.inspected)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("inspectedToContract() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateService_Generate(t *testing.T) {
	// Note: These tests would require mocking the inspector.InspectProject function
	// For now, we'll skip the actual execution since it requires a real project
	t.Skip("Skipping integration tests that require mocking inspector.InspectProject")

	// Create a temporary directory for test outputs
	tempDir := t.TempDir()

	_ = tempDir // avoid unused variable warning

	tests := []struct {
		name          string
		opts          GenerateOptions
		mockInspected *inspector.InspectedCLI
		wantErr       bool
		checkOutput   func(t *testing.T, outputPath string)
	}{
		{
			name: "successful generation",
			opts: GenerateOptions{
				ProjectPath: tempDir,
				Entrypoint:  "github.com/test/cmd.NewRootCmd",
			},
			mockInspected: &inspector.InspectedCLI{
				Use:   "testapp",
				Short: "A test application",
				Flags: []inspector.InspectedFlag{
					{
						Name:  "debug",
						Usage: "Enable debug mode",
						Type:  "bool",
					},
				},
				Commands: []inspector.InspectedCommand{
					{
						Use:   "serve",
						Short: "Start the server",
					},
				},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, outputPath string) {
				// Read the generated file
				data, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}

				// Unmarshal and verify content
				var generated contract.Contract
				err = yaml.Unmarshal(data, &generated)
				if err != nil {
					t.Fatalf("Failed to unmarshal YAML: %v", err)
				}

				if generated.Use != "testapp" {
					t.Errorf("generated.Use = %q, want %q", generated.Use, "testapp")
				}
				if generated.Short != "A test application" {
					t.Errorf("generated.Short = %q, want %q", generated.Short, "A test application")
				}
				if len(generated.Flags) != 1 {
					t.Errorf("len(generated.Flags) = %d, want 1", len(generated.Flags))
				}
				if len(generated.Flags) > 0 && generated.Flags[0].Name != "debug" {
					t.Errorf("generated.Flags[0].Name = %q, want %q", generated.Flags[0].Name, "debug")
				}
				if len(generated.Commands) != 1 {
					t.Errorf("len(generated.Commands) = %d, want 1", len(generated.Commands))
				}
				if len(generated.Commands) > 0 && generated.Commands[0].Use != "serve" {
					t.Errorf("generated.Commands[0].Use = %q, want %q", generated.Commands[0].Use, "serve")
				}
			},
		},
		{
			name: "absolute output path",
			opts: GenerateOptions{
				ProjectPath: tempDir,
				Entrypoint:  "github.com/test/cmd.NewRootCmd",
			},
			mockInspected: &inspector.InspectedCLI{
				Use:   "absapp",
				Short: "App with absolute path",
			},
			wantErr: false,
			checkOutput: func(t *testing.T, outputPath string) {
				// Verify file exists
				_, err := os.Stat(outputPath)
				if err != nil {
					t.Errorf("Failed to stat generated file: %v", err)
				}
			},
		},
	}

	_ = tests // avoid unused variable warning
}

func TestGenerateService_inspectedFlagsToContractFlags(t *testing.T) {
	service := NewGenerateService()

	inspectedFlags := []inspector.InspectedFlag{
		{
			Name:       "verbose",
			Shorthand:  "v",
			Usage:      "Verbose output",
			Type:       "bool",
			Persistent: true,
		},
		{
			Name:  "config",
			Usage: "Config file",
			Type:  "string",
		},
	}

	result := service.inspectedFlagsToContractFlags(inspectedFlags)

	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}

	if result[0].Name != "verbose" {
		t.Errorf("result[0].Name = %q, want %q", result[0].Name, "verbose")
	}
	if result[0].Shorthand != "v" {
		t.Errorf("result[0].Shorthand = %q, want %q", result[0].Shorthand, "v")
	}
	if result[0].Usage != "Verbose output" {
		t.Errorf("result[0].Usage = %q, want %q", result[0].Usage, "Verbose output")
	}
	if result[0].Type != "bool" {
		t.Errorf("result[0].Type = %q, want %q", result[0].Type, "bool")
	}
	if !result[0].Persistent {
		t.Error("result[0].Persistent = false, want true")
	}

	if result[1].Name != "config" {
		t.Errorf("result[1].Name = %q, want %q", result[1].Name, "config")
	}
	if result[1].Shorthand != "" {
		t.Errorf("result[1].Shorthand = %q, want %q", result[1].Shorthand, "")
	}
	if result[1].Usage != "Config file" {
		t.Errorf("result[1].Usage = %q, want %q", result[1].Usage, "Config file")
	}
	if result[1].Type != "string" {
		t.Errorf("result[1].Type = %q, want %q", result[1].Type, "string")
	}
	if result[1].Persistent {
		t.Error("result[1].Persistent = true, want false")
	}
}

func TestGenerateService_inspectedCommandsToContractCommands(t *testing.T) {
	service := NewGenerateService()

	inspectedCommands := []inspector.InspectedCommand{
		{
			Use:   "create",
			Short: "Create resources",
			Commands: []inspector.InspectedCommand{
				{
					Use:   "user",
					Short: "Create a user",
				},
			},
		},
		{
			Use:   "delete",
			Short: "Delete resources",
		},
	}

	result := service.inspectedCommandsToContractCommands(inspectedCommands)

	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}

	if result[0].Use != "create" {
		t.Errorf("result[0].Use = %q, want %q", result[0].Use, "create")
	}
	if result[0].Short != "Create resources" {
		t.Errorf("result[0].Short = %q, want %q", result[0].Short, "Create resources")
	}
	if len(result[0].Commands) != 1 {
		t.Errorf("len(result[0].Commands) = %d, want 1", len(result[0].Commands))
	}
	if len(result[0].Commands) > 0 && result[0].Commands[0].Use != "user" {
		t.Errorf("result[0].Commands[0].Use = %q, want %q", result[0].Commands[0].Use, "user")
	}

	if result[1].Use != "delete" {
		t.Errorf("result[1].Use = %q, want %q", result[1].Use, "delete")
	}
	if result[1].Short != "Delete resources" {
		t.Errorf("result[1].Short = %q, want %q", result[1].Short, "Delete resources")
	}
	if len(result[1].Commands) != 0 {
		t.Errorf("len(result[1].Commands) = %d, want 0", len(result[1].Commands))
	}
}
