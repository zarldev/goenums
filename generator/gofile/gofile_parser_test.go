package gofile_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/testdata"
	"github.com/zarldev/goenums/source"
)

func TestGoFileParser_Parse(t *testing.T) {
	t.Parallel()
	var testcases []testdata.InputOutputTest
	testcases = append(testcases, testdata.InputOutputTest{
		Name:   "not go code",
		Source: source.FromFileSystem(testdata.FS, "notgocode/notgocode.go"),
		Config: testdata.DefaultConfig,
		Err:    gofile.ErrParseGoSource,
	})
	testcases = append(testcases, testdata.InputOutputTest{
		Name:   "no valid enums found",
		Source: source.FromFileSystem(testdata.FS, "noenums/noenums.go"),
		Config: testdata.DefaultConfig,
		Err:    enum.ErrNoEnumsFound,
	})
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.SetParser(gofile.NewParser(
				gofile.WithParserConfiguration(tc.Config),
				gofile.WithSource(tc.Source)))
			parser := tc.Parser()
			representations, err := parser.Parse(context.Background())
			if !errors.Is(err, tc.Err) {
				t.Errorf("unexpected error: %v expected %v", err, tc.Err)
				return
			}
			if len(representations) != len(tc.GenerationRequests) {
				t.Errorf("expected %d enum representations, got %d", len(tc.GenerationRequests), len(representations))
				return
			}
			for _, rep := range representations {
				if rep.Package == "" {
					t.Errorf("missing package name in %s", tc.Name)
					return
				}
				validateTypeInfo(t, rep.EnumIota)
				if len(rep.EnumIota.Enums) == 0 {
					t.Errorf("no enum values in %s for type %s", tc.Name, rep.EnumIota.Type)
					return
				}
			}
		})
	}
}

func TestGoFileParser_ParseWithCancelContext(t *testing.T) {
	t.Parallel()
	parser := gofile.NewParser(gofile.WithSource(
		source.FromFileSystem(testdata.FS, "planets/planets.go")))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := parser.Parse(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestParser_ErrorHandling(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source enum.Source
		err    error
	}{
		{
			name:   "file not found",
			source: source.FromFile("nonexistent.go"),
			err:    gofile.ErrReadGoSource,
		},
		{
			name:   "invalid go syntax",
			source: source.FromFileSystem(testdata.FS, "notgocode/notgocode.go"),
			err:    gofile.ErrParseGoSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := gofile.NewParser(gofile.WithSource(tt.source))
			_, err := parser.Parse(context.Background())
			if !errors.Is(err, tt.err) {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.err)
			}
		})
	}
}

func TestParser_SpecificScenarios(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		file                 string
		expectedEnumIotas    int
		expectedEnumsInFirst int
		shouldHaveFields     bool
	}{
		{
			name:                 "skip values",
			file:                 "skipvalues/skipvalues.go",
			expectedEnumIotas:    1,
			expectedEnumsInFirst: 4, // V1, V3, V4, V7
			shouldHaveFields:     false,
		},
		{
			name:                 "with attributes",
			file:                 "attributes/planets.go",
			expectedEnumIotas:    1,
			expectedEnumsInFirst: 8,
			shouldHaveFields:     true,
		},
		{
			name:                 "quotes and aliases",
			file:                 "quotes/tickets.go",
			expectedEnumIotas:    1,
			expectedEnumsInFirst: 7, // unknown through completed
			shouldHaveFields:     true,
		},
		{
			name:                 "values only",
			file:                 "values/planets.go",
			expectedEnumIotas:    1,
			expectedEnumsInFirst: 9,
			shouldHaveFields:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := gofile.NewParser(
				gofile.WithSource(source.FromFileSystem(testdata.FS, tt.file)),
				gofile.WithParserConfiguration(testdata.DefaultConfig),
			)

			result, err := parser.Parse(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != tt.expectedEnumIotas {
				t.Errorf("expected %d enum iotas, got %d",
					tt.expectedEnumIotas, len(result))
				return
			}

			if len(result) > 0 && len(result[0].EnumIota.Enums) != tt.expectedEnumsInFirst {
				t.Errorf("expected %d enums in first iota, got %d",
					tt.expectedEnumsInFirst, len(result[0].EnumIota.Enums))
			}

			if len(result) > 0 && tt.shouldHaveFields {
				if len(result[0].EnumIota.Fields) == 0 {
					t.Error("expected enum iota to have fields, but it doesn't")
				}
			}
		})
	}
}

