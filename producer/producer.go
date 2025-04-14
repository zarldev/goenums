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
package producer

import (
	"context"
	"fmt"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/producer/config"
)

var (
	// ErrParserFailedToParse indicates a general parsing failure occurred.
	// This error wraps more specific parsing errors.
	ErrParserFailedToParse = fmt.Errorf("failed to parse")

	// ErrParserNoEnumsFound indicates the source was valid but contained no enums.
	// This typically means the input file doesn't contain enum-like constructs.
	ErrParserNoEnumsFound = fmt.Errorf("no enums found")

	// ErrGeneratorFailedToGenerate indicates output generation failed.
	// This error wraps more specific generation errors.
	ErrGeneratorFailedToGenerate = fmt.Errorf("failed to generate")
)

// Source represents an input source for enum definitions.
// It provides content to be parsed and identifies its origin.
type Source interface {
	// Content retrieves the raw data to be parsed
	Content() ([]byte, error)

	// Filename returns an identifier for the source
	Filename() string
}

// Producer orchestrates the enum generation workflow by connecting
// a parser and writer with configuration settings.
type Producer struct {
	Configuration config.Configuration
	parser        enum.Parser
	writer        enum.Writer
}

// NewProducer creates a Producer with the specified configuration and components.
// The producer will use the given parser to extract enum definitions and the
// writer to generate output artifacts.
func NewProducer(configuration config.Configuration,
	parser enum.Parser,
	generator enum.Writer) *Producer {
	return &Producer{
		Configuration: configuration,
		parser:        parser,
		writer:        generator,
	}
}

// ParseAndWrite executes the complete enum generation workflow:
// 1. Parse input to extract enum representations
// 2. Generate output from those representations
// It returns an error if either step fails.
func (g *Producer) ParseAndWrite(ctx context.Context) error {
	enums, err := g.parser.Parse(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrParserFailedToParse, err)
	}
	if err = g.writer.Write(ctx, enums); err != nil {
		return fmt.Errorf("%w: %w", ErrGeneratorFailedToGenerate, err)
	}
	return nil
}
