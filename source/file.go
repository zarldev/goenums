package source

import (
	"errors"
	"fmt"

	"github.com/zarldev/goenums/file"
)

var (
	// ErrReadFileSource is returned when there is an error reading the source file.
	ErrReadFileSource = errors.New("failed to read file source")
)

const (
	// MaxFileSize is the maximum allowed file size in bytes (default: 10MB)
	// Files larger than this will trigger an error to prevent resource exhaustion
	MaxFileSize = 10 * 1024 * 1024 // 10MB
)

// FromFile creates a new file-based Source implementation that reads
// enum definitions from a file at the specified path.
func FromFile(path string) *FileSource {
	return FromFileSystem(&file.OSReadWriteFileFS{}, path)
}

// FromFileSystem creates a new file-based Source implementation that reads
// enum definitions from a file at the specified path from the provided filesystem.
func FromFileSystem(fs file.ReadWriteCreateFileFS, path string) *FileSource {
	return &FileSource{
		Path: path,
		FS:   fs,
	}
}

// FileSource implements Source for file-based content sources.
// It reads enum definitions from a file on the local filesystem.
type FileSource struct {
	// Path is the filesystem path to the source file
	Path string
	FS   file.ReadWriteCreateFileFS
}

// Content reads and returns the file's contents as a byte slice.
// It fulfills the Source interface by providing the raw content
// to be parsed for enum definitions.
func (fs *FileSource) Content() ([]byte, error) {
	// Check file size before reading to prevent loading extremely large files
	fileInfo, err := fs.FS.Stat(fs.Path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s (stat): %w", ErrReadFileSource, fs.Path, err)
	}

	if fileInfo.Size() > MaxFileSize {
		return nil, fmt.Errorf("%w: %s exceeds maximum allowed size of %d bytes",
			ErrReadFileSource, fs.Path, MaxFileSize)
	}
	b, err := fs.FS.ReadFile(fs.Path)
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
