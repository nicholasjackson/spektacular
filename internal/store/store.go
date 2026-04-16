package store

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotFound is returned when the requested file or directory does not exist.
var ErrNotFound = errors.New("not found")

// Store provides read/write access to a project's data directory.
// All paths are relative to the store root.
type Store interface {
	// Root returns the absolute path to the store root directory.
	Root() string
	// Read returns the contents of the file at path.
	Read(path string) ([]byte, error)
	// Write creates or overwrites the file at path with content.
	// Parent directories are created automatically.
	Write(path string, content []byte) error
	// Delete removes the file at path. Returns nil if the file does not exist.
	Delete(path string) error
	// List returns the names of direct children within the directory at path.
	List(path string) ([]string, error)
	// Exists reports whether a file or directory exists at path.
	Exists(path string) bool
}

// FileStore implements Store over the local filesystem.
// All paths are resolved relative to root and must not escape it.
type FileStore struct {
	root string
}

// NewFileStore creates a FileStore rooted at root.
func NewFileStore(root string) *FileStore {
	return &FileStore{root: filepath.Clean(root)}
}

// Root returns the absolute path to the store root directory.
func (f *FileStore) Root() string {
	return f.root
}

// abs resolves a relative path against the root, rejecting path traversal.
func (f *FileStore) abs(path string) (string, error) {
	joined := filepath.Join(f.root, path)
	rel, err := filepath.Rel(f.root, joined)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", errors.New("path escapes store root")
	}
	return joined, nil
}

func (f *FileStore) Read(path string) ([]byte, error) {
	abs, err := f.abs(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

func (f *FileStore) Write(path string, content []byte) error {
	abs, err := f.abs(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return err
	}
	return os.WriteFile(abs, content, 0644)
}

func (f *FileStore) Delete(path string) error {
	abs, err := f.abs(path)
	if err != nil {
		return err
	}
	err = os.Remove(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (f *FileStore) List(path string) ([]string, error) {
	abs, err := f.abs(path)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names, nil
}

func (f *FileStore) Exists(path string) bool {
	abs, err := f.abs(path)
	if err != nil {
		return false
	}
	_, err = os.Stat(abs)
	return err == nil
}
