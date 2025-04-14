// The logging package provides customized structured logging functionality.
// It configures slog with custom formatting to produce cleaner, more readable log output
// by removing standard prefixes as the output is to a cli, and we can also have verbose
// output with a lower log level.
package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Configure sets up the default slog logger with appropriate settings.
// When verbose is true, the log level is set to Debug; otherwise, it defaults to Info.
// This function configures a custom text handler that writes to stdout.
func Configure(verbose bool) {
	w := os.Stdout
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
	h := slog.NewTextHandler(w, opts)
	return &simpleHandler{h: h, w: w, opts: opts}
}

// simpleHandler implements slog.Handler with modified output formatting.
// It wraps a standard text handler but simplifies the output.
type simpleHandler struct {
	h    slog.Handler
	w    io.Writer
	opts *slog.HandlerOptions
}

// Enabled reports whether the handler handles records at the given level.
// It delegates to the underlying handler's Enabled method.
func (h *simpleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

// WithAttrs returns a new Handler whose attributes include both
// the receiver's attributes and the arguments.
func (h *simpleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &simpleHandler{h: h.h.WithAttrs(attrs), w: h.w, opts: h.opts}
}

// WithGroup returns a new Handler with the given group name added to
// the receiver's existing groups.
func (h *simpleHandler) WithGroup(name string) slog.Handler {
	return &simpleHandler{h: h.h.WithGroup(name), w: h.w, opts: h.opts}
}

// Handle formats the log record, omitting the standard msg= prefix
// and filtering out system attributes like time and level.
func (h *simpleHandler) Handle(ctx context.Context, r slog.Record) error {
	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "" || a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == slog.SourceKey {
			return true
		}
		attrs = append(attrs, fmt.Sprintf("%s=%s", a.Key, a.Value.String()))
		return true
	})
	format := "%s\n"
	args := []any{r.Message}
	if len(attrs) > 0 {
		format = "%s %s\n"
		args = append(args, strings.Join(attrs, " "))
	}
	fmt.Fprintf(h.w, format, args...)
	return nil
}
