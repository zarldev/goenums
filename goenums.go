// The goenums tool addresses Go's lack of native enum support by generating
// type-safe wrappers around constant declarations.
// It currently has support for input from Go files wher it identifies iota-based
// constant groups in Go source files and produces an output Go files with
// helper code that provides:
//
// # Key Features
//
//   - Type-safe enum wrapper types
//   - Comprehensive string conversion and parsing
//   - JSON marshaling and unmarshaling
//   - SQL database integration via Scanner/Valuer interfaces
//   - Case-insensitive string parsing (optional)
//   - Validation methods for checking valid enum values
//   - Iteration support with automatic legacy fallback
//   - Exhaustive switch checking
//
// # Command Usage
//
//	goenums [options] file.go
//
// # Design Philosophy
//
// The tool follows a modular, interface-based architecture that separates:
// content sourcing, parsing, and code generation. This allows for future
// extensions to support different input formats or generation targets
// without changing the core workflow.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/internal/version"
	"github.com/zarldev/goenums/logging"
	"github.com/zarldev/goenums/producer"
	"github.com/zarldev/goenums/producer/config"
	"github.com/zarldev/goenums/producer/gofile"
	"github.com/zarldev/goenums/source"
)

// main is the entry point for the goenums command-line tool. It parses command-line
// flags, initializes the appropriate parser and writer based on input file type,
// and orchestrates the enum generation process.
func main() {
	var (
		help, vers, failfast, legacy, insensitive, verbose bool
		err                                                error
	)
	flag.BoolVar(&help, "help", false,
		"Print help information")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&vers, "version", false,
		"Print version information")
	flag.BoolVar(&vers, "v", false, "")
	flag.BoolVar(&failfast, "failfast", false,
		"Enable failfast mode - fail on generation of invalid enum while parsing (default: false)")
	flag.BoolVar(&failfast, "f", false, "")
	flag.BoolVar(&legacy, "legacy", false,
		"Generate legacy code without Go 1.23+ iterator support (default: false)")
	flag.BoolVar(&legacy, "l", false, "")
	flag.BoolVar(&insensitive, "insensitive", false,
		"Generate case insensitive string parsing (default: false)")
	flag.BoolVar(&insensitive, "i", false, "")
	flag.BoolVar(&verbose, "verbose", false,
		"Enable verbose mode - prints out the generated code (default: false)")
	flag.BoolVar(&verbose, "vv", false, "")
	flag.Parse()

	args := flag.Args()

	if help {
		printHelp()
		return
	}

	if vers {
		printVersion()
		return
	}

	if len(args) < 1 {
		slog.Error("error: you must provide a filename")
		return
	}

	filename := flag.Arg(0)

	config := config.Configuration{
		Failfast:    failfast,
		Insensitive: insensitive,
		Legacy:      legacy,
		Verbose:     verbose,
	}

	logging.Configure(config.Verbose)

	var (
		parser enum.Parser
		writer enum.Writer
	)
	slog.Info(asciiArt)
	slog.Info(fmt.Sprintf("Version: %s", version.CURRENT))
	slog.Debug("starting generation...")
	slog.Debug("config settings",
		slog.String("filename", filename),
		slog.Bool("failfast", config.Failfast),
		slog.Bool("legacy", config.Legacy),
		slog.Bool("insensitive", config.Insensitive),
		slog.Bool("verbose", config.Verbose))

	ext := filepath.Ext(filename)

	switch ext {
	case ".go":
		slog.Debug("initializing gofile parser and writer")
		parser = gofile.NewParser(config,
			source.NewFileSource(filename))
		writer = gofile.NewGenerator(config)
	default:
		slog.Error("error: only .go files are supported")
		return
	}
	slog.Debug("initializing producer")
	producer := producer.NewProducer(config, parser, writer)
	slog.Info("starting parsing and generation")
	if err = producer.ParseAndWrite(context.Background()); err != nil {
		slog.Error("failed to generate enums:",
			slog.String("filename", filename),
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("successfully generated enums")
}

// printHelp displays usage instructions and command-line options
// to assist users in understanding how to use the tool.
func printHelp() {
	printTitle()
	fmt.Println("Usage: goenums [options] filename")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

// printVersion displays the current version of the goenums tool.
// This is useful for troubleshooting and ensuring compatibility.
func printVersion() {
	printTitle()
	fmt.Printf("\t\tversion: %s\n", version.CURRENT)
}

// printTitle displays the ASCII art logo for the goenums tool.
// This provides visual branding for the command-line interface.
func printTitle() {
	fmt.Println(asciiArt)
}

// asciiArt is the graphical logo displayed in the terminal when
// the tool is run, providing visual identification.
const asciiArt = `   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/ `
