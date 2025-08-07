package discovery

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// InteractiveSelector allows users to select from multiple entrypoint candidates
type InteractiveSelector struct {
	input  io.Reader
	output io.Writer
}

// NewInteractiveSelector creates a new interactive selector
func NewInteractiveSelector(input io.Reader, output io.Writer) *InteractiveSelector {
	return &InteractiveSelector{
		input:  input,
		output: output,
	}
}

// SelectCandidate prompts the user to select from multiple candidates
func (s *InteractiveSelector) SelectCandidate(candidates []EntrypointCandidate) (*EntrypointCandidate, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates to select from")
	}

	if len(candidates) == 1 {
		return &candidates[0], nil
	}

	// Display candidates
	fmt.Fprintln(s.output, "\nMultiple entrypoints found. Please select one:")
	fmt.Fprintln(s.output)

	for i, candidate := range candidates {
		fmt.Fprintf(s.output, "%d. %s (confidence: %d%%)\n", i+1, candidate.Framework, candidate.Confidence)
		fmt.Fprintf(s.output, "   File: %s:%d\n", candidate.FilePath, candidate.LineNumber)
		fmt.Fprintf(s.output, "   Pattern: %s\n", candidate.Pattern)
		if candidate.FunctionSignature != "" {
			fmt.Fprintf(s.output, "   Function: %s\n", candidate.FunctionSignature)
		}
		if candidate.PackagePath != "" {
			fmt.Fprintf(s.output, "   Package: %s\n", candidate.PackagePath)
		}
		fmt.Fprintln(s.output)
	}

	// Get user selection
	reader := bufio.NewReader(s.input)

	for {
		fmt.Fprintf(s.output, "Enter selection (1-%d) or 'q' to quit: ", len(candidates))

		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		// Check for quit
		if strings.ToLower(input) == "q" {
			return nil, fmt.Errorf("selection cancelled by user")
		}

		// Try to parse as number
		num, err := strconv.Atoi(input)
		if err != nil {
			fmt.Fprintf(s.output, "Invalid input. Please enter a number between 1 and %d.\n", len(candidates))
			continue
		}

		// Check bounds
		if num < 1 || num > len(candidates) {
			fmt.Fprintf(s.output, "Invalid selection. Please enter a number between 1 and %d.\n", len(candidates))
			continue
		}

		return &candidates[num-1], nil
	}
}

// FormatSelectedEntrypoint formats the selected entrypoint for display
func FormatSelectedEntrypoint(candidate *EntrypointCandidate) string {
	entrypoint := candidate.PackagePath
	
	// For Cobra CLIs, try to determine the correct function name
	if candidate.Framework == "cobra" {
		functionName := determineCObraFunctionName(*candidate)
		if functionName != "" {
			entrypoint = candidate.PackagePath + "." + functionName
		}
	}
	
	return fmt.Sprintf("--entrypoint %s", entrypoint)
}
