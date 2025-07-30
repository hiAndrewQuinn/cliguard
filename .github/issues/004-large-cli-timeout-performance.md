# Issue: Large CLIs Cause Timeouts During Inspection

## Summary
When attempting to generate contracts for large, complex CLIs (like GitHub CLI), the inspection process times out. This suggests performance issues with the current inspection approach.

## Current Behavior
```bash
$ cliguard generate --project-path . --entrypoint "github.com/cli/cli/v2/pkg/cmd/root.NewCmdRoot"
# Times out after 30 seconds
Error: Command timed out after 30s
```

## Expected Behavior
Cliguard should handle large CLIs efficiently, possibly with:
- Progress indication
- Configurable timeouts
- Performance optimizations

## Root Cause Analysis

### Potential Causes
1. **Deep Command Trees**: Large CLIs may have deeply nested command structures
2. **Circular References**: Complex initialization might create circular dependencies
3. **Heavy Initialization**: Some CLIs perform expensive operations during command creation
4. **Import Resolution**: Large dependency trees take time to resolve

### Inspection Process Issues
The current inspector:
1. Builds a complete Go program
2. Compiles it with all dependencies
3. Executes it to extract structure

For large projects, compilation alone can be slow.

## Proposed Solutions

### Solution 1: Timeout Configuration
Add configurable timeout:

```go
// cmd/generate.go
cmd.Flags().Duration("timeout", 5*time.Minute, "Timeout for inspection process")
```

### Solution 2: Static Analysis Alternative
Instead of runtime inspection, use static analysis:

```go
// Alternative inspector using go/ast
func StaticInspect(projectPath, entrypoint string) (*InspectedCLI, error) {
    // Parse Go code without compilation
    // Extract command structure from AST
}
```

### Solution 3: Incremental Inspection
Break inspection into smaller parts:

```go
// Inspect commands lazily
type LazyInspector struct {
    root *cobra.Command
    inspected map[string]bool
}

func (i *LazyInspector) InspectCommand(cmd *cobra.Command) InspectedCommand {
    // Only inspect what's needed
    if i.inspected[cmd.Use] {
        return i.cache[cmd.Use]
    }
    // Inspect this command only
}
```

### Solution 4: Caching and Optimization

#### Build Cache
```go
// Cache compiled inspector binaries
type InspectorCache struct {
    dir string
}

func (c *InspectorCache) GetOrBuild(projectPath, entrypoint string) (string, error) {
    hash := calculateHash(projectPath, entrypoint)
    cachedBinary := filepath.Join(c.dir, hash)
    if exists(cachedBinary) {
        return cachedBinary, nil
    }
    // Build and cache
}
```

#### Parallel Processing
```go
// Inspect subcommands in parallel
func inspectCommandsParallel(cmds []*cobra.Command) []InspectedCommand {
    results := make([]InspectedCommand, len(cmds))
    var wg sync.WaitGroup
    
    for i, cmd := range cmds {
        wg.Add(1)
        go func(idx int, c *cobra.Command) {
            defer wg.Done()
            results[idx] = inspectCommand(c)
        }(i, cmd)
    }
    
    wg.Wait()
    return results
}
```

## Implementation Plan

### Phase 1: Quick Fixes
1. Add `--timeout` flag (default 5 minutes)
2. Add progress output during inspection
3. Better error messages on timeout

### Phase 2: Performance Analysis
1. Add profiling to identify bottlenecks
2. Benchmark different CLI sizes
3. Create performance test suite

### Phase 3: Optimization
1. Implement caching system
2. Add parallel command inspection
3. Consider static analysis alternative

## Test Cases

### Performance Benchmarks
Create `benchmark/` directory:

```go
// benchmark_test.go
func BenchmarkSmallCLI(b *testing.B) {
    // 5-10 commands
}

func BenchmarkMediumCLI(b *testing.B) {
    // 50-100 commands
}

func BenchmarkLargeCLI(b *testing.B) {
    // 500+ commands
}
```

### Timeout Test
```go
func TestInspectionTimeout(t *testing.T) {
    // Create CLI that delays during init
    // Verify timeout is respected
    // Verify error message is helpful
}
```

## Metrics to Track

1. **Inspection Time** by:
   - Number of commands
   - Depth of command tree
   - Number of flags
   - Project size

2. **Memory Usage** during inspection

3. **Compilation Time** vs **Execution Time**

## Documentation Updates

### Add Performance Guide
```markdown
## Performance Considerations

### Large CLIs
For CLIs with many commands, consider:

1. **Increase timeout**: `--timeout 10m`
2. **Use verbose mode**: `--verbose` to see progress
3. **Check compilation**: Ensure project builds quickly

### Optimization Tips
- Minimize initialization logic in command constructors
- Avoid circular dependencies
- Use lazy initialization where possible
```

## Affected Projects
Based on testing:
- GitHub CLI (gh) - Complex with many features
- Kubernetes (kubectl) - Very large command tree
- OpenShift (oc) - Built on top of kubectl

## Priority
Medium - This affects adoption for large, enterprise CLIs, but workarounds exist.

## Labels
- bug
- performance
- timeout
- large-cli
- optimization