package cmd

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// setupTestFixtures ensures test fixtures are properly initialized
func setupTestFixtures(t *testing.T) string {
	t.Helper()
	fixtureDir := filepath.Join("..", "test", "fixtures", "simple-cli")

	// Ensure go.mod is tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = fixtureDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to tidy test fixture: %v", err)
	}

	return fixtureDir
}
