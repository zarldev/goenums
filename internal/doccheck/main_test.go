package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckTarget(t *testing.T) {
	t.Parallel()
	siteDir := t.TempDir()
	const pageName = "examples/basic/index.html"
	pagePath := filepath.Join(siteDir, filepath.FromSlash(pageName))
	if err := os.MkdirAll(filepath.Dir(pagePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pagePath, []byte("page"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(siteDir, "usage"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "usage", "index.html"), []byte("usage"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "asset.css"), []byte("css"), 0o600); err != nil {
		t.Fatal(err)
	}

	root, err := os.OpenRoot(siteDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })

	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{name: "site route", target: "/goenums/usage/"},
		{name: "site asset", target: "/goenums/asset.css"},
		{name: "fragment only", target: "#section"},
		{name: "wrong base path", target: "/usage/", wantErr: true},
		{name: "missing route", target: "/goenums/missing/", wantErr: true},
		{name: "escapes site root", target: "../../etc/passwd", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkTarget(root, "/goenums/", pageName, tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkTarget(%q) error = %v, wantErr %v", tt.target, err, tt.wantErr)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	writeSite := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		mustWrite(t, dir, "index.html",
			`<a href="/goenums/usage/">usage</a>`+
				`<a href="asset.css">asset</a>`+
				`<a href="#top">anchor</a>`+
				`<a href="https://example.com">external</a>`)
		mustWrite(t, dir, "usage/index.html", `<a href="/goenums/">home</a>`)
		mustWrite(t, dir, "asset.css", "body{}")
		return dir
	}

	tests := []struct {
		name string
		args func(t *testing.T) []string
		want int
	}{
		{
			name: "all links resolve",
			args: func(t *testing.T) []string {
				return []string{"doccheck", writeSite(t), "/goenums/"}
			},
			want: 0,
		},
		{
			name: "broken link fails",
			args: func(t *testing.T) []string {
				dir := writeSite(t)
				mustWrite(t, dir, "broken.html", `<a href="/goenums/missing/">gone</a>`)
				return []string{"doccheck", dir, "/goenums/"}
			},
			want: 1,
		},
		{
			name: "wrong base path fails",
			args: func(t *testing.T) []string {
				dir := t.TempDir()
				mustWrite(t, dir, "index.html", `<a href="/elsewhere/">nope</a>`)
				return []string{"doccheck", dir, "/goenums/"}
			},
			want: 1,
		},
		{
			name: "wrong argument count",
			args: func(_ *testing.T) []string {
				return []string{"doccheck"}
			},
			want: 2,
		},
		{
			name: "missing site directory",
			args: func(t *testing.T) []string {
				return []string{"doccheck", filepath.Join(t.TempDir(), "absent"), "/goenums/"}
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := run(tt.args(t)); got != tt.want {
				t.Fatalf("run() = %d, want %d", got, tt.want)
			}
		})
	}
}

func mustWrite(t *testing.T, dir, name, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeBasePath(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		in   string
		want string
	}{
		{in: "/goenums/", want: "/goenums/"},
		{in: "goenums", want: "/goenums/"},
		{in: "/", want: "/"},
		{in: "", want: "/"},
	} {
		if got := normalizeBasePath(tt.in); got != tt.want {
			t.Errorf("normalizeBasePath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestShouldCheck(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		target string
		want   bool
	}{
		{target: "/goenums/usage/", want: true},
		{target: "relative/page", want: true},
		{target: "#anchor", want: false},
		{target: "https://example.com", want: false},
		{target: "mailto:docs@example.com", want: false},
		{target: "//example.com", want: false},
	} {
		if got := shouldCheck(tt.target); got != tt.want {
			t.Errorf("shouldCheck(%q) = %v, want %v", tt.target, got, tt.want)
		}
	}
}
