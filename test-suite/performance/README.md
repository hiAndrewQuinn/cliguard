# Cliguard Performance Benchmarks

This directory contains comprehensive performance benchmarks for testing cliguard's performance with CLIs of various sizes, from small (10 commands) to extra-large (500+ commands).

## Overview

Performance benchmarks help ensure cliguard maintains good performance as CLI complexity grows and help detect performance regressions during development.

### Benchmark Categories

- **Small CLIs**: 5-10 commands, minimal nesting
- **Medium CLIs**: 50-100 commands, some nesting  
- **Large CLIs**: 500+ commands, complex nesting

### Components Benchmarked

1. **Validator** (`internal/validator/validator_bench_test.go`)
   - Tests validation performance across different CLI sizes
   - Measures memory allocation patterns
   - Tests deeply nested command structures

2. **Inspector** (`internal/inspector/inspector_bench_test.go`) 
   - Tests CLI structure conversion performance
   - Benchmarks flag extraction operations
   - Measures conversion overhead for different CLI sizes

3. **Contract Parser** (`internal/contract/parser_bench_test.go`)
   - Tests YAML contract loading performance
   - Measures serialization/deserialization overhead
   - Tests complex contract structures

## Running Benchmarks

### Quick Benchmark Run

Run all benchmarks with default settings:

```bash
./run-benchmarks.sh
```

### Customized Benchmark Run

Customize benchmark parameters:

```bash
# Run with longer benchmark time
BENCHMARK_TIME=10s ./run-benchmarks.sh

# Run with more iterations for stable results  
BENCHMARK_COUNT=5 ./run-benchmarks.sh

# Specify custom output directory
OUTPUT_DIR=/tmp/my-benchmarks ./run-benchmarks.sh
```

### Individual Package Benchmarks

Run benchmarks for specific packages:

```bash
# Validator benchmarks only
go test -bench=. -benchmem ./internal/validator

# Inspector benchmarks only  
go test -bench=. -benchmem ./internal/inspector

# Contract parser benchmarks only
go test -bench=. -benchmem ./internal/contract
```

### Specific Benchmark Tests

Run individual benchmark functions:

```bash
# Test small CLI validation performance
go test -bench=BenchmarkValidateSmallCLI -benchtime=5s ./internal/validator

# Test large CLI conversion performance  
go test -bench=BenchmarkConvertLargeCLI -benchmem ./internal/inspector
```

## Real-World CLI Testing

### Automated Tests

Test against simulated real-world CLIs:

```bash
# Test all sizes (small, medium, large)
./test-real-world.sh

# Test specific CLI types
./test-with-helm.sh      # Medium-sized CLI (~75 commands)
./test-with-kubectl.sh   # Large CLI (500+ commands)
```

### Manual Testing

Create and test custom CLI projects:

```bash
# Create test CLI projects of different sizes
mkdir -p small medium large
cd small && go run ../utils/generate_cli.go small .
cd ../medium && go run ../utils/generate_cli.go medium .  
cd ../large && go run ../utils/generate_cli.go large .

# Test cliguard commands on generated projects
cd small && cliguard inspect --target . --entry "cmd.NewRootCmd"
cd ../medium && cliguard validate --contract contract.yaml --target .
```

## Performance Metrics

### Expected Performance Ranges

Based on benchmark results, typical performance ranges are:

**Validator Performance:**
- Small CLI (10 commands): ~19μs, ~7KB memory, ~358 allocs
- Medium CLI (50 commands): ~220μs, ~71KB memory, ~3,918 allocs  
- Large CLI (100 commands): ~920μs, ~332KB memory, ~14,512 allocs
- Extra Large CLI (500 commands): ~8.5ms, ~3MB memory, ~104,858 allocs

**Inspector Performance (Conversion):**
- Small CLI: ~4μs, ~10KB memory, ~51 allocs
- Medium CLI: ~23μs, ~49KB memory, ~225 allocs
- Large CLI: ~43μs, ~99KB memory, ~444 allocs  
- Extra Large CLI: ~360μs, ~461KB memory, ~2,178 allocs

**Contract Parser Performance:**  
- Small Contract: ~870μs, ~511KB memory, ~6,386 allocs
- Medium Contract: ~5ms, ~2.7MB memory, ~34,732 allocs
- Large Contract: ~10ms, ~5.5MB memory, ~70,163 allocs
- Extra Large Contract: ~53ms, ~28MB memory, ~354,006 allocs

### Performance Analysis

**Key Observations:**
1. **Linear scaling**: Performance scales roughly linearly with CLI size
2. **Memory efficiency**: Memory usage is proportional to CLI complexity
3. **Parsing overhead**: Contract parsing is the most expensive operation
4. **Validation cost**: Validation overhead is reasonable even for large CLIs

**Performance Bottlenecks:**
- Contract YAML parsing dominates execution time for large contracts
- Memory allocation increases significantly with CLI complexity
- Deep nesting adds validation complexity

