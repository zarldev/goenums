package gofile_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/testdata"
	"github.com/zarldev/goenums/source"
)

func TestGoFileParser_Parse(t *testing.T) {
	t.Parallel()
	testcases := slices.Clone(testdata.InputOutputTestCases)
	testcases = append(testcases, testdata.InputOutputTest{
		Name:   "invalid file",
		Source: source.FromFileSystem(testdata.FS, "invalid/invalid.go"),
		Config: testdata.DefaultConfig,
		Err:    gofile.ErrParseGoFile,
	})
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			parser := gofile.NewParser(
				gofile.WithParserConfig(tc.Config),
				gofile.WithSource(tc.Source))

			representations, err := parser.Parse(t.Context())
			if tc.Err != nil && !errors.Is(err, tc.Err) {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(representations) != len(tc.Representations) {
				t.Errorf("expected %d enum representations, got %d", len(tc.Representations), len(representations))
				return
			}

			for _, rep := range representations {
				if rep.PackageName == "" {
					t.Errorf("missing package name in %s", tc.Name)
					return
				}
				validateTypeInfo(t, rep.TypeInfo)
				if len(rep.Enums) == 0 {
					t.Errorf("no enum values in %s for type %s", tc.Name, rep.TypeInfo.Name)
					return
				}
				for _, enum := range rep.Enums {
					validateEnum(t, enum, rep.TypeInfo.Name)
				}
			}
		})
	}
}

func TestParseWithDifferentConfigs(t *testing.T) {
	t.Parallel()
	configs := []struct {
		name   string
		config config.Configuration
	}{
		{name: "failfast", config: config.Configuration{Failfast: true}},
		{name: "legacy", config: config.Configuration{Legacy: true}},
		{name: "insensitive", config: config.Configuration{Insensitive: true}},
		{name: "all", config: config.Configuration{Failfast: true, Legacy: true, Insensitive: true}},
	}

	sources := []struct {
		name   string
		source enum.Source
	}{
		{name: "planets", source: source.FromFileSystem(testdata.FS, "planets/planets.go")},
		{name: "status", source: source.FromFileSystem(testdata.FS, "status/status.go")},
		{name: "multiple", source: source.FromFileSystem(testdata.FS, "multiple/multiple.go")},
	}

	for _, src := range sources {
		for _, cfg := range configs {
			tname := fmt.Sprintf("%s_%s", src.name, cfg.name)
			t.Run(tname, func(t *testing.T) {
				t.Parallel()
				parser := gofile.NewParser(
					gofile.WithParserConfig(cfg.config),
					gofile.WithSource(src.source))
				representations, err := parser.Parse(t.Context())
				if err != nil {
					t.Errorf("parse error with config %+v: %v", cfg, err)
				}
				if len(representations) == 0 {
					t.Errorf("no representations found with config %+v", cfg)
				}
				for _, rep := range representations {
					if rep.Failfast != cfg.config.Failfast {
						t.Errorf("failfast config not propagated: expected %v, got %v", cfg.config.Failfast, rep.Failfast)
					}
					if rep.Legacy != cfg.config.Legacy {
						t.Errorf("legacy config not propagated: expected %v, got %v", cfg.config.Legacy, rep.Legacy)
					}
					if rep.CaseInsensitive != cfg.config.Insensitive {
						t.Errorf("insensitive config not propagated: expected %v, got %v", cfg.config.Insensitive, rep.CaseInsensitive)
					}
				}
			})
		}
	}
}

