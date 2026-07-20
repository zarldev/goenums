package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckTarget(t *testing.T) {
	t.Parallel()
	siteDir := t.TempDir()
	pagePath := filepath.Join(siteDir, "examples", "basic", "index.html")
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkTarget(siteDir, "/goenums/", pagePath, tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkTarget(%q) error = %v, wantErr %v", tt.target, err, tt.wantErr)
			}
		})
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
