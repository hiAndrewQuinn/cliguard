// Package service provides high-level business logic for cliguard operations.
// It acts as an orchestration layer that coordinates between the contract,
// inspector, and validator packages.
//
// The service package implements the primary operations that users interact
// with through the CLI or when using cliguard as a library. It handles the
// complete workflow of loading contracts, inspecting CLIs, and validating them.
//
// # Available Services
//
// ValidateService:
// Orchestrates the validation process by loading contracts, inspecting CLIs,
// and running validation:
//
//	svc := service.NewValidateService()
//	result, err := svc.Validate(service.ValidateOptions{
//	    ContractPath: "cliguard.yaml",
//	    ProjectPath:  ".",
//	    Constructor:  "cmd.NewRootCmd",
//	})
//	if err != nil {
//	    return err
//	}
//	if !result.IsValid {
//	    fmt.Println(result.Report)
//	}
//
// GenerateService:
// Generates contract files from existing CLI implementations:
//
//	svc := service.NewGenerateService()
//	err := svc.Generate(service.GenerateOptions{
//	    ProjectPath:  ".",
//	    Constructor:  "cmd.NewRootCmd",
//	    OutputPath:   "cliguard.yaml",
//	})
//
// # Service Configuration
//
// Services can be configured with custom implementations:
//
//	svc := &ValidateService{
//	    ContractLoader: myCustomLoader,
//	    Inspector:      myCustomInspector,
//	}
//
// # Error Handling
//
// Services provide rich error information with context:
//   - File I/O errors when loading contracts
//   - Build errors when inspecting projects
//   - Validation errors with detailed reports
//   - Configuration errors for invalid options
//
// # Testing Support
//
// Services are designed with testing in mind. All dependencies can be
// injected, making it easy to mock external interactions:
//
//	svc := &ValidateService{
//	    ContractLoader: func(path string) (*contract.Contract, error) {
//	        return &contract.Contract{Use: "test"}, nil
//	    },
//	    Inspector: func(path, constructor string) (*inspector.InspectedCLI, error) {
//	        return &inspector.InspectedCLI{Use: "test"}, nil
//	    },
//	}
package service
