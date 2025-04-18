//go:build prod
// +build prod

package testdata

import (
	"io"
	"io/fs"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator/config"
)

// In production builds, this is a stub implementation that doesn't embed any files

var (
	_ file.ReadStatFS        = TestDataFS{}
	_ file.CreateWriteFileFS = TestDataFS{}
)

type TestDataFS struct {
	write *file.MemFS
}

// Create implements file.CreateWriteFileFS.
func (f TestDataFS) Create(name string) (io.WriteCloser, error) {
	return f.write.Create(name)
}

// WriteFile implements file.CreateWriteFileFS.
func (f TestDataFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return f.write.WriteFile(name, data, perm)
}

// Open implements file.ReadStatFS.
func (f TestDataFS) Open(name string) (fs.File, error) {
	// In production, there are no embedded files to open
	return nil, fs.ErrNotExist
}

// Stat implements file.ReadStatFS.
func (f TestDataFS) Stat(name string) (fs.FileInfo, error) {
	// In production, there are no embedded files to stat
	return nil, fs.ErrNotExist
}

var FS = TestDataFS{
	write: file.NewMemFS(),
}

func (f TestDataFS) ReadFile(name string) ([]byte, error) {
	// In production, there are no embedded files to read
	return nil, fs.ErrNotExist
}

// Empty test cases in production build
var InputOutputTestCases = []struct {
	Name string

	Config              config.Configuration
	Source              enum.Source
	ExpectedFiles       []string
	RepresentationCount int
	Err                 error
}{}
