package source_test

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing/fstest"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/source"
)

// Example_fileSource demonstrates how to create and use a file-based source
// to read content from the filesystem.
func Example_fileSource() {
	// In a real application, you would use an actual file path:
	// src := source.FromFile("path/to/your/enums.go")

	// For demonstration, we'll use an in-memory fstest.MapFS
	mockFS := fstest.MapFS{
		"enums.go": &fstest.MapFile{
			Data: []byte(`package example

type color int

const (
	unknown color = iota // invalid
	red                  // Red
	green                // Green
	blue                 // Blue
)`),
		},
	}

	// Create a file source using our mock filesystem
	src := source.FromFileSystem(&mockFSAdapter{fs: mockFS}, "enums.go")

	// Get the filename (useful for error reporting)
	fmt.Println("Source filename:", src.Filename())

	// Read the content
	content, err := src.Content()
	if err != nil {
		fmt.Printf("Error reading content: %v\n", err)
		return
	}

	// Print the first line of content to demonstrate successful reading
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 {
		fmt.Println("First line of content:", lines[0])
	}

	// Output:
	// Source filename: enums.go
	// First line of content: package example
}

// Example_readerSource demonstrates how to create and use a reader-based source
// to process content from any io.Reader implementation.
func Example_readerSource() {
	// Create a simple string reader with some enum definitions
	enumContent := `package demo

type status int

const (
	unknown status = iota // invalid
	pending                // Pending
	active                 // Active
	completed              // Completed
)
`
	reader := strings.NewReader(enumContent)

	// Create a reader source
	src := source.FromReader(reader)

	// Get the generic filename
	fmt.Println("Source identifier:", src.Filename())

	// Read the content
	content, err := src.Content()
	if err != nil {
		fmt.Printf("Error reading content: %v\n", err)
		return
	}

	// Print content length to demonstrate successful reading
	fmt.Printf("Content read successfully (%d bytes)\n", len(content))

	// Print the second line to show we actually got content
	lines := strings.Split(string(content), "\n")
	if len(lines) > 1 {
		fmt.Println("Second line:", lines[1])
	}

	// Output:
	// Source identifier: reader
	// Content read successfully (181 bytes)
	// Second line:
}

// Example_processingMultipleSources shows how to handle multiple sources
// with the same processing logic, demonstrating the abstraction benefit.
func Example_processingMultipleSources() {
	// Define a simple processor function that works with any Source
	processSource := func(src enum.Source) error {
		fmt.Printf("Processing source: %s\n", src.Filename())

		content, err := src.Content()
		if err != nil {
			return fmt.Errorf("failed to read source %s: %w", src.Filename(), err)
		}

		// Count lines of code as a simple processing example
		lineCount := len(strings.Split(string(content), "\n"))
		fmt.Printf("Source contains %d lines\n", lineCount)

		return nil
	}

	// Create sources from different origins
	fileContent := `package colors
	
type color int

const (
	unknown color = iota
	red
	green
	blue
)`

	// Create a mock file source
	mockFS := fstest.MapFS{
		"colors.go": &fstest.MapFile{
			Data: []byte(fileContent),
		},
	}
	fileSource := source.FromFileSystem(&mockFSAdapter{fs: mockFS}, "colors.go")

	// Create a reader source
	readerSource := source.FromReader(strings.NewReader(fileContent))

	// Process both sources using the same logic
	fmt.Println("=== Processing multiple sources ===")
	_ = processSource(fileSource)
	fmt.Println("---")
	_ = processSource(readerSource)

	// Output:
	// === Processing multiple sources ===
	// Processing source: colors.go
	// Source contains 10 lines
	// ---
	// Processing source: reader
	// Source contains 10 lines
}

// mockFSAdapter adapts a fstest.MapFS to the file.ReadStatFS interface
type mockFSAdapter struct {
	fs fstest.MapFS
}

func (m *mockFSAdapter) ReadFile(name string) ([]byte, error) {
	return m.fs.ReadFile(name)
}

func (m *mockFSAdapter) Stat(name string) (os.FileInfo, error) {
	return m.fs.Stat(name)
}

func (m *mockFSAdapter) Open(name string) (fs.File, error) {
	// This returns fs.File as required by the interface, not io.Reader
	return m.fs.Open(name)
}
