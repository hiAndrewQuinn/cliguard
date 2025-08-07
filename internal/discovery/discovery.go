package discovery

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hiAndrewQuinn/cliguard/internal/filesystem"
)

// Discoverer finds CLI entrypoints in Go projects
type Discoverer struct {
	fs          filesystem.FileSystem
	projectPath string
	patterns    []Pattern
}

// NewDiscoverer creates a new entrypoint discoverer
func NewDiscoverer(projectPath string, fs filesystem.FileSystem) *Discoverer {
	if fs == nil {
		fs = &filesystem.OSFileSystem{}
	}
	return &Discoverer{
		fs:          fs,
		projectPath: projectPath,
		patterns:    GetCLIPatterns(),
	}
}

// DiscoverEntrypoints finds potential CLI entrypoints in the project
func (d *Discoverer) DiscoverEntrypoints() ([]EntrypointCandidate, error) {
	var candidates []EntrypointCandidate

	// First, find all Go files in the project
	goFiles, err := d.findGoFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to find Go files: %w", err)
	}

	// Debug: print found files
	// fmt.Printf("Found %d Go files\n", len(goFiles))

	// Check each file for patterns
	for _, file := range goFiles {
		fileCandidates, err := d.analyzeFile(file)
		if err != nil {
			// Continue with other files even if one fails
			continue
		}
		candidates = append(candidates, fileCandidates...)
	}

	// Sort by confidence (highest first) and prioritize non-test directories
	sort.Slice(candidates, func(i, j int) bool {
		// Check if files are in test directories
		iIsTest := strings.Contains(candidates[i].FilePath, "test") || 
			strings.Contains(candidates[i].FilePath, "fixtures") ||
			strings.Contains(candidates[i].FilePath, "test-suite")
		jIsTest := strings.Contains(candidates[j].FilePath, "test") || 
			strings.Contains(candidates[j].FilePath, "fixtures") ||
			strings.Contains(candidates[j].FilePath, "test-suite")
		
		// Prioritize non-test files
		if iIsTest != jIsTest {
			return !iIsTest // non-test files come first
		}
		
		// Then sort by confidence
		if candidates[i].Confidence != candidates[j].Confidence {
			return candidates[i].Confidence > candidates[j].Confidence
		}
		
		// Finally, prefer files in cmd/ directory
		iInCmd := strings.Contains(candidates[i].FilePath, "/cmd/")
		jInCmd := strings.Contains(candidates[j].FilePath, "/cmd/")
		if iInCmd != jInCmd {
			return iInCmd
		}
		
		return candidates[i].FilePath < candidates[j].FilePath
	})

	return candidates, nil
}

// findGoFiles finds all Go files in the project
func (d *Discoverer) findGoFiles() ([]string, error) {
	var goFiles []string

	// Debug project path
	// fmt.Printf("Walking project path: %s\n", d.projectPath)

	err := filepath.Walk(d.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and hidden directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		// Debug: print each file
		// if !info.IsDir() {
		// 	fmt.Printf("Checking file: %s\n", path)
		// }

		// Only process .go files
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		relPath, err := filepath.Rel(d.projectPath, path)
		if err != nil {
			return nil
		}

		goFiles = append(goFiles, relPath)
		return nil
	})

	return goFiles, err
}

// analyzeFile analyzes a single Go file for entrypoint patterns
func (d *Discoverer) analyzeFile(filePath string) ([]EntrypointCandidate, error) {
	var candidates []EntrypointCandidate

	absPath := filepath.Join(d.projectPath, filePath)
	content, err := d.fs.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// Parse the file to get package information
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, absPath, content, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	// Check imports to determine which patterns to apply
	imports := d.extractImports(node)
	applicablePatterns := d.getApplicablePatterns(imports)

	if len(applicablePatterns) == 0 {
		return candidates, nil
	}

	// Get the module path
	modulePath, err := d.getModulePath()
	if err != nil {
		modulePath = ""
	}

	// Calculate package path
	packagePath := d.calculatePackagePath(modulePath, filePath)

	// Scan file line by line for patterns
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		for _, pattern := range applicablePatterns {
			for _, codePattern := range pattern.CodePatterns {
				matched, err := regexp.MatchString(codePattern.Pattern, line)
				if err != nil {
					continue
				}

				if matched {
					candidate := EntrypointCandidate{
						FilePath:    filePath,
						LineNumber:  lineNumber,
						Line:        strings.TrimSpace(line),
						Framework:   pattern.Name,
						Pattern:     codePattern.Description,
						Confidence:  codePattern.Confidence,
						PackagePath: packagePath,
					}

					// Try to extract function signature
					if funcSig := d.extractFunctionSignature(string(content), lineNumber); funcSig != "" {
						candidate.FunctionSignature = funcSig
					}

					candidates = append(candidates, candidate)
				}
			}
		}
	}

	return candidates, nil
}

