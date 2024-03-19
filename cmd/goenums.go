package main

import (
	"fmt"
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
}

func printHelp() {
	fmt.Println("Usage: goenums <filename.go>")
}
