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
	stdstrings "strings"
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
	help, version, failfast, legacy, legacyText, insensitive, verbose, constraints, xExpConstraints bool
	output                                                                                          string
	interfaces                                                                                      string
	interfacesSet                                                                                   bool
}

func parseFlags(args []string) (flags, []string, *flag.FlagSet, error) {
	var f flags
	fs := flag.NewFlagSet("goenums", flag.ContinueOnError)
	fs.BoolVar(&f.help, "help", false,
		"Print help information")
	fs.BoolVar(&f.help, "h", false, "")
	fs.BoolVar(&f.version, "version", false,
		"Print version information")
	fs.BoolVar(&f.version, "v", false, "")
	fs.BoolVar(&f.failfast, "failfast", false,
		"Enable failfast mode - fail on generation of invalid enum while parsing (default: false)")
	fs.BoolVar(&f.failfast, "f", false, "")
	fs.BoolVar(&f.legacy, "legacy", false,
		"Generate legacy code without Go 1.23+ iterator support (default: false)")
	fs.BoolVar(&f.legacy, "l", false, "")
	fs.BoolVar(&f.legacyText, "legacy-text", false,
		"Generate legacy quoted MarshalText output (default: false)")
	fs.BoolVar(&f.insensitive, "insensitive", false,
		"Generate case insensitive string parsing (default: false)")
	fs.BoolVar(&f.insensitive, "i", false, "")
	fs.BoolVar(&f.verbose, "verbose", false,
		"Enable verbose mode - prints out the generated code (default: false)")
	fs.BoolVar(&f.verbose, "vv", false, "")
	fs.StringVar(&f.output, "output", "",
		"Specify the output format (default: go)")
	fs.StringVar(&f.output, "o", "", "")
	fs.StringVar(&f.interfaces, "interfaces", "",
		"Generate only the listed interface handlers: json,text,yaml,sql,binary (default: all)")
	fs.BoolVar(&f.constraints, "constraints", false,
		"Generate local numeric constraints instead of importing golang.org/x/exp/constraints (default: true)")
	fs.BoolVar(&f.constraints, "c", false, "")
	fs.BoolVar(&f.xExpConstraints, "x-exp-constraints", false,
		"Import golang.org/x/exp/constraints instead of generating local numeric constraints (default: false)")
	if err := fs.Parse(args); err != nil {
		return f, nil, fs, err
	}
	fs.Visit(func(selected *flag.Flag) {
		if selected.Name == "interfaces" {
			f.interfacesSet = true
		}
	})
	return f, fs.Args(), fs, nil
}

func parseHandlers(value string, explicitlySet bool) (config.Handlers, error) {
	all := config.Handlers{
		JSON:   true,
		Text:   true,
		SQL:    true,
		YAML:   true,
		Binary: true,
	}
	if value == "" {
		if explicitlySet {
			return config.Handlers{}, fmt.Errorf("invalid --interfaces value %q: empty interface name", value)
		}
		return all, nil
	}

	var handlers config.Handlers
	for _, rawPart := range stdstrings.Split(value, ",") {
		part := stdstrings.ToLower(stdstrings.TrimSpace(rawPart))
		if part == "" {
			return config.Handlers{}, fmt.Errorf("invalid --interfaces value %q: empty interface name", value)
		}
		switch part {
		case "json":
			handlers.JSON = true
		case "text":
			handlers.Text = true
		case "yaml":
			handlers.YAML = true
		case "sql":
			handlers.SQL = true
		case "binary":
			handlers.Binary = true
		default:
			return config.Handlers{}, fmt.Errorf(
				"invalid --interfaces value %q: unknown interface %q; valid values are json,text,yaml,sql,binary",
				value,
				part)
		}
	}
	return handlers, nil
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
	os.Exit(mainStatus())
}

func mainStatus() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	if err := run(ctx, os.Args[1:]); err != nil {
		return 1
	}
	return 0
}

