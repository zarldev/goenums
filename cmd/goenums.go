package main

import (
	"fmt"
	"go/format"
	"io"
	"log/slog"
	"os"

	"github.com/zarldev/goenums/pkg/generator"
)

func main() {
	var err error
	if len(os.Args) != 2 {
		printHelp()
		return
	}
	filename := os.Args[1]
	err = generator.ParseAndGenerate(filename)
	if err != nil {
		slog.Error("Failed to generate enums: %v", err)
		os.Exit(1)
	}
	file, err := os.Open(filename)
	if err != nil {
		slog.Error("Failed to open file: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err != nil {
			slog.Error("Failed to close file: %v", err)
			os.Exit(1)
		}
	}()

	b, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Failed to read file: %v", err)
		return
	}
	src, err := format.Source(b)
	if err != nil {
		slog.Error("Failed to format source: %v", err)
		return
	}
	err = os.WriteFile("_"+filename, src, 0644)
	if err != nil {
		slog.Error("Failed to write file: %v", err)
		return
	}
	fmt.Println("Enums generated successfully!")
}

func printHelp() {
	fmt.Println("Usage: goenums <filename.go>")
}
