# Cliguard Architecture Documentation

## Overview

Cliguard is a contract-based validation tool for Cobra CLIs. It ensures that command-line interfaces maintain their expected structure by validating them against YAML contract files.

## Core Components

### 1. Main Entry Point (`main.go`)
- Simple entry point that calls `cmd.Execute()`
- Current coverage: 0%

### 2. Command Layer (`cmd/`)
- **`root.go`**: Defines the root command and validate subcommand
  - `Execute()`: Main execution function (0% coverage)
  - `NewRootCmd()`: Creates the root cobra command (100% coverage)
  - `runValidate()`: Core validation logic (57.1% coverage)

### 3. Contract Parser (`internal/contract/`)
- **`types.go`**: Defines contract data structures (Contract, Command, Flag)
- **`parser.go`**: Loads and validates YAML contracts
  - `Load()`: Reads YAML file and performs validation (90.2% coverage)
  - Validates flag types, names, shorthands, and duplicates

### 4. Inspector (`internal/inspector/`)
- **`inspector.go`**: Generates and runs inspection code (9.2% coverage)
  - `InspectProject()`: Main inspection function (0% coverage)
  - Creates temporary Go program that imports target CLI
  - Runs inspector to extract CLI structure
  - Returns JSON representation of actual CLI
- **`types.go`**: Defines inspection result structures

### 5. Validator (`internal/validator/`)
- **`validator.go`**: Compares contract vs actual CLI (93.3% coverage)
  - `Validate()`: Main validation function
  - Checks commands, flags, types, descriptions
  - Returns detailed validation results
- **`types.go`**: Defines validation result structures

## Data Flow

1. **User Input**: User runs `cliguard validate` with flags:
   - `--project-path`: Target project location (defaults to current directory)
   - `--contract`: Contract file path (defaults to `cliguard.yaml`)
   - `--entrypoint`: Function that returns root command

2. **Contract Loading**: 
   - Loads YAML contract file
   - Validates contract structure
   - Checks for duplicate flags, invalid types, etc.

3. **Project Inspection**:
   - Creates temporary directory
   - Generates inspector Go program
   - Inspector imports target project
   - Runs inspector to extract CLI structure
   - Returns JSON representation

4. **Validation**:
   - Compares contract against actual structure
   - Checks all commands, flags, types, descriptions
   - Generates detailed error report

5. **Output**:
   - Success: "✅ Validation passed!"
   - Failure: Detailed error report with mismatches

## Key Design Decisions

1. **Dynamic Inspection**: Rather than static analysis, cliguard generates and runs code to inspect the actual CLI structure at runtime.

2. **Contract-First**: The YAML contract is the source of truth, making it easy to version control CLI interfaces.

3. **Comprehensive Validation**: Validates not just structure but also types, descriptions, and flag properties.

4. **Clear Error Reporting**: Provides detailed, actionable error messages for any mismatches.

## Current Test Coverage

- **main.go**: 0% (entry point)
- **cmd/**: 61.5% (missing Execute() and parts of runValidate())
- **contract/**: 90.2% (well tested)
- **inspector/**: 9.2% (core InspectProject function untested)
- **validator/**: 93.3% (well tested)

## Areas Needing Improvement

1. **Inspector Testing**: The InspectProject function is complex and completely untested
2. **Integration Testing**: Need end-to-end tests of the full workflow
3. **Error Path Coverage**: Many error conditions in runValidate are untested
4. **Main Function**: Entry point is untested