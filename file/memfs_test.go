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
		fs       file.ReadCreateWriteFileFS
		filename string
		content  string
		err      error
	}{
		{
			name: "successfully write and read a file",
			fs: func() file.ReadCreateWriteFileFS {
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
			fs:       func() file.ReadCreateWriteFileFS { return file.NewMemFS() }(),
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

func TestMemFS_Stat(t *testing.T) {
	t.Parallel()
	const content = "test content"
	const filename = "test"
	tests := []struct {
		name     string
		fs       file.ReadCreateWriteFileFS
		filename string
		err      error
	}{
		{
			name: "successfully stat a file",
			fs: func() file.ReadCreateWriteFileFS {
				fs := file.NewMemFS()
				err := fs.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Errorf("creating test file %q", err.Error())
					return nil
				}
				return fs
			}(),
			filename: filename,
		},
		{
			name:     "stat empty path",
			fs:       func() file.ReadCreateWriteFileFS { return file.NewMemFS() }(),
			filename: "",
			err:      fs.ErrInvalid,
		},
		{
			name:     "stat non-existent file",
			fs:       func() file.ReadCreateWriteFileFS { return file.NewMemFS() }(),
			filename: "non-existent",
			err:      fs.ErrNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fi, err := tt.fs.Stat(tt.filename)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("unexpected error: %s", err.Error())
					return
				}
				return
			}
			if fi == nil {
				t.Errorf("unexpected nil FileInfo")
				return
			}
			if fi.Name() != tt.filename {
				t.Errorf("unexpected FileInfo name: %s", fi.Name())
				return
			}
			if fi.Size() != int64(len(content)) {
				t.Errorf("unexpected FileInfo size: %d", fi.Size())
				return
			}
		})
	}
}

func TestMemFS_Create(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fs       file.ReadCreateWriteFileFS
		filename string
		err      error
	}{
		{
			name:     "successfully create a file",
			fs:       func() file.ReadCreateWriteFileFS { return file.NewMemFS() }(),
			filename: "test",
		},
		{
			name:     "create file with empty name",
			fs:       func() file.ReadCreateWriteFileFS { return file.NewMemFS() }(),
			filename: "",
			err:      fs.ErrInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.fs.Create(tt.filename)
			if !errors.Is(err, tt.err) {
				t.Errorf("unexpected error: %s", err.Error())
				return
			}
		})
	}
}

func TestMemFS_OpenAndStat(t *testing.T) {
	t.Parallel()
	const content = "test content"
	const filename = "test"
	tests := []struct {
		name     string
		fs       file.ReadCreateWriteFileFS
		filename string
		err      error
	}{
		{
			name: "successfully open and stat a file",
			fs: func() file.ReadCreateWriteFileFS {
				fsys := file.NewMemFS()
				err := fsys.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Errorf("creating test file %q", err.Error())
					return nil
				}
				return fsys
			}(),
			filename: filename,
			err:      nil,
		},
		{
			name:     "open empty path",
			fs:       func() file.ReadCreateWriteFileFS { return file.NewMemFS() }(),
			filename: "",
			err:      fs.ErrInvalid,
		},
		{
			name:     "open non-existent file",
			fs:       func() file.ReadCreateWriteFileFS { return file.NewMemFS() }(),
			filename: "non-existent",
			err:      fs.ErrNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			file, err := tt.fs.Open(tt.filename)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("unexpected error: %s", err.Error())
					return
				}
				return
			}
			if file == nil {
				t.Errorf("unexpected nil File")
				return
			}
			fi, err := file.Stat()
			if err != nil {
				t.Errorf("unexpected error: %s", err.Error())
				return
			}
			if fi == nil {
				t.Errorf("unexpected nil FileInfo")
				return
			}
			if fi.Name() != tt.filename {
				t.Errorf("unexpected FileInfo name: %s", fi.Name())
				return
			}
			if fi.Size() != int64(len(content)) {
				t.Errorf("unexpected FileInfo size: %d", fi.Size())
				return
			}
			if fi.Mode() != 0644 {
				t.Errorf("unexpected FileInfo mode: %d", fi.Mode())
				return
			}
		})
	}
}

func TestMemFS_Open(t *testing.T) {
	t.Parallel()
	fsys := file.NewMemFS()
	tests := []struct {
		name     string
		fs       file.ReadCreateWriteFileFS
		filename string
		err      error
	}{
		{
			name: "successfully open a file",
			fs: func() file.ReadCreateWriteFileFS {
				fsys.WriteFile("test", []byte("test content"), 0644)
				return fsys
			}(),
			filename: "test",
			err:      nil,
		},
		{
			name:     "open empty path",
			fs:       func() file.ReadCreateWriteFileFS { return file.NewMemFS() }(),
			filename: "",
			err:      fs.ErrInvalid,
		},
		{
			name:     "open non-existent file",
			fs:       func() file.ReadCreateWriteFileFS { return file.NewMemFS() }(),
			filename: "non-existent",
			err:      fs.ErrNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.fs.Open(tt.filename)
			if !errors.Is(err, tt.err) {
				t.Errorf("unexpected error: %s", err.Error())
				return
			}
		})
	}
}
