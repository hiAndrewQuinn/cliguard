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
		Use:   "test-cli",
		Short: fmt.Sprintf("Test CLI with %d commands for performance testing", numCommands),
		Long:  "A comprehensive test CLI designed for performance testing and validation.",
		Commands: make([]contract.Command, 0, numCommands),
		Flags: []contract.Flag{
			{
				Name:  "config",
				Type:  "string",
				Usage: "config file (default is $HOME/.test-cli.yaml)",
				Persistent: true,
			},
			{
				Name:  "verbose",
				Type:  "bool",
				Usage: "verbose output",
				Persistent: true,
			},
			{
				Name:  "debug",
				Type:  "bool",
				Usage: "enable debug mode",
				Persistent: true,
			},
			{
				Name:  "log-level",
				Type:  "string",
				Usage: "log level (debug, info, warn, error)",
				Persistent: true,
			},
			{
				Name:  "output",
				Type:  "string",
				Usage: "output format (json, yaml, text)",
				Persistent: true,
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
					Use:   fmt.Sprintf("sub-%d", j),
					Short: fmt.Sprintf("Subcommand %d of command %d", j, i),
					Flags: []contract.Flag{
						{
							Name:  "sub-option",
							Type:  "string",
							Usage: "Subcommand option",
						},
						{
							Name:  "sub-flag",
							Type:  "bool",
							Usage: "Subcommand flag",
						},
					},
				}
				cmd.Commands = append(cmd.Commands, subcmd)
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
		Use:   fmt.Sprintf("command-%03d", index),
		Short: fmt.Sprintf("Command %d for performance testing", index),
		Flags: make([]contract.Flag, 0, flagCount),
	}

	// Add standard flags
	cmd.Flags = append(cmd.Flags, contract.Flag{
		Name:  "input",
		Type:  "string",
		Usage: "Input file path",
	})

	cmd.Flags = append(cmd.Flags, contract.Flag{
		Name:  "output",
		Type:  "string",
		Usage: "Output file path",
	})

	cmd.Flags = append(cmd.Flags, contract.Flag{
		Name:  "force",
		Type:  "bool",
		Usage: "Force operation without confirmation",
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
		Name:  fmt.Sprintf("flag-%d", index),
		Type:  flagType,
		Usage: fmt.Sprintf("%s flag %d", flagType, index),
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
		Use:   "kubectl-like",
		Short: "A kubectl-like CLI for performance testing",
		Long:  "A comprehensive kubectl-like CLI tool designed for performance testing and validation.",
		Flags: []contract.Flag{
			{Name: "kubeconfig", Type: "string", Usage: "Path to kubeconfig file", Persistent: true},
			{Name: "context", Type: "string", Usage: "Kubernetes context", Persistent: true},
			{Name: "namespace", Type: "string", Usage: "Kubernetes namespace", Persistent: true},
			{Name: "output", Type: "string", Usage: "Output format", Persistent: true},
		},
		Commands: []contract.Command{
			{
				Use:   "get",
				Short: "Display resources",
				Flags: []contract.Flag{
					{Name: "all-namespaces", Type: "bool", Usage: "List across all namespaces"},
					{Name: "selector", Type: "string", Usage: "Label selector"},
					{Name: "watch", Type: "bool", Usage: "Watch for changes"},
				},
				Commands: []contract.Command{
					{Use: "pods", Short: "Get pods"},
					{Use: "services", Short: "Get services"},
					{Use: "deployments", Short: "Get deployments"},
					{Use: "configmaps", Short: "Get configmaps"},
					{Use: "secrets", Short: "Get secrets"},
				},
			},
			{
				Use:   "apply",
				Short: "Apply configuration",
				Flags: []contract.Flag{
					{Name: "filename", Type: "string", Usage: "File to apply"},
					{Name: "recursive", Type: "bool", Usage: "Process directory recursively"},
					{Name: "dry-run", Type: "string", Usage: "Dry run mode"},
				},
			},
			{
				Use:   "delete",
				Short: "Delete resources",
				Flags: []contract.Flag{
					{Name: "filename", Type: "string", Usage: "File containing resources"},
					{Name: "force", Type: "bool", Usage: "Force deletion"},
					{Name: "grace-period", Type: "int", Usage: "Grace period in seconds"},
				},
			},
			{
				Use:   "describe",
				Short: "Describe resources",
				Flags: []contract.Flag{
					{Name: "show-events", Type: "bool", Usage: "Show events"},
				},
			},
			{
				Use:   "logs",
				Short: "Print logs",
				Flags: []contract.Flag{
					{Name: "follow", Type: "bool", Usage: "Follow log output"},
					{Name: "tail", Type: "int", Usage: "Number of lines to show"},
					{Name: "since", Type: "string", Usage: "Show logs since duration"},
				},
			},
		},
	}
}

