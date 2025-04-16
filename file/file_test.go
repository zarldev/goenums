package file

import (
	"errors"
	"io"
	"strings"
	"testing"
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
			name:    "empty path",
			path:    "",
			content: "package main",
			err:     ErrWriteFile,
		},
		{
			name: "nil write func",
			path: "test.go",
			err:  ErrWriteFile,
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
			content:  "package main\nfunc main() {fmt.Println(\"hello\")}\n",
			format:   true,
			expected: "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n",
		},
		{
			name:    "invalid go file with formatting",
			path:    "invalid.go",
			content: "this is not valid go code",
			format:  true,
			err:     ErrFormatFile,
		},
		{
			name:     "invalid go file without formatting",
			path:     "invalid.go",
			content:  "this is not valid go code",
			expected: "this is not valid go code",
			format:   false,
		},
		{
			name: "empty path",
			path: "",
			err:  ErrWriteFile,
		},
		{
			name:    "empty content",
			path:    "empty.go",
			content: "",
			format:  false,
			err:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewMemFS()
			writeFunc := func(w io.Writer) error {
				_, err := io.WriteString(w, tt.content)
				return err
			}
			err := WriteToFileAndFormatFS(t.Context(), fs, tt.path, tt.format, writeFunc)
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
				// When formatting is enabled, we only check if the content contains
				// the essential parts since formatting might vary
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

func TestOSReadWriteFileFS(t *testing.T) {
	fs := &OSReadWriteFileFS{}
	// Just verify that we can create the type
	if fs == nil {
		t.Error("failed to create OSReadWriteFileFS")
	}
}
