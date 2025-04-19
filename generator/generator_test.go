package generator_test

import (
	"errors"
	"testing"

	"github.com/zarldev/goenums/generator"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/testdata"
	"github.com/zarldev/goenums/source"
)

func TestGenerator_ParseAndWrite(t *testing.T) {
	t.Parallel()
	testcases := testdata.InputOutputTestCases
	testcases = append(testcases, testdata.InputOutputTest{
		Name:   "not go code",
		Source: source.FromFileSystem(testdata.FS, "notgocode/notgocode.go"),
		Config: testdata.DefaultConfig,
		Err:    generator.ErrFailedToParse,
	})
	testcases = append(testcases, testdata.InputOutputTest{
		Name:   "no valid enums",
		Source: source.FromFileSystem(testdata.FS, "noenums/noenums.go"),
		Config: testdata.DefaultConfig,
		Err:    generator.ErrNoEnumsFound,
	})
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			parser := gofile.NewParser(
				gofile.WithParserConfig(tc.Config),
				gofile.WithSource(tc.Source))
			wri := gofile.NewWriter(
				gofile.WithWriterConfig(tc.Config),
				gofile.WithFileSystem(testdata.FS))

			p := generator.New(
				generator.WithConfig(tc.Config),
				generator.WithParser(parser),
				generator.WithWriter(wri))
			if err := p.ParseAndWrite(t.Context()); err != nil {
				if !errors.Is(err, tc.Err) {
					t.Errorf("failed to generate enums for %s, got %v", tc.Source.Filename(), err)
				}
			}
			for _, filename := range tc.ExpectedFiles {
				if _, err := testdata.FS.Stat(filename); err != nil {
					if !errors.Is(err, tc.Err) {
						t.Errorf("failed to find generated file %s, got %v", tc.ExpectedFiles, err)
					}
				}
			}
		})
	}
}
