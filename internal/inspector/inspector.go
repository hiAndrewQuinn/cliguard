package inspector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const inspectorTemplate = `package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	{{- if .ImportPath }}
	{{- if .ImportAlias }}
	{{ .ImportAlias }} "{{ .ImportPath }}"
	{{- else }}
	"{{ .ImportPath }}"
	{{- end }}
	{{- end }}
)

type InspectedCLI struct {
	Use      string              ` + "`json:\"use\"`" + `
	Short    string              ` + "`json:\"short\"`" + `
	Long     string              ` + "`json:\"long,omitempty\"`" + `
	Flags    []InspectedFlag     ` + "`json:\"flags,omitempty\"`" + `
	Commands []InspectedCommand  ` + "`json:\"commands,omitempty\"`" + `
}

type InspectedCommand struct {
	Use      string              ` + "`json:\"use\"`" + `
	Short    string              ` + "`json:\"short\"`" + `
	Long     string              ` + "`json:\"long,omitempty\"`" + `
	Flags    []InspectedFlag     ` + "`json:\"flags,omitempty\"`" + `
	Commands []InspectedCommand  ` + "`json:\"commands,omitempty\"`" + `
}

type InspectedFlag struct {
	Name       string ` + "`json:\"name\"`" + `
	Shorthand  string ` + "`json:\"shorthand,omitempty\"`" + `
	Usage      string ` + "`json:\"usage\"`" + `
	Type       string ` + "`json:\"type\"`" + `
	Persistent bool   ` + "`json:\"persistent\"`" + `
}

func main() {
	var rootCmd *cobra.Command
	
	{{ if .EntrypointFunc }}
	// Call the user's entrypoint function
	{{- if .ImportAlias }}
	rootCmd = {{ .ImportAlias }}.{{ .EntrypointFunc }}()
	{{- else }}
	rootCmd = {{ .EntrypointFunc }}()
	{{- end }}
	{{ else }}
	// Try to find the root command in common locations
	if cmd := findRootCommand(); cmd != nil {
		rootCmd = cmd
	} else {
		fmt.Fprintf(os.Stderr, "Could not find root command\n")
		os.Exit(1)
	}
	{{ end }}

	// Inspect the command tree
	cli := inspectCommand(rootCmd)
	
	// Output as JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cli); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}

func findRootCommand() *cobra.Command {
	// This is a placeholder - in real implementation, we'd use reflection
	// or require the user to specify the entrypoint
	return nil
}

func inspectCommand(cmd *cobra.Command) InspectedCLI {
	cli := InspectedCLI{
		Use:   cmd.Use,
		Short: cmd.Short,
		Long:  cmd.Long,
	}
	
	// Inspect local flags
	localFlags := inspectFlagSet(cmd.Flags(), false)
	
	// Inspect persistent flags
	persistentFlags := inspectFlagSet(cmd.PersistentFlags(), true)
	
	// Combine flags, avoiding duplicates
	flagMap := make(map[string]InspectedFlag)
	for _, f := range localFlags {
		flagMap[f.Name] = f
	}
	for _, f := range persistentFlags {
		if _, exists := flagMap[f.Name]; !exists {
			flagMap[f.Name] = f
		}
	}
	
	cli.Flags = make([]InspectedFlag, 0, len(flagMap))
	for _, f := range flagMap {
		cli.Flags = append(cli.Flags, f)
	}
	
	// Inspect subcommands
	for _, subcmd := range cmd.Commands() {
		if subcmd.Hidden {
			continue
		}
		cli.Commands = append(cli.Commands, inspectSubcommand(subcmd))
	}
	
	return cli
}

func inspectSubcommand(cmd *cobra.Command) InspectedCommand {
	command := InspectedCommand{
		Use:   cmd.Use,
		Short: cmd.Short,
		Long:  cmd.Long,
	}
	
	// Inspect local flags only (persistent flags are inherited)
	command.Flags = inspectFlagSet(cmd.Flags(), false)
	
	// Inspect subcommands
	for _, subcmd := range cmd.Commands() {
		if subcmd.Hidden {
			continue
		}
		command.Commands = append(command.Commands, inspectSubcommand(subcmd))
	}
	
	return command
}

func inspectFlagSet(flags *pflag.FlagSet, persistent bool) []InspectedFlag {
	var inspectedFlags []InspectedFlag
	
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		
		inspectedFlag := InspectedFlag{
			Name:       flag.Name,
			Shorthand:  flag.Shorthand,
			Usage:      flag.Usage,
			Type:       getFlagType(flag),
			Persistent: persistent,
		}
		
		inspectedFlags = append(inspectedFlags, inspectedFlag)
	})
	
	return inspectedFlags
}

func getFlagType(flag *pflag.Flag) string {
	// Get the type from the flag's value
	flagType := reflect.TypeOf(flag.Value).String()
	
	// Map common pflag types to simpler names
	typeMap := map[string]string{
		"*pflag.stringValue":     "string",
		"*pflag.boolValue":       "bool",
		"*pflag.intValue":        "int",
		"*pflag.int64Value":      "int64",
		"*pflag.float64Value":    "float64",
		"*pflag.durationValue":   "duration",
		"*pflag.stringSliceValue": "stringSlice",
	}
	
	if simpleType, ok := typeMap[flagType]; ok {
		return simpleType
	}
	
	return flagType
}
`

