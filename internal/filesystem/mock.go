package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MockFileSystem is a mock implementation for testing
type MockFileSystem struct {
	Files        map[string][]byte
	Directories  map[string]bool
	TempDirNum   int
	StatErrors   map[string]error
	MkdirTempErr error // Error to return from MkdirTemp
}

// NewMockFileSystem creates a new mock filesystem
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:       make(map[string][]byte),
		Directories: make(map[string]bool),
		StatErrors:  make(map[string]error),
	}
}

// MkdirTemp creates a mock temporary directory
func (fs *MockFileSystem) MkdirTemp(dir, pattern string) (string, error) {
	if fs.MkdirTempErr != nil {
		return "", fs.MkdirTempErr
	}
	fs.TempDirNum++
	tempDir := filepath.Join(dir, fmt.Sprintf("%s%d", pattern, fs.TempDirNum))
	fs.Directories[tempDir] = true
	return tempDir, nil
}

// RemoveAll removes a mock path
func (fs *MockFileSystem) RemoveAll(path string) error {
	// Remove the directory
	delete(fs.Directories, path)

	// Remove all files under this path
	for filePath := range fs.Files {
		if filepath.HasPrefix(filePath, path) {
			delete(fs.Files, filePath)
		}
	}

	return nil
}

// WriteFile writes to a mock file
func (fs *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	// Ensure parent directory exists
	dir := filepath.Dir(name)
	if !fs.directoryExists(dir) {
		return fmt.Errorf("directory %s does not exist", dir)
	}

	fs.Files[name] = data
	return nil
}

// ReadFile reads from a mock file
func (fs *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if data, ok := fs.Files[name]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

// Stat returns mock file info
func (fs *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if err, ok := fs.StatErrors[name]; ok {
		return nil, err
	}

	if _, ok := fs.Files[name]; ok {
		return &mockFileInfo{name: filepath.Base(name), isDir: false}, nil
	}

	if fs.directoryExists(name) {
		return &mockFileInfo{name: filepath.Base(name), isDir: true}, nil
	}

	return nil, os.ErrNotExist
}

// directoryExists checks if a directory exists in the mock filesystem
func (fs *MockFileSystem) directoryExists(path string) bool {
	if path == "/" || path == "." {
		return true
	}

	if exists, ok := fs.Directories[path]; ok && exists {
		return true
	}

	// Check if it's a parent of any existing directory
	for dir := range fs.Directories {
		if filepath.HasPrefix(dir, path) {
			return true
		}
	}

	// Check if it's a parent of any existing file
	for file := range fs.Files {
		if filepath.HasPrefix(file, path) {
			return true
		}
	}

	return false
}

// mockFileInfo implements os.FileInfo
type mockFileInfo struct {
	name  string
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }
