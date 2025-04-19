// Package enum defines the core interfaces and data structures for enum representation.
//
// This package forms the foundation of the goenums system by defining the domain model
// for enum types and values, along with the interfaces that components must implement
// to participate in the enum generation pipeline.
//
// The interfaces follow a clear separation of concerns:
//   - Parser: Extracts enum definitions from source content
//   - Writer: Generates output artifacts from enum representations
//   - Source: Provides raw content for parsing
//
// This design enables a modular system where different input formats and output targets
// can be supported without modifying the core workflow.
package enum

import (
	"context"
	"time"
)

// Parser defines the contract for components that convert source content into
// enum representations. Implementations of this interface analyze source code or other
// input formats and extract structured information about enum types and values.
// Different implementations can support various input languages or formats while
// producing the same standardized Representation output.
type Parser interface {
	// Parse analyzes the content from a source and returns structured enum representations.
	// It transforms input data into a format-agnostic model that can be used
	// for code generation. The context allows for cancellation and timeout control.
	Parse(ctx context.Context) ([]Representation, error)
}

// Source abstracts the origin of input content to be parsed for enum definitions.
// This interface decouples the parsing logic from the specific location or format
// of the input data, allowing for flexible input sources.
type Source interface {
	// Content returns the raw bytes to be parsed for enum definitions.
	// This method retrieves the complete content from whatever backing store
	// or input mechanism the implementation uses.
	Content() ([]byte, error)

	// Filename returns an identifier for the source, typically a file path.
	// Even for non-file sources, this should return a meaningful identifier
	Filename() string
}

// Writer defines the contract for components that transform enum representations
// into output artifacts. Implementations of this interface generate code or other
// output formats based on the standardized Representation model. This interface
// completes the pipeline from input source to output artifact.
type Writer interface {
	// Write generates output artifacts from enum representations.
	// It transforms the format-agnostic model into concrete output such as
	// source code files. The context allows for cancellation and timeout control.
	Write(ctx context.Context, enums []Representation) error
}

// Representation is a comprehensive model that encapsulates all information needed
// to generate an enum implementation.
//
// A Representation serves as the central data transfer object in the generation pipeline,
// containing everything a Writer needs to generate output artifacts without having to
// refer back to the original source.
type Representation struct {
	// Metadata about the generation process and the source of the enum definitions.
	Version        string
	GenerationTime time.Time

	// Configuration options for the generation process.
	Failfast        bool
	Legacy          bool
	CaseInsensitive bool

	// Information about the enum type being generated.
	PackageName    string
	TypeInfo       TypeInfo
	Enums          []Enum
	SourceFilename string
}

// Enum represents a single enum value within an enum type representation. It combines the
// core information about the enum constant (name, value, etc.), type-specific metadata,
// and the original raw content from which it was parsed. This structure provides a
// complete view of an enum value for code generation purposes.
type Enum struct {
	Info     Info
	TypeInfo TypeInfo
	Raw      Raw
}

// Raw contains the unprocessed textual content associated with an enum.
// It preserves the original comments and documentation from the input source.
type Raw struct {
	// Comment is the raw comment associated with the enum constant
	Comment string
	// TypeComment is the raw comment associated with the enum type declaration
	TypeComment string
}

// Info contains the core identifying information for an enum constant.
// It includes various name formats (camel case, lowercase, etc.) for versatility
// in different output contexts, the integer value of the enum, and a flag
// indicating whether this is a valid enum value or represents an invalid sentinel.
type Info struct {
	// Name is the original identifier for the enum constant
	Name string
	// Alias provides an optional alternative name for the enum
	Alias string
	// Aliases provides a list of alternative names for the enum
	Aliases []string
	// Camel is the camel-case representation of the name
	Camel string
	// Lower is the lowercase representation of the name
	Lower string
	// Upper is the uppercase representation of the name
	Upper string
	// Value is the integer value assigned to this enum constant
	Value int
	// Valid indicates whether this is a regular enum value (true) or an invalid sentinel (false)
	Valid bool
}

// TypeInfo contains metadata about the enum type itself rather than individual values.
// It captures naming information in various formats, index offset information, and
// details about any non-iota enum values. This information is essential for generating
// type declarations and shared enum functionality.
type TypeInfo struct {
	// Index is the starting offset value for the enum constants
	Index int
	// Name is the original type name
	Name string
	// Camel is the camel-case representation of the type name
	Camel string
	// Lower is the lowercase representation of the type name
	Lower string
	// Upper is the uppercase representation of the type name
	Upper string
	// Plural is the pluralized form of the type name
	Plural string
	// PluralCamel is the camel-case representation of the pluralized type name
	PluralCamel string
	// NameTypePair contains information about enum values that don't use iota identifer
	NameTypePair []NameTypePair
}

// NameTypePair represents a non-iota identified enum constant with explicit type and value.
// This structure captures constants that are defined as part of the incremental iota pattern.
type NameTypePair struct {
	// Name is the identifier of the enum constant
	Name string
	// Type is the explicit type of the enum constant
	Type string
	// Value is the explicit value expression of the enum constant
	Value string
}
