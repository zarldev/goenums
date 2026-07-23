package main

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

var linkPattern = regexp.MustCompile(`(?i)\b(?:href|src)=["']([^"']+)["']`)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: doccheck <site-directory> <base-path>")
		return 2
	}

	root, err := os.OpenRoot(args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer root.Close()

	failures, err := checkSite(root, normalizeBasePath(args[2]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, failure := range failures {
		fmt.Fprintln(os.Stderr, failure)
	}
	if len(failures) > 0 {
		return 1
	}
	return 0
}

// checkSite walks every HTML page under root and returns a message for each
// broken internal link. A non-nil error signals the walk itself failed.
func checkSite(root *os.Root, basePath string) ([]string, error) {
	var failures []string
	err := fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || path.Ext(name) != ".html" {
			return nil
		}
		content, err := root.ReadFile(name)
		if err != nil {
			return err
		}
		for _, match := range linkPattern.FindAllStringSubmatch(string(content), -1) {
			if target := match[1]; shouldCheck(target) {
				if err := checkTarget(root, basePath, name, target); err != nil {
					failures = append(failures, fmt.Sprintf("%s: %v", name, err))
				}
			}
		}
		return nil
	})
	return failures, err
}

func normalizeBasePath(basePath string) string {
	basePath = "/" + strings.Trim(basePath, "/")
	if basePath == "/" {
		return basePath
	}
	return basePath + "/"
}

func shouldCheck(target string) bool {
	if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "//") {
		return false
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return false
	}
	return parsed.Scheme == "" && parsed.Host == ""
}

// checkTarget resolves an internal link to a path within root and reports
// whether the target file exists. Paths are relative to the site root so all
// lookups stay inside it. pageName is the root-relative path of the page the
// link was found in.
func checkTarget(root *os.Root, basePath, pageName, target string) error {
	parsed, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", target, err)
	}
	linkPath := parsed.Path
	if linkPath == "" {
		return nil
	}

	var candidate string
	switch {
	case strings.HasPrefix(linkPath, "/"):
		if !strings.HasPrefix(linkPath, basePath) {
			return fmt.Errorf("link %q does not use base path %q", target, basePath)
		}
		candidate = strings.TrimPrefix(linkPath, basePath)
	default:
		candidate = path.Join(path.Dir(pageName), linkPath)
	}
	candidate = strings.Trim(candidate, "/")
	if candidate == "" {
		candidate = "."
	}

	if exists(root, candidate) || exists(root, path.Join(candidate, "index.html")) {
		return nil
	}
	if path.Ext(candidate) == "" && exists(root, candidate+".html") {
		return nil
	}
	return fmt.Errorf("broken internal link %q", target)
}

func exists(root *os.Root, name string) bool {
	if !fs.ValidPath(name) {
		return false
	}
	_, err := root.Stat(name)
	return err == nil
}
