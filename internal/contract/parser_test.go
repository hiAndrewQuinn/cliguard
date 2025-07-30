package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, c *Contract)
	}{
		{
			name: "valid_simple_contract",
			yamlContent: `
use: testcli
short: Test CLI application
flags:
  - name: config
    shorthand: c
    usage: Config file path
    type: string
    persistent: true
commands:
  - use: serve
    short: Start the server
    flags:
      - name: port
        shorthand: p
        usage: Port number
        type: int
`,
			wantErr: false,
			validate: func(t *testing.T, c *Contract) {
				if c.Use != "testcli" {
					t.Errorf("Use = %q, want %q", c.Use, "testcli")
				}
				if c.Short != "Test CLI application" {
					t.Errorf("Short = %q, want %q", c.Short, "Test CLI application")
				}
				if len(c.Flags) != 1 {
					t.Errorf("len(Flags) = %d, want 1", len(c.Flags))
				}
				if len(c.Commands) != 1 {
					t.Errorf("len(Commands) = %d, want 1", len(c.Commands))
				}
			},
		},
		{
			name: "valid_with_long_description",
			yamlContent: `
use: testcli
short: Test CLI
long: This is a longer description
`,
			wantErr: false,
			validate: func(t *testing.T, c *Contract) {
				if c.Long != "This is a longer description" {
					t.Errorf("Long = %q, want %q", c.Long, "This is a longer description")
				}
			},
		},
		{
			name: "invalid_empty_use",
			yamlContent: `
use: ""
short: Test CLI
`,
			wantErr:     true,
			errContains: "root command 'use' field cannot be empty",
		},
		{
			name: "invalid_flag_no_name",
			yamlContent: `
use: testcli
short: Test CLI
flags:
  - shorthand: c
    usage: Config file
    type: string
`,
			wantErr:     true,
			errContains: "flag name cannot be empty",
		},
		{
			name: "invalid_flag_no_type",
			yamlContent: `
use: testcli
short: Test CLI
flags:
  - name: config
    usage: Config file
`,
			wantErr:     true,
			errContains: "flag 'config': type cannot be empty",
		},
		{
			name: "invalid_flag_bad_type",
			yamlContent: `
use: testcli
short: Test CLI
flags:
  - name: config
    usage: Config file
    type: badtype
`,
			wantErr:     true,
			errContains: "flag 'config': invalid type 'badtype'",
		},
		{
			name: "invalid_duplicate_flag_names",
			yamlContent: `
use: testcli
short: Test CLI
flags:
  - name: config
    type: string
    usage: Config file
  - name: config
    type: bool
    usage: Another config
`,
			wantErr:     true,
			errContains: "duplicate flag name: config",
		},
		{
			name: "invalid_duplicate_flag_shorthands",
			yamlContent: `
use: testcli
short: Test CLI
flags:
  - name: config
    shorthand: c
    type: string
    usage: Config file
  - name: cache
    shorthand: c
    type: bool
    usage: Enable cache
`,
			wantErr:     true,
			errContains: "duplicate flag shorthand: c",
		},
		{
			name: "invalid_long_shorthand",
			yamlContent: `
use: testcli
short: Test CLI
flags:
  - name: config
    shorthand: cfg
    type: string
    usage: Config file
`,
			wantErr:     true,
			errContains: "flag shorthand must be a single character: cfg",
		},
		{
			name: "nested_commands",
			yamlContent: `
use: testcli
short: Test CLI
commands:
  - use: db
    short: Database commands
    commands:
      - use: migrate
        short: Run migrations
        flags:
          - name: force
            shorthand: f
            type: bool
            usage: Force migration
`,
			wantErr: false,
			validate: func(t *testing.T, c *Contract) {
				if len(c.Commands) != 1 {
					t.Fatalf("len(Commands) = %d, want 1", len(c.Commands))
				}
				if len(c.Commands[0].Commands) != 1 {
					t.Errorf("len(Commands[0].Commands) = %d, want 1", len(c.Commands[0].Commands))
				}
			},
		},
		{
			name: "all_flag_types",
			yamlContent: `
use: testcli
short: Test CLI
flags:
  - name: string-flag
    type: string
    usage: String flag
  - name: bool-flag
    type: bool
    usage: Bool flag
  - name: int-flag
    type: int
    usage: Int flag
  - name: int64-flag
    type: int64
    usage: Int64 flag
  - name: float64-flag
    type: float64
    usage: Float64 flag
  - name: duration-flag
    type: duration
    usage: Duration flag
  - name: slice-flag
    type: stringSlice
    usage: String slice flag
`,
			wantErr: false,
			validate: func(t *testing.T, c *Contract) {
				if len(c.Flags) != 7 {
					t.Errorf("len(Flags) = %d, want 7", len(c.Flags))
				}
			},
		},
		{
			name: "all_new_flag_types",
			yamlContent: `
use: testcli
short: Test CLI with new flag types
flags:
  # Integer variants
  - name: int8-flag
    type: int8
    usage: Int8 flag
  - name: int16-flag
    type: int16
    usage: Int16 flag
  - name: int32-flag
    type: int32
    usage: Int32 flag
  - name: uint-flag
    type: uint
    usage: Uint flag
  - name: uint8-flag
    type: uint8
    usage: Uint8 flag
  - name: uint16-flag
    type: uint16
    usage: Uint16 flag
  - name: uint32-flag
    type: uint32
    usage: Uint32 flag
  - name: uint64-flag
    type: uint64
    usage: Uint64 flag
  # Float variants
  - name: float32-flag
    type: float32
    usage: Float32 flag
  # Slice types
  - name: int-slice-flag
    type: intSlice
    usage: Int slice flag
  - name: int32-slice-flag
    type: int32Slice
    usage: Int32 slice flag
  - name: int64-slice-flag
    type: int64Slice
    usage: Int64 slice flag
  - name: uint-slice-flag
    type: uintSlice
    usage: Uint slice flag
  - name: float32-slice-flag
    type: float32Slice
    usage: Float32 slice flag
  - name: float64-slice-flag
    type: float64Slice
    usage: Float64 slice flag
  - name: bool-slice-flag
    type: boolSlice
    usage: Bool slice flag
  - name: duration-slice-flag
    type: durationSlice
    usage: Duration slice flag
  # Map types
  - name: string-to-string-flag
    type: stringToString
    usage: String to string map flag
  - name: string-to-int64-flag
    type: stringToInt64
    usage: String to int64 map flag
  # Network types
  - name: ip-flag
    type: ip
    usage: IP address flag
  - name: ip-slice-flag
    type: ipSlice
    usage: IP slice flag
  - name: ip-mask-flag
    type: ipMask
    usage: IP mask flag
  - name: ip-net-flag
    type: ipNet
    usage: IP net flag
  # Binary types
  - name: bytes-hex-flag
    type: bytesHex
    usage: Bytes hex flag
  - name: bytes-base64-flag
    type: bytesBase64
    usage: Bytes base64 flag
  # Special types
  - name: count-flag
    type: count
    usage: Count flag
`,
			wantErr: false,
			validate: func(t *testing.T, c *Contract) {
				if len(c.Flags) != 26 {
					t.Errorf("len(Flags) = %d, want 26", len(c.Flags))
				}
				// Verify all types are parsed correctly
				expectedTypes := []string{
					"int8", "int16", "int32", "uint", "uint8", "uint16", "uint32", "uint64",
					"float32", "intSlice", "int32Slice", "int64Slice", "uintSlice",
					"float32Slice", "float64Slice", "boolSlice", "durationSlice",
					"stringToString", "stringToInt64", "ip", "ipSlice", "ipMask", "ipNet",
					"bytesHex", "bytesBase64", "count",
				}
				if len(expectedTypes) != len(c.Flags) {
					t.Errorf("expectedTypes count mismatch: got %d, want %d", len(expectedTypes), len(c.Flags))
				}
				for i, expectedType := range expectedTypes {
					if i < len(c.Flags) && c.Flags[i].Type != expectedType {
						t.Errorf("Flag[%d].Type = %q, want %q", i, c.Flags[i].Type, expectedType)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "contract.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}

			// Test Load
			got, err := Load(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, want error containing %q", err, tt.errContains)
				}
			}

			if err == nil && tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() string
		wantErr     bool
		errContains string
	}{
		{
			name: "empty_contract_path",
			setupFunc: func() string {
				return ""
			},
			wantErr:     true,
			errContains: "contract path cannot be empty",
		},
		{
			name: "file_not_found",
			setupFunc: func() string {
				return "/nonexistent/file.yaml"
			},
			wantErr:     true,
			errContains: "failed to read contract file",
		},
		{
			name: "invalid_yaml",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				tmpFile := filepath.Join(tmpDir, "invalid.yaml")
				os.WriteFile(tmpFile, []byte("invalid: yaml: content:\n  - this is bad"), 0644)
				return tmpFile
			},
			wantErr:     true,
			errContains: "failed to parse contract YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contractPath := tt.setupFunc()

			_, err := Load(contractPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
