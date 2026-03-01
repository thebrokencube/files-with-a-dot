package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/list"
)

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func colorEnabled(noColorFlag bool) bool {
	if noColorFlag {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isTerminal()
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// isFilePath returns true if the value should be treated as a file path (not a shortname).
func isFilePath(s string) bool {
	if strings.HasSuffix(s, ".yml") || strings.HasSuffix(s, ".yaml") {
		return true
	}
	if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "/") || strings.HasPrefix(s, "~/") {
		return true
	}
	fi, err := os.Stat(s)
	return err == nil && fi.Mode().IsRegular()
}

// matchShortname finds an active entry matching the shortname.
// Pass 1: exact match on entry.Path. Pass 2: exact match on final path component.
func matchShortname(entries []list.Entry, shortname string) (list.Entry, error) {
	// Pass 1: exact match on full path
	for _, e := range entries {
		if e.Section == "active" && e.Path == shortname {
			return e, nil
		}
	}

	// Pass 2: match on final path component
	var matches []list.Entry
	for _, e := range entries {
		if e.Section == "active" {
			base := filepath.Base(e.Path)
			if base == shortname {
				matches = append(matches, e)
			}
		}
	}

	switch len(matches) {
	case 0:
		return list.Entry{}, nil
	case 1:
		return matches[0], nil
	default:
		var paths []string
		for _, m := range matches {
			paths = append(paths, m.Path)
		}
		return list.Entry{}, fmt.Errorf("ambiguous shortname %q matches: %s", shortname, strings.Join(paths, ", "))
	}
}

// activeShortnames returns all active entry paths, sorted alphabetically.
func activeShortnames(entries []list.Entry) []string {
	var paths []string
	for _, e := range entries {
		if e.Section == "active" {
			paths = append(paths, e.Path)
		}
	}
	sort.Strings(paths)
	return paths
}

// resolveFolioPath resolves a --folio flag value to an absolute folio.yml path.
func resolveFolioPath(value string) (string, error) {
	if isFilePath(value) {
		return value, nil
	}

	homeDir, err := home.Dir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve shortname: %w", err)
	}

	entries, err := list.Scan(homeDir)
	if err != nil {
		return "", fmt.Errorf("scanning FOLIO_HOME: %w", err)
	}

	match, err := matchShortname(entries, value)
	if err != nil {
		return "", err
	}
	if match.Path != "" {
		return filepath.Join(homeDir, "active", match.Path, "folio.yml"), nil
	}

	active := activeShortnames(entries)
	if len(active) > 0 {
		return "", fmt.Errorf("unknown project %q — active projects:\n  %s", value, strings.Join(active, "\n  "))
	}
	return "", fmt.Errorf("unknown project %q (no active projects)", value)
}

// resolveOrDie resolves *folioPath in place. Returns false and prints error on failure.
func resolveOrDie(folioPath *string) bool {
	resolved, err := resolveFolioPath(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	*folioPath = resolved
	return true
}
