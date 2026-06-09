// generator/gofile/gofile_writer_test.go
package gofile_test

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/testdata"
)

func TestWriter_Write(t *testing.T) {
	t.Parallel()
	testcases := slices.Clone(testdata.InputOutputTestCases)
	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			memfs := file.NewMemFS()
			writer := gofile.NewWriter(
				gofile.WithWriterConfiguration(tt.Config),
				gofile.WithFileSystem(memfs))

			err := writer.Write(t.Context(), tt.GenerationRequests)
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

// TestWriter_CaseInsensitivePreservesStringOutput asserts that when Insensitive
// is true, String() still returns the original alias casing, while the parse
// map keys are lower-cased.
func TestWriter_CaseInsensitivePreservesStringOutput(t *testing.T) {
	t.Parallel()
	memfs := file.NewMemFS()
	writer := gofile.NewWriter(
		gofile.WithWriterConfiguration(config.Configuration{Insensitive: true}),
		gofile.WithFileSystem(memfs))

	req := enum.GenerationRequest{
		Package:        "testpkg",
		SourceFilename: "status.go",
		OutputFilename: "status",
		Version:        "test",
		Configuration:  config.Configuration{Insensitive: true},
		EnumIota: enum.EnumIota{
			Type:       "status",
			StartIndex: 1,
			Enums: []enum.Enum{
				{Name: "pending", Index: 0, Aliases: []string{"Pending"}, Valid: true},
				{Name: "failed", Index: 1, Aliases: []string{"Failed"}, Valid: true},
			},
		},
	}

	if err := writer.Write(t.Context(), []enum.GenerationRequest{req}); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	content, err := memfs.ReadFile("status_enums.go")
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	s := string(content)

	// String() output must preserve the original alias casing.
	if !strings.Contains(s, `const statusNames = "PendingFailed"`) {
		t.Errorf("String() names constant should preserve original alias casing, got:\n%s", s)
	}

	// Parse map must use lower-cased keys for case-insensitive matching.
	if !strings.Contains(s, `"pending":`) {
		t.Errorf("parse map should contain lower-cased key 'pending'")
	}
	if !strings.Contains(s, `"failed":`) {
		t.Errorf("parse map should contain lower-cased key 'failed'")
	}

	// Must NOT contain the original casing in the parse map (that would mean
	// insensitive mode wasn't applied to the map keys).
	if strings.Contains(s, `"Pending":`) || strings.Contains(s, `"Failed":`) {
		lines := strings.Split(s, "\n")
		for _, line := range lines {
			if strings.Contains(line, `"Pending":`) || strings.Contains(line, `"Failed":`) {
				// Allow the names constant and names map which correctly use original casing
				if !strings.Contains(line, `statusNames`) {
					t.Errorf("parse map should not contain original-cased key: %s", line)
				}
			}
		}
	}
}

// TestWriter_ConstraintsNoDuplicateTypes asserts that when Constraints is true
// and multiple enums are generated in the same package, the output does not
// declare package-level types (float, integer, number) that would collide.
func TestWriter_ConstraintsNoDuplicateTypes(t *testing.T) {
	t.Parallel()
	memfs := file.NewMemFS()
	writer := gofile.NewWriter(
		gofile.WithWriterConfiguration(config.Configuration{Constraints: true}),
		gofile.WithFileSystem(memfs))

	reqs := []enum.GenerationRequest{
		{
			Package:        "testpkg",
			SourceFilename: "a.go",
			OutputFilename: "a",
			Version:        "test",
			Configuration:  config.Configuration{Constraints: true},
			EnumIota: enum.EnumIota{
				Type:       "alpha",
				StartIndex: 0,
				Enums:      []enum.Enum{{Name: "A", Index: 0, Valid: true}},
			},
		},
		{
			Package:        "testpkg",
			SourceFilename: "b.go",
			OutputFilename: "b",
			Version:        "test",
			Configuration:  config.Configuration{Constraints: true},
			EnumIota: enum.EnumIota{
				Type:       "beta",
				StartIndex: 0,
				Enums:      []enum.Enum{{Name: "B", Index: 0, Valid: true}},
			},
		},
	}

	if err := writer.Write(t.Context(), reqs); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	for _, name := range []string{"a_enums.go", "b_enums.go"} {
		content, err := memfs.ReadFile(name)
		if err != nil {
			t.Fatalf("read generated file %s: %v", name, err)
		}

		s := string(content)

		// Must NOT emit duplicate package-level constraint types.
		forbidden := []string{
			"type float interface",
			"type integer interface",
			"type number interface",
		}
		for _, f := range forbidden {
			if strings.Contains(s, f) {
				t.Errorf("%s must not contain %q (causes duplicate-type errors when multiple enums share a package)", name, f)
			}
		}

		// Must inline the constraint in the generic function instead.
		if !strings.Contains(s, "func numberTo") {
			t.Errorf("%s should contain a numberTo generic function", name)
		}
		if !bytes.Contains(content, []byte("int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64")) {
			t.Errorf("%s should inline the numeric constraint directly in the generic function signature", name)
		}
	}
}
