package file

import (
	"bytes"
	"io"
	"io/fs"
	"sync"
	"time"
)

// DefaultFilePerms are the default permissions
const DefaultFilePerms fs.FileMode = 0644

// Compile time check to ensure MemFS implements ReadWriteCreateFileFS
var _ ReadCreateWriteFileFS = (*MemFS)(nil)

// MemFS is a simple in-memory filesystem implementation
// used for testing purposes. It provides thread-safe access
// to files stored as byte buffers in memory.
type MemFS struct {
	mu    sync.RWMutex
	files map[string]*bytes.Buffer
}

// NewMemFS creates a new empty in-memory filesystem.
// It initializes an empty map of files that can be accessed
// through the ReadWriteCreateFileFS interface methods.
func NewMemFS() *MemFS {
	return &MemFS{
		files: make(map[string]*bytes.Buffer),
	}
}

// ReadFile implements ReadWriteFileFS.ReadFile by returning
// a copy of the file's contents from memory.
// If the file doesn't exist, it returns fs.ErrNotExist.
func (m *MemFS) ReadFile(name string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if name == "" {
		return nil, fs.ErrInvalid
	}
	if buf, ok := m.files[name]; ok {
		return buf.Bytes(), nil
	}
	return nil, fs.ErrNotExist
}

// WriteFile implements ReadWriteFileFS.WriteFile by storing
// the provided data in memory under the given name.
// File permissions are noted but not enforced in memory.
// If the file already exists, it will be overwritten.
func (m *MemFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if name == "" {
		return fs.ErrInvalid
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[name] = bytes.NewBuffer(data)
	return nil
}

// Open implements fs.FS.Open by providing a file handle
// for the specified file in memory.
// The returned file can be read but not written to.
func (m *MemFS) Open(name string) (fs.File, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if name == "" {
		return nil, fs.ErrInvalid
	}
	if buf, ok := m.files[name]; ok {
		return &memFile{
			name:   name,
			Reader: bytes.NewReader(buf.Bytes()),
			Buffer: buf,
		}, nil
	}
	return nil, fs.ErrNotExist
}

// Create implements ReadWriteCreateFileFS.Create by creating
// an empty file in memory and returning a writer to it.
// If a file with the same name exists, it will be truncated.
func (m *MemFS) Create(name string) (io.WriteCloser, error) {
	if name == "" {
		return nil, fs.ErrInvalid
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[name] = bytes.NewBuffer(nil)
	return &memFile{
		name:   name,
		Reader: bytes.NewReader(nil),
		Buffer: m.files[name],
	}, nil
}

// Stat implements fs.StatFS by returning file information
// for a file in the in-memory filesystem.
// Returns fs.ErrNotExist if the file doesn't exist.
func (m *MemFS) Stat(name string) (fs.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if name == "" {
		return nil, fs.ErrInvalid
	}
	f, ok := m.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return &memFileInfo{
		name: name,
		size: int64(f.Len()),
	}, nil
}

// memFileInfo implements fs.FileInfo for in-memory files.
// It provides file metadata for files stored in the MemFS in-memory filesystem.
//
// This struct satisfies the fs.FileInfo interface by providing basic file
// information such as name, size, and permissions. Since files are stored
// in memory, some properties like modification time are simulated.
type memFileInfo struct {
	// name is the filename or path identifier
	name string
	// size is the current size of the file in bytes
	size int64
}

func (m *memFileInfo) Name() string       { return m.name }
func (m *memFileInfo) Size() int64        { return m.size }
func (m *memFileInfo) Mode() fs.FileMode  { return DefaultFilePerms }
func (m *memFileInfo) ModTime() time.Time { return time.Now() }
func (m *memFileInfo) IsDir() bool        { return false }
func (m *memFileInfo) Sys() any           { return nil }

// memFile implements both fs.File and io.WriteCloser for in-memory files.
// It provides a file-like interface for data stored in memory buffers,
// supporting both reading and writing operations.
//
// The memFile struct bridges the gap between the standard library's file
// interfaces and in-memory storage, making it useful for testing and
// scenarios where actual filesystem access is not desired.
type memFile struct {
	// name is the identifier for this file
	name string
	// Reader provides read access to the file's content
	Reader *bytes.Reader
	// Buffer provides write access and stores the file's content
	Buffer *bytes.Buffer
}

func (f *memFile) Close() error { return nil }

func (f *memFile) Stat() (fs.FileInfo, error) {
	if f.Buffer == nil {
		return nil, fs.ErrInvalid
	}
	return &memFileInfo{
		name: f.name,
		size: int64(f.Buffer.Len()),
	}, nil
}

func (f *memFile) Write(p []byte) (int, error) {
	if f.Buffer == nil {
		return 0, fs.ErrInvalid
	}
	return f.Buffer.Write(p)
}

func (f *memFile) Read(p []byte) (int, error) {
	if f.Reader == nil {
		return 0, fs.ErrInvalid
	}
	return f.Reader.Read(p)
}
