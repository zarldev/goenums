package gofile_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/source"
)

var (
	testCases = []struct {
		name        string
		source      enum.Source
		config      config.Configuration
		expectError bool
		typeCount   int
	}{
		{
			name:      "Status-Strings-Validation",
			source:    source.FromFile("../../internal/testdata/validation-strings/status.go"),
			config:    config.Configuration{},
			typeCount: 1,
		},
		{
			name:      "Status-Validation",
			source:    source.FromFile("../../internal/testdata/validation/status.go"),
			config:    config.Configuration{},
			typeCount: 1,
		},
		{
			name:      "Planets",
			source:    source.FromFile("../../internal/testdata/planets/planets.go"),
			config:    config.Configuration{},
			typeCount: 1,
		},
		{
			name:      "Planets-Gravity-Only",
			source:    source.FromFile("../../internal/testdata/planets_gravity_only/planets.go"),
			config:    config.Configuration{},
			typeCount: 1,
		},
		{
			name:      "Planets-Simple",
			source:    source.FromFile("../../internal/testdata/planets_simple/planets.go"),
			config:    config.Configuration{},
			typeCount: 1,
		},
		{
			name:      "Discount-Types",
			source:    source.FromFile("../../internal/testdata/sale/discount.go"),
			config:    config.Configuration{},
			typeCount: 1,
		},
		{
			name:      "Order-Types",
			source:    source.FromFile("../../internal/testdata/orders/orders.go"),
			config:    config.Configuration{},
			typeCount: 1,
		},
		{
			name:      "Multiple-Types",
			source:    source.FromFile("../../internal/testdata/multiple/multiple.go"),
			config:    config.Configuration{},
			typeCount: 2,
		},
		{
			name:      "Ticket-Statuses-With-Spaces",
			source:    source.FromFile("../../internal/testdata/spaces/tickets.go"),
			config:    config.Configuration{},
			typeCount: 1,
		},
		{
			name:        "Non-Existent-File",
			source:      source.FromFile("../../internal/testdata/non_existent_file.go"),
			config:      config.Configuration{},
			expectError: true,
		},
	}
)

func TestParse(t *testing.T) {
	for _, tc := range testCases {
		if !tc.expectError {
			if _, err := os.Stat(tc.source.Filename()); err != nil {
				t.Fatalf("test file %s doesn't exist. make sure testdata is properly set up", tc.source.Filename())
			}
		}
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parser := gofile.NewParser(tc.config, tc.source)
			representations, err := parser.Parse(context.Background())

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error but got none for %s", tc.name)
				}
				return
			}

			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			if len(representations) != tc.typeCount {
				t.Errorf("expected %d enum types, got %d", tc.typeCount, len(representations))
			}

			for _, rep := range representations {
				if rep.PackageName == "" {
					t.Errorf("missing package name in %s", tc.name)
				}

				validateTypeInfo(t, rep.TypeInfo)

				if len(rep.Enums) == 0 {
					t.Errorf("no enum values in %s for type %s", tc.name, rep.TypeInfo.Name)
					continue
				}

				for _, enum := range rep.Enums {
					validateEnum(t, enum, rep.TypeInfo.Name)
				}
			}
		})
	}
}

func TestParseWithDifferentConfigs(t *testing.T) {
	configs := []config.Configuration{
		{Failfast: true},
		{Legacy: true},
		{Insensitive: true},
		{Failfast: true, Legacy: true, Insensitive: true},
	}

	sources := []enum.Source{
		source.FromFile("../../internal/testdata/planets/planets.go"),
		source.FromFile("../../internal/testdata/validation/status.go"),
		source.FromFile("../../internal/testdata/multiple/multiple.go"),
	}

	for _, src := range sources {
		for i, cfg := range configs {
			t.Run(strings.TrimSuffix(strings.TrimPrefix(src.Filename(), "../../internal/testdata/"), ".go")+
				"-config-"+string(rune('A'+i)), func(t *testing.T) {
				t.Parallel()

				parser := gofile.NewParser(cfg, src)
				representations, err := parser.Parse(context.Background())
				if err != nil {
					t.Fatalf("parse error with config %+v: %v", cfg, err)
				}

				if len(representations) == 0 {
					t.Errorf("no representations found with config %+v", cfg)
				}

				for _, rep := range representations {
					if rep.Failfast != cfg.Failfast {
						t.Errorf("failfast config not propagated: expected %v, got %v", cfg.Failfast, rep.Failfast)
					}
					if rep.Legacy != cfg.Legacy {
						t.Errorf("legacy config not propagated: expected %v, got %v", cfg.Legacy, rep.Legacy)
					}
					if rep.CaseInsensitive != cfg.Insensitive {
						t.Errorf("insensitive config not propagated: expected %v, got %v", cfg.Insensitive, rep.CaseInsensitive)
					}
				}
			})
		}
	}
}

func TestParseWithCancelledContext(t *testing.T) {
	t.Parallel()

	parser := gofile.NewParser(config.Configuration{},
		source.FromFile("../../internal/testdata/planets/planets.go"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := parser.Parse(ctx)
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
	if err != context.Canceled {
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

			tempFile, err := os.CreateTemp("", "gofile_parser_test_*.go")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			_, err = tempFile.WriteString(tc.source)
			if err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			tempFile.Close()

			parser := gofile.NewParser(config.Configuration{}, source.FromFile(tempFile.Name()))
			reps, err := parser.Parse(context.Background())
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			if len(reps) == 0 || len(reps[0].Enums) == 0 {
				t.Fatalf("no enums found in test source")
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
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			_, err = tempFile.WriteString(tc.source)
			if err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			tempFile.Close()

			parser := gofile.NewParser(config.Configuration{}, source.FromFile(tempFile.Name()))
			reps, err := parser.Parse(context.Background())
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			if len(reps) == 0 {
				t.Fatalf("no representations found in test source")
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
