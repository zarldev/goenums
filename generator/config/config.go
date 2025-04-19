// Package config defines configuration options for the enum generation process.
//
// The Configuration struct centralizes all options that influence the behavior
// of enum parsing and generation:
//
//   - Failfast: Strict validation mode for enum values
//   - Legacy: Compatibility mode for older Go versions
//   - Insensitive: Case flexibility in string parsing
//   - Verbose: Extended logging for debugging
//
// This package allows configuration to be passed consistently through the
// system, ensuring all components respect the same settings.
package config

// Configuration holds all the settings that control enum generation behavior.
// It is passed to both parsers and generators to ensure consistent behavior
// throughout the generation process.
type Configuration struct {
	// Failfast enables strict validation of enum values during parsing and generation.
	// When true, the system will return errors for invalid enum values rather than
	// silently handling them.
	Failfast bool

	// Insensitive enables case-insensitive matching when parsing enum string values.
	// When true, enum values can be matched regardless of case (e.g., "RED" == "red").
	Insensitive bool

	// Legacy enables compatibility with Go versions before 1.23.
	// When true, the generated code will not use features like range-over-func
	// that are only available in Go 1.21+.
	Legacy bool

	// Verbose enables detailed logging throughout the enum generation process.
	// When true, additional information about parsing and generation steps will
	// be logged, which is useful for debugging.
	Verbose bool

	// OutputFormat is the format of the output file.
	OutputFormat string
}
