package logging_test

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/zarldev/goenums/logging"
)

func TestLoggingOutput(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	tests := []struct {
		name     string
		verbose  bool
		logFunc  func(logger *slog.Logger)
		contains []string
		excludes []string
	}{
		{
			name:    "info level",
			verbose: false,
			logFunc: func(logger *slog.Logger) {
				logger.Info("Info message")
				logger.Debug("Debug message") // Should be filtered out
			},
			contains: []string{"Info message"},
			excludes: []string{"Debug message", "msg=", "level="},
		},
		{
			name:    "debug level",
			verbose: true,
			logFunc: func(logger *slog.Logger) {
				logger.Info("Info message")
				logger.Debug("Debug message")
			},
			contains: []string{"Info message", "Debug message"},
			excludes: []string{"msg=", "level="},
		},
		{
			name:    "with attributes",
			verbose: false,
			logFunc: func(logger *slog.Logger) {
				logger.Info("With attrs", "user", "admin", "id", 42)
			},
			contains: []string{"With attrs", "user: ", "admin", "id: ", "42"},
			excludes: []string{"msg ", "level "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if ctx.Err() != nil {
				t.Skip("context cancelled")
			}

			var buf bytes.Buffer
			level := slog.LevelInfo
			if tt.verbose {
				level = slog.LevelDebug
			}
			handler := logging.NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: level})
			logger := slog.New(handler)

			tt.logFunc(logger)

			output := buf.String()

			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("missing expected text %q in output: %q", s, output)
				}
			}

			for _, s := range tt.excludes {
				if strings.Contains(output, s) {
					t.Errorf("unexpected text %q present in output: %q", s, output)
				}
			}
		})
	}
}

func TestCustomHandler(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	var buf bytes.Buffer
	handler := logging.NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)
	buf.Reset()
	attrLogger := logger.With("component", "test")
	attrLogger.Info("With attributes")
	output := buf.String()
	if !strings.Contains(output, "With attributes") {
		t.Errorf("missing expected message text in output: %q", output)
	}
	if !strings.Contains(output, "component:") {
		t.Errorf("withattrs not propagating attributes to output: %q", output)
	}
	buf.Reset()
	groupLogger := logger.WithGroup("system")
	groupLogger.Info("With group", "status", "ok")
	output = buf.String()
	if !strings.Contains(output, "With group") {
		t.Errorf("missing expected message text in output: %q", output)
	}
	if !strings.Contains(output, "status:") {
		t.Errorf("withgroup not properly handling attributes in output: %q", output)
	}
	if !handler.Enabled(ctx, slog.LevelInfo) {
		t.Error("handler should be enabled for info level")
	}
}

// TestConfigure mutates the process-wide default logger, so it must not run
// in parallel with anything that logs through slog's package-level functions.
func TestConfigure(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	var buf bytes.Buffer
	logging.ConfigureWithWriter(&buf, false)
	slog.Info("info message")   //nolint:sloglint // the test verifies the configured default logger
	slog.Debug("debug message") //nolint:sloglint // the test verifies the configured default logger
	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("missing info message in output: %q", output)
	}
	if strings.Contains(output, "debug message") {
		t.Errorf("debug message should be filtered at info level: %q", output)
	}

	buf.Reset()
	logging.ConfigureWithWriter(&buf, true)
	slog.Debug("debug message") //nolint:sloglint // the test verifies the configured default logger
	output = buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("missing debug message in verbose output: %q", output)
	}

	logging.Configure(false)
	if !slog.Default().Enabled(t.Context(), slog.LevelInfo) {
		t.Error("default logger should be enabled for info level")
	}
	if slog.Default().Enabled(t.Context(), slog.LevelDebug) {
		t.Error("default logger should not be enabled for debug level when not verbose")
	}
}

func TestNewCustomTextHandlerNilOptions(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := slog.New(logging.NewCustomTextHandler(&buf, nil))
	logger.Info("info message")
	logger.Debug("debug message")
	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("missing info message in output: %q", output)
	}
	if strings.Contains(output, "debug message") {
		t.Errorf("nil options should default to info level: %q", output)
	}
}

func TestAttrFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attrs    []slog.Attr
		contains []string
		excludes []string
	}{
		{
			name:     "empty key prints value only",
			attrs:    []slog.Attr{slog.Any("", "bare value")},
			contains: []string{"bare value"},
			excludes: []string{":"},
		},
		{
			name:     "key longer than fixed width",
			attrs:    []slog.Attr{slog.String("averyverylongkey", "value")},
			contains: []string{"averyverylongkey:value"},
		},
		{
			name: "time level and source keys are skipped",
			attrs: []slog.Attr{
				slog.String(slog.TimeKey, "skipped"),
				slog.String(slog.LevelKey, "skipped"),
				slog.String(slog.SourceKey, "skipped"),
			},
			contains: []string{"message"},
			excludes: []string{"skipped"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			logger := slog.New(logging.NewCustomTextHandler(&buf, nil))
			logger.LogAttrs(t.Context(), slog.LevelInfo, "message", tt.attrs...)
			output := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("missing expected text %q in output: %q", s, output)
				}
			}
			for _, s := range tt.excludes {
				if strings.Contains(output, s) {
					t.Errorf("unexpected text %q present in output: %q", s, output)
				}
			}
		})
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write refused")
}

func TestHandleWriteError(t *testing.T) {
	t.Parallel()
	handler := logging.NewCustomTextHandler(errWriter{}, nil)
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "message", 0)
	err := handler.Handle(t.Context(), r)
	if !errors.Is(err, logging.ErrLogging) {
		t.Errorf("expected error wrapping ErrLogging, got %v", err)
	}
}
