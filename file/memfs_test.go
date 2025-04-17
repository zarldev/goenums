package file_test

import (
	"bytes"
	"errors"
	"io/fs"
	"testing"

	"github.com/zarldev/goenums/file"
)

func TestMemFS_WriteFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		content  string
		expected string
		setup    func(t *testing.T) *file.MemFS
		err      error
	}{
		{
			name:     "successfully write and read a file",
			filename: "test.txt",
			setup: func(t *testing.T) *file.MemFS {
				mfs := file.NewMemFS()
				_, err := mfs.Create("test.txt")
				if err != nil {
					t.Errorf("creating file %v", err)
					return nil
				}
				return mfs
			},
			content:  "test content",
			expected: "test content",
		},
		{
			name:     "file does not exist",
			filename: "nonexistent.txt",
			content:  "",
			expected: "",
			err:      fs.ErrInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := file.NewMemFS()
			if tt.setup != nil {
				fs = tt.setup(t)
			}
			if fs == nil {
				t.Errorf("setup failed")
				return
			}
			if err := fs.WriteFile(tt.filename, []byte(tt.content), 0644); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}
}

func TestMemFS_ReadFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fs       file.ReadWriteCreateFileFS
		filename string
		content  string
		err      error
	}{
		{
			name: "successfully write and read a file",
			fs: func() file.ReadWriteCreateFileFS {
				fs := file.NewMemFS()
				err := fs.WriteFile("test", []byte("test content"), 0644)
				if err != nil {
					t.Errorf("creating test file %q", err.Error())
					return nil
				}
				return fs
			}(),
			filename: "test",
			content:  "test content",
		},
		{
			name:     "read non-existent file",
			fs:       func() file.ReadWriteCreateFileFS { return file.NewMemFS() }(),
			filename: "",
			content:  "",
			err:      fs.ErrInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content, err := tt.fs.ReadFile(tt.filename)
			if !errors.Is(err, tt.err) {
				t.Errorf("unexpected error: %s", err.Error())
				return
			}
			if !bytes.Equal(content, []byte(tt.content)) {
				t.Errorf("unexpected content: %s", content)
			}
		})
	}
}
