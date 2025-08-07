package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"gopkg.in/yaml.v3"
)

// GenerateContract creates a contract YAML file for the specified CLI size
func GenerateContract(dir string, size CLISize) error {
	numCommands := getCommandCount(size)

	c := &contract.Contract{
		Name:        "test-cli",
		Version:     "1.0.0",
		Description: fmt.Sprintf("Test CLI with %d commands for performance testing", numCommands),
		Commands:    make([]contract.Command, 0, numCommands),
		GlobalFlags: []contract.Flag{
			{
				Name:        "config",
				Type:        "string",
				Description: "config file (default is $HOME/.test-cli.yaml)",
			},
			{
				Name:        "verbose",
				Type:        "bool",
				Description: "verbose output",
			},
			{
				Name:        "debug",
				Type:        "bool",
				Description: "enable debug mode",
			},
			{
				Name:        "log-level",
				Type:        "string",
				Description: "log level (debug, info, warn, error)",
			},
			{
				Name:        "output",
				Type:        "string",
				Description: "output format (json, yaml, text)",
			},
		},
	}

	// Generate commands
	for i := 0; i < numCommands; i++ {
		cmd := generateContractCommand(i, size == Large)

		// Add subcommands for every 10th command
		if i%10 == 0 && i > 0 {
			for j := 0; j < 5; j++ {
				subcmd := contract.Command{
					Name:        fmt.Sprintf("sub-%d", j),
					Description: fmt.Sprintf("Subcommand %d of command %d", j, i),
					Flags: []contract.Flag{
						{
							Name:        "sub-option",
							Type:        "string",
							Description: "Subcommand option",
						},
						{
							Name:        "sub-flag",
							Type:        "bool",
							Description: "Subcommand flag",
						},
					},
				}
				cmd.Subcommands = append(cmd.Subcommands, subcmd)
			}
		}

		c.Commands = append(c.Commands, cmd)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal contract: %w", err)
	}

	contractPath := filepath.Join(dir, "contract.yaml")
	if err := os.WriteFile(contractPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write contract file: %w", err)
	}

	return nil
}

func generateContractCommand(index int, addMoreFlags bool) contract.Command {
	flagCount := 5
	if addMoreFlags {
		flagCount = 15
	}

	cmd := contract.Command{
		Name:        fmt.Sprintf("command-%03d", index),
		Description: fmt.Sprintf("Command %d for performance testing", index),
		Flags:       make([]contract.Flag, 0, flagCount),
	}

	// Add standard flags
	cmd.Flags = append(cmd.Flags, contract.Flag{
		Name:        "input",
		Type:        "string",
		Description: "Input file path",
		Required:    true,
	})

	cmd.Flags = append(cmd.Flags, contract.Flag{
		Name:        "output",
		Type:        "string",
		Description: "Output file path",
	})

	cmd.Flags = append(cmd.Flags, contract.Flag{
		Name:        "force",
		Type:        "bool",
		Description: "Force operation without confirmation",
	})

	// Add additional flags
	for i := 3; i < flagCount; i++ {
		flag := generateContractFlag(i)
		cmd.Flags = append(cmd.Flags, flag)
	}

	return cmd
}

func generateContractFlag(index int) contract.Flag {
	flagTypes := []string{"string", "bool", "int", "float64", "stringSlice"}
	flagType := flagTypes[index%len(flagTypes)]

	flag := contract.Flag{
		Name:        fmt.Sprintf("flag-%d", index),
		Type:        flagType,
		Description: fmt.Sprintf("%s flag %d", flagType, index),
	}

	// Add validation for some flags
	if index%7 == 0 && flagType == "int" {
		min := 0
		max := 100
		flag.Validation = &contract.Validation{
			Min:     &min,
			Max:     &max,
			Message: "Value must be between 0 and 100",
		}
	} else if index%5 == 0 && flagType == "string" {
		flag.Validation = &contract.Validation{
			Pattern: "^[a-zA-Z0-9-]+$",
			Message: "Must contain only alphanumeric characters and hyphens",
		}
	}

	// Add default values for some flags
	if index%3 == 0 {
		switch flagType {
		case "string":
			defaultVal := fmt.Sprintf("default-%d", index)
			flag.Default = &defaultVal
		case "int":
			defaultVal := index * 10
			flag.Default = &defaultVal
		case "bool":
			defaultVal := false
			flag.Default = &defaultVal
		}
	}

	return flag
}

