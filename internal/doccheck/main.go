package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var linkPattern = regexp.MustCompile(`(?i)\b(?:href|src)=["']([^"']+)["']`)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: doccheck <site-directory> <base-path>")
		os.Exit(2)
	}

	siteDir, err := filepath.Abs(os.Args[1])
	if err != nil {
		fail(err)
	}
	basePath := normalizeBasePath(os.Args[2])

	var failures []string
	err = filepath.WalkDir(siteDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range linkPattern.FindAllStringSubmatch(string(content), -1) {
			if target := match[1]; shouldCheck(target) {
				if err := checkTarget(siteDir, basePath, path, target); err != nil {
					failures = append(failures, fmt.Sprintf("%s: %v", path, err))
				}
			}
		}
		return nil
	})
	if err != nil {
		fail(err)
	}
	if len(failures) > 0 {
		for _, failure := range failures {
			fmt.Fprintln(os.Stderr, failure)
		}
		os.Exit(1)
	}
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

func checkTarget(siteDir, basePath, pagePath, target string) error {
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
		candidate = filepath.Join(siteDir, strings.TrimPrefix(linkPath, basePath))
	default:
		candidate = filepath.Join(filepath.Dir(pagePath), linkPath)
	}

	if exists(candidate) || exists(filepath.Join(candidate, "index.html")) {
		return nil
	}
	if filepath.Ext(candidate) == "" && exists(candidate+".html") {
		return nil
	}
	return fmt.Errorf("broken internal link %q", target)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
