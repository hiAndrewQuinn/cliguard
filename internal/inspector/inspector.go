package inspector

import (
	"strings"
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
		// Basic types (existing)
		"*pflag.stringValue":      "string",
		"*pflag.boolValue":        "bool",
		"*pflag.intValue":         "int",
		"*pflag.int64Value":       "int64",
		"*pflag.float64Value":     "float64",
		"*pflag.durationValue":    "duration",
		"*pflag.stringSliceValue": "stringSlice",
		
		// Integer variants
		"*pflag.int8Value":        "int8",
		"*pflag.int16Value":       "int16",
		"*pflag.int32Value":       "int32",
		"*pflag.uint8Value":       "uint8",
		"*pflag.uint16Value":      "uint16",
		"*pflag.uint32Value":      "uint32",
		"*pflag.uint64Value":      "uint64",
		"*pflag.uintValue":        "uint",
		
		// Float variants
		"*pflag.float32Value":     "float32",
		
		// Slice types
		"*pflag.intSliceValue":     "intSlice",
		"*pflag.int32SliceValue":   "int32Slice",
		"*pflag.int64SliceValue":   "int64Slice",
		"*pflag.uintSliceValue":    "uintSlice",
		"*pflag.float32SliceValue": "float32Slice",
		"*pflag.float64SliceValue": "float64Slice",
		"*pflag.boolSliceValue":    "boolSlice",
		"*pflag.durationSliceValue": "durationSlice",
		
		// Map types
		"*pflag.stringToStringValue": "stringToString",
		"*pflag.stringToInt64Value":  "stringToInt64",
		
		// Network types
		"*pflag.ipValue":      "ip",
		"*pflag.ipSliceValue": "ipSlice",
		"*pflag.ipMaskValue":  "ipMask",
		"*pflag.ipNetValue":   "ipNet",
		
		// Binary types
		"*pflag.bytesHexValue":    "bytesHex",
		"*pflag.bytesBase64Value": "bytesBase64",
		
		// Special types
		"*pflag.countValue": "count",
	}
	
	if simpleType, ok := typeMap[flagType]; ok {
		return simpleType
	}
	
	return flagType
}
`

// InspectProject generates an inspector program and runs it to get the CLI structure
func InspectProject(projectPath, entrypoint string) (*InspectedCLI, error) {
	// Create inspector with default dependencies
	inspector := NewInspector(Config{
		ProjectPath: projectPath,
		Entrypoint:  entrypoint,
	})

	return inspector.Inspect()
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
