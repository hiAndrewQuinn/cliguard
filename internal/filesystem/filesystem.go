package filesystem

import (
	"os"
)

// FileSystem is an interface for file system operations
type FileSystem interface {
	MkdirTemp(dir, pattern string) (string, error)
	RemoveAll(path string) error
	WriteFile(name string, data []byte, perm os.FileMode) error
	ReadFile(name string) ([]byte, error)
	Stat(name string) (os.FileInfo, error)
}

// OSFileSystem is the real implementation using os package
type OSFileSystem struct{}

// MkdirTemp creates a temporary directory
func (fs *OSFileSystem) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

// RemoveAll removes a path and any children it contains
func (fs *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// WriteFile writes data to a file
func (fs *OSFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// ReadFile reads the contents of a file
func (fs *OSFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// Stat returns file info
func (fs *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
