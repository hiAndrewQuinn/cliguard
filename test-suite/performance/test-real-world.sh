#!/bin/bash

# Test cliguard performance with all real-world CLI simulations
# This runs tests against small, medium, and large CLI configurations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "================================"
echo "Real-World CLI Performance Tests"
echo "================================"
echo ""

# Make scripts executable
chmod +x "${SCRIPT_DIR}/test-with-kubectl.sh"
chmod +x "${SCRIPT_DIR}/test-with-helm.sh"

# Test with small CLI (10 commands)
echo "1. Testing with Small CLI (10 commands)..."
echo "-------------------------------------------"
TEST_DIR="${SCRIPT_DIR}/small/test-cli"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

cd "$TEST_DIR"
cat > generate.go << 'EOF'
package main

import (
    "log"
    "github.com/hiAndrewQuinn/cliguard/test-suite/performance/utils"
)

func main() {
    if err := utils.GenerateCLIProject(".", utils.Small); err != nil {
        log.Fatal(err)
    }
    if err := utils.GenerateContract(".", utils.Small); err != nil {
        log.Fatal(err)
    }
}
EOF

go mod init small-test
go mod edit -replace github.com/hiAndrewQuinn/cliguard="${SCRIPT_DIR}/../.."
go get github.com/hiAndrewQuinn/cliguard/test-suite/performance/utils
go run generate.go

START_TIME=$(date +%s%N)
if go run "${SCRIPT_DIR}/../../main.go" inspect --target . --entry "cmd.NewRootCmd" > /dev/null 2>&1; then
    END_TIME=$(date +%s%N)
    ELAPSED=$((($END_TIME - $START_TIME) / 1000000))
    echo -e "${GREEN}✓${NC} Small CLI inspection: ${ELAPSED}ms"
else
    echo -e "${RED}✗${NC} Small CLI inspection failed"
fi

# Test with medium CLI (Helm-like)
echo ""
echo "2. Testing with Medium CLI (Helm-like, ~75 commands)..."
echo "-------------------------------------------------------"
if "${SCRIPT_DIR}/test-with-helm.sh" > /tmp/helm-test.log 2>&1; then
    echo -e "${GREEN}✓${NC} Medium CLI test passed"
    grep "completed in" /tmp/helm-test.log | sed 's/^/   /'
else
    echo -e "${RED}✗${NC} Medium CLI test failed"
    tail -20 /tmp/helm-test.log
fi

# Test with large CLI (kubectl-like)
echo ""
echo "3. Testing with Large CLI (kubectl-like, 500+ commands)..."
echo "----------------------------------------------------------"
if "${SCRIPT_DIR}/test-with-kubectl.sh" > /tmp/kubectl-test.log 2>&1; then
    echo -e "${GREEN}✓${NC} Large CLI test passed"
    grep "completed in" /tmp/kubectl-test.log | sed 's/^/   /'
else
    echo -e "${RED}✗${NC} Large CLI test failed"
    echo -e "${YELLOW}⚠${NC} This is expected if the 30-second timeout is hit"
    tail -20 /tmp/kubectl-test.log
fi

# Performance comparison summary
echo ""
echo "================================"
echo "Performance Summary"
echo "================================"

# Extract timings if available
if [ -f /tmp/helm-test.log ]; then
    echo "Medium CLI (75 commands):"
    grep "completed in" /tmp/helm-test.log | sed 's/^/  /'
fi

if [ -f /tmp/kubectl-test.log ]; then
    echo "Large CLI (500 commands):"
    grep "completed in" /tmp/kubectl-test.log | sed 's/^/  /' || echo "  Timed out (>30s)"
fi

echo ""
echo "Note: Large CLIs may timeout due to runtime inspection limitations."
echo "This is a known issue that could be addressed with static analysis."
echo ""
echo "Real-world CLI tests completed!"