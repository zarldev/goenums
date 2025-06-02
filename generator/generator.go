// The generator package coordinates the workflow of enum parsing and generation.
//
// The generator package serves as the orchestration layer that connects the
// core components of the enum generation system:
//
// - Sources that provide input content
// - Parsers that extract enum representations
// - Writers that generate output artifacts
//
// By implementing a mediator pattern, the generator maintains separation
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
	"fmt"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/generator/gofile"
)

// Generator is the main orchestrator for the enum generation workflow.
// Generator orchestrates the enum generation workflow by connecting
// a parser and writer with configuration settings.
type Generator struct {
	Configuration config.Configuration
	parser        enum.Parser
	writer        enum.Writer
}

type GeneratorOption func(*Generator)

func WithConfig(configuration config.Configuration) func(*Generator) {
	return func(g *Generator) {
		g.Configuration = configuration
	}
}

func WithParser(parser enum.Parser) func(*Generator) {
	return func(g *Generator) {
		g.parser = parser
	}
}
func WithWriter(writer enum.Writer) func(*Generator) {
	return func(g *Generator) {
		g.writer = writer
	}
}

// New creates a Generator with the specified configuration and components.
// The generator will use the given parser to extract enum definitions and the
// writer to generate output artifacts.
func New(opts ...GeneratorOption) *Generator {
	g := Generator{
		Configuration: config.Configuration{},
		parser:        gofile.NewParser(),
		writer:        gofile.NewWriter(),
	}
	for _, opt := range opts {
		opt(&g)
	}
	return &g
}

// ParseAndWrite executes the complete enum generation workflow:
// 1. Parse input to extract enum representations
// 2. Generate output from those representations
// It returns an error if either step fails.
func (g *Generator) ParseAndWrite(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("%w: %w", enum.ErrParseSource, ctx.Err())
	}
	genr, err := g.parser.Parse(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", enum.ErrParseSource, err)
	}
	if ctx.Err() != nil {
		return fmt.Errorf("%w: %w", enum.ErrParseSource, ctx.Err())
	}
	if err = g.writer.Write(ctx, genr); err != nil {
		return fmt.Errorf("%w: %w", enum.ErrWriteOutput, err)
	}
	return nil
}
