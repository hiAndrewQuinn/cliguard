#!/bin/bash

# regenerate-test-projects.sh
# Regenerates the test-projects/ directory by cloning popular Go CLI projects
# Used for testing cliguard against real-world Cobra applications

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_PROJECTS_DIR="$SCRIPT_DIR/test-projects"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to clone a repository
clone_repo() {
    local repo_url="$1"
    local project_name="$2"
    local target_dir="$TEST_PROJECTS_DIR/$project_name"
    
    log_info "Cloning $project_name from $repo_url..."
    
    if [ -d "$target_dir" ]; then
        log_warning "$project_name already exists, removing..."
        rm -rf "$target_dir"
    fi
    
    if git clone --depth 1 "$repo_url" "$target_dir"; then
        log_success "Successfully cloned $project_name"
        return 0
    else
        log_error "Failed to clone $project_name"
        return 1
    fi
}

# Function to create a simple demo CLI project
create_demo_cli() {
    local demo_dir="$TEST_PROJECTS_DIR/demo-cli"
    
    log_info "Creating demo CLI project..."
    
    mkdir -p "$demo_dir"/{cmd,internal}
    
    # Create go.mod
    cat > "$demo_dir/go.mod" << 'EOF'
module github.com/hiAndrewQuinn/cliguard/test-projects/demo-cli

go 1.21

require github.com/spf13/cobra v1.8.0

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
EOF

    # Create main.go
    cat > "$demo_dir/main.go" << 'EOF'
package main

import (
	"github.com/hiAndrewQuinn/cliguard/test-projects/demo-cli/cmd"
)

func main() {
	cmd.Execute()
}
EOF

    # Create cmd/root.go
    cat > "$demo_dir/cmd/root.go" << 'EOF'
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "demo",
	Short: "A simple demo CLI for testing cliguard",
	Long:  `A simple demo CLI application built with Cobra for testing cliguard functionality.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from demo CLI!")
	},
}

// NewRootCmd returns the root command for use with cliguard
func NewRootCmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}
EOF

    log_success "Created demo CLI project"
}

main() {
    log_info "Starting test-projects regeneration..."
    
    # Remove existing test-projects directory if it exists
    if [ -d "$TEST_PROJECTS_DIR" ]; then
        log_warning "Removing existing test-projects directory..."
        rm -rf "$TEST_PROJECTS_DIR"
    fi
    
    # Create test-projects directory
    mkdir -p "$TEST_PROJECTS_DIR"
    
    # List of repositories to clone
    declare -A repositories=(
        ["cli"]="https://github.com/cli/cli.git"
        ["helm"]="https://github.com/helm/helm.git"
        ["hugo"]="https://github.com/gohugoio/hugo.git"
        ["cockroach"]="https://github.com/cockroachdb/cockroach.git"
        ["etcd"]="https://github.com/etcd-io/etcd.git"
        ["kubernetes"]="https://github.com/kubernetes/kubernetes.git"
    )
    
    local success_count=0
    local total_count=${#repositories[@]}
    
    # Clone each repository
    for project in "${!repositories[@]}"; do
        if clone_repo "${repositories[$project]}" "$project"; then
            ((success_count++))
        fi
    done
    
    # Create demo CLI project
    create_demo_cli
    ((success_count++))
    ((total_count++))
    
    # Copy the TEST_RESULTS.md if we have it as a reference
    if [ -f "$SCRIPT_DIR/TEST_RESULTS.md.backup" ]; then
        cp "$SCRIPT_DIR/TEST_RESULTS.md.backup" "$TEST_PROJECTS_DIR/TEST_RESULTS.md"
        log_info "Restored TEST_RESULTS.md from backup"
    else
        # Create a basic TEST_RESULTS.md template
        cat > "$TEST_PROJECTS_DIR/TEST_RESULTS.md" << 'EOF'
# Cliguard Testing Results on Popular Go Cobra Projects

## Summary

This document summarizes the results of testing cliguard's generate and validate commands on popular Go projects that use Cobra.

## Test Projects

### Available Projects

1. **GitHub CLI (gh)** - `github.com/cli/cli`
2. **Helm** - `github.com/helm/helm`  
3. **Hugo** - `github.com/gohugoio/hugo`
4. **CockroachDB** - `github.com/cockroachdb/cockroach`
5. **etcd** - `github.com/etcd-io/etcd`
6. **Kubernetes** - `github.com/kubernetes/kubernetes`
7. **Demo CLI** - Simple test project with standard Cobra setup

## Usage

Run cliguard against these projects to test compatibility and performance.

## Notes

- Projects are cloned with `--depth 1` for faster setup
- Remove this directory when context windows get too large
- Regenerate using `./regenerate-test-projects.sh`
EOF
        log_info "Created basic TEST_RESULTS.md template"
    fi
    
    # Summary
    echo
    log_success "Regeneration complete!"
    log_info "Successfully set up $success_count/$total_count projects"
    
    if [ $success_count -ne $total_count ]; then
        log_warning "Some projects failed to clone. Check your internet connection."
    fi
    
    log_info "Test projects are ready in: $TEST_PROJECTS_DIR"
}

# Show usage if -h or --help is passed
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    cat << 'EOF'
Usage: ./regenerate-test-projects.sh

Regenerates the test-projects/ directory by cloning popular Go CLI projects
that use Cobra. This is useful for testing cliguard against real-world 
applications and when the directory needs to be removed due to large 
context windows.

Projects cloned:
- GitHub CLI (cli/cli)
- Helm (helm/helm)
- Hugo (gohugoio/hugo)  
- CockroachDB (cockroachdb/cockroach)
- etcd (etcd-io/etcd)
- Kubernetes (kubernetes/kubernetes)
- Demo CLI (simple test project)

Options:
  -h, --help    Show this help message

The script will remove any existing test-projects/ directory and recreate 
it from scratch.
EOF
    exit 0
fi

main "$@"