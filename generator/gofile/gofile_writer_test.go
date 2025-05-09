// generator/gofile/gofile_writer_test.go
package gofile_test

import (
	"errors"
	"testing"

	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/testdata"
)

func TestWriter_Write(t *testing.T) {
	t.Parallel()
	for _, tt := range testdata.InputOutputTestCases {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			memfs := file.NewMemFS()
			writer := gofile.NewWriter(
				gofile.WithWriterConfiguration(tt.Config),
				gofile.WithFileSystem(memfs))

			err := writer.Write(t.Context(), tt.Representations)
			if err != nil && !errors.Is(err, tt.Err) {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.Validate != nil {
				tt.Validate(t, memfs)
			}
		})
	}
}
