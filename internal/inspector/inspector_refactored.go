package inspector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hiAndrewQuinn/cliguard/internal/errors"
	"github.com/hiAndrewQuinn/cliguard/internal/executor"
	"github.com/hiAndrewQuinn/cliguard/internal/filesystem"
)

// Config holds the configuration for the inspector
type Config struct {
	ProjectPath string
	Entrypoint  string
	Timeout     time.Duration
	FileSystem  filesystem.FileSystem
	Executor    executor.CommandExecutor
}

// Inspector provides CLI inspection functionality
type Inspector struct {
	config Config
}

// NewInspector creates a new Inspector with the given configuration
func NewInspector(config Config) *Inspector {
	// Set defaults if not provided
	if config.FileSystem == nil {
		config.FileSystem = &filesystem.OSFileSystem{}
	}
	if config.Executor == nil {
		config.Executor = &executor.OSExecutor{}
	}

	// Wrap executor with timeout if specified
	if config.Timeout > 0 {
		config.Executor = executor.NewTimeoutExecutor(config.Executor, config.Timeout)
	}

	return &Inspector{
		config: config,
	}
}

// EntrypointInfo contains parsed entrypoint information
type EntrypointInfo struct {
	ImportPath    string
	ImportAlias   string
	FunctionName  string
	IsMainPackage bool
}

// Inspect generates an inspector program and runs it to get the CLI structure
func (i *Inspector) Inspect() (*InspectedCLI, error) {
	// Create a temporary directory for the inspector
	tempDir, err := i.config.FileSystem.MkdirTemp("", "cliguard-inspector-*")
	if err != nil {
		return nil, errors.TempDirError{
			Operation: "create",
			Err:       err,
		}
	}
	defer func() {
		_ = i.config.FileSystem.RemoveAll(tempDir)
	}()

	// Parse the entrypoint
	entrypointInfo, err := i.parseEntrypoint(i.config.Entrypoint)
	if err != nil {
		return nil, err // parseEntrypoint already returns proper error
	}

	// Setup the temporary module
	if err := i.setupTempModule(tempDir, entrypointInfo); err != nil {
		return nil, fmt.Errorf("failed to setup temp module: %w", err)
	}

	// Generate the inspector code
	inspectorCode, err := i.generateInspectorCode(entrypointInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate inspector code: %w", err)
	}

	// Write the inspector program
	inspectorPath := filepath.Join(tempDir, "inspector.go")
	if err := i.config.FileSystem.WriteFile(inspectorPath, []byte(inspectorCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write inspector program: %w", err)
	}

	// Get dependencies
	if err := i.getDependencies(tempDir); err != nil {
		return nil, err // getDependencies already returns proper error
	}

	// Run the inspector
	output, err := i.runInspector(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run inspector: %w", err)
	}

	// Parse the output
	inspectedCLI, err := i.parseInspectorOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inspector output: %w", err)
	}
	return inspectedCLI, nil
}

// parseEntrypoint parses the entrypoint string into its components
func (i *Inspector) parseEntrypoint(entrypoint string) (*EntrypointInfo, error) {
	info := &EntrypointInfo{}

	if entrypoint == "" {
		return info, nil
	}

	parts := strings.Split(entrypoint, ".")
	if len(parts) < 2 {
		return nil, errors.EntrypointParseError{
			Entrypoint: entrypoint,
			Reason:     "invalid format",
		}
	}

	// Check if it's just package.Function (e.g., main.NewRootCmd)
	if len(parts) == 2 && parts[0] == "main" {
		info.IsMainPackage = true
		info.FunctionName = parts[1]
	} else {
		info.ImportPath = strings.Join(parts[:len(parts)-1], ".")
		info.FunctionName = parts[len(parts)-1]
		info.ImportAlias = "userCmd"
	}

	return info, nil
}

