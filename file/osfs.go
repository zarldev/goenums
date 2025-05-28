package file

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidPath = errors.New("invalid file path")

func validatePath(name string) error {
	cleaned := filepath.Clean(name)
	if strings.Contains(cleaned, "..") {
		return ErrInvalidPath
	}
	if filepath.IsAbs(cleaned) {
		return ErrInvalidPath
	}
	return nil
}

// compile-time check to ensure OSReadFileFS implements ReadFileFS
var _ fs.ReadFileFS = (*OSReadWriteFileFS)(nil)

// OSReadWriteFileFS is a type that implements fs.ReadFileFS using os.ReadFile.
type OSReadWriteFileFS struct {
}

// ReadFile reads the named file and returns the contents.
func (o *OSReadWriteFileFS) ReadFile(name string) ([]byte, error) {
	if err := validatePath(name); err != nil {
		return nil, err
	}
	return os.ReadFile(name) // #nosec G304 - path validated above
}

// Open opens the named file.
func (o *OSReadWriteFileFS) Open(name string) (fs.File, error) {
	if err := validatePath(name); err != nil {
		return nil, err
	}
	return os.Open(name) // #nosec G304 - path validated above
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
	if err := validatePath(name); err != nil {
		return nil, err
	}
	return os.Create(name) // #nosec G304 - path validated above
}
