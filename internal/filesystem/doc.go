// Package filesystem provides an abstraction layer for file system operations
// in cliguard, making file operations testable and mockable.
//
// The filesystem package allows cliguard to work with files in a way that
// can be easily tested without requiring actual file system access. This is
// particularly useful for unit tests and for implementing virtual file systems.
//
// # Basic Usage
//
// The default filesystem uses the OS file system:
//
//	fs := filesystem.NewOSFileSystem()
//	data, err := fs.ReadFile("cliguard.yaml")
//	if err != nil {
//	    return err
//	}
//
//	err = fs.WriteFile("output.yaml", data, 0644)
//	if err != nil {
//	    return err
//	}
//
// # Testing Support
//
// For testing, use the in-memory filesystem:
//
//	fs := filesystem.NewMemoryFileSystem()
//	fs.WriteFile("test.yaml", []byte("content"), 0644)
//
//	// Your code that uses the filesystem
//	data, err := fs.ReadFile("test.yaml")
//
// # File Operations
//
// The filesystem interface provides:
//   - ReadFile: Read file contents
//   - WriteFile: Write data to a file
//   - Exists: Check if a file or directory exists
//   - MkdirAll: Create directories recursively
//   - Remove: Delete files or directories
//   - Walk: Traverse directory trees
//   - Stat: Get file information
//
// # Path Handling
//
// The filesystem handles paths consistently:
//   - Normalizes path separators
//   - Resolves relative paths
//   - Handles symbolic links appropriately
//   - Provides clean path operations
//
// # Error Handling
//
// The filesystem provides standard error types:
//   - File not found errors
//   - Permission denied errors
//   - Path already exists errors
//   - Invalid path errors
//
// # Virtual File Systems
//
// Custom filesystems can be implemented for:
//   - In-memory testing
//   - Remote file systems
//   - Encrypted storage
//   - Layered file systems
//   - Read-only overlays
//
// # Security
//
// The filesystem implementation:
//   - Validates paths to prevent directory traversal
//   - Respects file permissions
//   - Provides safe defaults for file creation
//   - Handles temporary files securely
package filesystem
