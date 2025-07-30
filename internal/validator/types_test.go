package validator

import (
	"testing"
)

func TestValidationResult_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		result ValidationResult
		want   bool
	}{
		{
			name:   "no_errors",
			result: ValidationResult{Valid: true, Errors: []ValidationError{}},
			want:   true,
		},
		{
			name:   "nil_errors",
			result: ValidationResult{Valid: true, Errors: nil},
			want:   true,
		},
		{
			name: "has_errors",
			result: ValidationResult{
				Valid: false,
				Errors: []ValidationError{
					{Type: ErrorTypeMissing, Path: "test"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{Valid: true}

	// Initially should be valid
	if !result.IsValid() {
		t.Error("Result should initially be valid")
	}

	// Add an error
	result.AddError(ErrorTypeMissing, "test-path", "expected", "actual", "test message")

	// Should no longer be valid
	if result.IsValid() {
		t.Error("Result should not be valid after adding error")
	}

	if result.Valid {
		t.Error("Valid field should be false after adding error")
	}

	// Check the error was added correctly
	if len(result.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(result.Errors))
	}

	err := result.Errors[0]
	if err.Type != ErrorTypeMissing {
		t.Errorf("Error type = %v, want %v", err.Type, ErrorTypeMissing)
	}
	if err.Path != "test-path" {
		t.Errorf("Error path = %v, want %v", err.Path, "test-path")
	}
	if err.Expected != "expected" {
		t.Errorf("Error expected = %v, want %v", err.Expected, "expected")
	}
	if err.Actual != "actual" {
		t.Errorf("Error actual = %v, want %v", err.Actual, "actual")
	}
	if err.Message != "test message" {
		t.Errorf("Error message = %v, want %v", err.Message, "test message")
	}

	// Add another error
	result.AddError(ErrorTypeUnexpected, "another-path", "", "unexpected", "another message")
	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}
}