func TestParser_IotaOffsetOperators(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		src        string
		startIndex int
		err        error
	}{
		{
			name: "subtraction",
			src: `package validation
type status int
const (
	failed status = iota - 1
	pending
	active
)`,
			startIndex: -1,
		},
		{
			name: "multiplication",
			src: `package validation
type status int
const (
	failed status = iota * 2
	pending
	active
)`,
			err: gofile.ErrParseGoSource,
		},
		{
			name: "division",
			src: `package validation
type status int
const (
	failed status = iota / 2
	pending
	active
)`,
			err: gofile.ErrParseGoSource,
		},
		{
			name: "unsupported operator returns parse error",
			src: `package validation
type status int
const (
	failed status = iota << 3
	pending
	active
)`,
			err: gofile.ErrParseGoSource,
		},
		{
			name: "addition",
			src: `package validation
type status int
const (
	failed status = iota + 2
	pending
	active
)`,
			startIndex: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := gofile.NewParser(
				gofile.WithSource(source.FromReader(strings.NewReader(tt.src))),
				gofile.WithParserConfiguration(testdata.DefaultConfig),
			)
			result, err := parser.Parse(context.Background())
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("expected error %v, got %v", tt.err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != 1 {
				t.Fatalf("expected 1 generation request, got %d", len(result))
			}
			if result[0].EnumIota.StartIndex != tt.startIndex {
				t.Errorf("expected start index %d, got %d",
					tt.startIndex, result[0].EnumIota.StartIndex)
			}
			if got := result[0].EnumIota.Enums[0].Value; got != tt.startIndex {
				t.Errorf("expected first enum value %d, got %d", tt.startIndex, got)
			}
		})
	}
}

func TestParser_NoTypeDeclarations(t *testing.T) {
	t.Parallel()
	src := `package validation

func helper() {}
`
	parser := gofile.NewParser(
		gofile.WithSource(source.FromReader(strings.NewReader(src))),
		gofile.WithParserConfiguration(testdata.DefaultConfig),
	)
	_, err := parser.Parse(context.Background())
	if !errors.Is(err, enum.ErrNoEnumsFound) {
		t.Errorf("expected ErrNoEnumsFound, got %v", err)
	}
}

func TestParser_IgnoresUnrelatedTypeDeclarations(t *testing.T) {
	t.Parallel()
	src := `package validation

type helper struct { Name string }

type status int

const (
	pending status = iota
	active
)
`
	parser := gofile.NewParser(
		gofile.WithSource(source.FromReader(strings.NewReader(src))),
		gofile.WithParserConfiguration(testdata.DefaultConfig),
	)
	result, err := parser.Parse(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 generation request, got %d", len(result))
	}
	if result[0].EnumIota.Type != "status" {
		t.Fatalf("expected status enum, got %q", result[0].EnumIota.Type)
	}
	if len(result[0].EnumIota.Enums) != 2 {
		t.Fatalf("expected 2 enum values, got %d", len(result[0].EnumIota.Enums))
	}
}

func TestParser_PanicRecoveryReturnsParseError(t *testing.T) {
	t.Parallel()
	parser := gofile.NewParser(gofile.WithSource(panickingSource{}))
	_, err := parser.Parse(context.Background())
	if !errors.Is(err, gofile.ErrParseGoSource) {
		t.Fatalf("expected ErrParseGoSource, got %v", err)
	}
}

type panickingSource struct{}

func (panickingSource) Content() ([]byte, error) { panic("boom") }
func (panickingSource) Filename() string         { return "panic.go" }

// Benchmark tests
func BenchmarkParser_Parse(b *testing.B) {
	parser := gofile.NewParser(
		gofile.WithSource(source.FromFileSystem(testdata.FS, "attributes/planets.go")),
		gofile.WithParserConfiguration(testdata.DefaultConfig),
	)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_, err := parser.Parse(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions
func validateTypeInfo(t *testing.T, info enum.EnumIota) {
	t.Helper()
	if info.Type == "" {
		t.Error("type name is empty")
	}
}

func TestParser_FieldCountPolicy(t *testing.T) {
	t.Parallel()
	const input = `package testpkg

type status int // Label string, Code int

const (
	Pending status = iota // pending
)
`
	for _, tt := range []struct {
		name     string
		failfast bool
		wantErr  bool
	}{
		{name: "marks enum invalid", failfast: false},
		{name: "failfast returns error", failfast: true, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := gofile.NewParser(
				gofile.WithParserConfiguration(config.Configuration{Failfast: tt.failfast}),
				gofile.WithSource(source.FromReader(strings.NewReader(input))),
			)
			requests, err := parser.Parse(context.Background())
			if tt.wantErr {
				if !errors.Is(err, enum.ErrFieldCount) {
					t.Fatalf("Parse() error = %v, want ErrFieldCount", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(requests) != 1 || len(requests[0].EnumIota.Enums) != 1 {
				t.Fatalf("unexpected requests: %#v", requests)
			}
			if requests[0].EnumIota.Enums[0].Valid {
				t.Error("enum with mismatched field count must be invalid")
			}
		})
	}
}
