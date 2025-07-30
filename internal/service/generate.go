package service

import (
	"fmt"

	"github.com/hiAndrewQuinn/cliguard/internal/contract"
	"github.com/hiAndrewQuinn/cliguard/internal/inspector"
	"gopkg.in/yaml.v3"
)

// GenerateOptions contains options for the generate command
type GenerateOptions struct {
	ProjectPath string
	Entrypoint  string
}

// GenerateService handles the generation of contract files
type GenerateService struct{}

// NewGenerateService creates a new GenerateService
func NewGenerateService() *GenerateService {
	return &GenerateService{}
}

// Generate inspects a CLI and generates a contract YAML string
func (s *GenerateService) Generate(opts GenerateOptions) (string, error) {
	// Inspect the project to get the CLI structure
	inspectedCLI, err := inspector.InspectProject(opts.ProjectPath, opts.Entrypoint)
	if err != nil {
		return "", fmt.Errorf("failed to inspect project: %w", err)
	}

	// Convert inspected CLI to contract
	contract := s.inspectedToContract(inspectedCLI)

	// Marshal contract to YAML
	yamlData, err := yaml.Marshal(contract)
	if err != nil {
		return "", fmt.Errorf("failed to marshal contract to YAML: %w", err)
	}

	// Add comment header
	header := `# Cliguard contract file
# To use this contract, pipe this output to a file:
#   cliguard generate --project-path . > cliguard.yaml
#
`
	return header + string(yamlData), nil
}

// inspectedToContract converts an InspectedCLI to a Contract
func (s *GenerateService) inspectedToContract(inspected *inspector.InspectedCLI) *contract.Contract {
	return &contract.Contract{
		Use:      inspected.Use,
		Short:    inspected.Short,
		Long:     inspected.Long,
		Flags:    s.inspectedFlagsToContractFlags(inspected.Flags),
		Commands: s.inspectedCommandsToContractCommands(inspected.Commands),
	}
}

// inspectedFlagsToContractFlags converts InspectedFlag slice to Flag slice
func (s *GenerateService) inspectedFlagsToContractFlags(flags []inspector.InspectedFlag) []contract.Flag {
	var contractFlags []contract.Flag
	for _, f := range flags {
		contractFlags = append(contractFlags, contract.Flag{
			Name:       f.Name,
			Shorthand:  f.Shorthand,
			Usage:      f.Usage,
			Type:       f.Type,
			Persistent: f.Persistent,
		})
	}
	return contractFlags
}

// inspectedCommandsToContractCommands converts InspectedCommand slice to Command slice
func (s *GenerateService) inspectedCommandsToContractCommands(commands []inspector.InspectedCommand) []contract.Command {
	var contractCommands []contract.Command
	for _, cmd := range commands {
		contractCommands = append(contractCommands, s.inspectedCommandToContractCommand(cmd))
	}
	return contractCommands
}

// inspectedCommandToContractCommand converts a single InspectedCommand to Command
func (s *GenerateService) inspectedCommandToContractCommand(cmd inspector.InspectedCommand) contract.Command {
	return contract.Command{
		Use:      cmd.Use,
		Short:    cmd.Short,
		Long:     cmd.Long,
		Flags:    s.inspectedFlagsToContractFlags(cmd.Flags),
		Commands: s.inspectedCommandsToContractCommands(cmd.Commands),
	}
}
