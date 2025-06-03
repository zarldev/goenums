// The goenums tool addresses Go's lack of native enum support by generating
// type-safe wrappers around iota-based constant declarations.
//
// It analyzes Go source files to identify iota-based constant groups and produces
// comprehensive enum implementations with rich functionality including string
// conversion, JSON/database integration, validation, and iteration support.
//
// # Key Features
//
//   - Type-safe enum wrapper types with custom fields
//   - Comprehensive string conversion and parsing (with optional case-insensitive mode)
//   - JSON, Text, Binary, and YAML marshaling/unmarshaling
//   - SQL database integration via Scanner/Valuer interfaces
//   - Validation methods for checking valid enum values
//   - Go 1.23+ iterator support with automatic legacy fallback
//   - Exhaustive processing to ensure all values are handled
//   - Numeric parsing from underlying integer values
//   - Alias support for alternative enum names
//
// # Basic Usage
//
// Define an enum in your Go source file:
//
//	type Status int
//	const (
//	    Active Status = iota
//	    Inactive
//	    Pending
//	)
//
// Generate the enum implementation:
//
//	goenums status.go
//
// Use the generated enum:
//
//	status := Statuses.ACTIVE
//	fmt.Println(status.String())        // "Active"
//	parsed, _ := ParseStatus("Pending") // Statuses.PENDING
//	fmt.Println(status.IsValid())       // true
//
// # Advanced Features
//
// ## Custom Fields
//
// Add metadata to enum values using field syntax:
//
//	type HTTPStatus int // code[int], message[string]
//	const (
//	    OK HTTPStatus = iota         // 200, "Success"
//	    NotFound                     // 404, "Not Found"
//	    InternalError                // 500, "Internal Server Error"
//	)
//
// Access custom fields in generated code:
//
//	status := HTTPStatuses.OK
//	fmt.Println(status.Code())    // 200
//	fmt.Println(status.Message()) // "Success"
//
// ## JSON Integration
//
//	type Task struct {
//	    Name     string     `json:"name"`
//	    Priority Priority   `json:"priority"`
//	}
//
//	task := Task{Name: "Deploy", Priority: Priorities.HIGH}
//	json.Marshal(task) // {"name":"Deploy","priority":"High"}
//
// ## Database Integration
//
//	// Enums implement database/sql interfaces automatically
//	var status Status
//	err := db.QueryRow("SELECT status FROM tasks WHERE id = ?", id).Scan(&status)
//
//	_, err = db.Exec("INSERT INTO tasks (status) VALUES (?)", Statuses.ACTIVE)
//
// ## Iteration (Go 1.23+)
//
//	for status := range Statuses.All() {
//	    fmt.Println("Status:", status.String())
//	}
//
// # Command Line Options
//
//	goenums [options] file.go
//
//	-f, -failfast      Fail on invalid enum values during parsing
//	-l, -legacy        Generate code without Go 1.23+ iterator support
//	-i, -insensitive   Enable case-insensitive string parsing
//	-c, -constraints   Generate constraints locally instead of importing
//	-v, -version       Show version information
//	-h, -help          Show help information
//	-vv, -verbose      Enable verbose output
//	-o, -output        Specify output format (default: go)
//
// # Design Philosophy
//
// The tool follows a modular, interface-based architecture that separates
// content sourcing, parsing, and code generation. This design enables:
//
//   - Support for different input formats (currently Go, extensible to others)
//   - Multiple output targets (currently Go, extensible to other languages)
//   - Clean separation of concerns between components
//   - Easy testing and maintenance of individual components
//   - Future extensibility without breaking existing functionality
//
// The generated code prioritizes type safety, performance, and integration
// with standard Go interfaces and conventions.
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
	"text/template"

	"github.com/zarldev/goenums/enum"

	"github.com/zarldev/goenums/generator"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
	"github.com/zarldev/goenums/internal/version"
	"github.com/zarldev/goenums/logging"
	"github.com/zarldev/goenums/source"
	"github.com/zarldev/goenums/strings"
)

// Define flag groups
type flags struct {
	help, version, failfast, legacy, insensitive, verbose, constraints bool
	output                                                             string
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
	flag.BoolVar(&f.constraints, "constraints", false,
		"Specify whether to generate the float and integer constraints or import 'golang.org/x/exp/constraints' (default: false - imports)")
	flag.BoolVar(&f.constraints, "c", false, "")
	flag.Parse()
	return f, flag.Args()
}

const (
	colorReset       = "\033[0m"
	colorBlue        = "\033[34m"
	colorCyan        = "\033[36m"
	colorYellow      = "\033[33m"
	colorGreen       = "\033[32m"
	logoTemplateBody = colorBlue + `
   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/
` + colorReset
	versionTemplateBody = colorCyan + `
    https://zarldev.github.io/goenums ` + colorReset + colorGreen + `
       version :: {{.Version}}
` + colorReset
)

var (
	logoTemplate    = template.Must(template.New("logo").Parse(logoTemplateBody))
	versionTemplate = template.Must(template.New("version").Parse(versionTemplateBody))
)

// logo displays the goenums logo.
func logo() {
	err := logoTemplate.Execute(os.Stdout, nil)
	if err != nil {
		slog.Default().Error("Error executing logo template", slog.Any("error", err))
	}
}