// GenerateRealWorldContract creates contracts that simulate real-world CLIs
func GenerateRealWorldContract(dir string, cliType string) error {
	var c *contract.Contract

	switch cliType {
	case "kubectl":
		c = generateKubectlLikeContract()
	case "helm":
		c = generateHelmLikeContract()
	case "docker":
		c = generateDockerLikeContract()
	default:
		return fmt.Errorf("unknown CLI type: %s", cliType)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal contract: %w", err)
	}

	contractPath := filepath.Join(dir, fmt.Sprintf("%s-contract.yaml", cliType))
	if err := os.WriteFile(contractPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write contract file: %w", err)
	}

	return nil
}

func generateKubectlLikeContract() *contract.Contract {
	return &contract.Contract{
		Name:        "kubectl-like",
		Version:     "1.0.0",
		Description: "A kubectl-like CLI for performance testing",
		GlobalFlags: []contract.Flag{
			{Name: "kubeconfig", Type: "string", Description: "Path to kubeconfig file"},
			{Name: "context", Type: "string", Description: "Kubernetes context"},
			{Name: "namespace", Type: "string", Description: "Kubernetes namespace"},
			{Name: "output", Type: "string", Description: "Output format"},
		},
		Commands: []contract.Command{
			{
				Name:        "get",
				Description: "Display resources",
				Flags: []contract.Flag{
					{Name: "all-namespaces", Type: "bool", Description: "List across all namespaces"},
					{Name: "selector", Type: "string", Description: "Label selector"},
					{Name: "watch", Type: "bool", Description: "Watch for changes"},
				},
				Subcommands: []contract.Command{
					{Name: "pods", Description: "Get pods"},
					{Name: "services", Description: "Get services"},
					{Name: "deployments", Description: "Get deployments"},
					{Name: "configmaps", Description: "Get configmaps"},
					{Name: "secrets", Description: "Get secrets"},
				},
			},
			{
				Name:        "apply",
				Description: "Apply configuration",
				Flags: []contract.Flag{
					{Name: "filename", Type: "string", Description: "File to apply", Required: true},
					{Name: "recursive", Type: "bool", Description: "Process directory recursively"},
					{Name: "dry-run", Type: "string", Description: "Dry run mode"},
				},
			},
			{
				Name:        "delete",
				Description: "Delete resources",
				Flags: []contract.Flag{
					{Name: "filename", Type: "string", Description: "File containing resources"},
					{Name: "force", Type: "bool", Description: "Force deletion"},
					{Name: "grace-period", Type: "int", Description: "Grace period in seconds"},
				},
			},
			{
				Name:        "describe",
				Description: "Describe resources",
				Flags: []contract.Flag{
					{Name: "show-events", Type: "bool", Description: "Show events"},
				},
			},
			{
				Name:        "logs",
				Description: "Print logs",
				Flags: []contract.Flag{
					{Name: "follow", Type: "bool", Description: "Follow log output"},
					{Name: "tail", Type: "int", Description: "Number of lines to show"},
					{Name: "since", Type: "string", Description: "Show logs since duration"},
				},
			},
		},
	}
}

