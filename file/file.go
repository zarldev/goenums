// The file package provides utilities for file I/O operations with specific
// handling for Go source files, including automatic formatting.
package file

import (
	"fmt"
	"go/format"
	"io"
	"os"
)

// WriteToFile creates a file at the specified path and writes content to it
// using the provided write function. After writing, if requested it formats
// the file using Go's standard formatter if the content is valid Go code.
// The write function receives an io.Writer to write content to the file.
func WriteToFileAndFormat(fullPath string, format bool, writeFunc func(io.Writer) error) error {
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", fullPath, err)
	}
	err = func(file *os.File) error {
		defer file.Close()
		return writeFunc(file)
	}(f)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", fullPath, err)
	}
	if format {
		err = formatFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to format generated file %s: %w", fullPath, err)
		}
	}
	return nil
}

// formatFile applies Go's standard formatting to a file at the specified path.
// It reads the file content, formats it using go/format, and writes the
// formatted content back to the file. This ensures generated Go code
// follows standard formatting conventions.
func formatFile(filename string) error {
	f, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file for formatting %s: %w", filename, err)
	}
	b, err := format.Source(f)
	if err != nil {
		return fmt.Errorf("failed to format Go source in %s: %w\nInvalid code may have been generated", filename, err)
	}
	err = os.WriteFile(filename, b, 0644)
	if err != nil {
		return fmt.Errorf("failed to write formatted file %s: %w", filename, err)
	}
	return nil
}
