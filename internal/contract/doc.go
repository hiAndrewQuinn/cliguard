// Package contract defines the YAML contract structure for CLI validation.
// It provides types for representing expected CLI commands, flags, and their properties.
//
// The contract package is the core of cliguard's validation system. It defines
// how users specify the expected structure of their CLI applications through
// YAML configuration files.
//
// # Basic Usage
//
// Contracts are typically loaded from YAML files:
//
//	contract, err := Load("cliguard.yaml")
//	if err != nil {
//	    return err
//	}
//
// # Contract Structure
//
// A contract defines the expected structure of a CLI application:
//
//	use: myapp
//	short: My application description
//	long: |
//	  A longer description of my application
//	  that can span multiple lines.
//	flags:
//	  - name: config
//	    shorthand: c
//	    type: string
//	    usage: Config file path
//	    required: true
//	  - name: verbose
//	    shorthand: v
//	    type: bool
//	    usage: Enable verbose output
//	commands:
//	  - use: serve
//	    short: Start the server
//	    flags:
//	      - name: port
//	        shorthand: p
//	        type: int
//	        usage: Port to listen on
//	        default: "8080"
//
// # Flag Types
//
// The contract supports all standard Go flag types:
//   - bool: Boolean flags
//   - string: String values
//   - int, int8, int16, int32, int64: Integer types
//   - uint, uint8, uint16, uint32, uint64: Unsigned integer types
//   - float32, float64: Floating point types
//   - duration: Time durations (e.g., "1h30m")
//   - stringSlice: Arrays of strings
//   - intSlice: Arrays of integers
//   - boolSlice: Arrays of booleans
//
// # Validation
//
// Contracts are validated against actual CLI implementations using the
// validator package. This ensures that the CLI matches its specification.
package contract
