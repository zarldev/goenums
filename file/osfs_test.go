package file_test

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/zarldev/goenums/file"
)

func TestOSReadWriteFileFS_WriteAndRead(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	osfs := &file.OSReadWriteFileFS{}
	tests := []struct {
		name       string
		path       string
		content    []byte
		err        error
		errPathErr bool
	}{
		{
			name:    "basic write and read",
			path:    "test.txt",
			content: []byte("test content"),
		},
		{
			name:    "empty content",
			path:    "empty.txt",
			content: []byte{},
		},
		{
			name:    "binary content",
			path:    "binary.dat",
			content: []byte{0x00, 0x01, 0xFF, 0xFE},
		},
		{
			name:       "open non-existent file",
			path:       "",
			err:        &fs.PathError{},
			errPathErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fullPath := filepath.Join(tempDir, tt.path)
			if err := osfs.WriteFile(fullPath, tt.content, 0644); err != nil &&
				!errors.Is(err, tt.err) {
				if tt.errPathErr {
					perr := &fs.PathError{}
					if errors.As(err, &perr) {
						return
					}
					t.Errorf("expected a fs.PathError, got %T", err)
					return
				}
				t.Errorf("unexpected write error: %v", err)
				return
			}
			got, err := osfs.ReadFile(fullPath)
			if err != nil {
				t.Errorf("read file error: %v", err)
				return
			}
			if !bytes.Equal(got, tt.content) {
				t.Errorf("content mismatch: got %q, want %q", got, tt.content)
			}
		})
	}
}

func TestOSReadWriteFileFS_Create(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	osfs := &file.OSReadWriteFileFS{}
	tests := []struct {
		name       string
		path       string
		content    []byte
		err        error
		errPathErr bool
	}{
		{
			name:    "create and write",
			path:    "create.txt",
			content: []byte("created content"),
		},
		{
			name:    "create with empty content",
			path:    "create-empty.txt",
			content: []byte{},
		},
		{
			name:       "open non-existent file",
			path:       "",
			content:    []byte("test"),
			err:        fs.ErrNotExist,
			errPathErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fullPath := filepath.Join(tempDir, tt.path)
			w, err := osfs.Create(fullPath)
			if err != nil && !errors.Is(err, tt.err) {
				if tt.errPathErr {
					perr := &fs.PathError{}
					if errors.As(err, &perr) {
						return
					}
					t.Errorf("expected a PathError, got %T", err)
					return
				}
				t.Errorf("unexpected write error: %v", err)
				return
			}
			if _, err := w.Write(tt.content); err != nil {
				t.Errorf("unexpected write error: %q", err)
				return
			}
			if err := w.Close(); err != nil {
				t.Errorf("unexpected close error: %q", err)
				return
			}
			got, err := os.ReadFile(fullPath)
			if err != nil {
				t.Errorf("unexpected read error: %q", err)
				return
			}
			if !bytes.Equal(got, tt.content) {
				t.Errorf("content mismatch: got %q, want %q", got, tt.content)
			}
		})
	}
}

func TestOSReadWriteFileFS_Open(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	osfs := &file.OSReadWriteFileFS{}
	tests := []struct {
		name       string
		path       string
		content    []byte
		setup      func(*testing.T)
		err        error
		errPathErr bool
	}{
		{
			name:    "open existing file",
			path:    "open.txt",
			content: []byte("file to open"),
			setup: func(t *testing.T) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(tempDir, "open.txt"), []byte("file to open"), 0644); err != nil {
					t.Errorf("unexpected setup error %v", err)
					return
				}
			},
		},
		{
			name:       "open non-existent file",
			path:       "nonexistent.txt",
			err:        &fs.PathError{},
			errPathErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.setup != nil {
				tt.setup(t)
			}
			fullPath := filepath.Join(tempDir, tt.path)
			f, err := osfs.Open(fullPath)
			if err != nil && !errors.Is(err, tt.err) {
				if tt.errPathErr {
					perr := &fs.PathError{}
					if errors.As(err, &perr) {
						return
					}
					t.Errorf("expected a fs.PathError, got %T", err)
					return
				}
				t.Errorf("unexpected write error: %v", err)
				return
			}
			got, err := io.ReadAll(f)
			if err != nil {
				t.Errorf("unexpected readall error %q", err)
				return
			}
			if err := f.Close(); err != nil {
				t.Errorf("unexpected close error %q", err)
				return
			}
			if !bytes.Equal(got, tt.content) {
				t.Errorf("content mismatch: got %q, want %q", got, tt.content)
			}
		})
	}
}

func TestOSReadWriteFileFS_Stat(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	osfs := &file.OSReadWriteFileFS{}
	tests := []struct {
		name  string
		path  string
		setup func(t *testing.T, path string)
		isDir bool
		size  int
		err   error
	}{
		{
			name: "stat regular file",
			path: "stat.txt",
			setup: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("stat test"), 0644); err != nil {
					t.Errorf("unexpected setup error: %v", err)
				}
			},
			size: len([]byte("stat test")),
		},
		{
			name: "stat directory",
			path: "statdir",
			setup: func(t *testing.T, path string) {
				if err := os.Mkdir(path, 0755); err != nil {
					t.Errorf("unexpected setup error: %v", err)
				}
			},
			size:  40,
			isDir: true,
		},
		{
			name: "stat non-existent file",
			path: "nonexistent.txt",
			err:  fs.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fullPath := filepath.Join(tempDir, tt.path)
			if tt.setup != nil {
				tt.setup(t, fullPath)
			}
			info, err := osfs.Stat(fullPath)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("unexpected Stat error: %v", err)
					return
				}
				return
			}
			if info.Size() != int64(tt.size) {
				t.Errorf("size mismatch: got %d, want %d", info.Size(), tt.size)
			}
			if info.IsDir() != tt.isDir {
				t.Errorf("is a directory mismatch: got %v, want %v", info.IsDir(), tt.isDir)
			}
		})
	}
}
