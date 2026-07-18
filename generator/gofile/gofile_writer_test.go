// generator/gofile/gofile_writer_test.go
package gofile_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
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

			err := writer.Write(context.Background(), tt.GenerationRequests)
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

	if err := writer.Write(context.Background(), []enum.GenerationRequest{req}); err != nil {
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

	if err := writer.Write(context.Background(), reqs); err != nil {
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

func TestWriter_MatchUsesCompileTimeMatcherInterface(t *testing.T) {
	t.Parallel()
	memfs := file.NewMemFS()
	writer := gofile.NewWriter(gofile.WithFileSystem(memfs))

	req := enum.GenerationRequest{
		Package:        "testpkg",
		SourceFilename: "status.go",
		OutputFilename: "status",
		Version:        "test",
		EnumIota: enum.EnumIota{
			Type:       "status",
			StartIndex: 1,
			Enums: []enum.Enum{
				{Name: "unknown", Index: 0, Aliases: []string{"Unknown"}, Valid: false},
				{Name: "pending", Index: 1, Aliases: []string{"Pending"}, Valid: true},
				{Name: "failed", Index: 2, Aliases: []string{"Failed"}, Valid: true},
			},
		},
	}

	if err := writer.Write(context.Background(), []enum.GenerationRequest{req}); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	content, err := memfs.ReadFile("status_enums.go")
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	s := string(content)

	want := []string{
		"type StatusMatcher interface",
		"Pending()",
		"Failed()",
		"func MatchStatus(en Status, matcher StatusMatcher) error",
		"case Statuses.PENDING:",
		"matcher.Pending()",
		"case Statuses.FAILED:",
		"matcher.Failed()",
		`return fmt.Errorf("unhandled Status: %v", en)`,
		"func MustMatchStatus(en Status, matcher StatusMatcher)",
	}
	for _, fragment := range want {
		if !strings.Contains(s, fragment) {
			t.Errorf("generated output missing %q", fragment)
		}
	}

	if strings.Contains(s, "Unknown()") {
		t.Errorf("invalid enum values should not be part of the matcher interface")
	}
	if strings.Contains(s, "map[Status]func()") {
		t.Errorf("MatchStatus should use the matcher interface, not a handlers map")
	}
}

var errWriterFailure = errors.New("writer failure")

type failingWriteFS struct {
	*file.MemFS
}

func (f *failingWriteFS) Create(string) (io.WriteCloser, error) {
	return failingWriteCloser{}, nil
}

func (f *failingWriteFS) WriteFile(string, []byte, fs.FileMode) error {
	return errWriterFailure
}

type failingWriteCloser struct{}

func (failingWriteCloser) Write([]byte) (int, error) {
	return 0, errWriterFailure
}

func (failingWriteCloser) Close() error {
	return nil
}

func TestWriter_WriteReturnsOutputError(t *testing.T) {
	t.Parallel()
	fs := &failingWriteFS{MemFS: file.NewMemFS()}
	writer := gofile.NewWriter(gofile.WithFileSystem(fs))

	err := writer.Write(context.Background(), []enum.GenerationRequest{benchmarkRequest(1)})
	if !errors.Is(err, errWriterFailure) {
		t.Fatalf("Write() error = %v, want %v", err, errWriterFailure)
	}
}

func TestWriter_SerializationContracts(t *testing.T) {
	t.Parallel()
	fs := file.NewMemFS()
	request := benchmarkRequest(1)
	writer := gofile.NewWriter(gofile.WithFileSystem(fs))
	if err := writer.Write(context.Background(), []enum.GenerationRequest{request}); err != nil {
		t.Fatal(err)
	}
	content, err := fs.ReadFile("statuses_enums.go")
	if err != nil {
		t.Fatal(err)
	}
	generated := string(content)
	for _, want := range []string{
		"return json.Marshal(s.String())",
		"if err := json.Unmarshal(by, &value); err == nil",
		"ParseStatus(by)",
		"return []byte(s.String()), nil",
	} {
		if !strings.Contains(generated, want) {
			t.Errorf("generated output missing %q", want)
		}
	}
}

func TestWriter_ImportsFollowHandlers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		handlers  config.Handlers
		want      []string
		doNotWant []string
	}{
		{
			name:      "json only",
			handlers:  config.Handlers{JSON: true},
			want:      []string{`"encoding/json"`},
			doNotWant: []string{`"database/sql/driver"`},
		},
		{
			name:      "sql only",
			handlers:  config.Handlers{SQL: true},
			want:      []string{`"database/sql/driver"`},
			doNotWant: []string{`"encoding/json"`},
		},
		{
			name:      "no handlers",
			doNotWant: []string{`"encoding/json"`, `"database/sql/driver"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := file.NewMemFS()
			request := benchmarkRequest(1)
			request.Configuration.Handlers = tt.handlers
			writer := gofile.NewWriter(gofile.WithFileSystem(fs))
			if err := writer.Write(context.Background(), []enum.GenerationRequest{request}); err != nil {
				t.Fatal(err)
			}
			content, err := fs.ReadFile("statuses_enums.go")
			if err != nil {
				t.Fatal(err)
			}
			generated := string(content)
			for _, want := range tt.want {
				if !strings.Contains(generated, want) {
					t.Errorf("generated output missing import %s", want)
				}
			}
			for _, unwanted := range tt.doNotWant {
				if strings.Contains(generated, unwanted) {
					t.Errorf("generated output unexpectedly contains import %s", unwanted)
				}
			}
		})
	}
}

func TestWriter_LegacyTextMarshalCompatibility(t *testing.T) {
	t.Parallel()
	fs := file.NewMemFS()
	request := benchmarkRequest(1)
	request.Configuration.LegacyTextMarshal = true
	writer := gofile.NewWriter(gofile.WithFileSystem(fs))
	if err := writer.Write(context.Background(), []enum.GenerationRequest{request}); err != nil {
		t.Fatal(err)
	}
	content, err := fs.ReadFile("statuses_enums.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `return []byte("\"" + s.String() + "\"")`) {
		t.Fatal("legacy text marshal output did not include quoted MarshalText")
	}
}

func TestWriter_GeneratesSelectedInterfaces(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		configuration  config.Configuration
		present        []string
		absent         []string
		presentImports []string
		absentImports  []string
	}{
		{
			name:           "json and sql only",
			configuration:  config.Configuration{Handlers: config.Handlers{JSON: true, SQL: true}},
			present:        []string{"MarshalJSON", "UnmarshalJSON", "Scan(value any)", "Value() (driver.Value, error)"},
			absent:         []string{"MarshalText", "UnmarshalText", "MarshalYAML", "UnmarshalYAML", "MarshalBinary", "UnmarshalBinary"},
			presentImports: []string{`"encoding/json"`, `"database/sql/driver"`},
		},
		{
			name:          "text only",
			configuration: config.Configuration{Handlers: config.Handlers{Text: true}},
			present:       []string{"MarshalText", "UnmarshalText"},
			absent:        []string{"MarshalJSON", "UnmarshalJSON", "MarshalYAML", "UnmarshalYAML", "MarshalBinary", "UnmarshalBinary", "Scan(value any)", "Value() (driver.Value, error)"},
			absentImports: []string{`"encoding/json"`, `"database/sql/driver"`},
		},
		{
			name:          "text only legacy marshal",
			configuration: config.Configuration{LegacyTextMarshal: true, Handlers: config.Handlers{Text: true}},
			present:       []string{"MarshalText", "UnmarshalText", `return []byte("\"" + s.String() + "\"")`},
			absent:        []string{"MarshalJSON", "UnmarshalJSON", "MarshalYAML", "UnmarshalYAML", "MarshalBinary", "UnmarshalBinary", "Scan(value any)", "Value() (driver.Value, error)"},
		},
		{
			name:          "yaml only",
			configuration: config.Configuration{Handlers: config.Handlers{YAML: true}},
			present:       []string{"MarshalYAML", "UnmarshalYAML"},
			absent:        []string{"MarshalJSON", "UnmarshalJSON", "MarshalText", "UnmarshalText", "MarshalBinary", "UnmarshalBinary", "Scan(value any)", "Value() (driver.Value, error)"},
		},
		{
			name:          "binary only",
			configuration: config.Configuration{Handlers: config.Handlers{Binary: true}},
			present:       []string{"MarshalBinary", "UnmarshalBinary"},
			absent:        []string{"MarshalJSON", "UnmarshalJSON", "MarshalText", "UnmarshalText", "MarshalYAML", "UnmarshalYAML", "Scan(value any)", "Value() (driver.Value, error)"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := file.NewMemFS()
			request := benchmarkRequest(1)
			request.Configuration = tt.configuration
			writer := gofile.NewWriter(gofile.WithFileSystem(fs))
			if err := writer.Write(context.Background(), []enum.GenerationRequest{request}); err != nil {
				t.Fatal(err)
			}
			content, err := fs.ReadFile("statuses_enums.go")
			if err != nil {
				t.Fatal(err)
			}
			generated := string(content)
			for _, want := range append(tt.present, tt.presentImports...) {
				if !strings.Contains(generated, want) {
					t.Errorf("generated output missing %q", want)
				}
			}
			for _, unwanted := range append(tt.absent, tt.absentImports...) {
				if strings.Contains(generated, unwanted) {
					t.Errorf("generated output unexpectedly contains %q", unwanted)
				}
			}
		})
	}
}

func TestWriter_AliasEscapingAndCollisions(t *testing.T) {
	t.Parallel()
	t.Run("escapes aliases as Go string literals", func(t *testing.T) {
		t.Parallel()
		fs := file.NewMemFS()
		request := benchmarkRequest(1)
		request.EnumIota.Enums[0].Aliases = []string{"ready_to_ship", "ready-to-ship", "quote\"value", "line\nvalue"}
		writer := gofile.NewWriter(gofile.WithFileSystem(fs))
		if err := writer.Write(context.Background(), []enum.GenerationRequest{request}); err != nil {
			t.Fatal(err)
		}
		content, err := fs.ReadFile("statuses_enums.go")
		if err != nil {
			t.Fatal(err)
		}
		generated := string(content)
		for _, want := range []string{`"ready_to_ship"`, `"ready-to-ship"`, `"quote\"value"`, `"line\nvalue"`} {
			if !strings.Contains(generated, want) {
				t.Errorf("generated output missing escaped alias %s", want)
			}
		}
	})

	t.Run("rejects duplicate aliases", func(t *testing.T) {
		t.Parallel()
		request := benchmarkRequest(2)
		request.EnumIota.Enums[0].Aliases = []string{"same-alias"}
		request.EnumIota.Enums[1].Aliases = []string{"same-alias"}
		writer := gofile.NewWriter(gofile.WithFileSystem(file.NewMemFS()))
		err := writer.Write(context.Background(), []enum.GenerationRequest{request})
		if err == nil {
			t.Fatal("Write() accepted duplicate aliases")
		}
		if !strings.Contains(err.Error(), `duplicate enum alias "same-alias"`) {
			t.Fatalf("Write() error = %v", err)
		}
	})

	t.Run("case insensitive duplicate aliases", func(t *testing.T) {
		t.Parallel()
		request := benchmarkRequest(2)
		request.Configuration.Insensitive = true
		request.EnumIota.Enums[0].Aliases = []string{"READY_TO_SHIP"}
		request.EnumIota.Enums[1].Aliases = []string{"ready_to_ship"}
		writer := gofile.NewWriter(gofile.WithFileSystem(file.NewMemFS()))
		if err := writer.Write(context.Background(), []enum.GenerationRequest{request}); err == nil {
			t.Fatal("Write() accepted case-insensitive duplicate aliases")
		}
	})
}
