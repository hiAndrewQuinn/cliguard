package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test with help flag to avoid actually running validation
	os.Args = []string{"cliguard", "--help"}

	// We can't easily test main() directly since it calls os.Exit
	// This test mainly ensures the file compiles and imports work
	// The actual functionality is tested through cmd package tests
}
