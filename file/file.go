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

// ReadWriteCreateFileFS is an interface that combines file reading, writing, and creation operations.
// Implementations should provide thread-safe access to the filesystem and handle permissions appropriately.
type ReadWriteCreateFileFS interface {
	// ReadFile reads the entire file named by path and returns its contents.
	fs.ReadFileFS

	// Stat returns file information for the specified path.
	fs.StatFS

	// Create creates or truncates the named file and returns a writer to it.
	// If the file already exists, it is truncated.
	Create(name string) (io.WriteCloser, error)

	// WriteFile writes data to the named file, creating it if necessary.
	// If the file exists, it is truncated. Permissions are set according to perm.
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

var _ ReadWriteCreateFileFS = (*OSReadWriteFileFS)(nil)

// WriteToFileAndFormatFS creates a file at the specified path and writes content to it
// reading and writing from the provided filesystem.
func WriteToFileAndFormatFS(ctx context.Context, fs ReadWriteCreateFileFS, fullPath string, format bool, writeFunc func(io.Writer) error) error {
	if fullPath == "" {
		return fmt.Errorf("%w: %s", ErrCreateFile, "path cannot be empty")
	}
	if writeFunc == nil {
		return fmt.Errorf("%w: %s", ErrWriteFile, "must provide a writeable func")
	}
	if ctx.Err() != nil {
		return ctx.Err()
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
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err = formatFile(fs, fullPath); err != nil {
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
