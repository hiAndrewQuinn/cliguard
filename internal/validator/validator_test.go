package validator

import (
	"testing"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		expected *contract.Contract
		actual   *inspector.InspectedCLI
		wantErrs []ValidationError
	}{
		{
			name: "exact_match",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []contract.Flag{
					{Name: "config", Shorthand: "c", Usage: "Config file", Type: "string", Persistent: true},
				},
				Commands: []contract.Command{
					{
						Use:   "serve",
						Short: "Start server",
						Flags: []contract.Flag{
							{Name: "port", Shorthand: "p", Usage: "Port", Type: "int"},
						},
					},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []inspector.InspectedFlag{
					{Name: "config", Shorthand: "c", Usage: "Config file", Type: "string", Persistent: true},
				},
				Commands: []inspector.InspectedCommand{
					{
						Use:   "serve",
						Short: "Start server",
						Flags: []inspector.InspectedFlag{
							{Name: "port", Shorthand: "p", Usage: "Port", Type: "int"},
						},
					},
				},
			},
			wantErrs: nil,
		},
		{
			name: "root_use_mismatch",
			expected: &contract.Contract{
				Use:   "expectedcli",
				Short: "Test CLI",
			},
			actual: &inspector.InspectedCLI{
				Use:   "actualcli",
				Short: "Test CLI",
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeMismatch, Path: "root", Expected: "expectedcli", Actual: "actualcli"},
			},
		},
		{
			name: "root_short_mismatch",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Expected description",
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Actual description",
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeMismatch, Path: "root", Expected: "Expected description", Actual: "Actual description"},
			},
		},
		{
			name: "missing_flag",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []contract.Flag{
					{Name: "config", Type: "string", Usage: "Config"},
					{Name: "verbose", Type: "bool", Usage: "Verbose"},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []inspector.InspectedFlag{
					{Name: "config", Type: "string", Usage: "Config"},
				},
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeMissing, Path: "--verbose", Expected: "verbose"},
			},
		},
		{
			name: "unexpected_flag",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []contract.Flag{
					{Name: "config", Type: "string", Usage: "Config"},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []inspector.InspectedFlag{
					{Name: "config", Type: "string", Usage: "Config"},
					{Name: "debug", Type: "bool", Usage: "Debug"},
				},
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeUnexpected, Path: "--debug", Actual: "debug"},
			},
		},
		{
			name: "flag_type_mismatch",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []contract.Flag{
					{Name: "port", Type: "string", Usage: "Port"},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []inspector.InspectedFlag{
					{Name: "port", Type: "int", Usage: "Port"},
				},
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeInvalidType, Path: "--port", Expected: "string", Actual: "int"},
			},
		},
		{
			name: "flag_shorthand_mismatch",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []contract.Flag{
					{Name: "config", Shorthand: "c", Type: "string", Usage: "Config"},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []inspector.InspectedFlag{
					{Name: "config", Shorthand: "k", Type: "string", Usage: "Config"},
				},
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeMismatch, Path: "--config", Expected: "c", Actual: "k"},
			},
		},
		{
			name: "flag_persistence_mismatch",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []contract.Flag{
					{Name: "config", Type: "string", Usage: "Config", Persistent: true},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Flags: []inspector.InspectedFlag{
					{Name: "config", Type: "string", Usage: "Config", Persistent: false},
				},
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeMismatch, Path: "--config", Expected: "persistent", Actual: "local"},
			},
		},
		{
			name: "missing_command",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Commands: []contract.Command{
					{Use: "serve", Short: "Serve"},
					{Use: "migrate", Short: "Migrate"},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Commands: []inspector.InspectedCommand{
					{Use: "serve", Short: "Serve"},
				},
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeMissing, Path: "migrate", Expected: "migrate"},
			},
		},
		{
			name: "unexpected_command",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Commands: []contract.Command{
					{Use: "serve", Short: "Serve"},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Commands: []inspector.InspectedCommand{
					{Use: "serve", Short: "Serve"},
					{Use: "debug", Short: "Debug"},
				},
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeUnexpected, Path: "debug", Actual: "debug"},
			},
		},
		{
			name: "nested_command_validation",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Commands: []contract.Command{
					{
						Use:   "db",
						Short: "Database",
						Commands: []contract.Command{
							{Use: "migrate", Short: "Migrate DB"},
						},
					},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Commands: []inspector.InspectedCommand{
					{
						Use:   "db",
						Short: "Database",
						Commands: []inspector.InspectedCommand{
							{Use: "seed", Short: "Seed DB"},
						},
					},
				},
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeMissing, Path: "db migrate", Expected: "migrate"},
				{Type: ErrorTypeUnexpected, Path: "db seed", Actual: "seed"},
			},
		},
		{
			name: "command_flag_validation",
			expected: &contract.Contract{
				Use:   "testcli",
				Short: "Test CLI",
				Commands: []contract.Command{
					{
						Use:   "serve",
						Short: "Serve",
						Flags: []contract.Flag{
							{Name: "port", Type: "int", Usage: "Port"},
						},
					},
				},
			},
			actual: &inspector.InspectedCLI{
				Use:   "testcli",
				Short: "Test CLI",
				Commands: []inspector.InspectedCommand{
					{
						Use:   "serve",
						Short: "Serve",
						Flags: []inspector.InspectedFlag{
							{Name: "host", Type: "string", Usage: "Host"},
						},
					},
				},
			},
			wantErrs: []ValidationError{
				{Type: ErrorTypeMissing, Path: "serve --port", Expected: "port"},
				{Type: ErrorTypeUnexpected, Path: "serve --host", Actual: "host"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.expected, tt.actual)

			if tt.wantErrs == nil {
				if !result.IsValid() {
					t.Errorf("Validate() expected no errors, but got %d errors", len(result.Errors))
					for _, err := range result.Errors {
						t.Logf("  Error: %+v", err)
					}
				}
			} else {
				if result.IsValid() {
					t.Errorf("Validate() expected errors, but got none")
				}

				if len(result.Errors) != len(tt.wantErrs) {
					t.Errorf("Validate() got %d errors, want %d", len(result.Errors), len(tt.wantErrs))
					for i, err := range result.Errors {
						t.Logf("  Error %d: %+v", i, err)
					}
				}

				// Check that expected errors are present
				for _, wantErr := range tt.wantErrs {
					found := false
					for _, gotErr := range result.Errors {
						if errorsMatch(wantErr, gotErr) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error not found: %+v", wantErr)
					}
				}
			}
		})
	}
}

func TestValidationResult_PrintReport(t *testing.T) {
	// This is mainly for coverage - actual output is tested via integration tests
	result := &ValidationResult{
		Errors: []ValidationError{
			{Type: ErrorTypeMissing, Path: "test --flag", Expected: "flag", Message: "flag"},
			{Type: ErrorTypeUnexpected, Path: "test --extra", Actual: "extra", Message: "flag"},
			{Type: ErrorTypeMismatch, Path: "test", Expected: "exp", Actual: "act", Message: "Mismatch"},
			{Type: ErrorTypeInvalidType, Path: "test --type", Expected: "string", Actual: "int", Message: "Type mismatch"},
		},
	}

	// Just ensure it doesn't panic
	result.PrintReport()
}

func errorsMatch(want, got ValidationError) bool {
	// For matching, we only care about Type, Path, and Expected/Actual values
	return want.Type == got.Type &&
		want.Path == got.Path &&
		want.Expected == got.Expected &&
		want.Actual == got.Actual
}