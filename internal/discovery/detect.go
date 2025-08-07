package discovery

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/hiAndrewQuinn/cliguard/internal/filesystem"
)

// DetectEntrypointFramework detects the CLI framework used by the given entrypoint
func DetectEntrypointFramework(projectPath, entrypoint string, fs filesystem.FileSystem) (string, error) {
	if entrypoint == "" {
		return "", fmt.Errorf("entrypoint cannot be empty")
	}

	// Use default filesystem if none provided
	if fs == nil {
		fs = &filesystem.OSFileSystem{}
	}

	// Parse the entrypoint to get package and function
	parts := strings.Split(entrypoint, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid entrypoint format: %s (expected format: package.Function)", entrypoint)
	}

	// Get the package path and function name
	packagePath := strings.Join(parts[:len(parts)-1], ".")
	functionName := parts[len(parts)-1]

	// Convert package path to file path
	// Remove the module prefix to get the relative path
	modulePath, err := getModulePathFromFS(projectPath, fs)
	if err != nil {
		return "", fmt.Errorf("failed to get module path: %w", err)
	}

	// Calculate relative package path
	relPath := strings.TrimPrefix(packagePath, modulePath)
	relPath = strings.TrimPrefix(relPath, "/")

	// Look for Go files in the package directory
	packageDir := filepath.Join(projectPath, relPath)

	// Try common file names
	fileNames := []string{"root.go", "main.go", "cmd.go", "app.go"}

	// Also try the function name as a file
	if functionName != "" {
		fileNames = append([]string{strings.ToLower(functionName) + ".go"}, fileNames...)
	}

	for _, fileName := range fileNames {
		filePath := filepath.Join(packageDir, fileName)
		framework, err := detectFrameworkInFile(filePath, fs)
		if err == nil && framework != "" {
			return framework, nil
		}
	}

	// If not found in specific files, scan all .go files in the directory
	files, err := os.ReadDir(packageDir)
	if err != nil {
		return "", fmt.Errorf("failed to read package directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".go") && !strings.HasSuffix(file.Name(), "_test.go") {
			filePath := filepath.Join(packageDir, file.Name())
			framework, err := detectFrameworkInFile(filePath, fs)
			if err == nil && framework != "" {
				return framework, nil
			}
		}
	}

	return "", fmt.Errorf("could not detect CLI framework for entrypoint: %s", entrypoint)
}

// detectFrameworkInFile detects the CLI framework in a specific file
func detectFrameworkInFile(filePath string, fs filesystem.FileSystem) (string, error) {
	content, err := fs.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Parse the file to check imports
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ImportsOnly)
	if err != nil {
		return "", err
	}

	// Check imports for known CLI frameworks
	for _, imp := range node.Imports {
		if imp.Path != nil {
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Check for Cobra
			if strings.Contains(importPath, "github.com/spf13/cobra") {
				return "cobra", nil
			}

			// Check for urfave/cli
			if strings.Contains(importPath, "github.com/urfave/cli") {
				return "urfave/cli", nil
			}

			// Check for kingpin
			if strings.Contains(importPath, "kingpin") &&
				(strings.Contains(importPath, "gopkg.in/alecthomas") ||
					strings.Contains(importPath, "github.com/alecthomas")) {
				return "kingpin", nil
			}
		}
	}

	// Check for standard library flag package
	for _, imp := range node.Imports {
		if imp.Path != nil && strings.Trim(imp.Path.Value, `"`) == "flag" {
			// Look for flag.Parse() in the content to confirm it's actually a CLI
			if strings.Contains(string(content), "flag.Parse()") {
				return "flag", nil
			}
		}
	}

	return "", nil
}

// getModulePathFromFS reads the module path from go.mod using the filesystem interface
func getModulePathFromFS(projectPath string, fs filesystem.FileSystem) (string, error) {
	goModPath := filepath.Join(projectPath, "go.mod")
	content, err := fs.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("module path not found in go.mod")
}
