package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zarldev/goenums/generator/config"
)

func TestConfiguration(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	input := filepath.Join(tmp, "status.go")
	if err := os.WriteFile(input, []byte("package smoke\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		args []string
		want config.Configuration
	}{
		{
			name: "defaults",
			args: []string{input},
			want: config.Configuration{Filenames: []string{input}, Constraints: true, AccessorStyle: config.AccessorStyleUpper},
		}, {
			name: "short flags",
			args: []string{"-f", "-l", "-i", "-vv", "-c", "-o", "go", input},
			want: config.Configuration{
				Failfast:      true,
				Legacy:        true,
				Insensitive:   true,
				Verbose:       true,
				Constraints:   true,
				OutputFormat:  "go",
				AccessorStyle: config.AccessorStyleUpper,
				Filenames:     []string{input},
			}},
		{
			name: "legacy text flag",
			args: []string{"-legacy-text", input},
			want: config.Configuration{LegacyTextMarshal: true, Constraints: true, AccessorStyle: config.AccessorStyleUpper, Filenames: []string{input}},
		}, {
			name: "selected interfaces",
			args: []string{"--interfaces=json,sql", input},
			want: config.Configuration{Filenames: []string{input}, Constraints: true, AccessorStyle: config.AccessorStyleUpper, Handlers: config.Handlers{JSON: true, SQL: true}},
		}, {
			name: "selected interfaces case and spaces",
			args: []string{"--interfaces=JSON, sql", input},
			want: config.Configuration{Filenames: []string{input}, Constraints: true, AccessorStyle: config.AccessorStyleUpper, Handlers: config.Handlers{JSON: true, SQL: true}},
		}, {
			name: "selected interfaces duplicates",
			args: []string{"--interfaces=json,json", input},
			want: config.Configuration{Filenames: []string{input}, Constraints: true, AccessorStyle: config.AccessorStyleUpper, Handlers: config.Handlers{JSON: true}},
		}, {
			name: "legacy text with text interface",
			args: []string{"--interfaces=text", "-legacy-text", input},
			want: config.Configuration{
				LegacyTextMarshal: true,
				Constraints:       true,
				AccessorStyle:     config.AccessorStyleUpper, Filenames: []string{input},
				Handlers: config.Handlers{Text: true},
			},
		},
		{
			name: "legacy text without text interface",
			args: []string{"--interfaces=json,sql", "-legacy-text", input},
			want: config.Configuration{
				LegacyTextMarshal: true,
				Constraints:       true,
				AccessorStyle:     config.AccessorStyleUpper, Filenames: []string{input},
				Handlers: config.Handlers{JSON: true, SQL: true},
			},
		},
		{
			name: "go accessor style",
			args: []string{"--accessor-style=go", input},
			want: config.Configuration{Filenames: []string{input}, Constraints: true, AccessorStyle: config.AccessorStyleGo},
		},
		{
			name: "explicit upper accessor style",
			args: []string{"--accessor-style=upper", input},
			want: config.Configuration{Filenames: []string{input}, Constraints: true, AccessorStyle: config.AccessorStyleUpper},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := configuration(context.Background(), tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if got.Failfast != tt.want.Failfast || got.Legacy != tt.want.Legacy || got.Insensitive != tt.want.Insensitive ||
				got.Verbose != tt.want.Verbose || got.Constraints != tt.want.Constraints || got.LegacyTextMarshal != tt.want.LegacyTextMarshal ||
				got.OutputFormat != tt.want.OutputFormat || got.AccessorStyle != tt.want.AccessorStyle {
				t.Fatalf("configuration() = %#v, want %#v", got, tt.want)
			}
			if len(got.Filenames) != 1 || got.Filenames[0] != input {
				t.Fatalf("filenames = %#v, want %q", got.Filenames, input)
			}
			if tt.want.Handlers == (config.Handlers{}) {
				if !got.Handlers.JSON || !got.Handlers.Text || !got.Handlers.SQL || !got.Handlers.YAML || !got.Handlers.Binary {
					t.Fatalf("default handlers not enabled: %#v", got.Handlers)
				}
				return
			}
			if got.Handlers != tt.want.Handlers {
				t.Fatalf("handlers = %#v, want %#v", got.Handlers, tt.want.Handlers)
			}
		})
	}
}

func TestConfigurationControlFlow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{name: "help", args: []string{"-help"}},
		{name: "short help", args: []string{"-h"}},
		{name: "version", args: []string{"-version"}},
		{name: "short version", args: []string{"-v"}},
		{name: "missing args", args: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := configuration(context.Background(), tt.args)
			if !errors.Is(err, ErrComplete) {
				t.Fatalf("configuration() error = %v, want ErrComplete", err)
			}
		})
	}
}

func TestConfigurationErrors(t *testing.T) {
	t.Parallel()
	_, err := configuration(context.Background(), []string{filepath.Join(t.TempDir(), "missing.go")})
	if err == nil {
		t.Fatal("configuration() accepted missing input file")
	}

	_, err = configuration(context.Background(), []string{"-no-such-flag"})
	if err == nil {
		t.Fatal("configuration() accepted unknown flag")
	}

	input := filepath.Join(t.TempDir(), "status.go")
	if err := os.WriteFile(input, []byte("package smoke\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		value string
		want  string
	}{
		{value: "--interfaces=", want: "empty interface name"},
		{value: "--interfaces=json,,sql", want: "empty interface name"},
		{value: "--interfaces=json,xml", want: `unknown interface "xml"`},
		{value: "--accessor-style=", want: `invalid --accessor-style value ""`},
		{value: "--accessor-style=camel", want: `invalid --accessor-style value "camel"`},
		{value: "--accessor-style=GO", want: `invalid --accessor-style value "GO"`},
	} {
		_, err = configuration(context.Background(), []string{tt.value, input})
		if err == nil {
			t.Fatalf("configuration() accepted invalid interfaces arg %q", tt.value)
		}
		if !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("configuration() error = %q, want substring %q", err, tt.want)
		}
	}
}

func TestRun(t *testing.T) {
	tmp := t.TempDir()
	input := "status.go"
	if err := os.WriteFile(filepath.Join(tmp, input), []byte(`package smoke

type status int

const (
	pending status = iota
	done
)
`), 0o600); err != nil {
		t.Fatal(err)
	}

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Errorf("restore cwd: %v", err)
		}
	})

	if err := run(context.Background(), []string{"-c", input}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "statuses_enums.go")); err != nil {
		t.Fatalf("generated file missing: %v", err)
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	goInput := filepath.Join(tmp, "status.go")
	if err := os.WriteFile(goInput, []byte("package smoke\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	txtInput := filepath.Join(tmp, "status.txt")
	if err := os.WriteFile(txtInput, []byte("not go"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{name: "unsupported input", args: []string{txtInput}},
		{name: "unsupported output", args: []string{"-o", "json", goInput}},
		{name: "no enums", args: []string{goInput}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := run(context.Background(), tt.args); err == nil {
				t.Fatal("run() returned nil error")
			}
		})
	}
}
