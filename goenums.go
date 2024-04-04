package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/zarldev/goenums/pkg/generator"
)

const currentVersion = "v0.2.4"

func main() {
	var err error
	if len(os.Args) != 2 {
		printHelp()
		return
	}
	if len(os.Args) == 2 {
		cmd := os.Args[1]
		switch cmd {
		case "help", "--help", "-h":
			printHelp()
			return
		case "version":
			printVersion()
			return
		}
	}
	filename := os.Args[1]
	err = generator.ParseAndGenerate(filename)
	if err != nil {
		slog.Error("Failed to generate enums: %v", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`goenums is a tool to generate enums from a Go source file.
To generate enums, run the following command:
		goenums <filename>
	For example:
		goenums example.go
	To print the version, run:
		goenums version
	`)
}

func printVersion() {
	fmt.Printf("goenums %s\n", currentVersion)
}
