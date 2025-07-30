# Refactoring Plan for Maximum Test Coverage

## Goals
1. Achieve >95% test coverage across all packages
2. Maintain exact same functionality
3. Improve code maintainability and testability
4. Create clear separation of concerns

## Refactoring Strategy

### Phase 1: Dependency Injection & Interfaces

#### 1.1 Create Command Executor Interface
```go
// internal/executor/executor.go
type CommandExecutor interface {
    Command(name string, args ...string) ExecCommand
}

type ExecCommand interface {
    SetDir(dir string)
    Output() ([]byte, error)
    CombinedOutput() ([]byte, error)
}
```

This allows mocking of exec.Command in tests.

#### 1.2 Create File System Interface
```go
// internal/fs/filesystem.go
type FileSystem interface {
    MkdirTemp(dir, pattern string) (string, error)
    RemoveAll(path string) error
    WriteFile(name string, data []byte, perm os.FileMode) error
    ReadFile(name string) ([]byte, error)
    Stat(name string) (os.FileInfo, error)
}
```

This allows mocking file operations in tests.

### Phase 2: Refactor Inspector Module

#### 2.1 Split InspectProject into Smaller Functions
- `parseEntrypoint(entrypoint string) (EntrypointInfo, error)`
- `generateInspectorCode(info EntrypointInfo) (string, error)`
- `setupTempModule(fs FileSystem, exec CommandExecutor, tempDir string) error`
- `runInspector(fs FileSystem, exec CommandExecutor, tempDir string) ([]byte, error)`
- `parseInspectorOutput(output []byte) (*InspectedCLI, error)`

#### 2.2 Create Inspector Configuration
```go
type InspectorConfig struct {
    ProjectPath string
    Entrypoint  string
    FileSystem  FileSystem
    Executor    CommandExecutor
}
```

### Phase 3: Refactor Command Layer

#### 3.1 Extract runValidate Logic
- Create a `ValidateService` struct with dependencies
- Move validation logic out of cobra command
- Make Execute() testable by accepting io.Writer

#### 3.2 Create Configuration Loader
```go
type ConfigLoader interface {
    LoadContract(path string) (*contract.Contract, error)
    ResolveContractPath(projectPath, contractPath string) (string, error)
}
```

### Phase 4: Improve Error Handling

#### 4.1 Create Custom Error Types
```go
type ValidationError struct {
    Type    string
    Message string
    Details map[string]interface{}
}
```

#### 4.2 Remove os.Exit calls
- Return errors instead of calling os.Exit
- Let main() handle exit codes

### Phase 5: Test Infrastructure

#### 5.1 Create Test Builders
```go
// testutil/builders.go
type ContractBuilder struct {...}
type CLIBuilder struct {...}
```

#### 5.2 Create Mock Implementations
```go
// testutil/mocks.go
type MockExecutor struct {...}
type MockFileSystem struct {...}
```

#### 5.3 Fix Integration Test Fixtures
- Ensure test fixtures have proper go.mod files
- Create minimal but valid test projects

## Implementation Order

1. **Week 1: Foundation**
   - Create interfaces for executor and filesystem
   - Implement real and mock versions
   - Update existing code to use interfaces

2. **Week 2: Inspector Refactoring**
   - Break down InspectProject
   - Add comprehensive unit tests
   - Achieve >90% coverage for inspector

3. **Week 3: Command Layer**
   - Refactor runValidate
   - Make Execute() testable
   - Add success path tests

4. **Week 4: Integration & Polish**
   - Fix integration tests
   - Add end-to-end tests
   - Test main.go
   - Documentation updates

## Testing Strategy

### Unit Tests
- Mock all external dependencies
- Test each function in isolation
- Cover all error paths

### Integration Tests
- Use real filesystem for key scenarios
- Test with actual Go projects
- Verify end-to-end functionality

### Regression Tests
- Create test suite from current behavior
- Ensure refactoring doesn't break anything
- Test with cliguard validating itself

## Success Metrics

1. **Coverage**: >95% across all packages
2. **Performance**: No regression in execution time
3. **Functionality**: All existing features work identically
4. **Maintainability**: Clear separation of concerns
5. **Testability**: Easy to add new tests

## Risk Mitigation

1. **Incremental Changes**: Small, testable commits
2. **Continuous Testing**: Run tests after each change
3. **Behavior Preservation**: Extensive regression tests
4. **Rollback Plan**: Git tags at each phase completion