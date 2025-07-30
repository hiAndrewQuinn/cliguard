package inspector

import (
	"fmt"
	"strings"
	"testing"
	"text/template"
)

func TestGetModuleName(t *testing.T) {
	tests := []struct {
		name         string
		goModContent string
		want         string
	}{
		{
			name: "simple_module",
			goModContent: `module github.com/example/project

go 1.21
`,
			want: "github.com/example/project",
		},
		{
			name: "module_with_dependencies",
			goModContent: `module github.com/test/app

go 1.21

require (
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.16.0
)
`,
			want: "github.com/test/app",
		},
		{
			name: "module_with_comments",
			goModContent: `// This is a comment
module github.com/myorg/myapp

go 1.21
`,
			want: "github.com/myorg/myapp",
		},
		{
			name: "module_with_leading_whitespace",
			goModContent: `   module   github.com/space/project   

go 1.21
`,
			want: "github.com/space/project",
		},
		{
			name:         "empty_go_mod",
			goModContent: ``,
			want:         "",
		},
		{
			name: "no_module_directive",
			goModContent: `go 1.21

require (
	github.com/spf13/cobra v1.7.0
)
`,
			want: "",
		},
		{
			name: "malformed_module_directive",
			goModContent: `module

go 1.21
`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getModuleName([]byte(tt.goModContent))
			if got != tt.want {
				t.Errorf("getModuleName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInspectProject_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		entrypoint  string
		wantErr     bool
	}{
		{
			name:        "empty_project_path",
			projectPath: "",
			entrypoint:  "cmd.NewRootCmd",
			wantErr:     true,
		},
		{
			name:        "invalid_entrypoint_format",
			projectPath: "/tmp/test",
			entrypoint:  "invalid..format",
			wantErr:     true,
		},
		{
			name:        "non_existent_project",
			projectPath: "/non/existent/path",
			entrypoint:  "cmd.NewRootCmd",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InspectProject(tt.projectPath, tt.entrypoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("InspectProject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetFlagTypeMapping(t *testing.T) {
	// Test the getFlagType function indirectly through the template
	tests := []struct {
		name         string
		pflagType    string
		expectedType string
	}{
		// Basic types
		{"string type", "*pflag.stringValue", "string"},
		{"bool type", "*pflag.boolValue", "bool"},
		{"int type", "*pflag.intValue", "int"},
		{"int64 type", "*pflag.int64Value", "int64"},
		{"float64 type", "*pflag.float64Value", "float64"},
		{"duration type", "*pflag.durationValue", "duration"},
		{"stringSlice type", "*pflag.stringSliceValue", "stringSlice"},

		// Integer variants
		{"int8 type", "*pflag.int8Value", "int8"},
		{"int16 type", "*pflag.int16Value", "int16"},
		{"int32 type", "*pflag.int32Value", "int32"},
		{"uint type", "*pflag.uintValue", "uint"},
		{"uint8 type", "*pflag.uint8Value", "uint8"},
		{"uint16 type", "*pflag.uint16Value", "uint16"},
		{"uint32 type", "*pflag.uint32Value", "uint32"},
		{"uint64 type", "*pflag.uint64Value", "uint64"},

		// Float variants
		{"float32 type", "*pflag.float32Value", "float32"},

		// Slice types
		{"intSlice type", "*pflag.intSliceValue", "intSlice"},
		{"int32Slice type", "*pflag.int32SliceValue", "int32Slice"},
		{"int64Slice type", "*pflag.int64SliceValue", "int64Slice"},
		{"uintSlice type", "*pflag.uintSliceValue", "uintSlice"},
		{"float32Slice type", "*pflag.float32SliceValue", "float32Slice"},
		{"float64Slice type", "*pflag.float64SliceValue", "float64Slice"},
		{"boolSlice type", "*pflag.boolSliceValue", "boolSlice"},
		{"durationSlice type", "*pflag.durationSliceValue", "durationSlice"},

		// Map types
		{"stringToString type", "*pflag.stringToStringValue", "stringToString"},
		{"stringToInt64 type", "*pflag.stringToInt64Value", "stringToInt64"},

		// Network types
		{"ip type", "*pflag.ipValue", "ip"},
		{"ipSlice type", "*pflag.ipSliceValue", "ipSlice"},
		{"ipMask type", "*pflag.ipMaskValue", "ipMask"},
		{"ipNet type", "*pflag.ipNetValue", "ipNet"},

		// Binary types
		{"bytesHex type", "*pflag.bytesHexValue", "bytesHex"},
		{"bytesBase64 type", "*pflag.bytesBase64Value", "bytesBase64"},

		// Special types
		{"count type", "*pflag.countValue", "count"},

		// Unknown type should return as-is
		{"unknown type", "*pflag.unknownValue", "*pflag.unknownValue"},
	}

	// Extract the getFlagType function from the template for testing
	_, err := template.New("test").Parse(inspectorTemplate)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Find the typeMap in the template
	templateContent := inspectorTemplate
	if !strings.Contains(templateContent, "typeMap := map[string]string{") {
		t.Fatal("typeMap not found in template")
	}

	// Verify all test cases are in the typeMap
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For unknown types, we expect them to be returned as-is
			if tt.pflagType == "*pflag.unknownValue" {
				return
			}

			// Check if the mapping exists in the template
			expectedMapping := fmt.Sprintf(`"%s":`, tt.pflagType)
			if !strings.Contains(templateContent, expectedMapping) {
				t.Errorf("Type mapping for %s not found in template", tt.pflagType)
			}

			// Check if the expected type is correct
			expectedValue := fmt.Sprintf(`%s":`, tt.pflagType)
			expectedTypeValue := fmt.Sprintf(`"%s"`, tt.expectedType)
			if strings.Contains(templateContent, expectedValue) {
				// Find the line that contains this mapping
				lines := strings.Split(templateContent, "\n")
				found := false
				for _, line := range lines {
					if strings.Contains(line, expectedValue) && strings.Contains(line, expectedTypeValue) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected mapping %s -> %s not found in template", tt.pflagType, tt.expectedType)
				}
			}
		})
	}
}
