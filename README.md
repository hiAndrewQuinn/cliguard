# Cliguard

**Keep your CLI consistent. Catch breaking changes before they ship.**

Cliguard is a contract-based validation tool for Cobra CLIs that helps you maintain API stability by detecting unintended changes to commands, flags, and their configurations.

## Why Cliguard?

Your CLI is an API that users depend on. When commands change unexpectedly or flags disappear, it breaks user scripts and workflows. Cliguard prevents this by:

- 🔍 **Discovering** CLI entrypoints automatically in any Go project  
- 📄 **Generating** contracts from your existing CLI structure  
- ✅ **Validating** that your CLI matches its contract over time  

## Quick Start: 3 Commands to CLI Safety

The typical workflow is dead simple:

### 1. **Discover** - Find your CLI entrypoint
```bash
cliguard discover --project-path /path/to/your/cli
```

This scans your project and suggests the best entrypoint:
```
Found 2 potential CLI entrypoint(s):

1. cobra (confidence: 95%)
   Function: func NewRootCmd() *cobra.Command
   Package: github.com/myorg/mycli/cmd

Suggested entrypoint:
  --entrypoint github.com/myorg/mycli/cmd.NewRootCmd
```

### 2. **Generate** - Create your first contract
```bash
cliguard generate --entrypoint "github.com/myorg/mycli/cmd.NewRootCmd" > cliguard.yaml
```

This creates a YAML file defining your current CLI structure:
```yaml
use: mycli
short: My awesome CLI application
flags:
  - name: config
    usage: Config file path
    type: string
commands:
  - use: serve
    short: Start the web server
    flags:
      - name: port
        usage: Port to listen on
        type: int
```

### 3. **Validate** - Prevent breaking changes
```bash
cliguard validate --entrypoint "github.com/myorg/mycli/cmd.NewRootCmd"
```

Add this to your CI pipeline and catch problems early:
```
✅ Validation passed! CLI structure matches the contract.
```

## Installation

```bash
go install github.com/hiAndrewQuinn/cliguard@latest
```

Or build from source:
```bash
git clone https://github.com/hiAndrewQuinn/cliguard.git
cd cliguard && go build -o cliguard .
```

## The Discover → Generate → Validate Loop

Cliguard is designed around a simple workflow that fits naturally into development:

### When starting with an existing CLI
1. **Discover** your entrypoint (one time setup)
2. **Generate** your initial contract from current state  
3. **Validate** in CI to catch future changes

### When developing new features
1. **Validate** to check current state matches contract
2. Update your CLI code as needed
3. **Generate** a new contract if changes are intentional
4. **Validate** again to confirm everything matches

### In your CI pipeline
Just run **validate** - it will catch any unintended changes and fail the build if your CLI drifts from its contract.

## Examples

### Working with Unfamiliar Codebases

Don't know where a CLI's commands are defined? Discover them instantly:

```bash
cliguard discover --project-path ./kubernetes --interactive
```

Shows all potential entrypoints with confidence scores, letting you pick the right one interactively.

### Adding to CI/CD

**GitHub Actions:**
```yaml
- name: Validate CLI Contract  
  run: |
    go install github.com/hiAndrewQuinn/cliguard@latest
    cliguard validate --entrypoint "github.com/org/repo/cmd.NewRootCmd"
```

**Make target:**
```makefile
.PHONY: validate-cli
validate-cli:
	cliguard validate --entrypoint "github.com/org/repo/cmd.NewRootCmd"
```

### Dogfooding Example

Cliguard validates its own CLI structure. Try it:

```bash
git clone https://github.com/hiAndrewQuinn/cliguard.git
cd cliguard && go build -o cliguard .
./cliguard validate --entrypoint "github.com/hiAndrewQuinn/cliguard/cmd.NewRootCmd"
```

## Command Reference

### `cliguard discover`
Find CLI entrypoints in Go projects automatically.

```bash
cliguard discover --project-path /path/to/project
cliguard discover --project-path /path/to/project --interactive  # Pick from multiple options
```

**Supports:** Cobra, urfave/cli, standard library flag, Kingpin (discovery only for non-Cobra frameworks)

### `cliguard generate`  
Create contract files from existing CLIs.

