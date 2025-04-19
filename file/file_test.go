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
			name: "empty path",
			path: "",
			err:  ErrCreateFile,
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
			content:  `package main func main() {fmt.Println("hello")}`,
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