// setupTempModule sets up the temporary Go module
func (i *Inspector) setupTempModule(tempDir string, info *EntrypointInfo) error {
	// Initialize go module
	initCmd := i.config.Executor.Command("go", "mod", "init", "cliguard-inspector")
	initCmd.SetDir(tempDir)
	if output, err := initCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to init module: %w\nOutput: %s", err, output)
	}

	// Handle imports based on the entrypoint
	if info.IsMainPackage {
		// For main package, we need to handle it specially
		srcGoMod := filepath.Join(i.config.ProjectPath, "go.mod")
		if _, err := i.config.FileSystem.Stat(srcGoMod); err == nil {
			modContent, err := i.config.FileSystem.ReadFile(srcGoMod)
			if err != nil {
				return fmt.Errorf("failed to read target go.mod: %w", err)
			}

			moduleName := getModuleName(modContent)
			if moduleName != "" {
				info.ImportPath = moduleName
				info.ImportAlias = "userPkg"

				// Add replace directive
				if err := i.addReplaceDirective(tempDir, moduleName); err != nil {
					return fmt.Errorf("failed to add replace directive: %w", err)
				}
			}
		}
	} else if info.ImportPath != "" {
		// For non-main packages, add replace directive
		srcGoMod := filepath.Join(i.config.ProjectPath, "go.mod")
		if _, err := i.config.FileSystem.Stat(srcGoMod); err == nil {
			modContent, err := i.config.FileSystem.ReadFile(srcGoMod)
			if err != nil {
				return fmt.Errorf("failed to read target go.mod: %w", err)
			}

			moduleName := getModuleName(modContent)
			if moduleName != "" {
				if err := i.addReplaceDirective(tempDir, moduleName); err != nil {
					return fmt.Errorf("failed to add replace directive: %w", err)
				}
			}
		}
	}

	return nil
}

// addReplaceDirective adds a replace directive to go.mod
func (i *Inspector) addReplaceDirective(tempDir, moduleName string) error {
	// Make sure we have an absolute path
	absProjectPath := i.config.ProjectPath
	if !filepath.IsAbs(absProjectPath) {
		var err error
		absProjectPath, err = filepath.Abs(absProjectPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
	}

	replaceCmd := i.config.Executor.Command("go", "mod", "edit", "-replace",
		fmt.Sprintf("%s=%s", moduleName, absProjectPath))
	replaceCmd.SetDir(tempDir)
	if output, err := replaceCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add replace directive: %w\nOutput: %s", err, output)
	}
	return nil
}

// generateInspectorCode generates the inspector Go code
func (i *Inspector) generateInspectorCode(info *EntrypointInfo) (string, error) {
	tmpl, err := template.New("inspector").Parse(inspectorTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		ImportPath     string
		ImportAlias    string
		EntrypointFunc string
	}{
		ImportPath:     info.ImportPath,
		ImportAlias:    info.ImportAlias,
		EntrypointFunc: info.FunctionName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// getDependencies gets the Go module dependencies
func (i *Inspector) getDependencies(tempDir string) error {
	// Try to run go mod tidy with -e flag to ignore errors
	tidyCmd := i.config.Executor.Command("go", "mod", "tidy", "-e")
	tidyCmd.SetDir(tempDir)
	if _, err := tidyCmd.CombinedOutput(); err == nil {
		return nil
	}

	// Fall back to regular go get
	getCmd := i.config.Executor.Command("go", "get", "./...")
	getCmd.SetDir(tempDir)
	if output, err := getCmd.CombinedOutput(); err != nil {
		return errors.DependencyError{
			Operation: "go get ./...",
			Output:    string(output),
			Err:       err,
		}
	}
	return nil
}

// runInspector runs the inspector program
func (i *Inspector) runInspector(tempDir string) ([]byte, error) {
	runCmd := i.config.Executor.Command("go", "run", "inspector.go")
	runCmd.SetDir(tempDir)
	output, err := runCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("inspector execution failed: %w", err)
	}
	return output, nil
}

// parseInspectorOutput parses the JSON output from the inspector
func (i *Inspector) parseInspectorOutput(output []byte) (*InspectedCLI, error) {
	var cli InspectedCLI
	if err := json.Unmarshal(output, &cli); err != nil {
		return nil, fmt.Errorf("invalid JSON output: %w\n\nRaw output:\n%s", err, output)
	}
	return &cli, nil
}
