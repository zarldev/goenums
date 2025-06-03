package file_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/zarldev/goenums/file"
)

func TestWriteToFileAndFormatFS(t *testing.T) {
	t.Parallel()
	successWriteFunc := func(s string) func(io.Writer) error {
		return func(w io.Writer) error {
			_, err := io.WriteString(w, s)
			return err
		}
	}

	failureWriteFunc := func(string) func(io.Writer) error {
		return func(io.Writer) error {
			return errors.New("expected error")
		}
	}

	tests := []struct {
		name      string
		path      string
		writeFunc func(io.Writer) error
		ctx       func() context.Context
		format    bool
		expected  string
		err       error
	}{
		{
			name: "empty path",
			path: "",
			err:  file.ErrCreateFile,
		},
		{
			name: "nil write func",
			path: "test.go",
			err:  file.ErrWriteFile,
		},
		{
			name:      "valid go file without formatting",
			path:      "test.go",
			format:    false,
			writeFunc: successWriteFunc("package main\nfunc main() {\nfmt.Println(\"hello\")\n}\n"),
			expected:  "package main\nfunc main() {\nfmt.Println(\"hello\")\n}\n",
		},
		{
			name:      "valid go file with formatting",
			path:      "test.go",
			writeFunc: successWriteFunc("package main\nfunc main() {\nfmt.Println(\"hello\")\n}\n"),
			format:    true,
			expected:  "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n",
		},
		{
			name:      "invalid go file with formatting",
			path:      "invalid.go",
			writeFunc: successWriteFunc("this is not valid go code"),
			format:    true,
			err:       file.ErrFormatFile,
		},
		{
			name:      "invalid go file without formatting",
			path:      "invalid.go",
			writeFunc: successWriteFunc("this is not valid go code"),
			expected:  "this is not valid go code",
			format:    false,
		},
		{
			name:      "empty content",
			path:      "empty.go",
			writeFunc: successWriteFunc(""),
			format:    false,
			err:       nil,
		},
		{
			name:      "write function error",
			path:      "write-error.go",
			writeFunc: failureWriteFunc("write error"),
			err:       file.ErrWriteFile,
		},
		{
			name:      "context cancelled before format",
			path:      "context-cancel.go",
			writeFunc: successWriteFunc("package main\nfunc main() {\nfmt.Println(\"hello\")\n}\n"),
			format:    true,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			err: context.Canceled,
		},
		{
			name:      "context cancelled before create",
			path:      "context-cancel-create.go",
			writeFunc: successWriteFunc("package main\nfunc main() {\nfmt.Println(\"hello\")\n}\n"),
			format:    false,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			err: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			if tt.ctx != nil {
				ctx = tt.ctx()
			}
			fs := file.NewMemFS()
			err := file.WriteToFileAndFormatFS(ctx, fs, tt.path, tt.format, tt.writeFunc)
			if err != nil {
				if tt.err != nil && !errors.Is(err, tt.err) {
					t.Errorf("unexpected error: %v", err)
					return
				}
				return
			}
			got, err := fs.ReadFile(tt.path)
			if err != nil {
				t.Errorf("failed to read file: %v", err)
				return
			}
			gotContent := string(got)
			if tt.format {
				if !strings.Contains(gotContent, "package main") {
					t.Errorf(`formatted content missing expected text:
					got: %q
					want to contain: package main`, gotContent)
				}
				return
			}
			if gotContent != tt.expected {
				t.Errorf(`content mismatch:
					got:  %q
					want: %q`, gotContent, tt.expected)
			}
		})
	}
}
