# Issue: Network Issues Prevent Testing and Development

## Summary
During testing, network issues (DNS resolution failures, proxy access) prevented downloading Go modules, making it impossible to test cliguard with external projects. The tool should be more resilient to network issues and provide better offline support.

## Current Behavior
```bash
$ go mod download
Error: dial tcp: lookup proxy.golang.org on [fe80::...]:53: write: operation not permitted
```

This blocks:
- Testing with external projects
- Building test CLIs
- Running the inspector on projects with dependencies

## Expected Behavior
Cliguard should:
1. Work offline when possible
2. Provide clear network error messages
3. Support vendored dependencies
4. Cache dependencies effectively

## Root Cause
The inspector process:
1. Creates a temporary module
2. Runs `go get` to fetch dependencies
3. Builds and executes the inspector

This requires network access for any project with external dependencies.

## Proposed Solutions

### Solution 1: Vendor Support
Detect and use vendored dependencies:

```go
// internal/inspector/inspector.go
func (i *Inspector) prepareModule() error {
    // Check for vendor directory
    vendorPath := filepath.Join(i.config.ProjectPath, "vendor")
    if info, err := os.Stat(vendorPath); err == nil && info.IsDir() {
        // Use -mod=vendor flag
        i.buildFlags = append(i.buildFlags, "-mod=vendor")
        return nil
    }
    // Continue with go get
}
```

### Solution 2: Offline Mode
Add `--offline` flag:

```go
// Skip dependency fetching in offline mode
if i.config.Offline {
    // Assume dependencies are available
    // Use -mod=readonly
    return i.buildWithExistingDeps()
}
```

### Solution 3: Better Error Messages
Improve network error detection and messaging:

```go
func wrapNetworkError(err error) error {
    if isNetworkError(err) {
        return fmt.Errorf(`network error detected: %w

Possible solutions:
1. Check your internet connection
2. Configure proxy settings: export GOPROXY=...
3. Use vendored dependencies: go mod vendor
4. Try offline mode: --offline (if dependencies are cached)

For more help: https://docs.cliguard.dev/network-issues`, err)
    }
    return err
}
```

### Solution 4: Dependency Caching
Implement smart caching:

```go
type DependencyCache struct {
    dir string
}

func (d *DependencyCache) Get(module, version string) (string, bool) {
    // Check local cache first
}

func (d *DependencyCache) Store(module, version, path string) error {
    // Cache for future use
}
```

## Implementation Plan

### Phase 1: Vendor Support
1. Detect vendor directories
2. Use `-mod=vendor` when available
3. Document vendor usage

### Phase 2: Offline Mode
1. Add `--offline` flag
2. Skip network operations
3. Use cached dependencies

### Phase 3: Better Errors
1. Detect network errors
2. Provide actionable messages
3. Link to documentation

### Phase 4: Caching
1. Implement module cache
2. Share cache across inspections
3. Add cache management commands

## Test Cases

### Offline Test
```go
func TestOfflineMode(t *testing.T) {
    // Disable network
    os.Setenv("GOPROXY", "off")
    
    // Should work with vendored deps
    _, err := InspectProject("./testdata/vendored", "pkg.NewCmd")
    assert.NoError(t, err)
    
    // Should fail gracefully without vendor
    _, err = InspectProject("./testdata/no-vendor", "pkg.NewCmd")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "offline mode")
}
```

### Network Error Test
```go
func TestNetworkErrorMessages(t *testing.T) {
    // Simulate network failure
    os.Setenv("GOPROXY", "http://invalid.proxy")
    
    _, err := InspectProject("./testdata/external-deps", "pkg.NewCmd")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "network error detected")
    assert.Contains(t, err.Error(), "Possible solutions")
}
```

## Documentation Updates

### Add Network Troubleshooting Guide
```markdown
## Network Troubleshooting

### Common Issues

#### DNS Resolution Failures
```
Error: lookup proxy.golang.org: no such host
```

Solutions:
1. Check DNS settings
2. Try different DNS servers
3. Use vendored dependencies

#### Proxy Issues
```
Error: connectex: A connection attempt failed
```

Solutions:
1. Configure GOPROXY: `export GOPROXY=https://proxy.golang.org,direct`
2. Use corporate proxy: `export HTTPS_PROXY=http://corp-proxy:8080`
3. Use offline mode with vendored deps

### Offline Usage

#### Vendor Dependencies
```bash
# Vendor your dependencies first
go mod vendor

# Cliguard will automatically use vendored deps
cliguard generate --project-path .
```

#### Offline Mode
```bash
# Use offline mode (requires cached/vendored deps)
cliguard generate --project-path . --offline
```
```

### Add to README
```markdown
## Offline Support

Cliguard supports offline operation:

1. **Vendored Dependencies**: Automatically detected and used
2. **Offline Mode**: Use `--offline` flag
3. **Caching**: Dependencies are cached for reuse
```

## Environment Variables
Document supported environment variables:

```markdown
## Environment Variables

- `GOPROXY`: Go module proxy (default: https://proxy.golang.org)
- `GOPRIVATE`: Private modules that bypass proxy
- `GONOPROXY`: Modules that bypass proxy
- `GONOSUMDB`: Modules that bypass checksum database
- `HTTPS_PROXY`: HTTPS proxy server
- `NO_PROXY`: Domains that bypass proxy
```

## Priority
High - Network issues severely impact usability and testing, especially in corporate environments.

## Labels
- bug
- network
- offline-support
- resilience
- developer-experience