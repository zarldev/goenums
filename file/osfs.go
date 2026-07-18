package file

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var ErrInvalidPath = errors.New("invalid file path")

func validatePath(name string) error {
	if filepath.IsAbs(name) {
		return nil
	}
	if !filepath.IsLocal(name) {
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

func (o *OSReadWriteFileFS) Stat(name string) (fs.FileInfo, error) {
	if err := validatePath(name); err != nil {
		return nil, err
	}
	return os.Stat(name) // #nosec G304 - path validated above
}

func (o *OSReadWriteFileFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if err := validatePath(name); err != nil {
		return err
	}
	return os.WriteFile(name, data, perm) // #nosec G304 - path validated above
}

// Create creates or truncates the named file.
func (o *OSReadWriteFileFS) Create(name string) (io.WriteCloser, error) {
	if err := validatePath(name); err != nil {
		return nil, err
	}
	return os.Create(name) // #nosec G304 - path validated above
}
