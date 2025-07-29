package validator

import "fmt"

// ValidationResult holds the results of validating a CLI against a contract
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// ValidationError represents a single validation failure
type ValidationError struct {
	Type     ErrorType
	Path     string
	Expected string
	Actual   string
	Message  string
}

// ErrorType defines the type of validation error
type ErrorType string

const (
	ErrorTypeMissing     ErrorType = "missing"
	ErrorTypeUnexpected  ErrorType = "unexpected"
	ErrorTypeMismatch    ErrorType = "mismatch"
	ErrorTypeInvalidType ErrorType = "invalid_type"
)

// IsValid returns true if there are no validation errors
func (vr *ValidationResult) IsValid() bool {
	return len(vr.Errors) == 0
}

// AddError adds a new validation error to the result
func (vr *ValidationResult) AddError(errorType ErrorType, path, expected, actual, message string) {
	vr.Errors = append(vr.Errors, ValidationError{
		Type:     errorType,
		Path:     path,
		Expected: expected,
		Actual:   actual,
		Message:  message,
	})
	vr.Valid = false
}

// PrintReport prints a human-readable validation report
func (vr *ValidationResult) PrintReport() {
	for _, err := range vr.Errors {
		switch err.Type {
		case ErrorTypeMissing:
			fmt.Printf("- %s: Missing %s\n", err.Path, err.Message)
			if err.Expected != "" {
				fmt.Printf("    Expected: %s\n", err.Expected)
			}
		case ErrorTypeUnexpected:
			fmt.Printf("- %s: Unexpected %s\n", err.Path, err.Message)
			if err.Actual != "" {
				fmt.Printf("    Found: %s\n", err.Actual)
			}
		case ErrorTypeMismatch:
			fmt.Printf("- %s: %s\n", err.Path, err.Message)
			fmt.Printf("    Expected: %s\n", err.Expected)
			fmt.Printf("    Actual:   %s\n", err.Actual)
		case ErrorTypeInvalidType:
			fmt.Printf("- %s: %s\n", err.Path, err.Message)
			fmt.Printf("    Expected type: %s\n", err.Expected)
			fmt.Printf("    Actual type:   %s\n", err.Actual)
		}
	}
}