## Profiling and Optimization

### CPU Profiling

The benchmark script generates CPU profiles for analysis:

```bash
# Run benchmarks (profiles saved automatically)
./run-benchmarks.sh

# Analyze CPU profiles
go tool pprof benchmark-results/profiles_*/validator_cpu.prof
go tool pprof benchmark-results/profiles_*/inspector_cpu.prof  
go tool pprof benchmark-results/profiles_*/contract_cpu.prof
```

**Useful pprof commands:**
- `top10` - Show top 10 functions by CPU usage
- `list <function>` - Show source code for function
- `web` - Generate SVG visualization (requires graphviz)
- `png` - Generate PNG visualization

### Memory Profiling

Analyze memory allocation patterns:

```bash
# Memory profiles are also generated automatically
go tool pprof benchmark-results/profiles_*/validator_mem.prof

# In pprof shell:
(pprof) top10
(pprof) list main.function_name
(pprof) web
```

### Performance Monitoring

**Regression Detection:**
- Use `benchstat` to compare benchmark runs over time
- Set up CI to run benchmarks and detect significant changes
- Monitor key metrics: ns/op, B/op, allocs/op

**Installing benchstat:**
```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

**Comparing benchmark runs:**
```bash
benchstat old_results.txt new_results.txt
```

## Known Performance Issues

### Large CLI Timeout Issue

**Problem**: CLIs with 500+ commands may hit timeout limits during runtime inspection.

**Symptoms:**
- Inspection takes >30 seconds
- Process may timeout or hang
- Particularly affects CLIs like kubectl with deep command hierarchies

**Workarounds:**
1. **Increase timeout**: Use `--timeout` flag when available
2. **Staged inspection**: Break large CLIs into smaller parts
3. **Static analysis**: Consider static AST parsing instead of runtime inspection

**Future Solutions:**
- Implement static analysis mode
- Add incremental inspection with caching
- Optimize command tree traversal algorithms
- Add parallel processing for independent command branches

### Memory Usage

**Contract Parsing**: Large contracts (500+ commands) can use 25MB+ memory during parsing.

**Mitigation strategies:**
- Implement streaming YAML parsing for large contracts
- Add contract validation before full parsing
- Consider contract splitting for very large CLIs

## Directory Structure

```
test-suite/performance/
├── README.md                    # This documentation
├── run-benchmarks.sh           # Main benchmark runner
├── test-real-world.sh          # Real-world CLI testing
├── test-with-kubectl.sh        # kubectl-like CLI test
├── test-with-helm.sh           # Helm-like CLI test
├── utils/                      
│   ├── generate_cli.go         # CLI project generator
│   └── generate_contract.go    # Contract generator
├── small/                      # Small CLI tests
├── medium/                     # Medium CLI tests
└── large/                      # Large CLI tests
```

## Contributing

### Adding New Benchmarks

1. **Create benchmark function** following Go conventions:
   ```go
   func BenchmarkYourFunction(b *testing.B) {
       // Setup
       b.ResetTimer()
       b.ReportAllocs()
       for i := 0; i < b.N; i++ {
           // Code to benchmark
       }
   }
   ```

2. **Add to appropriate test file**:
   - Validator benchmarks: `internal/validator/validator_bench_test.go`
   - Inspector benchmarks: `internal/inspector/inspector_bench_test.go`  
   - Contract benchmarks: `internal/contract/parser_bench_test.go`

3. **Update documentation** with expected performance ranges

### Performance Testing Best Practices

1. **Consistent environment**: Run benchmarks on dedicated hardware
2. **Multiple runs**: Use `BENCHMARK_COUNT=5` or higher for stable results
3. **Adequate duration**: Use `BENCHMARK_TIME=10s` for accurate measurements
4. **Memory profiling**: Always include `-benchmem` for allocation tracking
5. **Baseline comparison**: Keep baseline results for regression detection

### Reporting Performance Issues

When reporting performance issues, include:
- Benchmark results showing the problem
- System specifications (CPU, memory, OS)
- CLI characteristics (size, nesting depth, flag count)
- Profile data if available
- Reproduction steps

## Automation and CI Integration

### GitHub Actions Integration

Add to `.github/workflows/benchmarks.yml`:

```yaml
name: Performance Benchmarks

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    - name: Run Benchmarks
      run: |
        cd test-suite/performance
        ./run-benchmarks.sh
    - name: Upload Results
      uses: actions/upload-artifact@v3
      with:
        name: benchmark-results
        path: test-suite/performance/benchmark-results/
```

### Performance Monitoring Dashboard

Consider integrating with performance monitoring tools:
- **benchstat**: For statistical comparison
- **Continuous benchmarking**: Tools like gobench.org
- **Custom dashboard**: Parse benchmark results into time series data

This comprehensive performance testing suite ensures cliguard maintains excellent performance across all CLI sizes and helps identify optimization opportunities.