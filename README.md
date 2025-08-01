# Cliguard

A contract-based validation tool for Cobra CLIs that ensures your command structure remains consistent over time.

## Overview

Cliguard validates Go CLIs built with [Cobra](https://github.com/spf13/cobra) against a YAML contract file. It helps maintain API stability by detecting unintended changes to commands, flags, and their configurations. It can also generate contract files from existing CLIs, making it easy to get started.

## Features

- **Entrypoint discovery**: Automatically find CLI entrypoints in Go projects across multiple frameworks
- **Contract generation**: Automatically generate contract files from existing Cobra CLIs
- **Contract-based validation**: Define your expected CLI structure in a simple YAML file
- **Comprehensive checking**: Validates commands, subcommands, flags, types, and descriptions
- **CI/CD friendly**: Exit codes and clear output make it perfect for automated pipelines
- **Dogfooding**: Cliguard validates its own CLI structure
- **Convenient defaults**: No need to specify `--project-path` when running from your project directory

## Installation

```bash
go install github.com/hiAndrewQuinn/cliguard@latest
```

Or build from source:

```bash
git clone https://github.com/hiAndrewQuinn/cliguard.git
cd cliguard
go build -o cliguard .
```

## Quick Start

### Example: Cliguard Validating Itself

Cliguard uses itself to ensure its own CLI structure remains consistent. Here's how:

1. Clone and build Cliguard:

```bash
git clone https://github.com/hiAndrewQuinn/cliguard.git
cd cliguard
go build -o cliguard .
```

2. Look at Cliguard's own contract file:

```bash
cat cliguard.yaml
```

Output:
```yaml
# Cliguard contract for the cliguard CLI itself (dogfooding!)
use: cliguard
short: A contract-based validation tool for Cobra CLIs

commands:
  - use: validate
    short: Validate a Cobra CLI against a contract file
    flags:
      - name: project-path
        usage: Path to the root of the target Go project (defaults to current directory)
        type: string
      - name: contract
        usage: Path to the contract file (defaults to cliguard.yaml in project path)
        type: string
      - name: entrypoint
        usage: The function that returns the root command (e.g., github.com/user/repo/cmd.NewRootCmd)
        type: string
```

3. Run Cliguard on itself:

```bash
# From the cliguard directory, project-path defaults to current directory
./cliguard validate --entrypoint "github.com/hiAndrewQuinn/cliguard/cmd.NewRootCmd"
```

Output:
```
Loading contract from: /home/andrew/Code/cliguard/cliguard.yaml
Inspecting CLI structure in: /home/andrew/Code/cliguard
Validating CLI structure against contract...
✅ Validation passed! CLI structure matches the contract.
```

### Using Cliguard in Your Project

1. Generate a contract file from your existing CLI:

```bash
# From your project directory
cliguard generate --entrypoint "github.com/myorg/myapp/cmd.NewRootCmd"
```

This creates a `cliguard.yaml` file like:

```yaml
use: myapp
short: My awesome CLI application

flags:
  - name: config
    shorthand: c
    usage: Config file path
    type: string
    persistent: true

commands:
  - use: serve
    short: Start the web server
    flags:
      - name: port
        shorthand: p
        usage: Port to listen on
        type: int
  - use: migrate
    short: Run database migrations
```

2. Run validation to ensure your CLI structure remains consistent:

```bash
# From your project directory
cliguard validate --entrypoint "github.com/myorg/myapp/cmd.NewRootCmd"
```

## Contract Schema

The contract file (`cliguard.yaml`) mirrors Cobra's command structure:

### Root Level

```yaml
use: string        # Command name (required)
short: string      # Short description (required)
long: string       # Long description (optional)
flags: []Flag      # Root command flags (optional)
commands: []Command # Subcommands (optional)
```

### Command Structure

```yaml
commands:
  - use: string      # Command name (required)
    short: string    # Short description (required)
    long: string     # Long description (optional)
    flags: []Flag    # Command-specific flags (optional)
    commands: []Command # Nested subcommands (optional)
```

### Flag Structure

```yaml
flags:
  - name: string       # Flag name (required)
    shorthand: string  # Single character shorthand (optional)
    usage: string      # Help text (required)
    type: string       # Flag type (required)
    persistent: bool   # Is persistent flag (optional, default: false)
```

Supported flag types:
- `string`
- `bool`
- `int`
- `int64`
- `float64`
- `duration`
- `stringSlice`

## Usage

### Discover Entrypoints

Find CLI entrypoints in a Go project by analyzing common CLI framework patterns. Supports:
- Cobra (`github.com/spf13/cobra`)
- urfave/cli (`github.com/urfave/cli`)
- Standard library flag package
- Kingpin (`github.com/alecthomas/kingpin`)

```bash
cliguard discover --project-path /path/to/project
```

Use interactive mode to select from multiple candidates:

```bash
cliguard discover --project-path /path/to/project --interactive
# or
cliguard discover --project-path /path/to/project -i
```

Example output:
```
Searching for CLI entrypoints in: .

Found 3 potential CLI entrypoint(s):

1. cobra (confidence: 95%)
   File: cmd/root.go:15
   Pattern: Function returning root cobra.Command
   Code: func NewRootCmd() *cobra.Command {
   Function: func NewRootCmd() *cobra.Command
   Package: github.com/myorg/myproject/cmd

2. cobra (confidence: 85%)
   File: cmd/root.go:20
   Pattern: Cobra Execute function
   Code: func Execute() {
   Function: func Execute()
   Package: github.com/myorg/myproject/cmd

Suggested entrypoint:
  --entrypoint github.com/myorg/myproject/cmd.NewRootCmd
```

### Generate a Contract

Generate a contract file from an existing Cobra CLI:

```bash
# From your project directory (no --project-path needed)
cliguard generate

# Or specify a different project
cliguard generate --project-path /path/to/project
```

The generated contract is printed to stdout, so pipe it to a file:

```bash
# From your project directory
cliguard generate > cliguard.yaml

# Or with a specific project path
cliguard generate --project-path /path/to/project > my-contract.yaml
```

For projects where the root command is returned by a function:

```bash
# From your project directory
cliguard generate --entrypoint "github.com/org/project/cmd.NewRootCmd" > cliguard.yaml
```

### Basic Validation

```bash
# From your project directory (no --project-path needed)
cliguard validate

# Or validate a different project
cliguard validate --project-path /path/to/project
```

### Custom Contract Location

```bash
cliguard validate --project-path /path/to/project --contract /path/to/contract.yaml
```

### Specifying Entrypoint

For projects where the root command is returned by a function:

```bash
# From your project directory
cliguard validate --entrypoint "github.com/org/project/cmd.NewRootCmd"
```

## How It Works

1. **Contract Loading**: Cliguard reads and validates your YAML contract
2. **Inspector Generation**: Creates a temporary Go program that imports your CLI
3. **Structure Extraction**: Runs the inspector to extract the actual CLI structure
4. **Validation**: Compares the actual structure against the contract
5. **Reporting**: Provides clear feedback on any discrepancies

## CI/CD Integration

Cliguard provides comprehensive CI/CD integration examples and workflows for various platforms.

### Quick Start

Add Cliguard to your CI pipeline to catch breaking changes:

```yaml
# GitHub Actions example
- name: Validate CLI Contract
  run: |
    go install github.com/hiAndrewQuinn/cliguard@latest
    cliguard validate --entrypoint "github.com/org/repo/cmd.NewRootCmd"
```

### Full Examples

- **GitHub Actions**: See [`.github/workflows/`](.github/workflows/) for validation and auto-generation workflows
- **Platform Examples**: Browse [`examples/`](examples/) for GitLab CI, CircleCI, Jenkins, and more
- **Documentation**: Read the comprehensive [CI/CD Integration Guide](docs/ci-cd-integration.md)

### Supported Platforms

- ✅ GitHub Actions (full workflow examples)
- ✅ GitLab CI
- ✅ CircleCI  
- ✅ Jenkins Pipeline
- ✅ Docker-based validation
- ✅ Makefile integration

## Example Output

Success:
```
✅ Validation passed! CLI structure matches the contract.
```

Failure:
```
❌ Validation failed!

- root: Mismatch in short description
    Expected: A simple test CLI
    Actual:   A test CLI application
- --verbose: Missing flag
    Expected: verbose
- server --port: Flag type mismatch
    Expected type: string
    Actual type:   int
```

## Dogfooding

Cliguard validates its own CLI structure. See our [`cliguard.yaml`](./cliguard.yaml) contract file.

## Development

### Running Tests

Cliguard includes a comprehensive test suite. Use the Makefile for common operations:

```bash
# Run all tests
make test

# Run only integration tests
make test-integration

# Set up test fixtures
make test-fixtures

# Clean and regenerate test fixtures
make clean-fixtures
```

### Test Fixtures

The project includes test fixtures in `test/fixtures/` for integration testing. These fixtures are automatically maintained by the test helper functions.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT