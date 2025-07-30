package contract

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a contract file
func Load(contractPath string) (*Contract, error) {
	if contractPath == "" {
		return nil, fmt.Errorf("contract path cannot be empty")
	}

	absPath, err := filepath.Abs(contractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve contract path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract file: %w", err)
	}

	var contract Contract
	if err := yaml.Unmarshal(data, &contract); err != nil {
		return nil, fmt.Errorf("failed to parse contract YAML: %w", err)
	}

	if err := validate(&contract); err != nil {
		return nil, fmt.Errorf("contract validation failed: %w", err)
	}

	return &contract, nil
}

// validate performs basic validation on the contract
func validate(contract *Contract) error {
	if contract.Use == "" {
		return fmt.Errorf("root command 'use' field cannot be empty")
	}

	// Validate all flags
	if err := validateFlags(contract.Flags); err != nil {
		return fmt.Errorf("root command flags: %w", err)
	}

	// Validate all subcommands recursively
	for _, cmd := range contract.Commands {
		if err := validateCommand(&cmd, contract.Use); err != nil {
			return err
		}
	}

	return nil
}

func validateCommand(cmd *Command, parentPath string) error {
	if cmd.Use == "" {
		return fmt.Errorf("command under '%s': 'use' field cannot be empty", parentPath)
	}

	currentPath := parentPath + " " + cmd.Use

	if err := validateFlags(cmd.Flags); err != nil {
		return fmt.Errorf("command '%s' flags: %w", currentPath, err)
	}

	for _, subcmd := range cmd.Commands {
		if err := validateCommand(&subcmd, currentPath); err != nil {
			return err
		}
	}

	return nil
}

func validateFlags(flags []Flag) error {
	seenNames := make(map[string]bool)
	seenShorthands := make(map[string]bool)

	for _, flag := range flags {
		if flag.Name == "" {
			return fmt.Errorf("flag name cannot be empty")
		}

		if seenNames[flag.Name] {
			return fmt.Errorf("duplicate flag name: %s", flag.Name)
		}
		seenNames[flag.Name] = true

		if flag.Shorthand != "" {
			if len(flag.Shorthand) != 1 {
				return fmt.Errorf("flag shorthand must be a single character: %s", flag.Shorthand)
			}
			if seenShorthands[flag.Shorthand] {
				return fmt.Errorf("duplicate flag shorthand: %s", flag.Shorthand)
			}
			seenShorthands[flag.Shorthand] = true
		}

		if flag.Type == "" {
			return fmt.Errorf("flag '%s': type cannot be empty", flag.Name)
		}

		// Validate flag type
		validTypes := map[string]bool{
			// Basic types (existing)
			"string": true, "bool": true, "int": true, "int64": true,
			"float64": true, "duration": true, "stringSlice": true,

			// Integer variants
			"int8": true, "int16": true, "int32": true,
			"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,

			// Float variants
			"float32": true,

			// Slice types
			"intSlice": true, "int32Slice": true, "int64Slice": true,
			"uintSlice": true, "float32Slice": true, "float64Slice": true,
			"boolSlice": true, "durationSlice": true,

			// Map types
			"stringToString": true, "stringToInt64": true,

			// Network types
			"ip": true, "ipSlice": true, "ipMask": true, "ipNet": true,

			// Binary types
			"bytesHex": true, "bytesBase64": true,

			// Special types
			"count": true,
		}
		if !validTypes[flag.Type] {
			return fmt.Errorf("flag '%s': invalid type '%s'", flag.Name, flag.Type)
		}
	}

	return nil
}
