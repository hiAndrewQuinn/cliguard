package discovery

import (
	"bytes"
	"strings"
	"testing"
)

func TestSelectCandidate(t *testing.T) {
	tests := []struct {
		name           string
		candidates     []EntrypointCandidate
		input          string
		wantErr        bool
		expectedIndex  int
		expectedOutput []string
	}{
		{
			name: "single candidate returns immediately",
			candidates: []EntrypointCandidate{
				{
					FilePath:   "cmd/root.go",
					Framework:  "cobra",
					Confidence: 95,
				},
			},
			input:         "",
			wantErr:       false,
			expectedIndex: 0,
		},
		{
			name: "user selects first option",
			candidates: []EntrypointCandidate{
				{
					FilePath:    "cmd/root.go",
					Framework:   "cobra",
					Confidence:  95,
					LineNumber:  10,
					Pattern:     "Function returning root cobra.Command",
					PackagePath: "github.com/test/project/cmd",
				},
				{
					FilePath:    "main.go",
					Framework:   "flag",
					Confidence:  70,
					LineNumber:  20,
					Pattern:     "Flag parsing call",
					PackagePath: "github.com/test/project",
				},
			},
			input:         "1\n",
			wantErr:       false,
			expectedIndex: 0,
			expectedOutput: []string{
				"Multiple entrypoints found. Please select one:",
				"1. cobra (confidence: 95%)",
				"File: cmd/root.go:10",
				"2. flag (confidence: 70%)",
				"File: main.go:20",
				"Enter selection (1-2) or 'q' to quit:",
			},
		},
		{
			name: "user quits",
			candidates: []EntrypointCandidate{
				{FilePath: "cmd/root.go", Framework: "cobra", Confidence: 95},
				{FilePath: "main.go", Framework: "flag", Confidence: 70},
			},
			input:   "q\n",
			wantErr: true,
			expectedOutput: []string{
				"Enter selection (1-2) or 'q' to quit:",
			},
		},
		{
			name: "invalid input then valid",
			candidates: []EntrypointCandidate{
				{FilePath: "cmd/root.go", Framework: "cobra", Confidence: 95},
				{FilePath: "main.go", Framework: "flag", Confidence: 70},
			},
			input:         "invalid\n3\n2\n",
			wantErr:       false,
			expectedIndex: 1,
			expectedOutput: []string{
				"Invalid input. Please enter a number between 1 and 2.",
				"Invalid selection. Please enter a number between 1 and 2.",
			},
		},
		{
			name:       "empty candidates returns error",
			candidates: []EntrypointCandidate{},
			input:      "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}

			selector := NewInteractiveSelector(input, output)

			result, err := selector.SelectCandidate(tt.candidates)

			if (err != nil) != tt.wantErr {
				t.Errorf("SelectCandidate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil {
				// Find index of result
				resultIndex := -1
				for i, c := range tt.candidates {
					if c.FilePath == result.FilePath && c.Framework == result.Framework {
						resultIndex = i
						break
					}
				}

				if resultIndex != tt.expectedIndex {
					t.Errorf("Expected candidate at index %d, got index %d",
						tt.expectedIndex, resultIndex)
				}
			}

			// Check output contains expected strings
			outputStr := output.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nFull output:\n%s",
						expected, outputStr)
				}
			}
		})
	}
}

func TestFormatSelectedEntrypoint(t *testing.T) {
	tests := []struct {
		name      string
		candidate EntrypointCandidate
		expected  string
	}{
		{
			name: "NewRootCmd function",
			candidate: EntrypointCandidate{
				FunctionSignature: "func NewRootCmd() *cobra.Command",
				PackagePath:       "github.com/test/project/cmd",
			},
			expected: "--entrypoint github.com/test/project/cmd.NewRootCmd",
		},
		{
			name: "no function signature",
			candidate: EntrypointCandidate{
				PackagePath: "github.com/test/project",
			},
			expected: "--entrypoint github.com/test/project",
		},
		{
			name: "different function name",
			candidate: EntrypointCandidate{
				FunctionSignature: "func Execute()",
				PackagePath:       "github.com/test/project/cmd",
			},
			expected: "--entrypoint github.com/test/project/cmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSelectedEntrypoint(&tt.candidate)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
