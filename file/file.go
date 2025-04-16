// The file package provides utilities for file I/O operations with specific
// handling for Go source files, including automatic formatting.
package file

import (
	"context"
	"errors"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"os"
)

var (
	// ErrFormatFile indicates an error occurred while formatting a Go file.
	ErrFormatFile = errors.New("failed to format Go file")
	// ErrWriteFile indicates an error occurred while writing to a file.
	ErrWriteFile = errors.New("failed to write to file")
	// ErrCreateFile indicates an error occurred while creating a file.
	ErrCreateFile = errors.New("failed to create file")
	// ErrReadFile indicates an error occurred while reading a file.
	ErrReadFile = errors.New("failed to read file")
)

// compile-time check to ensure OSReadFileFS implements fs.ReadFileFS
var _ fs.ReadFileFS = (*OSReadWriteFileFS)(nil)

// OSReadWriteFileFS is a type that implements fs.ReadFileFS using os.ReadFile.
type OSReadWriteFileFS struct {
}

// ReadFile reads the named file and returns the contents.
func (o *OSReadWriteFileFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// Open opens the named file.
func (o *OSReadWriteFileFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

// Stat returns the FileInfo for the named file.
func (o *OSReadWriteFileFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// WriteFile writes data to a file named by filename with the provided permissions.
func (o *OSReadWriteFileFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// Create creates or truncates the named file.
func (o *OSReadWriteFileFS) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

// ReadWriteCreateFileFS is an interface that combines fs.ReadFileFS and WriteFile.
type ReadWriteCreateFileFS interface {
	fs.ReadFileFS
	Create(name string) (io.WriteCloser, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

// WriteToFileAndFormat creates a file at the specified path and writes content to it
// using the provided write function. After writing, if requested it formats
// the file using Go's standard formatter. This function handles file creation,
// writing, and formatting errors. The content is written using the provided writeFunc.
func WriteToFileAndFormat(ctx context.Context, fullPath string, format bool, writeFunc func(io.Writer) error) error {
	return WriteToFileAndFormatFS(ctx, &OSReadWriteFileFS{}, fullPath, format, writeFunc)
}

// WriteToFileAndFormatFS creates a file at the specified path and writes content to it
// reading and writing from the provided filesystem.
func WriteToFileAndFormatFS(ctx context.Context, fs ReadWriteCreateFileFS, fullPath string, format bool, writeFunc func(io.Writer) error) error {
	if fullPath == "" {
		return fmt.Errorf("%w: %s", ErrWriteFile, "path cannot be empty")
	}
	if writeFunc == nil {
		return fmt.Errorf("%w: %s", ErrWriteFile, "must provide a writeable func")
	}
	f, err := fs.Create(fullPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrCreateFile, fullPath, err)
	}
	defer f.Close()
	if err := writeFunc(f); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrWriteFile, fullPath, err)
	}
	if format {
		err = formatFile(fs, fullPath)
		if err != nil {
			return fmt.Errorf("%w: %s: %w", ErrFormatFile, fullPath, err)
		}
	}
	return nil
}

// formatFile applies Go's standard formatting to a file at the specified path.
// It reads the file content, formats it using go/format, and writes the
// formatted content back to the file using the provided filesystem. This
// ensures generated Go code follows standard formatting conventions.
func formatFile(fs ReadWriteCreateFileFS, filename string) error {
	f, err := fs.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrReadFile, filename, err)
	}
	b, err := format.Source(f)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrFormatFile, filename, err)
	}
	if err = fs.WriteFile(filename, b, 0644); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrWriteFile, filename, err)
	}
	return nil
}
