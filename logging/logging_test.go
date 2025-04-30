package logging_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

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
			contains: []string{"With attrs", "user=", "admin", "id=", "42"},
			excludes: []string{"msg=", "level="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if ctx.Err() != nil {
				t.Skip("context cancelled")
			}

			var buf bytes.Buffer
			logging.ConfigureWithWriter(&buf, tt.verbose)
			logger := slog.Default()

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
	if !strings.Contains(output, "component=") {
		t.Errorf("withattrs not propagating attributes to output: %q", output)
	}
	buf.Reset()
	groupLogger := logger.WithGroup("system")
	groupLogger.Info("With group", "status", "ok")
	output = buf.String()
	if !strings.Contains(output, "With group") {
		t.Errorf("missing expected message text in output: %q", output)
	}
	if !strings.Contains(output, "status=") {
		t.Errorf("withgroup not properly handling attributes in output: %q", output)
	}
	if !handler.Enabled(ctx, slog.LevelInfo) {
		t.Error("handler should be enabled for info level")
	}
}
