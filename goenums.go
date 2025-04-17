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
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/zarldev/goenums/enum"

	"github.com/zarldev/goenums/generator"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/version"
	"github.com/zarldev/goenums/logging"
	"github.com/zarldev/goenums/source"
	"github.com/zarldev/goenums/strings"
)

// asciiArt is the goenums logo
const asciiArt = `   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/ `

// Define flag groups
type flags struct {
	help, version, failfast, legacy, insensitive, verbose bool
	output                                                string
}

func parseFlags() (flags, []string) {
	var f flags
	flag.BoolVar(&f.help, "help", false,
		"Print help information")
	flag.BoolVar(&f.help, "h", false, "")
	flag.BoolVar(&f.version, "version", false,
		"Print version information")
	flag.BoolVar(&f.version, "v", false, "")
	flag.BoolVar(&f.failfast, "failfast", false,
		"Enable failfast mode - fail on generation of invalid enum while parsing (default: false)")
	flag.BoolVar(&f.failfast, "f", false, "")
	flag.BoolVar(&f.legacy, "legacy", false,
		"Generate legacy code without Go 1.23+ iterator support (default: false)")
	flag.BoolVar(&f.legacy, "l", false, "")
	flag.BoolVar(&f.insensitive, "insensitive", false,
		"Generate case insensitive string parsing (default: false)")
	flag.BoolVar(&f.insensitive, "i", false, "")
	flag.BoolVar(&f.verbose, "verbose", false,
		"Enable verbose mode - prints out the generated code (default: false)")
	flag.BoolVar(&f.verbose, "vv", false, "")
	flag.StringVar(&f.output, "output", "",
		"Specify the output format (default: go)")
	flag.StringVar(&f.output, "o", "", "")
	flag.Parse()
	return f, flag.Args()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()
	f, args := parseFlags()

	if f.help {
		printHelp()
		return
	}

	if f.version {
		printVersion()
		return
	}

	if len(args) < 1 {
		slog.Error("you must provide a filename")
		os.Exit(1)
	}

	// Process file input - now supports comma-separated lists
	filenames := args

	// Validate that all files exist
	for _, filename := range filenames {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			slog.Error("input file does not exist", slog.String("filename", filename))
			os.Exit(1)
		}
	}

	// Validate that all files exist
	for _, filename := range filenames {
		filename = strings.TrimSpace(filename)
		if filename == "" {
			continue
		}

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			slog.Error("input file does not exist", slog.String("filename", filename))
			os.Exit(1)
		}
	}

	config := config.Configuration{
		Failfast:    f.failfast,
		Insensitive: f.insensitive,
		Legacy:      f.legacy,
		Verbose:     f.verbose,
		Output:      f.output,
	}

	logging.Configure(config.Verbose)

	slog.Info(asciiArt)
	slog.Info(fmt.Sprintf("\t\tversion: %s", version.CURRENT))
	slog.Debug("starting generation...")
	slog.Debug("config settings",
		slog.Int("file_count", len(filenames)),
		slog.String("files", buildFileList(filenames)),
		slog.String("output", config.Output),
		slog.Bool("failfast", config.Failfast),
		slog.Bool("legacy", config.Legacy),
		slog.Bool("insensitive", config.Insensitive),
		slog.Bool("verbose", config.Verbose))

	for _, filename := range filenames {
		filename = strings.TrimSpace(filename)
		if filename == "" {
			continue
		}
		slog.Info("processing file", slog.String("filename", filename))
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
			slog.Error("only .go files are supported")
			return
		}

		switch config.Output {
		case "", "go":
			slog.Debug("initializing gofile writer")
			writer = gofile.NewWriter(config)
		default:
			slog.Error("only outputting to go files is supported")
			return
		}

		slog.Debug("initializing generator")
		gen := generator.New(config, parser, writer)
		slog.Info("starting parsing and generation")
		if err := gen.ParseAndWrite(ctx); err != nil {
			if errors.Is(err, generator.ErrFailedToParse) {
				slog.Error("unable to parse file", slog.String("filename", filename))
				slog.Error("please ensure that the file is a valid input file")
				slog.Error("for the selected parser")
			}
			if errors.Is(err, generator.ErrNoEnumsFound) {
				slog.Error("no enums found in file", slog.String("filename", filename))
				slog.Error("please ensure that the file contains enum definitions")
			}
			if errors.Is(err, generator.ErrGeneratorFailedToGenerate) {
				slog.Error("could not generate output")
				slog.Error("please ensure that the output destination is writable")
				slog.Error("and that input enums contain only valid characters")
			}
			slog.Error("could not generate enums", slog.String("error", err.Error()))
			slog.Error("exiting")
			os.Exit(1)
		}
		slog.Info("successfully generated enums")
	}
}

// printHelp displays usage instructions and command-line options
func printHelp() {
	logo()
	fmt.Println("Usage: goenums [options] file.go[,file2.go,...]")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

// printVersion displays the current version of the goenums tool.
func printVersion() {
	logo()
	fmt.Printf("\t\tversion: %s\n", version.CURRENT)
	if version.BUILD != "" {
		fmt.Printf("\t\tbuild: %s\n", version.BUILD)
	}
	if version.COMMIT != "" {
		fmt.Printf("\t\tcommit: %s\n", version.COMMIT)
	}
}

// logo displays the ASCII art logo for the goenums tool.
func logo() {
	fmt.Println(asciiArt)
}

func buildFileList(filenames []string) string {
	if len(filenames) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(filenames[0])
	for _, filename := range filenames[1:] {
		builder.WriteString(", ")
		builder.WriteString(filename)
	}
	return builder.String()
}
