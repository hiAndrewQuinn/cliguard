# Issue: CLIs That Build Commands in init() Functions Are Incompatible

## Summary
Cliguard requires a function that returns `*cobra.Command`, but many CLIs (like etcdctl) build their command structure using global variables populated in `init()` functions. This pattern makes them incompatible with cliguard's current design.

## Current Behavior
```go
// etcdctl pattern - incompatible with cliguard
var rootCmd = &cobra.Command{
    Use: "etcdctl",
}

func init() {
    rootCmd.AddCommand(newGetCommand())
    rootCmd.AddCommand(newPutCommand())
    // ... more initialization
}

func Start() error {
    return rootCmd.Execute()
}
```

Cliguard cannot use this because there's no function that returns the command.

## Expected Behavior
Cliguard should support CLIs that use the init() pattern, which is common in the Go ecosystem.

## Root Cause
The inspector template in `/internal/inspector/inspector.go` requires an entrypoint function that returns `*cobra.Command`. It cannot access global variables that are populated during init().

## Proposed Solutions

### Solution 1: Global Variable Detection
Enhance the inspector to detect and use global `*cobra.Command` variables:

```go
// In inspector template
func findRootCommand() *cobra.Command {
    // Use reflection to find exported variables of type *cobra.Command
    pkg := reflect.ValueOf({{ .PackageAlias }})
    // Look for common names: rootCmd, RootCmd, root, Root
    for _, name := range []string{"rootCmd", "RootCmd", "root", "Root"} {
        if field := pkg.FieldByName(name); field.IsValid() {
            if cmd, ok := field.Interface().(*cobra.Command); ok {
                return cmd
            }
        }
    }
    return nil
}
```

### Solution 2: Execution-Based Inspection
Instead of requiring a function, allow specifying a CLI binary and intercept its command structure:

```go
// New inspector mode
type Config struct {
    ProjectPath string
    Entrypoint  string
    BinaryPath  string  // Alternative to Entrypoint
    GlobalVar   string  // Name of global command variable
}
```

### Solution 3: Wrapper Generation
Provide a tool to generate wrapper functions for init-based CLIs:

```bash
cliguard wrap --package go.etcd.io/etcd/etcdctl/v3/ctlv3 --var rootCmd --output wrapper.go
```

Generates:
```go
package wrapper

import "go.etcd.io/etcd/etcdctl/v3/ctlv3"

func GetRootCmd() *cobra.Command {
    // Force init() to run
    _ = ctlv3.Start()
    return ctlv3.rootCmd  // Assuming it's accessible
}
```

## Implementation Plan

### Phase 1: Document Workaround
Add to documentation how users can create wrapper functions:

```go
// wrapper/cmd.go
package wrapper

import (
    "github.com/spf13/cobra"
    "original/package"
)

var rootCmdRef *cobra.Command

func init() {
    // Capture the command during init
    rootCmdRef = package.GetInternalRootCmd()
}

func NewRootCmd() *cobra.Command {
    return rootCmdRef
}
```

### Phase 2: Global Variable Support
1. Add `--global-var` flag to generate/validate commands
2. Update inspector template to support global variable access
3. Handle init() execution timing

### Phase 3: Binary Inspection Mode
1. Build the CLI binary
2. Use build tags or reflection to extract command structure
3. Support `--binary` flag as alternative to `--entrypoint`

## Test Cases

### Test Init-Based CLI
Create test case in `test-suite/edge-cases/init-based/`:

```go
// cmd/root.go
package cmd

var RootCmd = &cobra.Command{
    Use: "initcli",
}

func init() {
    RootCmd.AddCommand(&cobra.Command{
        Use:   "subcommand",
        Short: "A subcommand",
    })
}

func Execute() error {
    return RootCmd.Execute()
}
```

### Test Global Variable Detection
```go
func TestGlobalVarDetection(t *testing.T) {
    // Test that inspector can find and use global command variables
}
```

## Affected Projects
Based on testing:
- etcd (etcdctl)
- Many older Go CLIs that predate the "New" pattern
- CLIs that use init() for complex initialization

## Documentation Updates

### Add to README.md:
```markdown
## Compatibility

Cliguard requires a function that returns `*cobra.Command`. If your CLI uses global variables populated in init(), you have options:

1. **Create a wrapper function**:
```go
func NewRootCmd() *cobra.Command {
    return rootCmd  // your global variable
}
```

2. **Use upcoming global variable support** (planned):
```bash
cliguard generate --project-path . --global-var rootCmd
```
```

## Priority
Medium - This is a common pattern, but there are workarounds. Long-term support would improve adoption.

## Labels
- enhancement
- compatibility
- init-pattern
- documentation