func TestParseWithCancelledContext(t *testing.T) {
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

func validateTypeInfo(t *testing.T, info enum.TypeInfo) {
	t.Helper()

	if info.Name == "" {
		t.Error("type name is empty")
	}

	if info.Lower == "" {
		t.Errorf("type lower name is empty for %s", info.Name)
	}

	if info.Upper == "" {
		t.Errorf("type upper name is empty for %s", info.Name)
	}

	if info.Camel == "" {
		t.Errorf("type camel name is empty for %s", info.Name)
	}

	if info.Plural == "" {
		t.Errorf("type plural name is empty for %s", info.Name)
	}

	if info.PluralCamel == "" {
		t.Errorf("type pluralCamel name is empty for %s", info.Name)
	}
}

func validateEnum(t *testing.T, e enum.Enum, typeName string) {
	t.Helper()
	if e.Info.Name == "" {
		t.Error("enum name is empty")
	}
	if e.Info.Lower == "" {
		t.Errorf("enum lower name is empty for %s", e.Info.Name)
	}
	if e.Info.Upper == "" {
		t.Errorf("enum upper name is empty for %s", e.Info.Name)
	}
	if e.Info.Camel == "" {
		t.Errorf("enum camel name is empty for %s", e.Info.Name)
	}
	if e.TypeInfo.Name != typeName {
		t.Errorf("type name mismatch in enum: got %s, expected %s", e.TypeInfo.Name, typeName)
	}
}

func TestFormattingFunctions(t *testing.T) {
	t.Parallel()

	valueTests := []struct {
		name     string
		source   string
		typeName string
		enumName string
		comment  string
		expected string
	}{
		{
			name: "Integer Formatting",
			source: `package test
type ValueType int
const (
    TestInt ValueType = iota // 42
)`,
			typeName: "ValueType",
			enumName: "TestInt",
			comment:  "42",
			expected: "42",
		},
		{
			name: "Float Formatting",
			source: `package test
type ValueType int
const (
    TestFloat ValueType = iota // 3.14
)`,
			typeName: "ValueType",
			enumName: "TestFloat",
			comment:  "3.14",
			expected: "3.14",
		},
		{
			name: "String Formatting",
			source: `package test
type ValueType int
const (
    TestString ValueType = iota // "test string"
)`,
			typeName: "ValueType",
			enumName: "TestString",
			comment:  `"test string"`,
			expected: `"test string"`,
		},
		{
			name: "Bool Formatting",
			source: `package test
type ValueType int
const (
    TestBool ValueType = iota // true
)`,
			typeName: "ValueType",
			enumName: "TestBool",
			comment:  "true",
			expected: "true",
		},
	}

	for _, tc := range valueTests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tempDir := t.TempDir()
			tempFile, err := os.CreateTemp(tempDir, "gofile_parser_test_*.go")
			if err != nil {
				t.Errorf("failed to create temp file: %v", err)
				return
			}
			_, err = tempFile.WriteString(tc.source)
			if err != nil {
				t.Errorf("failed to write to temp file: %v", err)
			}
			tempFile.Close()

			parser := gofile.NewParser(gofile.WithSource(source.FromFile(tempFile.Name())))
			reps, err := parser.Parse(context.Background())
			if err != nil {
				t.Errorf("failed to parse: %v", err)
			}

			if len(reps) == 0 || len(reps[0].Enums) == 0 {
				t.Errorf("no enums found in test source")
			}

			enum := reps[0].Enums[0]
			if enum.Raw.Comment != tc.comment {
				t.Errorf("raw comment mismatch: got %q, expected %q", enum.Raw.Comment, tc.comment)
			}
		})
	}
}

func TestNameTypePairParsing(t *testing.T) {
	t.Parallel()

	pairTests := []struct {
		name      string
		source    string
		pairCount int
		pairs     [][2]string
	}{
		{
			name: "Bracket Format",
			source: `package test
type testType int // id[int], name[string]
const (
    TestValue testType = iota
)`,
			pairCount: 2,
			pairs: [][2]string{
				{"id", "int"},
				{"name", "string"},
			},
		},
		{
			name: "Parenthesis Format",
			source: `package test

type testType int // code(int), message(string)
const (
    TestValue testType = iota
)`,
			pairCount: 2,
			pairs: [][2]string{
				{"code", "int"},
				{"message", "string"},
			},
		},
		{
			name: "Space Format",
			source: `package test

type testType int // status int, desc string
const (
    TestValue testType = iota
)`,
			pairCount: 2,
			pairs: [][2]string{
				{"status", "int"},
				{"desc", "string"},
			},
		},
		{
			name: "Mixed Formats",
			source: `package test

type testType int // id[int], name(string), age int
const (
    TestValue TestType = iota
)`,
			pairCount: 0,
			pairs:     [][2]string{},
		},
	}

	for _, tc := range pairTests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempFile, err := os.CreateTemp("", "gofile_parser_test_*.go")
			if err != nil {
				t.Errorf("failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			_, err = tempFile.WriteString(tc.source)
			if err != nil {
				t.Errorf("failed to write to temp file: %v", err)
			}
			tempFile.Close()

			parser := gofile.NewParser(gofile.WithSource(source.FromFile(tempFile.Name())))
			reps, err := parser.Parse(context.Background())
			if err != nil {
				t.Errorf("failed to parse: %v", err)
			}
			if len(reps) == 0 {
				t.Errorf("no representations found in test source")
			}
			pairs := reps[0].TypeInfo.NameTypePair
			if len(pairs) != tc.pairCount {
				t.Errorf("expected %d name-type pairs, got %d", tc.pairCount, len(pairs))
				return
			}

			for i, expected := range tc.pairs {
				if i >= len(pairs) {
					t.Errorf("missing pair %d: %v", i, expected)
					continue
				}
				if pairs[i].Name != expected[0] || pairs[i].Type != expected[1] {
					t.Errorf("pair %d mismatch: expected {%s, %s}, got {%s, %s}",
						i, expected[0], expected[1], pairs[i].Name, pairs[i].Type)
				}
			}
		})
	}
}