// extractImports extracts import paths from the parsed file
func (d *Discoverer) extractImports(node *ast.File) []string {
	var imports []string
	for _, imp := range node.Imports {
		if imp.Path != nil {
			// Remove quotes from import path
			importPath := strings.Trim(imp.Path.Value, `"`)
			imports = append(imports, importPath)
		}
	}
	return imports
}

// getApplicablePatterns returns patterns that match the file's imports
func (d *Discoverer) getApplicablePatterns(imports []string) []Pattern {
	var applicable []Pattern
	seen := make(map[string]bool) // Track which patterns we've already added

	for _, pattern := range d.patterns {
		if seen[pattern.Name] {
			continue
		}
		
		for _, requiredImport := range pattern.Imports {
			found := false
			for _, fileImport := range imports {
				if strings.Contains(fileImport, requiredImport) {
					applicable = append(applicable, pattern)
					seen[pattern.Name] = true
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	return applicable
}

// getModulePath reads the go.mod file to get the module path
func (d *Discoverer) getModulePath() (string, error) {
	goModPath := filepath.Join(d.projectPath, "go.mod")
	content, err := d.fs.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("module path not found in go.mod")
}

// calculatePackagePath calculates the full package path for a file
func (d *Discoverer) calculatePackagePath(modulePath, filePath string) string {
	if modulePath == "" {
		return ""
	}

	dir := filepath.Dir(filePath)
	if dir == "." {
		return modulePath
	}

	return modulePath + "/" + strings.ReplaceAll(dir, string(filepath.Separator), "/")
}

// extractFunctionSignature tries to extract the function signature containing the line
func (d *Discoverer) extractFunctionSignature(content string, lineNumber int) string {
	lines := strings.Split(content, "\n")
	if lineNumber <= 0 || lineNumber > len(lines) {
		return ""
	}

	// Compile regex once
	funcStartRegex := regexp.MustCompile(`^func\s+`)
	
	// Look backwards for function declaration
	for i := lineNumber - 1; i >= 0 && i >= lineNumber-10; i-- {
		line := lines[i]
		if funcStartRegex.MatchString(strings.TrimSpace(line)) {
			// Extract function signature
			funcRegex := regexp.MustCompile(`func\s+(\w+)?\s*\([^)]*\)\s*([^{]*)`)
			if matches := funcRegex.FindStringSubmatch(line); len(matches) > 0 {
				return strings.TrimSpace(matches[0])
			}
		}
	}

	return ""
}

// PrintCandidates prints the discovered candidates in a user-friendly format
func PrintCandidates(w io.Writer, candidates []EntrypointCandidate, force bool) {
	if len(candidates) == 0 {
		fmt.Fprintln(w, "No CLI entrypoints found.")
		fmt.Fprintln(w, "Try specifying the entrypoint manually with --entrypoint flag.")
		return
	}

	fmt.Fprintf(w, "Found %d potential CLI entrypoint(s):\n\n", len(candidates))

	for i, candidate := range candidates {
		fmt.Fprintf(w, "%d. %s (confidence: %d%%)\n", i+1, candidate.Framework, candidate.Confidence)
		fmt.Fprintf(w, "   File: %s:%d\n", candidate.FilePath, candidate.LineNumber)
		fmt.Fprintf(w, "   Pattern: %s\n", candidate.Pattern)
		fmt.Fprintf(w, "   Code: %s\n", candidate.Line)
		
		if candidate.FunctionSignature != "" {
			fmt.Fprintf(w, "   Function: %s\n", candidate.FunctionSignature)
		}
		
		if candidate.PackagePath != "" {
			fmt.Fprintf(w, "   Package: %s\n", candidate.PackagePath)
		}
		
		// Add warning for non-Cobra frameworks
		if candidate.Framework != "cobra" {
			fmt.Fprintf(w, "\n   ⚠️  Note: cliguard currently only generates and validates Cobra CLIs.\n")
			fmt.Fprintf(w, "   Support for %s is coming soon!", candidate.Framework)
			if force {
				fmt.Fprintf(w, "\n   (Use --force flag with generate/validate to proceed anyway)")
			}
		}
		
		fmt.Fprintln(w)
	}

	// Suggest the most likely entrypoint
	if candidates[0].Confidence >= 85 {
		fmt.Fprintln(w, "Suggested entrypoint:")
		if candidates[0].FunctionSignature != "" && strings.Contains(candidates[0].FunctionSignature, "NewRootCmd") {
			fmt.Fprintf(w, "  --entrypoint %s.NewRootCmd\n", candidates[0].PackagePath)
		} else {
			fmt.Fprintf(w, "  --entrypoint %s\n", candidates[0].PackagePath)
		}
		
		// Add warning if suggested entrypoint is not Cobra
		if candidates[0].Framework != "cobra" {
			fmt.Fprintf(w, "\n⚠️  Note: This %s entrypoint is not currently supported by cliguard.\n", candidates[0].Framework)
			if force {
				fmt.Fprintln(w, "Use --force flag with generate/validate to proceed anyway.")
			}
		}
	}
}