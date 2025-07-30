# Issue: Non-Standard Build Systems Cause Compatibility Issues

## Summary
Projects using custom build systems (like CockroachDB's Bazel setup) are difficult or impossible to use with cliguard. The tool assumes standard Go toolchain, but many large projects use Bazel, Make, or other build systems with specific requirements.

## Current Behavior
```bash
# CockroachDB requires specific Bazel version
$ ./dev build
ERROR: You're not using Bazelisk!
/path/to/bazel-cockroachdb/7.6.0: No such file or directory
```

This prevents:
- Building the project
- Running cliguard inspection
- Generating contracts

## Expected Behavior
Cliguard should:
1. Support common build systems (Bazel, Make)
2. Allow custom build commands
3. Work with pre-built binaries
4. Provide clear guidance for complex build setups

## Root Cause
Current assumptions:
1. Standard `go build` works
2. Direct `go get` for dependencies
3. No special build requirements

Reality for large projects:
1. Custom toolchains (Bazel, specific Go versions)
2. Generated code that must be built first
3. Complex dependency management
4. Build-time configuration

## Proposed Solutions

### Solution 1: Custom Build Commands
Allow users to specify build commands:

```bash
cliguard generate \
  --project-path . \
  --entrypoint "pkg.NewCmd" \
  --build-cmd "bazel build //cmd/cli:cli"
```

Implementation:
```go
type Config struct {
    ProjectPath string
    Entrypoint  string
    BuildCmd    []string  // Custom build command
    BuildEnv    []string  // Environment variables
}

func (i *Inspector) build() error {
    if len(i.config.BuildCmd) > 0 {
        return i.runCustomBuild()
    }
    return i.standardGoBuild()
}
```

### Solution 2: Binary Mode
Support pre-built binaries:

```bash
# Build first
bazel build //cmd/cli:cli

# Use built binary
cliguard generate \
  --binary bazel-bin/cmd/cli/cli \
  --extract-only
```

### Solution 3: Build System Adapters
Create adapters for common build systems:

```go
// internal/builders/bazel.go
type BazelBuilder struct {
    workspace string
}

func (b *BazelBuilder) DetectBuildFile() bool {
    // Look for WORKSPACE or BUILD.bazel
}

func (b *BazelBuilder) BuildTarget(target string) error {
    // Run bazel build
}
```

### Solution 4: Build Configuration File
Support `.cliguard.build.yaml`:

```yaml
# .cliguard.build.yaml
build:
  system: bazel
  target: //cmd/cli:cli
  env:
    - USE_BAZELISK=1
  pre_build:
    - ./dev generate

inspection:
  timeout: 10m
  entrypoint: "github.com/org/project/pkg/cli.NewRootCmd"
```

## Implementation Plan

### Phase 1: Custom Build Support
1. Add `--build-cmd` flag
2. Support environment variables
3. Document custom build usage

### Phase 2: Binary Extraction
1. Add `--binary` mode
2. Extract structure without building
3. Support various binary formats

### Phase 3: Build System Detection
1. Auto-detect common build systems
2. Provide appropriate defaults
3. Clear messages about detected system

### Phase 4: Build Adapters
1. Implement Bazel adapter
2. Implement Make adapter
3. Plugin system for custom adapters

## Test Cases

### Bazel Project Test
```go
func TestBazelProject(t *testing.T) {
    // Mock project with WORKSPACE file
    projectDir := setupBazelProject(t)
    
    config := Config{
        ProjectPath: projectDir,
        BuildCmd:    []string{"bazel", "build", "//cmd:cli"},
        Entrypoint:  "pkg.NewCmd",
    }
    
    _, err := InspectProject(config)
    assert.NoError(t, err)
}
```

### Custom Build Test
```go
func TestCustomBuildCommand(t *testing.T) {
    config := Config{
        BuildCmd: []string{"make", "build-cli"},
        BuildEnv: []string{"CGO_ENABLED=0"},
    }
    
    // Should use custom command instead of go build
}
```

## Documentation Updates

### Add Build Systems Guide
```markdown
## Build System Support

### Standard Go Projects
No configuration needed - cliguard uses standard Go toolchain.

### Bazel Projects
```bash
# Option 1: Custom build command
cliguard generate \
  --project-path . \
  --entrypoint "pkg.NewCmd" \
  --build-cmd "bazel build //cmd/cli:cli"

# Option 2: Pre-built binary
bazel build //cmd/cli:cli
cliguard generate --binary bazel-bin/cmd/cli/cli_/cli
```

### Make Projects
```bash
# Use make target
cliguard generate \
  --project-path . \
  --entrypoint "cmd.NewRootCmd" \
  --build-cmd "make cli"
```

### Complex Build Requirements
Create `.cliguard.build.yaml`:
```yaml
build:
  system: custom
  commands:
    - make generate
    - make deps
    - make build
  env:
    - CGO_ENABLED=0
    - GOOS=linux
```
```

### Add Compatibility Matrix
```markdown
## Build System Compatibility

| Build System | Support | Notes |
|--------------|---------|-------|
| Go Modules | ✅ Native | Default |
| Bazel | ⚠️ Custom Command | Use --build-cmd |
| Make | ⚠️ Custom Command | Use --build-cmd |
| Gradle | ❌ Not Supported | Java build system |
| Buck | ❌ Not Supported | - |
```

## Examples for Common Projects

### CockroachDB
```bash
# Requires Bazelisk
cliguard generate \
  --project-path . \
  --entrypoint "github.com/cockroachdb/cockroach/pkg/cli.CockroachCmd" \
  --build-cmd "./dev build short" \
  --timeout 20m
```

### Kubernetes
```bash
# Uses Make
cliguard generate \
  --project-path . \
  --entrypoint "k8s.io/kubectl/pkg/cmd.NewKubectlCommand" \
  --build-cmd "make kubectl"
```

## Priority
Medium - Affects large enterprise projects, but workarounds exist. Important for wider adoption.

## Labels
- enhancement
- compatibility
- build-systems
- bazel
- enterprise