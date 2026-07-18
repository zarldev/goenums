package gofile_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/source"
)

func BenchmarkParser(b *testing.B) {
	for _, size := range []int{8, 128, 1024} {
		sourceCode := benchmarkSource(size)
		b.Run(fmt.Sprintf("values_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(sourceCode)))
			for range b.N {
				parser := gofile.NewParser(gofile.WithSource(source.FromReader(strings.NewReader(sourceCode))))
				requests, err := parser.Parse(context.Background())
				if err != nil {
					b.Fatal(err)
				}
				if len(requests) != 1 {
					b.Fatalf("got %d generation requests, want 1", len(requests))
				}
			}
		})
	}
}

func BenchmarkWriter(b *testing.B) {
	for _, size := range []int{8, 128, 1024} {
		request := benchmarkRequest(size)
		b.Run(fmt.Sprintf("values_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				fs := file.NewMemFS()
				writer := gofile.NewWriter(gofile.WithFileSystem(fs))
				if err := writer.Write(context.Background(), []enum.GenerationRequest{request}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkParseAndWrite(b *testing.B) {
	for _, size := range []int{8, 128, 1024} {
		sourceCode := benchmarkSource(size)
		b.Run(fmt.Sprintf("values_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(sourceCode)))
			for range b.N {
				parser := gofile.NewParser(gofile.WithSource(source.FromReader(strings.NewReader(sourceCode))))
				writer := gofile.NewWriter(gofile.WithFileSystem(file.NewMemFS()))
				gen := generator.New(generator.WithParser(parser), generator.WithWriter(writer))
				if err := gen.ParseAndWrite(context.Background()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func benchmarkSource(size int) string {
	var sourceCode strings.Builder
	sourceCode.WriteString("package benchmark\n\ntype status int\n\nconst (\n")
	for i := range size {
		if i == 0 {
			fmt.Fprintf(&sourceCode, "\tStatus%d status = iota\n", i)
			continue
		}
		fmt.Fprintf(&sourceCode, "\tStatus%d\n", i)
	}
	sourceCode.WriteString(")\n")
	return sourceCode.String()
}

func benchmarkRequest(size int) enum.GenerationRequest {
	enums := make([]enum.Enum, size)
	for i := range enums {
		enums[i] = enum.Enum{
			Name:    fmt.Sprintf("Status%d", i),
			Index:   i,
			Value:   i,
			Aliases: []string{fmt.Sprintf("status-%d", i)},
			Valid:   true,
		}
	}

	configuration := config.Configuration{
		Handlers: config.Handlers{JSON: true, Text: true, YAML: true, SQL: true, Binary: true},
	}
	return enum.GenerationRequest{
		Package:        "benchmark",
		EnumIota:       enum.EnumIota{Type: "status", Enums: enums},
		Version:        "benchmark",
		SourceFilename: "status.go",
		OutputFilename: "statuses",
		Configuration:  configuration,
	}
}

func BenchmarkWriterHandlers(b *testing.B) {
	cases := []struct {
		name          string
		configuration config.Configuration
	}{
		{name: "all", configuration: config.Configuration{Constraints: true, Handlers: allBenchmarkHandlers()}},
		{name: "json_sql", configuration: config.Configuration{Constraints: true, Handlers: config.Handlers{JSON: true, SQL: true}}},
		{name: "text", configuration: config.Configuration{Constraints: true, Handlers: config.Handlers{Text: true}}},
		{name: "x_exp_constraints", configuration: config.Configuration{Constraints: false, Handlers: allBenchmarkHandlers()}},
	}
	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			request := benchmarkRequest(128)
			request.Configuration = tt.configuration
			b.ReportAllocs()
			for range b.N {
				fs := file.NewMemFS()
				writer := gofile.NewWriter(gofile.WithFileSystem(fs))
				if err := writer.Write(context.Background(), []enum.GenerationRequest{request}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func allBenchmarkHandlers() config.Handlers {
	return config.Handlers{JSON: true, Text: true, YAML: true, SQL: true, Binary: true}
}
