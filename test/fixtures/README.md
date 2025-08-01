# Test Fixtures

This directory contains test fixtures used for integration testing of cliguard.

## Structure

```
fixtures/
└── simple-cli/         # A minimal Cobra CLI for testing
    ├── cmd/
    │   └── root.go     # Root command implementation
    ├── main.go         # Entry point
    ├── go.mod          # Go module file
    ├── go.sum          # Go dependencies
    └── cliguard.yaml   # Contract file (auto-generated)
```

## simple-cli

A minimal test CLI that demonstrates:
- Root command with persistent flags (`--config`)
- Nested command structure (`server` -> `start`)
- Command-specific flags (`server --port`)
- Print-and-exit implementation (no blocking operations)

### Commands

- `simple-cli --config <file>` - Set config file (persistent flag)
- `simple-cli server start` - Start the server (prints config and exits)

### Implementation Details

The `server start` command uses a print-and-exit pattern to avoid blocking during tests:
```go
fmt.Printf("Starting server on port %d\n", port)
if configFile != "" {
    fmt.Printf("Using config file: %s\n", configFile)
}
fmt.Println("Server configuration complete. (In a real implementation, the server would start here)")
```

This approach ensures integration tests can validate the CLI structure without dealing with long-running processes.

## Maintenance

Test fixtures are automatically maintained by:

1. **Test Helper Function**: `setupTestFixtures()` in `cmd/test_utils.go` runs `go mod tidy` before each test
2. **Makefile Targets**:
   - `make test-fixtures` - Set up/refresh fixtures
   - `make clean-fixtures` - Clean and regenerate from scratch

## Adding New Fixtures

To add a new test fixture:

1. Create a new directory under `fixtures/`
2. Add a minimal Cobra CLI implementation
3. Ensure `go mod init` and `go mod tidy` are run
4. Generate the contract file using cliguard
5. Update test helpers if needed