package file

import (
	"io"
	"io/fs"
	"os"
)

// compile-time check to ensure OSReadFileFS implements ReadFileFS
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
