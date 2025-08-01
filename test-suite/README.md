# Cliguard Test Suite

This directory contains a comprehensive test suite for cliguard, designed to test various Cobra CLI patterns and edge cases.

## Structure

```
test-suite/
├── basic/              # Basic functionality tests
│   ├── simple-cli/     # Minimal CLI with root command only
│   ├── subcommands/    # CLI with basic subcommands
│   └── flags/          # CLI with various flag configurations
├── edge-cases/         # Edge case tests
│   ├── nested/         # Deeply nested command structures
│   ├── many-flags/     # Commands with many flags
│   ├── flag-types/     # All supported flag types
│   ├── dynamic/        # Dynamically added commands
│   └── unicode/        # Unicode in names/descriptions
├── validation/         # Contract validation tests
│   ├── breaking/       # Tests for breaking changes
│   ├── additions/      # Tests for additions
│   └── compatible/     # Tests for compatible changes
└── performance/        # Performance testing
    ├── small/          # 5-10 commands
    ├── medium/         # 50-100 commands
    └── large/          # 500+ commands
```

## Test Cases

### Basic Tests

1. **simple-cli**: Minimal CLI with just a root command
2. **subcommands**: CLI with 2-3 levels of subcommands
3. **flags**: CLI demonstrating various flag patterns

### Edge Cases

1. **nested**: Tests deeply nested command structures (5+ levels)
2. **many-flags**: Commands with 20+ flags each
3. **flag-types**: Tests all Cobra flag types (string, int, bool, duration, etc.)
4. **dynamic**: Commands added conditionally or dynamically
5. **unicode**: Unicode characters in command names and descriptions

### Validation Tests

1. **breaking**: Pairs of CLIs where v2 introduces breaking changes
2. **additions**: Pairs of CLIs where v2 adds new features
3. **compatible**: Pairs of CLIs with compatible changes

### Performance Tests

1. **small**: Small CLI (5-10 commands)
2. **medium**: Medium CLI (50-100 commands)
3. **large**: Large CLI (500+ commands)

## Running Tests

Use the test runner script:

```bash
./run-tests.sh
```

Or run individual tests:

```bash
cd basic/simple-cli
go build
# project-path defaults to current directory
cliguard generate --entrypoint "github.com/cliguard/test/simple.NewRootCmd" > contract.yaml
cliguard validate --entrypoint "github.com/cliguard/test/simple.NewRootCmd" --contract contract.yaml
```