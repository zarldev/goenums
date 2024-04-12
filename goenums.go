package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/zarldev/goenums/pkg/generator"
)

func main() {
	var (
		help, version, failfast bool
		err                     error
	)
	flag.BoolVar(&help, "help", false, "Print help information")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&version, "version", false, "Print version information")
	flag.BoolVar(&version, "v", false, "")
	flag.BoolVar(&failfast, "failfast", false, "Enable failfast mode - fail on generation of invalid enum while parsing (default: false)")
	flag.BoolVar(&failfast, "f", false, "")
	flag.Parse()

	args := flag.Args()

	if help {
		printHelp()
		return
	}

	if version {
		printVersion()
		return
	}

	if len(args) < 1 {
		slog.Error("Error: you must provide a filename")
		return
	}

	filename := flag.Arg(0)
	err = generator.ParseAndGenerate(filename, failfast)
	if err != nil {
		slog.Error("Failed to generate enums: %v", err)
		os.Exit(1)
	}
}

func printHelp() {
	printTitle()
	fmt.Println("Usage: goenums [options] filename")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

const currentVersion = "v0.2.8"

func printVersion() {
	printTitle()
	fmt.Printf("\t\tversion: %s\n", currentVersion)
}

func printTitle() {
	fmt.Println(asciiArt)
}

var asciiArt = `   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/`
