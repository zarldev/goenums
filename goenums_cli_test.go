package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIInputPathBoundary(t *testing.T) {
	root := t.TempDir()
	binary := buildCLI(t, root)
	fixture, err := os.ReadFile("internal/testdata/aliases/orders.go")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("internal/testdata/aliases/orders_enums.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		argument string
	}{
		{"relative", "orders.go"},
		{"symlink", "linked.go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := filepath.Join(root, tc.name)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "orders.go"), fixture, 0o600); err != nil {
				t.Fatal(err)
			}
			if tc.name == "symlink" {
				if err := os.Symlink("orders.go", filepath.Join(dir, "linked.go")); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			}
			command := exec.CommandContext(t.Context(), binary, "-f", tc.argument)
			command.Dir = dir
			if out, err := command.CombinedOutput(); err != nil {
				t.Fatalf("generate: %v\n%s", err, out)
			}
			generated, err := os.ReadFile(filepath.Join(dir, "orders_enums.go"))
			if err != nil {
				t.Fatal(err)
			}
			// The generated command header records the user-selected path.
			expected := bytes.Replace(want, []byte("goenums -f orders.go"), []byte("goenums -f "+tc.argument), 1)
			if !bytes.Equal(generated, expected) {
				t.Fatalf("generated output differs from committed fixture (got %d bytes, want %d)", len(generated), len(expected))
			}
		})
	}
}

func TestCLIInputErrors(t *testing.T) {
	root := t.TempDir()
	binary := buildCLI(t, root)
	for _, tc := range []struct {
		name    string
		path    string
		message string
	}{
		{"nonlocal relative", "../missing.go", "invalid file path"},
		{"missing", "missing.go", "input file does not exist"},
		{"directory", ".", "input file is not regular"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			command := exec.CommandContext(t.Context(), binary, tc.path)
			command.Dir = root
			out, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(out), tc.message) {
				t.Fatalf("want %q error, got %v: %s", tc.message, err, out)
			}
		})
	}
	// Absolute input passes preflight, then reaches the existing output writer,
	// which independently rejects an absolute output destination.
	fixture, err := os.ReadFile("internal/testdata/aliases/orders.go")
	if err != nil {
		t.Fatal(err)
	}
	absolute := filepath.Join(root, "orders.go")
	if err := os.WriteFile(absolute, fixture, 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(t.Context(), binary, "-f", absolute)
	command.Dir = root
	out, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "failed to write output") || !strings.Contains(string(out), "invalid path") {
		t.Fatalf("absolute input should reach the existing writer: %v: %s", err, out)
	}
}

func TestCLIInputPermissionError(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission test requires non-root Unix user")
	}
	root := t.TempDir()
	binary := buildCLI(t, root)
	private := filepath.Join(root, "private")
	if err := os.Mkdir(private, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(private, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(private, 0o700) })
	command := exec.CommandContext(t.Context(), binary, filepath.Join(private, "orders.go"))
	out, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "permission denied") {
		t.Fatalf("want permission error, got %v: %s", err, out)
	}
}

func buildCLI(t *testing.T, root string) string {
	t.Helper()
	binary := filepath.Join(root, "goenums")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "goenums.go")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	return binary
}
