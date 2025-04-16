// The producer package coordinates the workflow of enum parsing and generation.
//
// The producer package serves as the orchestration layer that connects the
// core components of the enum generation system:
//
// - Sources that provide input content
// - Parsers that extract enum representations
// - Writers that generate output artifacts
//
// By implementing a mediator pattern, the producer maintains separation
// between components while coordinating their interaction in a cohesive
// workflow. This allows each component to focus on its specialized task
// without needing to know about the others.
//
// The abstraction enables the system to support various input formats,
// parsing strategies, and output formats while maintaining a consistent
// generation process.
package generator

import (
	"context"
	"errors"
	"fmt"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/config"
)

var (
	// ErrParserFailedToParse indicates a general parsing failure occurred.
	// This error wraps more specific parsing errors.
	ErrParserFailedToParse = errors.New("failed to parse")
	// ErrParserNoEnumsFound indicates the source was valid but contained no enums.
	// This typically means the input file doesn't contain enum-like constructs.
	ErrParserNoEnumsFound = errors.New("no enums found")
	// ErrGeneratorFailedToGenerate indicates output generation failed.
	// This error wraps more specific generation errors.
	ErrGeneratorFailedToGenerate = errors.New("failed to generate")
	// ErrNoEnumsFound indicates no enums were found in the provided sources.
	ErrNoEnumsFound = errors.New("no enums found")
)

// Source represents an input source for enum definitions.
// It provides content to be parsed and identifies its origin.
type Source interface {
	// Content retrieves the raw data to be parsed
	Content() ([]byte, error)

	// Filename returns an identifier for the source
	Filename() string
}

// Generator orchestrates the enum generation workflow by connecting
// a parser and writer with configuration settings.
type Generator struct {
	Configuration config.Configuration
	parser        enum.Parser
	writer        enum.Writer
}

// New creates a Producer with the specified configuration and components.
// The producer will use the given parser to extract enum definitions and the
// writer to generate output artifacts.
func New(configuration config.Configuration,
	parser enum.Parser,
	generator enum.Writer) *Generator {
	return &Generator{
		Configuration: configuration,
		parser:        parser,
		writer:        generator,
	}
}

// ParseAndWrite executes the complete enum generation workflow:
// 1. Parse input to extract enum representations
// 2. Generate output from those representations
// It returns an error if either step fails.
func (g *Generator) ParseAndWrite(ctx context.Context) error {
	enums, err := g.parser.Parse(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrParserFailedToParse, err)
	}
	if len(enums) == 0 {
		return ErrNoEnumsFound
	}
	if err = g.writer.Write(ctx, enums); err != nil {
		return fmt.Errorf("%w: %w", ErrGeneratorFailedToGenerate, err)
	}
	return nil
}
