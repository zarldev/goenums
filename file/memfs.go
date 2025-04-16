package file

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"sync"
)

// MemFS is a simple in-memory filesystem implementation
// this is used for testing purposes
type MemFS struct {
	mu    sync.RWMutex
	files map[string]*bytes.Buffer
}

// NewMemFS creates a new MemFS
func NewMemFS() *MemFS {
	return &MemFS{
		files: make(map[string]*bytes.Buffer),
	}
}

// ReadFile implements ReadWriteFileFS.ReadFile
func (m *MemFS) ReadFile(name string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if buf, ok := m.files[name]; ok {
		return buf.Bytes(), nil
	}
	return nil, fs.ErrNotExist
}

// WriteFile implements ReadWriteFileFS.WriteFile
func (m *MemFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.files[name] = bytes.NewBuffer(data)
	return nil
}

// Open implements fs.FS.Open
func (m *MemFS) Open(name string) (fs.File, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if buf, ok := m.files[name]; ok {
		return &memFile{
			Reader: bytes.NewReader(buf.Bytes()),
			Buffer: buf,
		}, nil
	}
	return nil, fs.ErrNotExist
}

// Create implements ReadWriteCreateFileFS.Create
func (m *MemFS) Create(name string) (io.WriteCloser, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[name] = bytes.NewBuffer(nil)
	return &memFile{
		Reader: bytes.NewReader(nil),
		Buffer: m.files[name],
	}, nil
}

type memFile struct {
	Reader *bytes.Reader
	Buffer *bytes.Buffer
}

func (f *memFile) Close() error               { return nil }
func (f *memFile) Stat() (fs.FileInfo, error) { return nil, errors.New("Stat not implemented") }
func (f *memFile) Write(p []byte) (n int, err error) {
	if f.Buffer == nil {
		f.Buffer = bytes.NewBuffer(nil)
	}
	return f.Buffer.Write(p)
}
func (f *memFile) Read(p []byte) (n int, err error) { return f.Reader.Read(p) }
