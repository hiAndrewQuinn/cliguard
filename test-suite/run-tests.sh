#!/bin/bash

# Cliguard Test Suite Runner
# This script runs all test cases and reports results

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CLIGUARD="${SCRIPT_DIR}/../cliguard"

# Test counter
PASSED=0
FAILED=0
SKIPPED=0

# Function to run a test
run_test() {
    local test_name=$1
    local test_dir=$2
    local entrypoint=$3
    local expected_result=${4:-"pass"}
    
    echo -e "\n${YELLOW}Running test: ${test_name}${NC}"
    echo "Directory: ${test_dir}"
    echo "Entrypoint: ${entrypoint}"
    
    # Generate contract
    if $CLIGUARD generate --project-path "$test_dir" --entrypoint "$entrypoint" > "$test_dir/contract.yaml" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} Generate succeeded"
    else
        if [ "$expected_result" = "fail-generate" ]; then
            echo -e "${GREEN}✓${NC} Generate failed as expected"
            ((PASSED++))
            return
        else
            echo -e "${RED}✗${NC} Generate failed unexpectedly"
            ((FAILED++))
            return
        fi
    fi
    
    # Validate contract
    if $CLIGUARD validate --project-path "$test_dir" --entrypoint "$entrypoint" --contract "$test_dir/contract.yaml" 2>&1 | grep -q "Validation passed"; then
        if [ "$expected_result" = "pass" ]; then
            echo -e "${GREEN}✓${NC} Validation passed as expected"
            ((PASSED++))
        else
            echo -e "${RED}✗${NC} Validation passed but should have failed"
            ((FAILED++))
        fi
    else
        if [ "$expected_result" = "fail-validate" ]; then
            echo -e "${GREEN}✓${NC} Validation failed as expected"
            ((PASSED++))
        else
            echo -e "${RED}✗${NC} Validation failed unexpectedly"
            ((FAILED++))
        fi
    fi
}

# Function to run a breaking change test
run_breaking_test() {
    local test_name=$1
    local v1_dir=$2
    local v1_entrypoint=$3
    local v2_dir=$4
    local v2_entrypoint=$5
    
    echo -e "\n${YELLOW}Running breaking change test: ${test_name}${NC}"
    
    # Generate v1 contract
    if $CLIGUARD generate --project-path "$v1_dir" --entrypoint "$v1_entrypoint" > "$v1_dir/contract.yaml" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} V1 contract generated"
    else
        echo -e "${RED}✗${NC} Failed to generate v1 contract"
        ((FAILED++))
        return
    fi
    
    # Validate v2 against v1 contract (should fail)
    if $CLIGUARD validate --project-path "$v2_dir" --entrypoint "$v2_entrypoint" --contract "$v1_dir/contract.yaml" 2>&1 | grep -q "Validation failed"; then
        echo -e "${GREEN}✓${NC} Breaking changes detected correctly"
        ((PASSED++))
    else
        echo -e "${RED}✗${NC} Failed to detect breaking changes"
        ((FAILED++))
    fi
}

echo "=========================================="
echo "       Cliguard Test Suite Runner         "
echo "=========================================="

# Basic tests
run_test "Simple CLI" "$SCRIPT_DIR/basic/simple-cli" "github.com/cliguard/test/simple/cmd.NewRootCmd"
run_test "Subcommands CLI" "$SCRIPT_DIR/basic/subcommands" "github.com/cliguard/test/subcommands/cmd.NewRootCmd"

# Edge case tests
run_test "Flag Types CLI" "$SCRIPT_DIR/edge-cases/flag-types" "github.com/cliguard/test/flagtypes/cmd.NewRootCmd" "fail-validate"

# Breaking change tests
run_breaking_test "Breaking Changes" \
    "$SCRIPT_DIR/validation/breaking/v1" "github.com/cliguard/test/breaking-v1/cmd.NewRootCmd" \
    "$SCRIPT_DIR/validation/breaking/v2" "github.com/cliguard/test/breaking-v2/cmd.NewRootCmd"

# Summary
echo -e "\n=========================================="
echo "              Test Summary                "
echo "=========================================="
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo -e "${YELLOW}Skipped:${NC} $SKIPPED"
echo -e "Total: $((PASSED + FAILED + SKIPPED))"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed!${NC}"
    exit 1
fi