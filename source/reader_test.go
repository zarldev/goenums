package source_test

import (
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/source"
)

type errorReader struct{}

func (e errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestReaderSource_Content(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		reader   io.Reader
		expected string
		err      error
	}{
		{
			name:     "successfully read content",
			reader:   strings.NewReader("test content"),
			expected: "test content",
			err:      nil,
		},
		{
			name:     "error reading content",
			reader:   errorReader{},
			expected: "",
			err:      source.ErrReadSource,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := source.FromReader(tt.reader)
			if src.Filename() != "reader" {
				t.Errorf("expected reader filename got %q", src.Filename())
				return
			}
			content, err := src.Content()
			if err != nil && !errors.Is(err, tt.err) {
				t.Errorf("unexpected error: %v", err)
				return
			}
			contentStr := string(content)
			if contentStr != tt.expected {
				t.Errorf("got %q, want %q", contentStr, tt.expected)
			}
		})
	}
}

func TestFileSource_Content(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "successfully read file content",
			testFunc: func(t *testing.T) {
				// Use MemFS for testing
				memfs := file.NewMemFS()
				content := "package main\nfunc main() {}"
				filename := "test.go"

				// Write test content
				err := memfs.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Errorf("failed to write test file: %v", err)
					return
				}

				// Test FromFileSystem
				src := source.FromFileSystem(memfs, filename)
				if src.Filename() != filename {
					t.Errorf("expected filename %q, got %q", filename, src.Filename())
				}

				got, err := src.Content()
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if string(got) != content {
					t.Errorf("got %q, want %q", string(got), content)
				}
			},
		},
		{
			name: "file too large",
			testFunc: func(t *testing.T) {
				// Create a test filesystem that reports a large file size
				testFS := &testFS{
					statFunc: func(name string) (fs.FileInfo, error) {
						return &testFileInfo{size: source.MaxFileSize + 1}, nil
					},
				}

				src := source.FromFileSystem(testFS, "large.go")
				_, err := src.Content()
				if err == nil || !errors.Is(err, source.ErrReadFileSource) {
					t.Errorf("expected ErrReadFileSource for large file, got %v", err)
				}
			},
		},
		{
			name: "stat error",
			testFunc: func(t *testing.T) {
				testFS := &testFS{
					statFunc: func(name string) (fs.FileInfo, error) {
						return nil, errors.New("stat error")
					},
				}

				src := source.FromFileSystem(testFS, "error.go")
				_, err := src.Content()
				if err == nil || !errors.Is(err, source.ErrReadFileSource) {
					t.Errorf("expected ErrReadFileSource for stat error, got %v", err)
				}
			},
		},
		{
			name: "read error",
			testFunc: func(t *testing.T) {
				testFS := &testFS{
					statFunc: func(name string) (fs.FileInfo, error) {
						return &testFileInfo{size: 100}, nil
					},
					readFileFunc: func(name string) ([]byte, error) {
						return nil, errors.New("read error")
					},
				}

				src := source.FromFileSystem(testFS, "read-error.go")
				_, err := src.Content()
				if err == nil || !errors.Is(err, source.ErrReadFileSource) {
					t.Errorf("expected ErrReadFileSource for read error, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

var errTestFs = errors.New("test filesystem error")

// Test filesystem for testing
type testFS struct {
	statFunc     func(name string) (fs.FileInfo, error)
	readFileFunc func(name string) ([]byte, error)
}

func (t *testFS) Stat(name string) (fs.FileInfo, error) {
	if t.statFunc != nil {
		return t.statFunc(name)
	}
	return nil, errTestFs
}

func (t *testFS) ReadFile(name string) ([]byte, error) {
	if t.readFileFunc != nil {
		return t.readFileFunc(name)
	}
	return nil, errTestFs
}

func (t *testFS) Open(name string) (fs.File, error) {
	return nil, errTestFs
}

// Test file info for testing
type testFileInfo struct {
	size int64
}

func (t *testFileInfo) Name() string       { return "test" }
func (t *testFileInfo) Size() int64        { return t.size }
func (t *testFileInfo) Mode() fs.FileMode  { return 0644 }
func (t *testFileInfo) ModTime() time.Time { return time.Now() }
func (t *testFileInfo) IsDir() bool        { return false }
func (t *testFileInfo) Sys() any           { return nil }
