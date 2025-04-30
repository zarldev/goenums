// Package logging provides customized structured logging functionality.
// It configures slog with custom formatting to produce cleaner, more readable log output
// by removing standard prefixes as the output is to a cli, and we can also have verbose
// output with a lower log level.
package logging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
)

var (
	// ErrLogging is a sentinel error used to identify logging-related errors.
	ErrLogging = errors.New("logging error")
)

// Configure sets up the default slog logger with appropriate settings.
// When verbose is true, the log level is set to Debug; otherwise, it defaults to Info.
// This function configures a custom text handler that writes to stdout.
func Configure(verbose bool) {
	ConfigureWithWriter(os.Stdout, verbose)
}

// ConfigureWithWriter sets up the default slog logger with a custom writer.
// This allows redirecting logs to different destinations (files, buffers, etc.)
func ConfigureWithWriter(w io.Writer, verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	handler := NewCustomTextHandler(w, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// NewCustomTextHandler creates a text handler with custom formatting that omits
// the standard "msg=" prefix from log output.
func NewCustomTextHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &logger{
		w:     w,
		level: opts.Level.Level(),
		group: "",
	}
}

// logger implements slog.Handler with very direct, simple handling
type logger struct {
	w     io.Writer
	level slog.Level
	attrs []slog.Attr
	group string
}

// Enabled reports whether the handler handles records at the given level.
func (h *logger) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// WithAttrs returns a new Handler whose attributes include both
// the receiver's attributes and the arguments.
func (h *logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &logger{
		w:     h.w,
		level: h.level,
		group: h.group,
	}
	newHandler.attrs = slices.Clone(h.attrs)
	newHandler.attrs = append(newHandler.attrs, attrs...)
	return newHandler
}

// WithGroup returns a new Handler with the given group name added to
// the receiver's existing groups.
func (h *logger) WithGroup(name string) slog.Handler {
	return &logger{
		w:     h.w,
		level: h.level,
		attrs: h.attrs,
		group: name,
	}
}

// formatAttr formats a single attribute with consistent, fixed-width spacing
func formatAttr(a slog.Attr) string {
	const (
		fixedWidth = 12
		halfWidth  = fixedWidth / 2
	)

	if a.Key == "" {
		padding := strings.Repeat(" ", halfWidth)
		return fmt.Sprintf("%s%v", padding, a.Value.Any())
	}
	spaces := fixedWidth - len(a.Key)
	if spaces < 0 {
		spaces = 0
	}
	padding := strings.Repeat(" ", spaces)
	return fmt.Sprintf("%s:%s%v", a.Key, padding, a.Value.Any())
}

// Handle formats and outputs the log record
func (h *logger) Handle(ctx context.Context, r slog.Record) error {
	var allAttrs []string

	// Process existing attributes
	for _, attr := range h.attrs {
		allAttrs = append(allAttrs, formatAttr(attr))
	}

	// Process record attributes
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == slog.SourceKey {
			return true
		}
		allAttrs = append(allAttrs, formatAttr(a))
		return true
	})

	var builder strings.Builder
	if r.Message != "" {
		_, _ = builder.WriteString(r.Message + "\n")
	}
	if len(allAttrs) > 0 {
		_, _ = builder.WriteString(strings.Join(allAttrs, "\n"))
	}
	_, _ = builder.WriteString("\n")

	if _, err := fmt.Fprint(h.w, builder.String()); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrLogging, "printing log message", err)
	}
	return nil
}
