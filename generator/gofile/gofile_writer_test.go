// generator/gofile/gofile_writer_test.go
package gofile_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
)

func TestWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		config          config.Configuration
		representations []enum.Representation
		err             error
		validate        func(t *testing.T, fs *file.MemFS)
	}{
		{
			name:   "simple enum representation",
			config: config.Configuration{},
			representations: []enum.Representation{
				{
					Version:        "v1.0.0",
					GenerationTime: time.Now(),
					PackageName:    "colors",
					SourceFilename: "colors.go",
					TypeInfo: enum.TypeInfo{
						Name:        "color",
						Camel:       "Color",
						Lower:       "color",
						Upper:       "COLOR",
						Plural:      "colors",
						PluralCamel: "Colors",
						NameTypePair: []enum.NameTypePair{
							{
								Name:  "Hex",
								Type:  "string",
								Value: `""`,
							},
							{
								Name:  "RGB",
								Type:  "int",
								Value: "0",
							},
						},
					},
					Enums: []enum.Enum{
						{
							Info: enum.Info{
								Name:  "red",
								Camel: "Red",
								Lower: "red",
								Upper: "RED",
								Alias: "Red",
								Value: 0,
								Valid: true,
							},
							TypeInfo: enum.TypeInfo{
								Name:  "color",
								Camel: "Color",
								NameTypePair: []enum.NameTypePair{
									{
										Name:  "Hex",
										Type:  "string",
										Value: `"#FF0000"`,
									},
									{
										Name:  "RGB",
										Type:  "int",
										Value: "16711680",
									},
								},
							},
						},
						{
							Info: enum.Info{
								Name:  "green",
								Camel: "Green",
								Lower: "green",
								Upper: "GREEN",
								Alias: "Green",
								Value: 1,
								Valid: true,
							},
							TypeInfo: enum.TypeInfo{
								Name:  "color",
								Camel: "Color",
								NameTypePair: []enum.NameTypePair{
									{
										Name:  "Hex",
										Type:  "string",
										Value: `"#00FF00"`,
									},
									{
										Name:  "RGB",
										Type:  "int",
										Value: "65280",
									},
								},
							},
						},
						{
							Info: enum.Info{
								Name:  "blue",
								Camel: "Blue",
								Lower: "blue",
								Upper: "BLUE",
								Alias: "Blue",
								Value: 2,
								Valid: true,
							},
							TypeInfo: enum.TypeInfo{
								Name:  "color",
								Camel: "Color",
								NameTypePair: []enum.NameTypePair{
									{
										Name:  "Hex",
										Type:  "string",
										Value: `"#0000FF"`,
									},
									{
										Name:  "RGB",
										Type:  "int",
										Value: "255",
									},
								},
							},
						},
					},
				},
			},

			validate: func(t *testing.T, fs *file.MemFS) {
				// Check if the expected file exists
				content, err := fs.ReadFile("color_enums.go")
				if err != nil {
					t.Errorf("failed to read generated file: %v", err)
				}

				// Basic validation of content
				mustContain := []string{
					"package colors",
					"type Color struct",
					"color",
					"Hex string",
					"RGB int",
					"var Colors = colorContainer{",
					"RED: Color{",
					"color: red,",
					`Hex:   "#FF0000",`,
					"RGB:   16711680,",
					"GREEN: Color{",
					"color: green,",
					`Hex:   "#00FF00",`,
					"RGB:   65280,",
					"color: blue,",
					`Hex:   "#0000FF",`,
					"RGB:   255",
				}
				contentStr := string(content)
				for _, s := range mustContain {

					if !strings.Contains(contentStr, s) {
						t.Errorf("content missing expected string: %q", s)
					}
				}
			},
		},
		{
			name: "case insensitive enum",
			config: config.Configuration{
				Insensitive: true,
			},
			representations: []enum.Representation{
				{
					Version:         "v1.0.0",
					GenerationTime:  time.Now(),
					PackageName:     "days",
					SourceFilename:  "days.go",
					CaseInsensitive: true,
					TypeInfo: enum.TypeInfo{
						Name:        "day",
						Camel:       "Day",
						Lower:       "day",
						Upper:       "DAY",
						Plural:      "days",
						PluralCamel: "Days",
					},
					Enums: []enum.Enum{
						{
							Info: enum.Info{
								Name:  "monday",
								Upper: "MONDAY",
								Alias: "Monday",
								Value: 0,
								Valid: true,
							},
							TypeInfo: enum.TypeInfo{
								Name:  "day",
								Camel: "Day",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, fs *file.MemFS) {
				content, err := fs.ReadFile("day_enums.go")
				if err != nil {
					t.Errorf("failed to read generated file: %v", err)
				}
				if !strings.Contains(string(content), "strings") {
					t.Error("expected strings import for case insensitive enum")
				}
				if !strings.Contains(string(content), "strings.ToLower") {
					t.Error("missing case insensitive conversion code")
				}
				if !strings.Contains(string(content), `"monday": Days.MONDAY,`) {
					t.Error("missing lowercase variant in name map")
				}
			},
		},
		{
			name:   "invalid output filename with space",
			config: config.Configuration{},
			representations: []enum.Representation{
				{
					TypeInfo: enum.TypeInfo{
						Lower: "invalid name",
					},
					SourceFilename: "test.go",
				},
			},
			err: gofile.ErrWriteGoFile,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			memfs := file.NewMemFS()
			writer := gofile.NewWriter(
				gofile.WithWriterConfig(tt.config),
				gofile.WithFileSystem(memfs))
			err := writer.Write(t.Context(), tt.representations)
			if err != nil && !errors.Is(err, tt.err) {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.validate != nil {
				tt.validate(t, memfs)
			}
		})
	}
}
