package generator_test

import (
	"errors"
	"testing"

	"github.com/zarldev/goenums/generator"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/testdata"
)

func TestGenerator_ParseAndGenerate(t *testing.T) {
	t.Parallel()
	for _, tc := range testdata.InputOutputTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			parser := gofile.NewParser(tc.Config, gofile.WithSource(tc.Source))
			wri := gofile.NewWriter(tc.Config, gofile.WithFileSystem(testdata.FS))
			p := generator.New(tc.Config, parser, wri)
			if err := p.ParseAndWrite(t.Context()); err != nil {
				if !errors.Is(err, tc.Err) {
					t.Errorf("failed to generate enums for %s, got %v", tc.Source.Filename(), err)
				}
			}
		})
	}
	for _, tc := range testdata.InputOutputTestCases {
		t.Run(tc.Name, func(t *testing.T) {
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
