package gofile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/zarldev/goenums/enum"
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
			representations, err := parser.Parse(t.Context())
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
	ctx, cancel := context.WithCancel(t.Context())
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
			_, err := parser.Parse(t.Context())
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

			result, err := parser.Parse(t.Context())
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

// Benchmark tests
func BenchmarkParser_Parse(b *testing.B) {
	parser := gofile.NewParser(
		gofile.WithSource(source.FromFileSystem(testdata.FS, "attributes/planets.go")),
		gofile.WithParserConfiguration(testdata.DefaultConfig),
	)
	ctx := b.Context()

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