type versionData struct {
	Version string
	Build   string
	Commit  string
}

// printVersion displays the current version of the goenums tool.
func printVersion() {
	data := versionData{
		Version: strings.ReplaceAll(version.CURRENT, "'", ""),
		Build:   strings.ReplaceAll(version.BUILD, "'", ""),
		Commit:  strings.ReplaceAll(version.COMMIT, "'", ""),
	}
	err := logoTemplate.Execute(os.Stdout, nil)
	if err != nil {
		slog.Default().Error("Error executing logo template", slog.Any("error", err))
	}
	err = versionTemplate.Execute(os.Stdout, data)
	if err != nil {
		slog.Default().Error("Error executing logo template", slog.Any("error", err))
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logging.Configure(false)
	// Setup signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()
	config, err := configuration(ctx)
	if err != nil {
		return
	}
	logging.Configure(config.Verbose)

	logo()
	slog.Default().Info(fmt.Sprintf("\t\tversion: %s", version.CURRENT))
	slog.Default().Debug("starting generation...")
	slog.Default().Debug("config settings",
		slog.Int("file_count", len(config.Filenames)),
		slog.String("files", buildFileList(config.Filenames)),
		slog.String("output", config.OutputFormat),
		slog.Bool("failfast", config.Failfast),
		slog.Bool("legacy", config.Legacy),
		slog.Bool("insensitive", config.Insensitive),
		slog.Bool("verbose", config.Verbose))

	for _, filename := range config.Filenames {
		filename = strings.TrimSpace(filename)
		if filename == "" {
			continue
		}
		slog.Default().Info("processing file", slog.String("filename", filename))
		var (
			parser enum.Parser
			writer enum.Writer
		)

		inExt := filepath.Ext(filename)
		switch inExt {
		case ".go":
			slog.Default().Debug("initializing go parser")
			parser = gofile.NewParser(
				gofile.WithParserConfiguration(config),
				gofile.WithSource(source.FromFile(filename)))
		default:
			slog.Default().Error("only .go files are supported")
			return
		}

		switch config.OutputFormat {
		case "", "go":
			slog.Default().Debug("initializing gofile writer")
			writer = gofile.NewWriter(gofile.WithWriterConfiguration(config))
		default:
			slog.Default().Error("only outputting to go files is supported")
			return
		}

		slog.Default().Debug("initializing generator")
		gen := generator.New(
			generator.WithConfig(config),
			generator.WithParser(parser),
			generator.WithWriter(writer))
		slog.Default().Info("starting parsing and generation")
		if err := gen.ParseAndWrite(ctx); err != nil {
			if errors.Is(err, enum.ErrParseSource) {
				slog.Default().Error("unable to parse file", slog.String("filename", filename))
				slog.Default().Error("please ensure that the file is a valid input file")
				slog.Default().Error("for the selected parser")
			}
			if errors.Is(err, enum.ErrNoEnumsFound) {
				slog.Default().Error("no enums found in file", slog.String("filename", filename))
				slog.Default().Error("please ensure that the file contains enum definitions")
			}
			if errors.Is(err, enum.ErrWriteOutput) {
				slog.Default().Error("could not generate output")
				slog.Default().Error("please ensure that the output destination is writable")
				slog.Default().Error("and that input enums contain only valid characters")
			}
			slog.Default().Error("could not generate enums", slog.String("error", err.Error()))
			slog.Default().Error("exiting")
			return
		}
		slog.Default().Info("successfully generated enums")
	}
}

var ErrComplete = errors.New("completed")

func configuration(ctx context.Context) (config.Configuration, error) {
	f, args := parseFlags()

	if f.help {
		printHelp()
		return config.Configuration{}, ErrComplete
	}

	if f.version {
		printVersion()
		return config.Configuration{}, ErrComplete
	}

	if len(args) < 1 {
		slog.Default().ErrorContext(ctx, "you must specify at least one input file")
		return config.Configuration{}, ErrComplete
	}

	filenames := args

	for _, filename := range filenames {
		filename = strings.TrimSpace(filename)
		if filename == "" {
			continue
		}

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			slog.Default().ErrorContext(ctx, "input file does not exist", slog.String("filename", filename))
			return config.Configuration{}, fmt.Errorf("input file does not exist %s", filename)
		}
	}

	config := config.Configuration{
		Failfast:     f.failfast,
		Insensitive:  f.insensitive,
		Legacy:       f.legacy,
		Verbose:      f.verbose,
		OutputFormat: f.output,
		Filenames:    filenames,
		Constraints:  f.constraints,
		Handlers: config.Handlers{
			JSON:   true,
			Text:   true,
			SQL:    true,
			YAML:   true,
			Binary: true,
		},
	}
	return config, nil
}

// printHelp displays usage instructions and command-line options
func printHelp() {
	logo()
	slog.Default().Info("Usage: goenums [options] file.go[,file2.go,...]")
	slog.Default().Info("Options:")
	flag.PrintDefaults()
}

func buildFileList(filenames []string) string {
	if len(filenames) == 0 {
		return ""
	}
	var builder strings.EnumBuilder
	builder.WriteString(filenames[0])
	for _, filename := range filenames[1:] {
		builder.WriteString(", ")
		builder.WriteString(filename)
	}
	return builder.String()
}
