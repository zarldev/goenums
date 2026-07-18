// The file package provides utilities for file I/O operations with specific
// handling for Go source files, including automatic formatting.
package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	goformat "go/format"
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

type ReadStatFS interface {
	// ReadFile reads the entire file named by path and returns its contents.
	fs.ReadFileFS

	// Stat returns file information for the specified path.
	fs.StatFS
}

type CreateWriteFileFS interface {
	// Create creates or truncates the named file and returns a writer to it.
	Create(name string) (io.WriteCloser, error)
	// WriteFile writes data to the named file, creating it if necessary.
	// If the file exists, it is truncated. Permissions are set according to perm.
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

// ReadCreateWriteFileFS is an interface that combines file reading, writing, and creation operations.
// Implementations should provide thread-safe access to the filesystem and handle permissions appropriately.
type ReadCreateWriteFileFS interface {
	ReadStatFS
	CreateWriteFileFS
}

var _ ReadCreateWriteFileFS = (*OSReadWriteFileFS)(nil)

// WriteToFileAndFormatFS creates a file at the specified path and writes content to it
// reading and writing from the provided filesystem.
func WriteToFileAndFormatFS(ctx context.Context, fs ReadCreateWriteFileFS, fullPath string, format bool, writeFunc func(io.Writer) error) error {
	if fullPath == "" {
		return fmt.Errorf("%w: %s", ErrCreateFile, "path cannot be empty")
	}
	if writeFunc == nil {
		return fmt.Errorf("%w: %s", ErrWriteFile, "must provide a writeable func")
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var output bytes.Buffer
	if err := writeFunc(&output); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrWriteFile, fullPath, err)
	}

	data := output.Bytes()
	if format {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		formatted, err := goformat.Source(data)
		if err != nil {
			return fmt.Errorf("%w: %s: %w", ErrFormatFile, fullPath, err)
		}
		data = formatted
	}

	if err := fs.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrWriteFile, fullPath, err)
	}
	return nil
}
