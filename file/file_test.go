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
	tests := []struct {
		name     string
		path     string
		content  string
		format   bool
		expected string
		err      error
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
			name:     "valid go file without formatting",
			path:     "test.go",
			content:  "package main\nfunc main() {\nfmt.Println(\"hello\")\n}\n",
			format:   false,
			expected: "package main\nfunc main() {\nfmt.Println(\"hello\")\n}\n",
		},
		{
			name:     "valid go file with formatting",
			path:     "test.go",
			content:  `package main func main() {fmt.Println("hello")}`,
			format:   true,
			expected: "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n",
		},
		{
			name:    "invalid go file with formatting",
			path:    "invalid.go",
			content: "this is not valid go code",
			format:  true,
			err:     file.ErrFormatFile,
		},
		{
			name:     "invalid go file without formatting",
			path:     "invalid.go",
			content:  "this is not valid go code",
			expected: "this is not valid go code",
			format:   false,
		},
		{
			name:    "empty content",
			path:    "empty.go",
			content: "",
			format:  false,
			err:     nil,
		},
		{
			name:    "write function error",
			path:    "write-error.go",
			content: "", // Will be ignored since writeFunc will error
			format:  false,
			err:     file.ErrWriteFile,
		},
		{
			name:    "context cancelled before format",
			path:    "context-cancel.go",
			content: "package main\nfunc main() {}",
			format:  true,
			err:     nil, // Will be handled specially
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := file.NewMemFS()

			var writeFunc func(io.Writer) error
			if tt.name == "write function error" {
				// Test case for write function error
				writeFunc = func(w io.Writer) error {
					return errors.New("simulated write error")
				}
			} else if tt.name == "nil write func" {
				// Test case for nil writeFunc
				writeFunc = nil
			} else {
				writeFunc = func(w io.Writer) error {
					_, err := io.WriteString(w, tt.content)
					return err
				}
			}

			// Handle context cancellation test case
			ctx := t.Context()
			if tt.name == "context cancelled before format" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				// Write the file first, then cancel context before format
				tempWriteFunc := func(w io.Writer) error {
					_, err := io.WriteString(w, tt.content)
					cancel() // Cancel context after write but before format
					return err
				}
				err := file.WriteToFileAndFormatFS(ctx, fs, tt.path, tt.format, tempWriteFunc)
				// Should get context.Canceled error
				if err == nil || !errors.Is(err, context.Canceled) {
					t.Errorf("expected context.Canceled error, got %v", err)
				}
				return
			}

			err := file.WriteToFileAndFormatFS(ctx, fs, tt.path, tt.format, writeFunc)
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
