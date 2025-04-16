package source

import (
	"errors"
	"fmt"
	"os"
)

var (
	// ErrReadFileSource is returned when there is an error reading the source file.
	ErrReadFileSource = errors.New("failed to read file source")
)

// FromFile creates a new file-based Source implementation that reads
// enum definitions from a file at the specified path.
func FromFile(path string) *FileSource {
	return &FileSource{Path: path}
}

// FileSource implements Source for file-based content sources.
// It reads enum definitions from a file on the local filesystem.
type FileSource struct {
	// Path is the filesystem path to the source file
	Path string
}

// Content reads and returns the file's contents as a byte slice.
// It fulfills the Source interface by providing the raw content
// to be parsed for enum definitions.
func (fs *FileSource) Content() ([]byte, error) {
	b, err := os.ReadFile(fs.Path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrReadFileSource, fs.Path, err)
	}
	return b, nil
}

// Filename returns the path of the source file.
// This identifies the source in error messages and generated
// code documentation.
func (fs *FileSource) Filename() string {
	return fs.Path
}
