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
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator"
	producer "github.com/zarldev/goenums/generator"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/version"
	"github.com/zarldev/goenums/logging"
	"github.com/zarldev/goenums/source"
)

// asciiArt is the goenums logo
const asciiArt = `   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/ `

func main() {
	var (
		help, vers, failfast, legacy, insensitive, verbose bool
		err                                                error
		output                                             string
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
	flag.StringVar(&output, "output", "",
		"Specify the output format (default: go)")
	flag.StringVar(&output, "o", "", "")
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
		Output:      output,
	}

	logging.Configure(config.Verbose)

	slog.Info(asciiArt)
	slog.Info(fmt.Sprintf("Version: %s", version.CURRENT))
	slog.Debug("starting generation...")
	slog.Debug("config settings",
		slog.String("filename", filename),
		slog.Bool("failfast", config.Failfast),
		slog.Bool("legacy", config.Legacy),
		slog.Bool("insensitive", config.Insensitive),
		slog.Bool("verbose", config.Verbose))

	var (
		parser enum.Parser
		writer enum.Writer
	)

	inExt := filepath.Ext(filename)
	switch inExt {
	case ".go":
		slog.Debug("initializing go parser")
		parser = gofile.NewParser(config,
			source.FromFile(filename))
	default:
		slog.Error("error: only .go files are supported")
		return
	}

	switch config.Output {
	case "", "go":
		slog.Debug("initializing gofile writer")
		writer = gofile.NewWriter(config)
	default:
		slog.Error("error: only outputting to go files is supported")
		return
	}

	slog.Debug("initializing producer")
	gen := generator.New(config, parser, writer)
	slog.Info("starting parsing and generation")
	if err = gen.ParseAndWrite(context.Background()); err != nil {
		slog.Error("failed to generate enums")
		if errors.Is(err, producer.ErrParserFailedToParse) {
			slog.Error("failed to parse file", slog.String("filename", filename))
			slog.Error("please ensure that the file is a valid input file")
			slog.Error("for the selected parser")
		}
		if errors.Is(err, producer.ErrParserNoEnumsFound) {
			slog.Error("no enums found in file", slog.String("filename", filename))
			slog.Error("please ensure that the file contains enum definitions")
		}
		if errors.Is(err, producer.ErrGeneratorFailedToGenerate) {
			slog.Error("failed to generate enums")
			slog.Error("please ensure that the output directory is writable")
			slog.Error("and that input enums contain only valid characters")
		}
		slog.Error("exiting")
		os.Exit(1)
	}
	slog.Info("successfully generated enums")
}

// printHelp displays usage instructions and command-line options
func printHelp() {
	logo()
	fmt.Println("Usage: goenums [options] filename")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

// printVersion displays the current version of the goenums tool.
func printVersion() {
	logo()
	fmt.Printf("\t\tversion: %s\n", version.CURRENT)
}

// logo displays the ASCII art logo for the goenums tool.
func logo() {
	fmt.Println(asciiArt)
}
