package gofile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/zarldev/goenums/enum"
	"github.com/zarldev/goenums/file"
	"github.com/zarldev/goenums/generator/config"
	"github.com/zarldev/goenums/strings"
)

var _ enum.Writer = &Writer{}

var (
	// ErrWriteGoFile is returned when an error occurs while writing the go file.
	ErrWriteGoFile = errors.New("error writing go file")
)

// Writer implements enum.Writer for go source files.
// It writes enum definitions to a file on provided filesystem,
// with the specified configuration.
type Writer struct {
	Configuration config.Configuration
	w             io.Writer
	fs            file.ReadCreateWriteFileFS
	templateErr   error
}

// WriterOption is a function that configures a Writer.
type WriterOption func(*Writer)

// WithFileSystem sets the filesystem to use for writing files.
func WithFileSystem(fs file.ReadCreateWriteFileFS) func(*Writer) {
	return func(w *Writer) {
		w.fs = fs
	}
}

// WithWriterConfiguration sets the configuration for the writer.
func WithWriterConfiguration(configuration config.Configuration) func(*Writer) {
	return func(w *Writer) {
		w.Configuration = configuration
	}
}

// NewWriter creates a new go file writer with the specified configuration and filesystem.
// The writer will write enum definitions to the provided filesystem.
// When no options are provided, it will write to stdout.
func NewWriter(opts ...WriterOption) *Writer {
	w := Writer{
		Configuration: config.Configuration{},
		fs:            &file.OSReadWriteFileFS{},
		w:             os.Stdout,
	}
	for _, opt := range opts {
		opt(&w)
	}
	return &w
}

func (g *Writer) Write(ctx context.Context,
	reqs []enum.GenerationRequest) error {
	for _, req := range reqs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !req.IsValid() {
			return fmt.Errorf("invalid enum: %s", req.SourceFilename)
		}
		dirPath := filepath.Dir(req.SourceFilename)
		if !filepath.IsLocal(dirPath) {
			return fmt.Errorf("invalid path: %s", dirPath)
		}
		outFilename := fmt.Sprintf("%s_enums.go", req.OutputFilename)
		if strings.Contains(outFilename, " ") || strings.Contains(outFilename, "/") {
			return fmt.Errorf("%w: '%s' contains invalid characters", ErrWriteGoFile, outFilename)
		}
		fullPath := filepath.Clean(filepath.Join(dirPath, outFilename))
		renderReq := newRenderRequest(req)
		err := file.WriteToFileAndFormatFS(ctx, g.fs, fullPath, true,
			func(w io.Writer) error {
				var output bytes.Buffer
				output.Grow(estimateGeneratedSize(renderReq))
				g.w = &output
				g.templateErr = nil
				g.writeEnumGenerationRequest(renderReq)
				if g.templateErr != nil {
					return g.templateErr
				}
				_, err := io.Copy(w, &output)
				return err
			})
		if err != nil {
			return fmt.Errorf("%w: %s: %w", ErrWriteGoFile, fullPath, err)
		}
	}
	return nil
}

func (g *Writer) writeEnumGenerationRequest(req renderRequest) {
	g.writeGeneratedComments(req.GenerationRequest)
	g.writePackageAndImports(req.GenerationRequest)
	g.writeWrapperDefinition(req.GenerationRequest)
	g.writeContainerDefinition(req)
	g.writeInvalidEnumDefinition(req.GenerationRequest)
	g.writeAllFunction(req)
	g.writeParseFunction(req.GenerationRequest)
	g.writeStringParsingMethod(req)
	g.writeNumberParsingMethods(req)
	g.writeExhaustiveFunction(req)
	if !req.Configuration.Legacy {
		g.writeMatchFunction(req)
	}
	g.writeIsValidFunction(req)
	if req.Configuration.Handlers.JSON {
		g.writeJSONMarshalMethod(req.GenerationRequest)
		g.writeJSONUnmarshalMethod(req.GenerationRequest)
	}
	if req.Configuration.Handlers.Text {
		g.writeTextMarshalMethod(req.GenerationRequest)
		g.writeTextUnmarshalMethod(req.GenerationRequest)
	}
	if req.Configuration.Handlers.SQL {
		g.writeScanMethod(req.GenerationRequest)
		g.writeValueMethod(req.GenerationRequest)
	}
	if req.Configuration.Handlers.Binary {
		g.writeBinaryMarshalMethod(req.GenerationRequest)
		g.writeBinaryUnmarshalMethod(req.GenerationRequest)
	}
	if req.Configuration.Handlers.YAML {
		g.writeYAMLMarshalMethod(req.GenerationRequest)
		g.writeYAMLUnmarshalMethod(req.GenerationRequest)
	}
	g.writeStringMethod(req)
	g.writeCompileCheck(req.GenerationRequest)
}

func (g *Writer) writeTemplate(t *template.Template, d any) {
	if g.templateErr != nil {
		return
	}
	if err := t.Execute(g.w, d); err != nil {
		g.templateErr = fmt.Errorf("execute template %s: %w", t.Name(), err)
		slog.Default().Error("error writing template", "template", t.Name(), "error", err)
	}
}
