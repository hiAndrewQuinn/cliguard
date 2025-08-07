#!/bin/bash

# Cliguard Performance Benchmark Suite
# This script runs comprehensive performance benchmarks for cliguard

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BENCHMARK_TIME=${BENCHMARK_TIME:-10s}
BENCHMARK_COUNT=${BENCHMARK_COUNT:-3}
OUTPUT_DIR=${OUTPUT_DIR:-"./benchmark-results"}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_FILE="${OUTPUT_DIR}/benchmark_${TIMESTAMP}.txt"

# Print colored output
print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Start benchmarking
print_header "Cliguard Performance Benchmarks"
echo "Timestamp: $(date)"
echo "Benchmark time: $BENCHMARK_TIME"
echo "Benchmark count: $BENCHMARK_COUNT"
echo "Results will be saved to: $RESULTS_FILE"
echo ""

# Initialize results file
{
    echo "Cliguard Performance Benchmark Results"
    echo "======================================"
    echo "Timestamp: $(date)"
    echo "Go Version: $(go version)"
    echo "OS: $(uname -s) $(uname -r)"
    echo "CPU: $(uname -m)"
    echo ""
} > "$RESULTS_FILE"

# Function to run benchmarks for a package
run_package_benchmarks() {
    local package=$1
    local description=$2
    
    print_header "$description"
    
    echo "Running benchmarks for: $package"
    echo "" | tee -a "$RESULTS_FILE"
    echo "$description" | tee -a "$RESULTS_FILE"
    echo "----------------------------------------" | tee -a "$RESULTS_FILE"
    
    # Run benchmarks with memory profiling
    if go test -bench=. -benchmem -benchtime="$BENCHMARK_TIME" -count="$BENCHMARK_COUNT" "$package" 2>&1 | tee -a "$RESULTS_FILE"; then
        print_success "Completed $description benchmarks"
    else
        print_error "Failed to run $description benchmarks"
        return 1
    fi
    
    echo "" | tee -a "$RESULTS_FILE"
}

# Run validator benchmarks
run_package_benchmarks \
    "./internal/validator" \
    "Validator Performance Benchmarks"

# Run inspector benchmarks
run_package_benchmarks \
    "./internal/inspector" \
    "Inspector Performance Benchmarks"

# Run contract parser benchmarks
run_package_benchmarks \
    "./internal/contract" \
    "Contract Parser Performance Benchmarks"

# Run CPU profiling for the largest benchmarks
print_header "CPU Profiling for Large CLIs"
echo "Generating CPU profiles for analysis..."

# Create profiles directory
PROFILES_DIR="${OUTPUT_DIR}/profiles_${TIMESTAMP}"
mkdir -p "$PROFILES_DIR"

# Run validator profiling
echo "Profiling validator with large CLI..."
go test -bench=BenchmarkValidateExtraLargeCLI \
    -benchtime=10s \
    -cpuprofile="${PROFILES_DIR}/validator_cpu.prof" \
    ./internal/validator 2>&1 | tail -5

# Run inspector profiling
echo "Profiling inspector with large project..."
go test -bench=BenchmarkInspectLargeProject \
    -benchtime=10s \
    -cpuprofile="${PROFILES_DIR}/inspector_cpu.prof" \
    ./internal/inspector 2>&1 | tail -5

# Run contract parser profiling
echo "Profiling contract parser with large contract..."
go test -bench=BenchmarkLoadExtraLargeContract \
    -benchtime=10s \
    -cpuprofile="${PROFILES_DIR}/contract_cpu.prof" \
    ./internal/contract 2>&1 | tail -5

print_success "CPU profiles saved to: $PROFILES_DIR"

# Memory profiling
print_header "Memory Profiling for Large CLIs"
echo "Generating memory profiles for analysis..."

# Run validator memory profiling
echo "Memory profiling validator..."
go test -bench=BenchmarkValidateExtraLargeCLI \
    -benchtime=10s \
    -memprofile="${PROFILES_DIR}/validator_mem.prof" \
    ./internal/validator 2>&1 | tail -5

print_success "Memory profiles saved to: $PROFILES_DIR"

# Generate comparison report if previous results exist
print_header "Benchmark Comparison"

LATEST_PREVIOUS=$(ls -t "$OUTPUT_DIR"/benchmark_*.txt 2>/dev/null | sed -n '2p')
if [ -n "$LATEST_PREVIOUS" ]; then
    echo "Comparing with previous run: $(basename "$LATEST_PREVIOUS")"
    
    # Use benchstat if available
    if command -v benchstat &> /dev/null; then
        echo "" | tee -a "$RESULTS_FILE"
        echo "Benchstat Comparison" | tee -a "$RESULTS_FILE"
        echo "----------------------------------------" | tee -a "$RESULTS_FILE"
        benchstat "$LATEST_PREVIOUS" "$RESULTS_FILE" | tee -a "$RESULTS_FILE"
        print_success "Comparison complete"
    else
        print_warning "benchstat not found. Install with: go install golang.org/x/perf/cmd/benchstat@latest"
    fi
else
    echo "No previous results found for comparison"
fi

# Summary
print_header "Benchmark Summary"

echo "Results saved to: $RESULTS_FILE"
echo "CPU/Memory profiles saved to: $PROFILES_DIR"
echo ""

# Extract and display key metrics
echo "Key Performance Metrics:"
echo "------------------------"

# Extract validator metrics
echo ""
echo "Validator Performance:"
grep -E "BenchmarkValidate.*-[0-9]+" "$RESULTS_FILE" | tail -4 || true

# Extract inspector metrics
echo ""
echo "Inspector Performance:"
grep -E "BenchmarkInspect.*-[0-9]+" "$RESULTS_FILE" | tail -3 || true

# Extract contract parser metrics
echo ""
echo "Contract Parser Performance:"
grep -E "BenchmarkLoad.*-[0-9]+" "$RESULTS_FILE" | tail -4 || true

echo ""
print_success "Performance benchmarks completed successfully!"

# Provide analysis tips
echo ""
echo "To analyze CPU profiles:"
echo "  go tool pprof ${PROFILES_DIR}/validator_cpu.prof"
echo "  go tool pprof ${PROFILES_DIR}/inspector_cpu.prof"
echo "  go tool pprof ${PROFILES_DIR}/contract_cpu.prof"
echo ""
echo "To analyze memory profiles:"
echo "  go tool pprof ${PROFILES_DIR}/validator_mem.prof"
echo ""
echo "Useful pprof commands:"
echo "  top10    - Show top 10 functions by CPU/memory"
echo "  list <func> - Show source code for function"
echo "  web      - Generate SVG visualization (requires graphviz)"

exit 0