```bash
cliguard generate --entrypoint "github.com/org/repo/cmd.NewRootCmd" > cliguard.yaml
cliguard generate --project-path /different/path --entrypoint "..." > contract.yaml
```

**Tip:** If you're in your project directory, `--project-path` defaults to current directory.

### `cliguard validate`
Validate CLI structure against contracts.

```bash
cliguard validate --entrypoint "github.com/org/repo/cmd.NewRootCmd"
cliguard validate --contract custom-contract.yaml --entrypoint "..."
```

**Returns:** Exit code 0 for success, non-zero for validation failures or errors.

## Contract File Format

Contracts are simple YAML files that mirror Cobra's structure:

```yaml
use: myapp                    # Root command name
short: Short description      # Required
long: Longer description      # Optional

flags:                        # Root-level flags
  - name: config             # Flag name
    shorthand: c             # Single character (optional)  
    usage: Config file path  # Help text
    type: string             # Flag type
    persistent: true         # Inherited by subcommands (optional)

commands:                     # Subcommands
  - use: serve
    short: Start the server
    flags:
      - name: port
        shorthand: p
        usage: Port number
        type: int
    commands:                 # Nested subcommands work too
      - use: status
        short: Check server status
```

**Supported flag types:** `string`, `bool`, `int`, `int64`, `float64`, `duration`, `stringSlice`

## CLI Framework Support

- ✅ **Cobra** - Full support (discover, generate, validate)
- ⏳ **urfave/cli** - Discovery only, generation/validation coming soon  
- ⏳ **Standard library flag** - Discovery only, generation/validation coming soon
- ⏳ **Kingpin** - Discovery only, generation/validation coming soon

Use `--force` with non-Cobra frameworks to experiment (results may be unreliable).

## Real-World Output

**Success:**
```
Loading contract from: cliguard.yaml
Inspecting CLI structure in: /path/to/project  
Validating CLI structure against contract...
✅ Validation passed! CLI structure matches the contract.
```

**Failure:**
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

## Advanced Usage

### Working with Complex Projects

For projects where auto-detection isn't perfect:

```bash
# Force operation with unsupported frameworks
cliguard generate --force --entrypoint "github.com/org/app/cmd.NewApp"

# Use custom contract locations
cliguard validate --contract ./configs/my-contract.yaml --entrypoint "..."

# Interactive selection for complex codebases  
cliguard discover --project-path ./large-project --interactive
```

### Integration Examples

See [`examples/`](examples/) directory for complete CI/CD integration examples:
- GitHub Actions workflows
- GitLab CI configurations  
- CircleCI, Jenkins, and Docker examples
- Makefile integration patterns

## Troubleshooting

**"No entrypoints found"** - Your CLI might use an unsupported framework, or the entrypoint function might have an unusual name. Try `--interactive` mode or check the [supported frameworks](#cli-framework-support).

**"Framework not supported"** - Use `--force` to experiment, but results may be unreliable. Consider contributing support for your framework!

**Validation fails unexpectedly** - Check if your CLI uses dynamic command registration or framework features not captured in contracts.

## Development

### Test Projects

Cliguard includes a collection of real-world CLI projects for testing against popular Go applications. These projects can consume significant disk space and context windows, so they're gitignored and can be regenerated as needed.

#### Regenerating Test Projects

```bash
./regenerate-test-projects.sh
```

This script clones the following projects for testing:
- **GitHub CLI** - `github.com/cli/cli`
- **Helm** - `github.com/helm/helm`  
- **Hugo** - `github.com/gohugoio/hugo`
- **CockroachDB** - `github.com/cockroachdb/cockroach`
- **etcd** - `github.com/etcd-io/etcd`
- **Kubernetes** - `github.com/kubernetes/kubernetes`
- **Demo CLI** - Simple test project with standard Cobra setup

The script will remove any existing `test-projects/` directory and recreate it from scratch. Use `./regenerate-test-projects.sh --help` for more details.

#### When to Regenerate

- When context windows become too large due to the test projects
- After removing the test-projects directory to save disk space  
- When setting up a new development environment
- When you need fresh copies of the upstream projects

## Contributing

We welcome contributions! Areas where help is especially needed:

- **New framework support** (urfave/cli, kingpin, etc.)
- **CI/CD integration examples** for other platforms
- **Documentation improvements**

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

MIT License - see [LICENSE](LICENSE) file.