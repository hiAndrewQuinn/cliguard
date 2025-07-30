# Issue: Complex Flag Types Generate Invalid Contract Types

## Summary
When generating contracts for CLIs that use advanced flag types (maps, slices, IP addresses, etc.), cliguard outputs internal pflag type names (e.g., `*pflag.stringToStringValue`) instead of normalized type names. This causes the generated contracts to fail validation immediately.

## Current Behavior
```yaml
# Generated contract shows:
flags:
  - name: string-to-string
    usage: String to string map flag
    type: '*pflag.stringToStringValue'  # ❌ Invalid type
```

## Expected Behavior
```yaml
# Should generate:
flags:
  - name: string-to-string
    usage: String to string map flag
    type: stringToString  # ✅ Normalized type
```

## Root Cause
The issue is in `/internal/inspector/inspector.go` at line 176-184 in the `getFlagType()` function. The current `typeMap` only includes 7 basic types but misses many pflag types.

## Affected Flag Types
Based on testing, these flag types are affected:
- `*pflag.stringToStringValue` (map[string]string)
- `*pflag.stringToInt64Value` (map[string]int64)
- `*pflag.countValue` (count flags)
- `*pflag.int8Value`, `*pflag.int16Value`, `*pflag.int32Value`
- `*pflag.uint8Value`, `*pflag.uint16Value`, `*pflag.uint32Value`, `*pflag.uint64Value`, `*pflag.uintValue`
- `*pflag.float32Value`
- `*pflag.intSliceValue`, `*pflag.int32SliceValue`, `*pflag.int64SliceValue`
- `*pflag.uintSliceValue`, `*pflag.float32SliceValue`, `*pflag.float64SliceValue`
- `*pflag.boolSliceValue`, `*pflag.durationSliceValue`
- `*pflag.ipValue`, `*pflag.ipSliceValue`, `*pflag.ipMaskValue`, `*pflag.ipNetValue`
- `*pflag.bytesHexValue`, `*pflag.bytesBase64Value`

## Fix Implementation

### 1. Update Type Mapping
Expand the `typeMap` in `/internal/inspector/inspector.go`:

```go
typeMap := map[string]string{
    // Basic types (existing)
    "*pflag.stringValue":       "string",
    "*pflag.boolValue":         "bool",
    "*pflag.intValue":          "int",
    "*pflag.int64Value":        "int64",
    "*pflag.float64Value":      "float64",
    "*pflag.durationValue":     "duration",
    "*pflag.stringSliceValue":  "stringSlice",
    
    // Integer variants
    "*pflag.int8Value":         "int8",
    "*pflag.int16Value":        "int16",
    "*pflag.int32Value":        "int32",
    "*pflag.uint8Value":        "uint8",
    "*pflag.uint16Value":       "uint16",
    "*pflag.uint32Value":       "uint32",
    "*pflag.uint64Value":       "uint64",
    "*pflag.uintValue":         "uint",
    
    // Float variants
    "*pflag.float32Value":      "float32",
    
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
```

### 2. Update Contract Validation
Update `validTypes` in `/internal/contract/parser.go` line 109 to include all new types:

```go
validTypes := map[string]bool{
    // Basic types (existing)
    "string": true, "bool": true, "int": true, "int64": true,
    "float64": true, "duration": true, "stringSlice": true,
    
    // Add all new types...
    "int8": true, "int16": true, "int32": true,
    "uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
    "float32": true,
    "intSlice": true, "int32Slice": true, "int64Slice": true,
    "uintSlice": true, "float32Slice": true, "float64Slice": true,
    "boolSlice": true, "durationSlice": true,
    "stringToString": true, "stringToInt64": true,
    "ip": true, "ipSlice": true, "ipMask": true, "ipNet": true,
    "bytesHex": true, "bytesBase64": true,
    "count": true,
}
```

## Test Case
Create test in `/internal/inspector/inspector_test.go`:

```go
func TestComplexFlagTypes(t *testing.T) {
    tests := []struct {
        pflagType    string
        expectedType string
    }{
        {"*pflag.stringToStringValue", "stringToString"},
        {"*pflag.countValue", "count"},
        {"*pflag.ipValue", "ip"},
        // ... test all types
    }
    
    for _, tt := range tests {
        t.Run(tt.pflagType, func(t *testing.T) {
            // Mock flag with specific type
            flag := &pflag.Flag{Value: mockValue(tt.pflagType)}
            got := getFlagType(flag)
            if got != tt.expectedType {
                t.Errorf("getFlagType() = %v, want %v", got, tt.expectedType)
            }
        })
    }
}
```

## Prevention Strategy

1. **Comprehensive Testing**: Add test cases for all pflag types
2. **Type Discovery**: Create a script to extract all pflag.Value implementations
3. **Validation**: Ensure contract validation accepts all normalized types
4. **Documentation**: Document supported flag types in README

## Reproduction Steps

1. Create a CLI with complex flag types:
```go
cmd.Flags().StringToStringVar(&m, "map", nil, "String to string map")
cmd.Flags().CountVarP(&c, "count", "c", "Count flag")
```

2. Generate contract:
```bash
cliguard generate --project-path . --entrypoint "pkg.NewCmd" > contract.yaml
```

3. Observe invalid types in contract.yaml

4. Validation fails:
```bash
cliguard validate --project-path . --entrypoint "pkg.NewCmd" --contract contract.yaml
# Error: invalid type '*pflag.stringToStringValue'
```

## Priority
High - This prevents cliguard from working with any CLI using advanced flag types, which are common in production CLIs.

## Labels
- bug
- type-system
- validation
- pflag-compatibility