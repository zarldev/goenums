package file_test

import (
	"bytes"
	"errors"
	"io"
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

func TestMemFile_ReadWrite(t *testing.T) {
	t.Parallel()
	fs := file.NewMemFS()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "write and read from memFile",
			testFunc: func(t *testing.T) {
				f, err := fs.Create("test.txt")
				if err != nil {
					t.Errorf("Create failed: %v", err)
					return
				}

				// Test Write
				data := []byte("test content")
				n, err := f.Write(data)
				if err != nil {
					t.Errorf("Write failed: %v", err)
					return
				}
				if n != len(data) {
					t.Errorf("Write returned %d, expected %d", n, len(data))
				}

				// Test Close
				if err := f.Close(); err != nil {
					t.Errorf("Close failed: %v", err)
				}

				// Test Read through Open
				f2, err := fs.Open("test.txt")
				if err != nil {
					t.Errorf("Open failed: %v", err)
					return
				}

				readBuf := make([]byte, len(data))
				n, err = f2.Read(readBuf)
				if err != nil {
					t.Errorf("Read failed: %v", err)
					return
				}
				if n != len(data) {
					t.Errorf("Read returned %d, expected %d", n, len(data))
				}
				if !bytes.Equal(readBuf, data) {
					t.Errorf("Read data mismatch: got %q, want %q", readBuf, data)
				}

				// Test Stat on file
				info, err := f2.Stat()
				if err != nil {
					t.Errorf("Stat failed: %v", err)
					return
				}
				if info.Size() != int64(len(data)) {
					t.Errorf("Stat size mismatch: got %d, want %d", info.Size(), len(data))
				}
				if info.Name() != "test.txt" {
					t.Errorf("Stat name mismatch: got %q, want %q", info.Name(), "test.txt")
				}
				if info.IsDir() {
					t.Errorf("Stat IsDir should be false")
				}
				if info.Mode() != 0644 {
					t.Errorf("Stat Mode mismatch: got %v, want %v", info.Mode(), 0644)
				}

				f2.Close()
			},
		},
		{
			name: "memFile with empty content",
			testFunc: func(t *testing.T) {
				// Test reading from empty file
				emptyFile, err := fs.Create("empty.txt")
				if err != nil {
					t.Errorf("Create empty file failed: %v", err)
					return
				}
				emptyFile.Close()

				openEmpty, err := fs.Open("empty.txt")
				if err != nil {
					t.Errorf("Open empty file failed: %v", err)
					return
				}

				buf := make([]byte, 10)
				n, err := openEmpty.Read(buf)
				if n != 0 {
					t.Errorf("Read from empty file returned %d bytes, expected 0", n)
				}
				if !errors.Is(err, io.EOF) {
					t.Errorf("Read from empty file should return EOF, got %v", err)
				}

				openEmpty.Close()
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

func TestMemFileInfo_Methods(t *testing.T) {
	t.Parallel()
	fs := file.NewMemFS()

	// Create a file to test FileInfo methods
	err := fs.WriteFile("info-test.txt", []byte("test content for info"), 0644)
	if err != nil {
		t.Errorf("WriteFile failed: %v", err)
		return
	}

	info, err := fs.Stat("info-test.txt")
	if err != nil {
		t.Errorf("Stat failed: %v", err)
		return
	}

	// Test all FileInfo methods
	if info.Name() != "info-test.txt" {
		t.Errorf("Name() = %q, want %q", info.Name(), "info-test.txt")
	}

	if info.Size() != 21 {
		t.Errorf("Size() = %d, want %d", info.Size(), 21)
	}

	if info.Mode() != 0644 {
		t.Errorf("Mode() = %v, want %v", info.Mode(), 0644)
	}

	if info.IsDir() {
		t.Errorf("IsDir() = true, want false")
	}

	if info.Sys() != nil {
		t.Errorf("Sys() = %v, want nil", info.Sys())
	}

	// ModTime should return a time (we just check it's not zero)
	if info.ModTime().IsZero() {
		t.Errorf("ModTime() returned zero time")
	}
}
