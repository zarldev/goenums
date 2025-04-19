package gofile_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/testdata"
	"github.com/zarldev/goenums/source"
)

func TestGoFileParser_Parse(t *testing.T) {
	t.Parallel()
	testcases := slices.Clone(testdata.InputOutputTestCases)
	testcases = append(testcases, testdata.InputOutputTest{
		Name:   "not go code",
		Source: source.FromFileSystem(testdata.FS, "notgocode/notgocode.go"),
		Config: testdata.DefaultConfig,
		Err:    gofile.ErrParseGoFile,
	})
	testcases = append(testcases, testdata.InputOutputTest{
		Name:   "no valid enums found",
		Source: source.FromFileSystem(testdata.FS, "noenums/noenums.go"),
		Config: testdata.DefaultConfig,
	})
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			parser := gofile.NewParser(
				gofile.WithParserConfig(tc.Config),
				gofile.WithSource(tc.Source))

			representations, err := parser.Parse(t.Context())
			if !errors.Is(err, tc.Err) {
				t.Errorf("unexpected error: %v expected %v", err, tc.Err)
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
