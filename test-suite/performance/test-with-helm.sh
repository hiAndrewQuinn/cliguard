#!/bin/bash

# Test cliguard performance with a Helm-like CLI
# This simulates testing against a medium-sized, real-world CLI

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="${SCRIPT_DIR}/medium/helm-test"

echo "Testing cliguard with Helm-like CLI (~30 commands)..."
echo "====================================="

# Create test directory
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# Generate Helm-like CLI project
echo "Generating Helm-like CLI project..."
cat > "$TEST_DIR/generate.go" << 'EOF'
package main

import (
    "fmt"
    "log"
    "github.com/hiAndrewQuinn/cliguard/test-suite/performance/utils"
)

func main() {
    if err := utils.GenerateCLIProject(".", utils.Medium); err != nil {
        log.Fatal(err)
    }
    if err := utils.GenerateRealWorldContract(".", "helm"); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Generated Helm-like test CLI")
}
EOF

cd "$TEST_DIR"
go mod init helm-test
go mod edit -replace github.com/hiAndrewQuinn/cliguard="${SCRIPT_DIR}/../.."
go get github.com/hiAndrewQuinn/cliguard/test-suite/performance/utils
go run generate.go

# Time the inspection
echo ""
echo "Running cliguard inspect on Helm-like CLI..."
START_TIME=$(date +%s%N)

if go run "${SCRIPT_DIR}/../../main.go" inspect --target . --entry "cmd.NewRootCmd" > inspect-output.json 2>&1; then
    END_TIME=$(date +%s%N)
    ELAPSED=$((($END_TIME - $START_TIME) / 1000000))
    echo "✓ Inspection completed in ${ELAPSED}ms"
    
    # Show statistics
    echo ""
    echo "CLI Statistics:"
    jq '{name, command_count: .commands | length, global_flags: .global_flags | length}' inspect-output.json
else
    echo "✗ Inspection failed"
    exit 1
fi

# Time the validation
echo ""
echo "Running cliguard validate on Helm-like CLI..."
START_TIME=$(date +%s%N)

if go run "${SCRIPT_DIR}/../../main.go" validate \
    --contract helm-contract.yaml \
    --target . \
    --entry "cmd.NewRootCmd" > validate-output.txt 2>&1; then
    END_TIME=$(date +%s%N)
    ELAPSED=$((($END_TIME - $START_TIME) / 1000000))
    echo "✓ Validation completed in ${ELAPSED}ms"
    
    # Check validation result
    if grep -q "valid" validate-output.txt; then
        echo "✓ CLI is valid according to contract"
    else
        echo "⚠ Validation issues found:"
        cat validate-output.txt
    fi
else
    echo "✗ Validation failed"
    exit 1
fi

# Generate contract from CLI
echo ""
echo "Running cliguard generate on Helm-like CLI..."
START_TIME=$(date +%s%N)

if go run "${SCRIPT_DIR}/../../main.go" generate \
    --target . \
    --entry "cmd.NewRootCmd" > generated-contract.yaml 2>&1; then
    END_TIME=$(date +%s%N)
    ELAPSED=$((($END_TIME - $START_TIME) / 1000000))
    echo "✓ Contract generation completed in ${ELAPSED}ms"
    
    # Show contract statistics
    echo ""
    echo "Generated Contract Statistics:"
    echo "- Size: $(wc -l < generated-contract.yaml) lines"
    echo "- Commands: $(grep -c "^  - name:" generated-contract.yaml || echo 0)"
else
    echo "✗ Contract generation failed"
    exit 1
fi

echo ""
echo "Helm-like CLI test completed successfully!"
echo "========================================="