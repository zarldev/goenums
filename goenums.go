package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/zarldev/goenums/pkg/generator"
)

var (
	fileName   = flag.String("file", "", "Path to the file to generate enums from")
	valuerType = flag.String("valuer", "string", "Type of the valuer interface implementation, support int and string, default to string")
)

func main() {
	flag.Parse()

	if fileName == nil || *fileName == "" {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	options := &generator.GenerateOptions{
		FileName:   *fileName,
		ValuerType: *valuerType,
	}

	err := generator.ParseAndGenerate(options)
	if err != nil {
		slog.Error("Failed to generate enums: %v", err)
		os.Exit(1)
	}
}