func run(ctx context.Context, args []string) error {
	logging.Configure(false)
	config, err := configuration(ctx, args)
	if err != nil {
		if errors.Is(err, ErrComplete) {
			return nil
		}
		return err
	}
	logging.Configure(config.Verbose)

	logo()
	slog.Default().InfoContext(ctx, fmt.Sprintf("		version: %s", version.CURRENT))
	slog.Default().DebugContext(ctx, "starting generation...")
	slog.Default().DebugContext(ctx, "config settings",
		slog.Int("file_count", len(config.Filenames)),
		slog.String("files", buildFileList(config.Filenames)),
		slog.String("output", config.OutputFormat),
		slog.Bool("failfast", config.Failfast),
		slog.Bool("legacy", config.Legacy),
		slog.Bool("legacy_text", config.LegacyTextMarshal),
		slog.Bool("handler_json", config.Handlers.JSON),
		slog.Bool("handler_text", config.Handlers.Text),
		slog.Bool("handler_yaml", config.Handlers.YAML),
		slog.Bool("handler_sql", config.Handlers.SQL),
		slog.Bool("handler_binary", config.Handlers.Binary),
		slog.Bool("insensitive", config.Insensitive),
		slog.Bool("verbose", config.Verbose))

	for _, filename := range config.Filenames {
		filename = strings.TrimSpace(filename)
		if filename == "" {
			continue
		}
		slog.Default().InfoContext(ctx, "processing file", slog.String("filename", filename))
		var (
			parser enum.Parser
			writer enum.Writer
		)

		inExt := filepath.Ext(filename)
		switch inExt {
		case ".go":
			slog.Default().DebugContext(ctx, "initializing go parser")
			parser = gofile.NewParser(
				gofile.WithParserConfiguration(config),
				gofile.WithSource(source.FromFile(filename)))
		default:
			err := fmt.Errorf("unsupported input file extension %q", inExt)
			slog.Default().ErrorContext(ctx, "only .go files are supported")
			return err
		}

		switch config.OutputFormat {
		case "", "go":
			slog.Default().DebugContext(ctx, "initializing gofile writer")
			writer = gofile.NewWriter(gofile.WithWriterConfiguration(config))
		default:
			err := fmt.Errorf("unsupported output format %q", config.OutputFormat)
			slog.Default().ErrorContext(ctx, "only outputting to go files is supported")
			return err
		}

		slog.Default().DebugContext(ctx, "initializing generator")
		gen := generator.New(
			generator.WithConfig(config),
			generator.WithParser(parser),
			generator.WithWriter(writer))
		slog.Default().InfoContext(ctx, "starting parsing and generation")
		if err := gen.ParseAndWrite(ctx); err != nil {
			if errors.Is(err, enum.ErrParseSource) {
				slog.Default().ErrorContext(ctx, "unable to parse file", slog.String("filename", filename))
				slog.Default().ErrorContext(ctx, "please ensure that the file is a valid input file")
				slog.Default().ErrorContext(ctx, "for the selected parser")
			}
			if errors.Is(err, enum.ErrNoEnumsFound) {
				slog.Default().ErrorContext(ctx, "no enums found in file", slog.String("filename", filename))
				slog.Default().ErrorContext(ctx, "please ensure that the file contains enum definitions")
			}
			if errors.Is(err, enum.ErrWriteOutput) {
				slog.Default().ErrorContext(ctx, "could not generate output")
				slog.Default().ErrorContext(ctx, "please ensure that the output destination is writable")
				slog.Default().ErrorContext(ctx, "and that input enums contain only valid characters")
			}
			slog.Default().ErrorContext(ctx, "could not generate enums", slog.String("error", err.Error()))
			slog.Default().ErrorContext(ctx, "exiting")
			return err
		}
		slog.Default().InfoContext(ctx, "successfully generated enums")
	}
	return nil
}

var ErrComplete = errors.New("completed")

func configuration(ctx context.Context, args []string) (config.Configuration, error) {
	f, args, fs, err := parseFlags(args)
	if err != nil {
		return config.Configuration{}, err
	}

	if f.help {
		printHelp(fs)
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

	handlers, err := parseHandlers(f.interfaces, f.interfacesSet)
	if err != nil {
		return config.Configuration{}, err
	}

	filenames := args

	for _, filename := range filenames {
		filename = strings.TrimSpace(filename)
		if filename == "" {
			continue
		}

		cleanFilename := filepath.Clean(filename)
		if _, err := os.Stat(cleanFilename); os.IsNotExist(err) {
			slog.Default().ErrorContext(ctx, "input file does not exist", slog.String("filename", filename))
			return config.Configuration{}, fmt.Errorf("input file does not exist %s", filename)
		}
	}

	config := config.Configuration{
		Failfast:          f.failfast,
		Insensitive:       f.insensitive,
		Legacy:            f.legacy,
		LegacyTextMarshal: f.legacyText,
		Verbose:           f.verbose,
		OutputFormat:      f.output,
		Filenames:         filenames,
		Constraints:       !f.xExpConstraints,
		Handlers:          handlers,
	}
	return config, nil
}

// printHelp displays usage instructions and command-line options
func printHelp(fs *flag.FlagSet) {
	logo()
	slog.Default().Info("Usage: goenums [options] file.go[,file2.go,...]")
	slog.Default().Info("Options:")
	fs.PrintDefaults()
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
