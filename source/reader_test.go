package source_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/zarldev/goenums/source"
)

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
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
