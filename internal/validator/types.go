package validator

import "fmt"

// ValidationResult holds the results of validating a CLI against a contract
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// ValidationError represents a single validation failure
type ValidationError struct {
	Type        ErrorType
	Path        string
	Expected    string
	Actual      string
	Message     string
	Description string // Additional descriptive text for the error
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
		Type:        errorType,
		Path:        path,
		Expected:    expected,
		Actual:      actual,
		Message:     message,
		Description: message, // For backward compatibility, use message as description
	})
	vr.Valid = false
}

// AddErrorWithDescription adds a new validation error with a separate description
func (vr *ValidationResult) AddErrorWithDescription(errorType ErrorType, path, expected, actual, message, description string) {
	vr.Errors = append(vr.Errors, ValidationError{
		Type:        errorType,
		Path:        path,
		Expected:    expected,
		Actual:      actual,
		Message:     message,
		Description: description,
	})
	vr.Valid = false
}

// PrintReport prints a human-readable validation report
func (vr *ValidationResult) PrintReport() {
	// Group errors by type for better organization
	var missingErrors, unexpectedErrors, mismatchErrors, invalidTypeErrors []ValidationError

	for _, err := range vr.Errors {
		switch err.Type {
		case ErrorTypeMissing:
			missingErrors = append(missingErrors, err)
		case ErrorTypeUnexpected:
			unexpectedErrors = append(unexpectedErrors, err)
		case ErrorTypeMismatch:
			mismatchErrors = append(mismatchErrors, err)
		case ErrorTypeInvalidType:
			invalidTypeErrors = append(invalidTypeErrors, err)
		}
	}

	// Print missing errors
	if len(missingErrors) > 0 {
		fmt.Println("\n❌ Missing items:")
		for _, err := range missingErrors {
			fmt.Printf("   • %s: %s\n", err.Description, err.Path)
			if err.Expected != "" {
				fmt.Printf("     Add to contract: %s\n", err.Expected)
			}
		}
	}

	// Print unexpected errors
	if len(unexpectedErrors) > 0 {
		fmt.Println("\n❌ Unexpected items:")
		for _, err := range unexpectedErrors {
			fmt.Printf("   • %s\n", err.Path)
			fmt.Printf("     %s\n", err.Message)
			if err.Actual != "" {
				fmt.Printf("     Found: %s\n", err.Actual)
			}
		}
	}

	// Print mismatch errors
	if len(mismatchErrors) > 0 {
		fmt.Println("\n❌ Mismatches:")
		for _, err := range mismatchErrors {
			fmt.Printf("   • %s\n", err.Path)
			fmt.Printf("     Contract: %s\n", err.Expected)
			fmt.Printf("     Actual:   %s\n", err.Actual)
			if err.Description != "" {
				fmt.Printf("     %s\n", err.Description)
			}
		}
	}

	// Print invalid type errors
	if len(invalidTypeErrors) > 0 {
		fmt.Println("\n❌ Invalid types:")
		for _, err := range invalidTypeErrors {
			fmt.Printf("   • %s\n", err.Path)
			fmt.Printf("     %s\n", err.Message)
			fmt.Printf("     Expected type: %s\n", err.Expected)
			fmt.Printf("     Actual type:   %s\n", err.Actual)
		}
	}

	// Print summary
	fmt.Printf("\nTotal errors: %d\n", len(vr.Errors))
}
