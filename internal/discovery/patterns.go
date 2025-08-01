package discovery

// Pattern represents a CLI framework pattern to search for
type Pattern struct {
	Name        string
	Description string
	// Import paths to look for in go files
	Imports []string
	// Code patterns to search for
	CodePatterns []CodePattern
	// Common file paths where entrypoints might be found
	FilePaths []string
}

// CodePattern represents a code pattern to search for
type CodePattern struct {
	// Pattern to search for (regex)
	Pattern string
	// Description of what this pattern finds
	Description string
	// Confidence level (0-100) that this indicates an entrypoint
	Confidence int
}

// GetCLIPatterns returns patterns for various CLI frameworks
func GetCLIPatterns() []Pattern {
	return []Pattern{
		{
			Name:        "cobra",
			Description: "Cobra CLI framework",
			Imports:     []string{"github.com/spf13/cobra"},
			CodePatterns: []CodePattern{
				{
					Pattern:     `func\s+NewRootCmd\s*\(\s*\)\s*\*cobra\.Command`,
					Description: "Function returning root cobra.Command",
					Confidence:  95,
				},
				{
					Pattern:     `rootCmd\s*:=\s*&cobra\.Command`,
					Description: "Root command initialization",
					Confidence:  90,
				},
				{
					Pattern:     `Execute\s*\(\s*\)`,
					Description: "Cobra Execute function",
					Confidence:  85,
				},
			},
			FilePaths: []string{
				"cmd/root.go",
				"cmd/*.go",
				"internal/cmd/*.go",
				"main.go",
			},
		},
		{
			Name:        "urfave/cli",
			Description: "urfave/cli framework",
			Imports:     []string{"github.com/urfave/cli/v2", "github.com/urfave/cli"},
			CodePatterns: []CodePattern{
				{
					Pattern:     `app\s*:=\s*&cli\.App`,
					Description: "CLI app initialization",
					Confidence:  90,
				},
				{
					Pattern:     `cli\.NewApp\s*\(\s*\)`,
					Description: "New CLI app creation",
					Confidence:  90,
				},
				{
					Pattern:     `app\.Run\s*\(`,
					Description: "CLI app run call",
					Confidence:  85,
				},
			},
			FilePaths: []string{
				"main.go",
				"cmd/main.go",
				"app/*.go",
			},
		},
		{
			Name:        "flag",
			Description: "Standard library flag package",
			Imports:     []string{"flag"},
			CodePatterns: []CodePattern{
				{
					Pattern:     `flag\.Parse\s*\(\s*\)`,
					Description: "Flag parsing call",
					Confidence:  70,
				},
				{
					Pattern:     `flag\.(String|Int|Bool|Float64)\s*\(`,
					Description: "Flag definition",
					Confidence:  60,
				},
			},
			FilePaths: []string{
				"main.go",
			},
		},
		{
			Name:        "kingpin",
			Description: "Kingpin CLI framework",
			Imports:     []string{"gopkg.in/alecthomas/kingpin.v2", "github.com/alecthomas/kingpin"},
			CodePatterns: []CodePattern{
				{
					Pattern:     `kingpin\.(New|Application)\s*\(`,
					Description: "Kingpin app creation",
					Confidence:  90,
				},
				{
					Pattern:     `app\s*:=\s*kingpin\.`,
					Description: "Kingpin app initialization",
					Confidence:  85,
				},
			},
			FilePaths: []string{
				"main.go",
				"cmd/*.go",
			},
		},
	}
}

// EntrypointCandidate represents a potential entrypoint found in the code
type EntrypointCandidate struct {
	// File path where the candidate was found
	FilePath string
	// Line number where the pattern was found
	LineNumber int
	// The actual line of code
	Line string
	// Framework detected (cobra, urfave/cli, etc.)
	Framework string
	// Pattern that matched
	Pattern string
	// Confidence score (0-100)
	Confidence int
	// Full function signature if detected
	FunctionSignature string
	// Package path (e.g., github.com/user/repo/cmd)
	PackagePath string
}