func generateHelmLikeContract() *contract.Contract {
	return &contract.Contract{
		Use:   "helm-like",
		Short: "A Helm-like CLI for performance testing",
		Long:  "A comprehensive Helm-like CLI tool designed for performance testing and validation.",
		Flags: []contract.Flag{
			{Name: "debug", Type: "bool", Usage: "Enable debug output", Persistent: true},
			{Name: "kube-context", Type: "string", Usage: "Kubernetes context", Persistent: true},
			{Name: "namespace", Type: "string", Usage: "Kubernetes namespace", Persistent: true},
		},
		Commands: []contract.Command{
			{
				Use:   "install",
				Short: "Install a chart",
				Flags: []contract.Flag{
					{Name: "values", Type: "stringSlice", Usage: "Values files"},
					{Name: "set", Type: "stringSlice", Usage: "Set values"},
					{Name: "dry-run", Type: "bool", Usage: "Simulate install"},
					{Name: "wait", Type: "bool", Usage: "Wait for deployment"},
					{Name: "timeout", Type: "string", Usage: "Timeout duration"},
				},
			},
			{
				Use:   "upgrade",
				Short: "Upgrade a release",
				Flags: []contract.Flag{
					{Name: "install", Type: "bool", Usage: "Install if not present"},
					{Name: "force", Type: "bool", Usage: "Force resource updates"},
					{Name: "reset-values", Type: "bool", Usage: "Reset values"},
				},
			},
			{
				Use:   "list",
				Short: "List releases",
				Flags: []contract.Flag{
					{Name: "all", Type: "bool", Usage: "Show all releases"},
					{Name: "deployed", Type: "bool", Usage: "Show deployed releases"},
					{Name: "failed", Type: "bool", Usage: "Show failed releases"},
				},
			},
			{
				Use:   "repo",
				Short: "Manage repositories",
				Commands: []contract.Command{
					{
						Use:   "add",
						Short: "Add a repository",
						Flags: []contract.Flag{
							{Name: "username", Type: "string", Usage: "Repository username"},
							{Name: "password", Type: "string", Usage: "Repository password"},
						},
					},
					{Use: "list", Short: "List repositories"},
					{Use: "update", Short: "Update repositories"},
					{Use: "remove", Short: "Remove a repository"},
				},
			},
		},
	}
}

func generateDockerLikeContract() *contract.Contract {
	return &contract.Contract{
		Use:   "docker-like",
		Short: "A Docker-like CLI for performance testing",
		Long:  "A comprehensive Docker-like CLI tool designed for performance testing and validation.",
		Flags: []contract.Flag{
			{Name: "host", Type: "string", Usage: "Docker host", Persistent: true},
			{Name: "tls", Type: "bool", Usage: "Use TLS", Persistent: true},
			{Name: "log-level", Type: "string", Usage: "Log level", Persistent: true},
		},
		Commands: []contract.Command{
			{
				Use:   "run",
				Short: "Run a container",
				Flags: []contract.Flag{
					{Name: "detach", Type: "bool", Usage: "Run in background"},
					{Name: "interactive", Type: "bool", Usage: "Keep STDIN open"},
					{Name: "tty", Type: "bool", Usage: "Allocate pseudo-TTY"},
					{Name: "env", Type: "stringSlice", Usage: "Environment variables"},
					{Name: "volume", Type: "stringSlice", Usage: "Bind mount volumes"},
					{Name: "publish", Type: "stringSlice", Usage: "Publish ports"},
					{Name: "name", Type: "string", Usage: "Container name"},
				},
			},
			{
				Use:   "ps",
				Short: "List containers",
				Flags: []contract.Flag{
					{Name: "all", Type: "bool", Usage: "Show all containers"},
					{Name: "quiet", Type: "bool", Usage: "Only display IDs"},
					{Name: "filter", Type: "stringSlice", Usage: "Filter output"},
				},
			},
			{
				Use:   "images",
				Short: "List images",
				Flags: []contract.Flag{
					{Name: "all", Type: "bool", Usage: "Show all images"},
					{Name: "digests", Type: "bool", Usage: "Show digests"},
					{Name: "quiet", Type: "bool", Usage: "Only show IDs"},
				},
			},
			{
				Use:   "build",
				Short: "Build an image",
				Flags: []contract.Flag{
					{Name: "tag", Type: "stringSlice", Usage: "Name and tag"},
					{Name: "file", Type: "string", Usage: "Dockerfile path"},
					{Name: "no-cache", Type: "bool", Usage: "Do not use cache"},
				},
			},
		},
	}
}
