// goenums is a tool to generate type-safe enums in from your idiomatic iota based enums.
// It generates a new file with the pluralised name of your input file with the suffix "_enums.go".
// Access to the enum values is done through the container struct which is the pluralised
// name of the enum type. All the enum values are constants and can be accessed through the container struct.
// The generated enum wrapper type will implement the interfaces:
// fmt.Stringer, json.Marshaler, json.Unmarshaler, sql.Scanner, driver.Valuer.
// Parse function to convert any type to the enum type as best as possible.
// An All function to return all the enum values as a slice.
// Failfast mode can be enabled to fail on generation of invalid enum while parsing
// rather than returning the zero value for the enum.
//
// Usage:
//
//	goenums [options] filename
//
// Options:
//
//	-f, -failfast   Enable failfast mode - fail on generation of invalid enum while parsing (default: false)
//
// This can also be used in a go generate directive.
// Example:
// //go:generate goenums -f status.go
//
// This will generate a new file called statuses_enums.go in the same directory as the input file.
// The generated file will contain the enum wrapper type and the container struct.
//
// https://www.zarl.dev
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/zarldev/goenums/pkg/generator"
)

const VERSION = "v0.3.5"

func main() {
	var (
		help, version, failfast bool
		err                     error
	)
	flag.BoolVar(&help, "help", false,
		"Print help information")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&version, "version", false,
		"Print version information")
	flag.BoolVar(&version, "v", false, "")
	flag.BoolVar(&failfast, "failfast", false,
		"Enable failfast mode - fail on generation of invalid enum while parsing (default: false)")
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

func printVersion() {
	printTitle()
	fmt.Printf("\t\tversion: %s\n", VERSION)
}

var asciiArt = `   ____ _____  ___  ____  __  ______ ___  _____
  / __ '/ __ \/ _ \/ __ \/ / / / __ '__ \/ ___/
 / /_/ / /_/ /  __/ / / / /_/ / / / / / (__  ) 
 \__, /\____/\___/_/ /_/\__,_/_/ /_/ /_/____/  
/____/`

func printTitle() {
	fmt.Println(asciiArt)
}