// InspectProject generates an inspector program and runs it to get the CLI structure
func InspectProject(projectPath, entrypoint string) (*InspectedCLI, error) {
	// Create a temporary directory for the inspector
	tempDir, err := os.MkdirTemp("", "cliguard-inspector-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Parse the entrypoint to extract import path and function
	var importPath, entrypointFunc, importAlias string
	isMainPackage := false

	if entrypoint != "" {
		parts := strings.Split(entrypoint, ".")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid entrypoint format, expected: package.Function")
		}

		// Check if it's just package.Function (e.g., main.NewRootCmd)
		if len(parts) == 2 && parts[0] == "main" {
			// For main package, we'll handle this later
			isMainPackage = true
			entrypointFunc = parts[1]
		} else {
			importPath = strings.Join(parts[:len(parts)-1], ".")
			entrypointFunc = parts[len(parts)-1]
			// Use an alias to avoid naming conflicts
			importAlias = "userCmd"
		}
	}

	// Initialize go module in temp directory first
	initCmd := exec.Command("go", "mod", "init", "cliguard-inspector")
	initCmd.Dir = tempDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to init module: %w\nOutput: %s", err, output)
	}

	// Handle imports and code copying based on the entrypoint
	if isMainPackage {
		// For main package, we need to copy the project as a module
		// First copy go.mod if it exists
		srcGoMod := filepath.Join(projectPath, "go.mod")
		if _, err := os.Stat(srcGoMod); err == nil {
			modContent, err := os.ReadFile(srcGoMod)
			if err != nil {
				return nil, fmt.Errorf("failed to read target go.mod: %w", err)
			}

			moduleName := getModuleName(modContent)
			if moduleName != "" {
				// Update our inspector to import the module and create an alias
				importPath = moduleName
				importAlias = "userPkg"

				// Add replace directive
				replaceCmd := exec.Command("go", "mod", "edit", "-replace", fmt.Sprintf("%s=%s", moduleName, projectPath))
				replaceCmd.Dir = tempDir
				if output, err := replaceCmd.CombinedOutput(); err != nil {
					return nil, fmt.Errorf("failed to add replace directive: %w\nOutput: %s", err, output)
				}
			}
		}
	} else if importPath != "" {
		// For non-main packages, add replace directive
		srcGoMod := filepath.Join(projectPath, "go.mod")
		if _, err := os.Stat(srcGoMod); err == nil {
			modContent, err := os.ReadFile(srcGoMod)
			if err != nil {
				return nil, fmt.Errorf("failed to read target go.mod: %w", err)
			}

			replaceCmd := exec.Command("go", "mod", "edit", "-replace", fmt.Sprintf("%s=%s", getModuleName(modContent), projectPath))
			replaceCmd.Dir = tempDir
			if output, err := replaceCmd.CombinedOutput(); err != nil {
				return nil, fmt.Errorf("failed to add replace directive: %w\nOutput: %s", err, output)
			}
		}
	}

	// Now generate the inspector program with the correct import details
	tmpl, err := template.New("inspector").Parse(inspectorTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		ImportPath     string
		ImportAlias    string
		EntrypointFunc string
	}{
		ImportPath:     importPath,
		ImportAlias:    importAlias,
		EntrypointFunc: entrypointFunc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Write the inspector program
	inspectorPath := filepath.Join(tempDir, "inspector.go")
	if err := os.WriteFile(inspectorPath, buf.Bytes(), 0644); err != nil {
		return nil, fmt.Errorf("failed to write inspector program: %w", err)
	}

	// Get dependencies
	getCmd := exec.Command("go", "get", "./...")
	getCmd.Dir = tempDir
	if output, err := getCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w\nOutput: %s", err, output)
	}

	// Run the inspector
	runCmd := exec.Command("go", "run", "inspector.go")
	runCmd.Dir = tempDir
	output, err := runCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("inspector failed: %w\nStderr: %s", err, exitErr.Stderr)
		}
		return nil, fmt.Errorf("failed to run inspector: %w", err)
	}

	// Parse the JSON output
	var cli InspectedCLI
	if err := json.Unmarshal(output, &cli); err != nil {
		return nil, fmt.Errorf("failed to parse inspector output: %w\nOutput: %s", err, output)
	}

	return &cli, nil
}

// getModuleName extracts the module name from go.mod content
func getModuleName(goModContent []byte) string {
	lines := strings.Split(string(goModContent), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}
