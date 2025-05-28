package generator_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/testdata"
	"github.com/zarldev/goenums/source"
)

func TestParserContract_Parse(t *testing.T) {
	t.Parallel()
	contractTests := slices.Clone(testdata.InputOutputTestCases)
	testcase := struct {
		Name   string
		Parser func(cfg config.Configuration, src enum.Source) enum.Parser
		Err    error
	}{
		Name: "gofile parser full contract test",
		Parser: func(cfg config.Configuration, src enum.Source) enum.Parser {
			return gofile.NewParser(
				gofile.WithParserConfiguration(cfg),
				gofile.WithSource(src))
		},
		Err: enum.ErrParseValue,
	}
	t.Run(testcase.Name, func(t *testing.T) {
		t.Parallel()
		for _, ct := range contractTests {
			ct.SetParser(testcase.Parser(ct.Config, ct.Source))
			t.Run(ct.Name, func(t *testing.T) {
				t.Parallel()
				parser := ct.Parser()
				representations, err := parser.Parse(t.Context())
				if !errors.Is(err, ct.Err) {
					t.Errorf("unexpected error: %v expected %v", err, ct.Err)
					return
				}
				if len(representations) != len(ct.GenerationRequests) {
					t.Errorf("expected %d enum representations, got %d", len(ct.GenerationRequests), len(representations))
					return
				}
				for _, rep := range representations {
					if rep.Package == "" {
						t.Errorf("missing package name in %s", ct.Name)
						return
					}
					validateTypeInfo(t, rep.EnumIota)
					if len(rep.EnumIota.Enums) == 0 {
						t.Errorf("no enum values in %s for type %s", ct.Name, rep.EnumIota.Type)
						return
					}
				}
			})
			t.Run(ct.Name+" with canceled context", func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				_, err := ct.Parser().Parse(ctx)
				if err != nil && !errors.Is(err, context.Canceled) {
					t.Errorf("expected context.Canceled error, got %v", err)
				}
			})
		}
	})
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
			expectedEnumsInFirst: 9, // unknown through neptune
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
			expectedEnumsInFirst: 9, // unknown through neptune
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
		gofile.WithSource(source.FromFileSystem(testdata.FS, "planets/planets.go")),
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
