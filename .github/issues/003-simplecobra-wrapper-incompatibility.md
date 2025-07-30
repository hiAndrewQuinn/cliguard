# Issue: CLIs Using Cobra Wrappers (like SimpleCobra) Are Incompatible

## Summary
Some projects like Hugo use custom wrappers around Cobra (e.g., `github.com/bep/simplecobra`) instead of using Cobra directly. These wrappers have different APIs and command structures, making them incompatible with cliguard's current implementation.

## Current Behavior
```go
// Hugo's pattern using SimpleCobra
func newExec() (*simplecobra.Exec, error) {
    rootCmd := &rootCommand{
        commands: []simplecobra.Commander{
            newHugoBuildCmd(),
            newVersionCmd(),
            // ...
        },
    }
    return simplecobra.New(rootCmd)
}
```

This doesn't return `*cobra.Command`, so cliguard cannot inspect it.

## Expected Behavior
Cliguard should either:
1. Support popular Cobra wrappers directly, or
2. Provide clear documentation about incompatibility and alternatives

## Root Cause
Cliguard's inspector is tightly coupled to the `cobra.Command` type and its specific API. Wrappers like SimpleCobra use different types and interfaces that aren't compatible.

## Analysis of SimpleCobra

SimpleCobra uses:
- `simplecobra.Commander` interface instead of `*cobra.Command`
- Different command registration patterns
- Wrapped execution model

The wrapper ultimately creates Cobra commands internally, but they're not directly accessible.

## Proposed Solutions

### Solution 1: Wrapper-Specific Adapters
Create adapter packages for popular wrappers:

```go
// internal/adapters/simplecobra/adapter.go
package simplecobra

func ExtractCobraCommand(exec *simplecobra.Exec) (*cobra.Command, error) {
    // Use reflection or wrapper's API to extract underlying cobra.Command
}
```

### Solution 2: Plugin Architecture
Allow users to provide custom extractors:

```go
// Custom extractor interface
type CommandExtractor interface {
    Extract(entrypoint string) (*cobra.Command, error)
}

// User provides implementation
type SimpleCobra地Extractor struct{}

func (e *SimpleCobraExtractor) Extract(entrypoint string) (*cobra.Command, error) {
    // Custom logic to extract command from SimpleCobra
}
```

### Solution 3: Documentation and Best Practices
Document known incompatibilities and provide migration guides:

```markdown
## Known Incompatibilities

### SimpleCobra (used by Hugo)
SimpleCobra wraps Cobra commands in a custom interface. To use cliguard with SimpleCobra projects:

1. **Option A**: Create a parallel Cobra structure for contract generation
2. **Option B**: Wait for SimpleCobra adapter support (planned)
3. **Option C**: Use cliguard's manual contract creation
```

## Implementation Considerations

### Detection
Add wrapper detection to provide better error messages:

```go
func detectWrapper(projectPath string) (string, error) {
    // Check imports for known wrappers
    if hasImport("github.com/bep/simplecobra") {
        return "SimpleCobra", nil
    }
    // ... check other wrappers
}
```

### Error Messages
Improve error messages when wrapper is detected:

```
Error: Detected SimpleCobra wrapper which is not currently supported.
SimpleCobra is a custom wrapper around Cobra used by projects like Hugo.

Options:
1. Create a standard Cobra command structure for contract generation
2. Manually create a contract file based on your CLI structure
3. See https://github.com/yourusername/cliguard/issues/3 for updates

For more information: https://docs.cliguard.dev/wrappers
```

## Test Cases

### Mock Wrapper Test
Create test wrapper in `test-suite/edge-cases/wrapper/`:

```go
// mockwrapper/wrapper.go
package mockwrapper

type Command interface {
    Execute() error
}

type Exec struct {
    root Command
}

func New(cmd Command) *Exec {
    return &Exec{root: cmd}
}
```

### Detection Test
```go
func TestWrapperDetection(t *testing.T) {
    tests := []struct {
        imports []string
        expected string
    }{
        {[]string{"github.com/bep/simplecobra"}, "SimpleCobra"},
        {[]string{"github.com/spf13/cobra"}, ""},
    }
    // Test detection logic
}
```

## Documentation Updates

### Add Compatibility Matrix
```markdown
## CLI Framework Compatibility

| Framework | Status | Notes |
|-----------|--------|-------|
| Cobra | ✅ Supported | Full support |
| SimpleCobra | ❌ Not Supported | Used by Hugo, [tracking issue](#3) |
| urfave/cli | ❌ Not Supported | Different framework entirely |
| Kong | ❌ Not Supported | Different framework entirely |
```

## Affected Projects
- Hugo (uses SimpleCobra)
- Any project using custom Cobra wrappers

## Priority
Low - This affects a small number of projects, and they can work around it by creating standard Cobra commands for contract generation.

## Labels
- enhancement
- compatibility
- wrapper-support
- documentation
- simplecobra