func generateHelmLikeContract() *contract.Contract {
	return &contract.Contract{
		Name:        "helm-like",
		Version:     "3.0.0",
		Description: "A Helm-like CLI for performance testing",
		GlobalFlags: []contract.Flag{
			{Name: "debug", Type: "bool", Description: "Enable debug output"},
			{Name: "kube-context", Type: "string", Description: "Kubernetes context"},
			{Name: "namespace", Type: "string", Description: "Kubernetes namespace"},
		},
		Commands: []contract.Command{
			{
				Name:        "install",
				Description: "Install a chart",
				Flags: []contract.Flag{
					{Name: "values", Type: "stringSlice", Description: "Values files"},
					{Name: "set", Type: "stringSlice", Description: "Set values"},
					{Name: "dry-run", Type: "bool", Description: "Simulate install"},
					{Name: "wait", Type: "bool", Description: "Wait for deployment"},
					{Name: "timeout", Type: "string", Description: "Timeout duration"},
				},
			},
			{
				Name:        "upgrade",
				Description: "Upgrade a release",
				Flags: []contract.Flag{
					{Name: "install", Type: "bool", Description: "Install if not present"},
					{Name: "force", Type: "bool", Description: "Force resource updates"},
					{Name: "reset-values", Type: "bool", Description: "Reset values"},
				},
			},
			{
				Name:        "list",
				Description: "List releases",
				Flags: []contract.Flag{
					{Name: "all", Type: "bool", Description: "Show all releases"},
					{Name: "deployed", Type: "bool", Description: "Show deployed releases"},
					{Name: "failed", Type: "bool", Description: "Show failed releases"},
				},
			},
			{
				Name:        "repo",
				Description: "Manage repositories",
				Subcommands: []contract.Command{
					{
						Name:        "add",
						Description: "Add a repository",
						Flags: []contract.Flag{
							{Name: "username", Type: "string", Description: "Repository username"},
							{Name: "password", Type: "string", Description: "Repository password"},
						},
					},
					{Name: "list", Description: "List repositories"},
					{Name: "update", Description: "Update repositories"},
					{Name: "remove", Description: "Remove a repository"},
				},
			},
		},
	}
}

func generateDockerLikeContract() *contract.Contract {
	return &contract.Contract{
		Name:        "docker-like",
		Version:     "20.0.0",
		Description: "A Docker-like CLI for performance testing",
		GlobalFlags: []contract.Flag{
			{Name: "host", Type: "string", Description: "Docker host"},
			{Name: "tls", Type: "bool", Description: "Use TLS"},
			{Name: "log-level", Type: "string", Description: "Log level"},
		},
		Commands: []contract.Command{
			{
				Name:        "run",
				Description: "Run a container",
				Flags: []contract.Flag{
					{Name: "detach", Type: "bool", Description: "Run in background"},
					{Name: "interactive", Type: "bool", Description: "Keep STDIN open"},
					{Name: "tty", Type: "bool", Description: "Allocate pseudo-TTY"},
					{Name: "env", Type: "stringSlice", Description: "Environment variables"},
					{Name: "volume", Type: "stringSlice", Description: "Bind mount volumes"},
					{Name: "publish", Type: "stringSlice", Description: "Publish ports"},
					{Name: "name", Type: "string", Description: "Container name"},
				},
			},
			{
				Name:        "ps",
				Description: "List containers",
				Flags: []contract.Flag{
					{Name: "all", Type: "bool", Description: "Show all containers"},
					{Name: "quiet", Type: "bool", Description: "Only display IDs"},
					{Name: "filter", Type: "stringSlice", Description: "Filter output"},
				},
			},
			{
				Name:        "images",
				Description: "List images",
				Flags: []contract.Flag{
					{Name: "all", Type: "bool", Description: "Show all images"},
					{Name: "digests", Type: "bool", Description: "Show digests"},
					{Name: "quiet", Type: "bool", Description: "Only show IDs"},
				},
			},
			{
				Name:        "build",
				Description: "Build an image",
				Flags: []contract.Flag{
					{Name: "tag", Type: "stringSlice", Description: "Name and tag"},
					{Name: "file", Type: "string", Description: "Dockerfile path"},
					{Name: "no-cache", Type: "bool", Description: "Do not use cache"},
				},
			},
		},
	}
}
